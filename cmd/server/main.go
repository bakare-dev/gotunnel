package main

import (
	"log"
	"net"
	"os"

	"github.com/bakare-dev/gotunnel/internal/protocol"
	"github.com/bakare-dev/gotunnel/internal/server"
)

func main() {
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
	log.Println("tunnel listening on", tunnelAddr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go handleClient(conn, router, public)
	}
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
			_ = sess.WriteFrame(&protocol.Frame{Type: protocol.MsgHandshakeAck})

		case protocol.MsgAuth:
			_ = sess.WriteFrame(&protocol.Frame{Type: protocol.MsgAuthOK})

			port := router.AllocatePort(sess)
			go public.Listen(port)

			_ = sess.WriteFrame(&protocol.Frame{
				Type:    protocol.MsgBindOK,
				Payload: protocol.EncodeUint16(uint16(port)),
			})

			log.Println("client bound to public port", port)
			goto FORWARD
		}
	}

FORWARD:
	for {
		frame, err := sess.ReadFrame()
		if err != nil {
			router.Remove(sess.PublicPort)
			return
		}
		_ = sess.HandleFrame(frame)
	}
}
