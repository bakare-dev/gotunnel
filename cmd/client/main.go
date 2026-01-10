package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bakare-dev/gotunnel/internal/client"
	"github.com/bakare-dev/gotunnel/internal/protocol"
)

func main() {
	serverAddr := flag.String("server", "localhost:9000", "tunnel server address")
	localAddr := flag.String("local", "", "local service to expose (e.g. localhost:6001)")
	token := flag.String("token", "dev-token", "auth token")
	noReconnect := flag.Bool("no-reconnect", false, "disable auto-reconnect on connection loss")

	tlsEnabled := flag.Bool("tls", false, "enable TLS encryption")
	tlsCA := flag.String("tls-ca", "certs/ca-cert.pem", "path to CA certificate")

	flag.Parse()

	if *localAddr == "" {
		log.Fatal("missing --local flag (e.g. --local localhost:6001)")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("\n│ INFO  │ Received shutdown signal")
		cancel()
	}()

	reconnectConfig := client.DefaultReconnectConfig()
	tlsConfig := client.TLSConfig{
		Enabled: *tlsEnabled,
		CAFile:  *tlsCA,
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("│ INFO  │ Shutting down...")
			return
		default:
		}

		conn, sess, publicPort, err := client.ConnectWithRetry(ctx, *serverAddr, *localAddr, *token, tlsConfig, reconnectConfig)
		if err != nil {
			log.Printf("│ ERROR │ Failed to connect: %v", err)
			return
		}

		printBanner(*serverAddr, publicPort, *localAddr, !*noReconnect, *tlsEnabled)

		err = runSession(ctx, conn, sess, *localAddr)

		fmt.Println("\n" + sess.Metrics.Summary())

		if err != nil && err != context.Canceled {
			log.Printf("│ ERROR │ Session lost: %v", err)
		}

		select {
		case <-ctx.Done():
			log.Println("│ INFO  │ Shutdown complete")
			return
		default:
		}

		if *noReconnect {
			log.Println("│ INFO  │ Auto-reconnect disabled, exiting")
			return
		}

		log.Println("│ INFO  │ Connection lost, attempting to reconnect...")
		time.Sleep(2 * time.Second)
	}
}

func runSession(ctx context.Context, conn *net.Conn, sess *protocol.Session, localAddr string) error {
	defer (*conn).Close()
	defer sess.Close()

	forwarder := client.NewForwarder(sess, localAddr)
	defer forwarder.Close()

	done := make(chan error, 1)

	go func() {
		for {
			select {
			case <-ctx.Done():
				done <- ctx.Err()
				return
			default:
			}

			frame, err := sess.ReadFrame()
			if err != nil {
				done <- err
				return
			}

			if frame.Type == protocol.MsgHeartbeat {
				continue
			}

			forwarder.HandleFrame(frame)
		}
	}()

	return <-done
}

func printBanner(server string, publicPort uint16, localAddr string, reconnectEnabled, tlsEnabled bool) {
	reconnectStatus := "enabled"
	if !reconnectEnabled {
		reconnectStatus = "disabled"
	}

	tlsStatus := "disabled"
	if tlsEnabled {
		tlsStatus = "enabled ✓"
	}

	banner := `
╔════════════════════════════════════════════════════════════╗
║                   GoTunnel v0.1.0                          ║
║                 Secure TCP Tunneling                       ║
╚════════════════════════════════════════════════════════════╝

Session Status         online
Version                0.1.0
Tunnel Server          %s
TLS Encryption         %s
Auto-Reconnect         %s

Forwarding             tcp://localhost:%d → %s

HTTP Requests
─────────────────────────────────────────────────────────────
`
	fmt.Printf(banner, server, tlsStatus, reconnectStatus, publicPort, localAddr)
	fmt.Printf("Connected at %s\n\n", time.Now().Format("2006-01-02 15:04:05"))
}
