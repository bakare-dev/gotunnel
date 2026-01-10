package client

import (
	"time"

	"github.com/bakare-dev/gotunnel/internal/protocol"
	"github.com/bakare-dev/gotunnel/internal/tunnel"
)

func (f *Forwarder) HandleFrame(frame *protocol.Frame) {
	switch frame.Type {

	case protocol.MsgStreamOpen:
		f.mu.Lock()
		f.httpLogs[frame.StreamID] = &tunnel.HTTPLog{
			StartTime: time.Now(),
		}
		f.mu.Unlock()
		f.openStream(frame.StreamID)

	case protocol.MsgStreamData:
		f.writeToLocal(frame.StreamID, frame.Payload)

	case protocol.MsgStreamClose:
		f.closeStream(frame.StreamID)
	}
}
