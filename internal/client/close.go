package client

import "log"

func (f *Forwarder) closeStream(streamID uint32) {
	f.mu.Lock()
	conn, ok := f.conns[streamID]
	if ok {
		conn.Close()
		delete(f.conns, streamID)
	}
	f.mu.Unlock()

	log.Println("stream", streamID, "closed")
}
