package client

import (
	"log"
	"net"

	"github.com/bakare-dev/gotunnel/internal/protocol"
)

func (f *Forwarder) pipeLocalToTunnel(streamID uint32, conn net.Conn) {
	buf := make([]byte, 4096)

	for {
		n, err := conn.Read(buf)
		if err != nil {
			break
		}

		err = f.sess.WriteFrame(&protocol.Frame{
			Type:     protocol.MsgStreamData,
			StreamID: streamID,
			Payload:  buf[:n],
		})
		if err != nil {
			log.Println("write to tunnel failed:", err)
			break
		}
	}

	_ = f.sess.WriteFrame(&protocol.Frame{
		Type:     protocol.MsgStreamClose,
		StreamID: streamID,
	})
}
