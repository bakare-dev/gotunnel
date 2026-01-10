package client

import (
	"log"

	"github.com/bakare-dev/gotunnel/internal/tunnel"
)

func (f *Forwarder) writeToLocal(streamID uint32, data []byte) {
	f.mu.Lock()
	conn, ok := f.conns[streamID]
	httpLog, hasLog := f.httpLogs[streamID]
	f.mu.Unlock()

	if !ok {
		return
	}

	if hasLog && httpLog.Request == nil {
		httpLog.Request = tunnel.ParseHTTPRequest(data)
	}

	if _, err := conn.Write(data); err != nil {
		log.Printf("│ ERROR │ [Stream %d] Failed to write to local: %v", streamID, err)
	}
}
