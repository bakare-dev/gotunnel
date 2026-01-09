package client

import "github.com/bakare-dev/gotunnel/internal/protocol"

func (f *Forwarder) HandleFrame(frame *protocol.Frame) {
	switch frame.Type {

	case protocol.MsgStreamOpen:
		f.openStream(frame.StreamID)

	case protocol.MsgStreamData:
		f.writeToLocal(frame.StreamID, frame.Payload)

	case protocol.MsgStreamClose:
		f.closeStream(frame.StreamID)
	}
}
