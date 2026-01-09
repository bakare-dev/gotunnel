package protocol

func (s *Session) HandleFrame(f *Frame) error {
	switch f.Type {

	case MsgStreamOpen:
		s.streams.Open()

	case MsgStreamData:
		stream, ok := s.streams.Get(f.StreamID)
		if !ok {
			return ErrInvalidLength
		}
		stream.in <- f.Payload

	case MsgStreamClose:
		s.streams.Close(f.StreamID)
	}

	return nil
}
