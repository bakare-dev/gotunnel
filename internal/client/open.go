package client

import (
	"log"
	"net"
)

func (f *Forwarder) openStream(streamID uint32) {
	conn, err := net.Dial("tcp", f.targetAddr)
	if err != nil {
		log.Printf("│ ERROR │ [Stream %d] Failed to connect to %s: %v", streamID, f.targetAddr, err)
		return
	}

	f.mu.Lock()
	f.conns[streamID] = conn
	f.mu.Unlock()

	go f.pipeLocalToTunnel(streamID, conn)
}
