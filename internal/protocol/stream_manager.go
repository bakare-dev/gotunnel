package protocol

import "sync"

type StreamManager struct {
	mu      sync.Mutex
	streams map[uint32]*Stream
	nextID  uint32
}

func NewStreamManager() *StreamManager {
	return &StreamManager{
		streams: make(map[uint32]*Stream),
		nextID:  1,
	}
}

func (m *StreamManager) Open() *Stream {
	m.mu.Lock()
	defer m.mu.Unlock()

	id := m.nextID
	m.nextID++

	stream := NewStream(id)
	m.streams[id] = stream
	return stream
}

func (m *StreamManager) Get(id uint32) (*Stream, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, ok := m.streams[id]
	return s, ok
}

func (m *StreamManager) Close(id uint32) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if s, ok := m.streams[id]; ok {
		s.Close()
		delete(m.streams, id)
	}
}
