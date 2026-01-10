package protocol

type Stream struct {
	ID     uint32
	In     chan []byte
	closed chan struct{}
}

func NewStream(id uint32) *Stream {
	return &Stream{
		ID:     id,
		In:     make(chan []byte, 16),
		closed: make(chan struct{}),
	}
}

func (s *Stream) Close() {
	select {
	case <-s.closed:
	default:
		close(s.closed)
		close(s.In)
	}
}

func (s *Stream) Done() <-chan struct{} {
	return s.closed
}
