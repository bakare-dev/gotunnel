package protocol

type Stream struct {
	ID     uint32
	in     chan []byte
	closed chan struct{}
}

func NewStream(id uint32) *Stream {
	return &Stream{
		ID:     id,
		in:     make(chan []byte, 16),
		closed: make(chan struct{}),
	}
}

func (s *Stream) Close() {
	close(s.closed)
}
