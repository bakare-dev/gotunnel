package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/bakare-dev/gotunnel/internal/protocol"
)

type ReconnectConfig struct {
	MaxRetries     int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	BackoffFactor  float64
}

func DefaultReconnectConfig() ReconnectConfig {
	return ReconnectConfig{
		MaxRetries:     10,
		InitialBackoff: 1 * time.Second,
		MaxBackoff:     30 * time.Second,
		BackoffFactor:  2.0,
	}
}

type TLSConfig struct {
	Enabled bool
	CAFile  string
}

func ConnectWithRetry(ctx context.Context, serverAddr, localAddr, token string, tlsCfg TLSConfig, config ReconnectConfig) (*net.Conn, *protocol.Session, uint16, error) {
	backoff := config.InitialBackoff

	for attempt := 1; attempt <= config.MaxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return nil, nil, 0, ctx.Err()
		default:
		}

		log.Printf("│ INFO  │ Connection attempt %d/%d...", attempt, config.MaxRetries)

		conn, sess, port, err := attemptConnection(serverAddr, localAddr, token, tlsCfg)
		if err == nil {
			log.Printf("│ INFO  │ Connected successfully")
			return conn, sess, port, nil
		}

		log.Printf("│ WARN  │ Connection failed: %v", err)

		if attempt < config.MaxRetries {
			log.Printf("│ INFO  │ Retrying in %v...", backoff)

			select {
			case <-ctx.Done():
				return nil, nil, 0, ctx.Err()
			case <-time.After(backoff):
			}

			backoff = time.Duration(float64(backoff) * config.BackoffFactor)
			if backoff > config.MaxBackoff {
				backoff = config.MaxBackoff
			}
		}
	}

	return nil, nil, 0, fmt.Errorf("failed to connect after %d attempts", config.MaxRetries)
}

func attemptConnection(serverAddr, localAddr, token string, tlsCfg TLSConfig) (*net.Conn, *protocol.Session, uint16, error) {
	var conn net.Conn
	var err error

	if tlsCfg.Enabled {
		caCert, err := os.ReadFile(tlsCfg.CAFile)
		if err != nil {
			return nil, nil, 0, fmt.Errorf("failed to read CA cert: %w", err)
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, nil, 0, fmt.Errorf("failed to parse CA cert")
		}

		config := &tls.Config{
			RootCAs:    caCertPool,
			ServerName: "localhost",
		}

		conn, err = tls.DialWithDialer(
			&net.Dialer{Timeout: 10 * time.Second},
			"tcp",
			serverAddr,
			config,
		)
		if err != nil {
			return nil, nil, 0, err
		}

		log.Println("│ INFO  │ TLS connection established")
	} else {

		conn, err = net.DialTimeout("tcp", serverAddr, 10*time.Second)
		if err != nil {
			return nil, nil, 0, err
		}
	}

	sess := protocol.NewSession(conn, conn)

	hs := &protocol.Handshake{
		Role:         protocol.RoleClient,
		Capabilities: protocol.CapHeartbeat,
		ExposeAddr:   localAddr,
	}

	payload, err := hs.Encode()
	if err != nil {
		conn.Close()
		return nil, nil, 0, err
	}

	if err := sess.WriteFrame(&protocol.Frame{
		Type:    protocol.MsgHandshake,
		Payload: payload,
	}); err != nil {
		conn.Close()
		return nil, nil, 0, err
	}

	frame, err := sess.ReadFrame()
	if err != nil || frame.Type != protocol.MsgHandshakeAck {
		conn.Close()
		return nil, nil, 0, fmt.Errorf("handshake rejected")
	}

	if err := sess.WriteFrame(&protocol.Frame{
		Type:    protocol.MsgAuth,
		Payload: protocol.EncodeAuth(token),
	}); err != nil {
		conn.Close()
		return nil, nil, 0, err
	}

	frame, err = sess.ReadFrame()
	if err != nil || frame.Type != protocol.MsgAuthOK {
		conn.Close()
		return nil, nil, 0, fmt.Errorf("authentication rejected")
	}

	frame, err = sess.ReadFrame()
	if err != nil || frame.Type != protocol.MsgBindOK {
		conn.Close()
		return nil, nil, 0, fmt.Errorf("failed to bind")
	}

	publicPort := protocol.DecodeUint16(frame.Payload)

	return &conn, sess, publicPort, nil
}
