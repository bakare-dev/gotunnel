package server

import (
	"log"
	"net"

	"github.com/bakare-dev/gotunnel/internal/protocol"
)

func (p *PublicListener) handleConn(conn net.Conn) {
	defer conn.Close()

	port := ExtractLocalPort(conn)

	sess, ok := p.router.Get(port)
	if !ok {
		log.Println("no session for port", port)
		return
	}

	stream := sess.Streams().Open()
	log.Println("opened stream", stream.ID, "for port", port)

	_ = sess.WriteFrame(&protocol.Frame{
		Type:     protocol.MsgStreamOpen,
		StreamID: stream.ID,
	})

	buf := make([]byte, 4096)

	for {
		n, err := conn.Read(buf)
		if err != nil {
			break
		}

		_ = sess.WriteFrame(&protocol.Frame{
			Type:     protocol.MsgStreamData,
			StreamID: stream.ID,
			Payload:  buf[:n],
		})
	}

	_ = sess.WriteFrame(&protocol.Frame{
		Type:     protocol.MsgStreamClose,
		StreamID: stream.ID,
	})
}
