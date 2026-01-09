package client

import (
	"log"
	"net"
)

func (f *Forwarder) openStream(streamID uint32) {
	conn, err := net.Dial("tcp", f.targetAddr)
	if err != nil {
		log.Println("local connect failed:", err)
		return
	}

	f.mu.Lock()
	f.conns[streamID] = conn
	f.mu.Unlock()

	log.Println("stream", streamID, "connected to", f.targetAddr)

	go f.pipeLocalToTunnel(streamID, conn)
}
