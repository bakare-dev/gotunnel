package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/bakare-dev/gotunnel/internal/protocol"
	"github.com/bakare-dev/gotunnel/internal/server"
)

func main() {

	tunnelAddr := flag.String("addr", ":9000", "tunnel listen address")
	startPort := flag.Int("start-port", 10000, "starting port for public listeners")

	tlsEnabled := flag.Bool("tls", false, "enable TLS encryption")
	tlsCert := flag.String("tls-cert", "certs/server-cert.pem", "path to TLS certificate")
	tlsKey := flag.String("tls-key", "certs/server-key.pem", "path to TLS private key")

	flag.Parse()

	printBanner(*tlsEnabled)

	router := server.NewRouter(*startPort)
	public := server.NewPublicListener(router)

	var ln net.Listener
	var err error

	if *tlsEnabled {

		cert, err := tls.LoadX509KeyPair(*tlsCert, *tlsKey)
		if err != nil {
			log.Fatalf("Failed to load TLS certificate: %v", err)
		}

		config := &tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS12,
		}

		ln, err = tls.Listen("tcp", *tunnelAddr, config)
		if err != nil {
			log.Fatal(err)
		}

		log.Println("│ INFO  │ TLS enabled ✓")
	} else {
		ln, err = net.Listen("tcp", *tunnelAddr)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("│ WARN  │ TLS disabled - connection is NOT encrypted")
	}

	log.Println("│ INFO  │ Server started")
	log.Printf("│ INFO  │ Tunnel port: %s\n", *tunnelAddr)
	log.Println("│ INFO  │ Ready for connections")
	log.Println("─────────────────────────────────────────────────────────────")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("\n│ INFO  │ Received shutdown signal")
		cancel()
		ln.Close()
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			conn, err := ln.Accept()
			if err != nil {
				select {
				case <-ctx.Done():
					return
				default:
					continue
				}
			}
			go handleClient(conn, router, public, ctx)
		}
	}()

	<-ctx.Done()
	log.Println("│ INFO  │ Shutting down server...")

	router.CloseAll()

	log.Println("│ INFO  │ Server shutdown complete")
}

func printBanner(tlsEnabled bool) {
	tlsStatus := ""
	if tlsEnabled {
		tlsStatus = " (TLS Enabled)"
	}

	banner := fmt.Sprintf(`
╔════════════════════════════════════════════════════════════╗
║              GoTunnel Server v0.1.0%s                        ║
╚════════════════════════════════════════════════════════════╝
`, tlsStatus)
	fmt.Println(banner)
}

func handleClient(conn net.Conn, router *server.Router, public *server.PublicListener, ctx context.Context) {
	defer conn.Close()

	sess := protocol.NewSession(conn, conn)

	for {
		select {
		case <-ctx.Done():
			sess.Close()
			return
		default:
		}

		frame, err := sess.ReadFrame()
		if err != nil {
			return
		}

		switch frame.Type {

		case protocol.MsgHandshake:
			if err := sess.ProcessHandshake(frame); err != nil {
				log.Printf("│ ERROR │ Handshake failed: %v", err)
				return
			}
			_ = sess.WriteFrame(&protocol.Frame{Type: protocol.MsgHandshakeAck})

		case protocol.MsgAuth:
			if err := sess.ProcessAuth(frame); err != nil {
				log.Printf("│ ERROR │ Auth failed: %v", err)
				return
			}
			_ = sess.WriteFrame(&protocol.Frame{Type: protocol.MsgAuthOK})

			port := router.AllocatePort(sess)
			go public.Listen(port)

			_ = sess.WriteFrame(&protocol.Frame{
				Type:    protocol.MsgBindOK,
				Payload: protocol.EncodeUint16(uint16(port)),
			})

			log.Printf("│ INFO  │ Client bound to public port %d", port)
			log.Printf("│ INFO  │ Exposing: %s → :%d", sess.ExposeAddr, port)
			log.Println("─────────────────────────────────────────────────────────────")
			goto FORWARD
		}
	}

FORWARD:
	for {
		select {
		case <-ctx.Done():
			router.Remove(sess.PublicPort)
			sess.Close()
			log.Printf("│ INFO  │ Client session closed (port %d)", sess.PublicPort)
			return
		default:
		}

		frame, err := sess.ReadFrame()
		if err != nil {
			router.Remove(sess.PublicPort)
			log.Printf("│ INFO  │ Client disconnected (port %d)", sess.PublicPort)
			return
		}

		if frame.Type == protocol.MsgHeartbeat {
			continue
		}

		_ = sess.HandleFrame(frame)
	}
}
