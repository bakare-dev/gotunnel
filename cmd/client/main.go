package main

import (
	"flag"
	"log"
	"net"

	"github.com/bakare-dev/gotunnel/internal/client"
	"github.com/bakare-dev/gotunnel/internal/protocol"
)

func main() {
	serverAddr := flag.String("server", "localhost:9000", "tunnel server address")
	localAddr := flag.String("local", "", "local service to expose (e.g. localhost:6001)")
	token := flag.String("token", "dev-token", "auth token")

	flag.Parse()

	if *localAddr == "" {
		log.Fatal("missing --local flag (e.g. --local localhost:6001)")
	}

	conn, err := net.Dial("tcp", *serverAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	sess := protocol.NewSession(conn, conn)

	hs := &protocol.Handshake{
		Role:         protocol.RoleClient,
		Capabilities: protocol.CapHeartbeat,
		ExposeAddr:   *localAddr,
	}

	payload, err := hs.Encode()
	if err != nil {
		log.Fatal(err)
	}

	if err := sess.WriteFrame(&protocol.Frame{
		Type:    protocol.MsgHandshake,
		Payload: payload,
	}); err != nil {
		log.Fatal(err)
	}

	frame, err := sess.ReadFrame()
	if err != nil || frame.Type != protocol.MsgHandshakeAck {
		log.Fatal("handshake rejected by server")
	}

	if err := sess.WriteFrame(&protocol.Frame{
		Type:    protocol.MsgAuth,
		Payload: protocol.EncodeAuth(*token),
	}); err != nil {
		log.Fatal(err)
	}

	frame, err = sess.ReadFrame()
	if err != nil || frame.Type != protocol.MsgAuthOK {
		log.Fatal("authentication rejected by server")
	}

	frame, err = sess.ReadFrame()
	if err != nil || frame.Type != protocol.MsgBindOK {
		log.Fatal("failed to bind public port")
	}

	publicPort := protocol.DecodeUint16(frame.Payload)

	log.Printf(
		"Tunnel established: tcp://localhost:%d -> %s\n",
		publicPort,
		*localAddr,
	)

	forwarder := client.NewForwarder(sess, *localAddr)

	for {
		frame, err := sess.ReadFrame()
		if err != nil {
			log.Println("client session error:", err)
			return
		}

		if frame.Type == protocol.MsgHeartbeat {
			continue
		}

		forwarder.HandleFrame(frame)
	}
}
