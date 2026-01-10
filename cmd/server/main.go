package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/bakare-dev/gotunnel/internal/protocol"
	"github.com/bakare-dev/gotunnel/internal/server"
)

func main() {
	printBanner()

	tunnelAddr := os.Getenv("TUNNEL_LISTEN")
	if tunnelAddr == "" {
		tunnelAddr = ":9000"
	}

	startPort := 10000

	router := server.NewRouter(startPort)
	public := server.NewPublicListener(router)

	ln, err := net.Listen("tcp", tunnelAddr)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("│ INFO  │ Server started")
	log.Printf("│ INFO  │ Tunnel port: %s\n", tunnelAddr)
	log.Println("│ INFO  │ Ready for connections")
	log.Println("─────────────────────────────────────────────────────────────")

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go handleClient(conn, router, public)
	}
}

func printBanner() {
	banner := `
╔════════════════════════════════════════════════════════════╗
║              GoTunnel Server v0.1.0                        ║
╚════════════════════════════════════════════════════════════╝
`
	fmt.Println(banner)
}

func handleClient(conn net.Conn, router *server.Router, public *server.PublicListener) {
	defer conn.Close()

	sess := protocol.NewSession(conn, conn)

	for {
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
