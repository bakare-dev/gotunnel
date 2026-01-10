package protocol

import (
	"io"
	"log"
	"sync"
	"time"

	"github.com/bakare-dev/gotunnel/internal/metrics"
)

type SessionState uint8

const (
	StateInit SessionState = iota
	StateHandshaken
	StateAuthenticated
)

type Session struct {
	r io.Reader
	w io.Writer

	state SessionState

	Role         PeerRole
	Capabilities Capability

	ExposeAddr string
	PublicPort int

	streams *StreamManager
	Metrics *metrics.Metrics

	lastSeen time.Time
	mu       sync.Mutex
	writeMu  sync.Mutex
	closed   chan struct{}
	once     sync.Once
}

func NewSession(r io.Reader, w io.Writer) *Session {
	return &Session{
		r:       r,
		w:       w,
		state:   StateInit,
		streams: NewStreamManager(),
		Metrics: metrics.New(),
		closed:  make(chan struct{}),
	}
}

func (s *Session) ReadFrame() (*Frame, error) {
	select {
	case <-s.closed:
		return nil, ErrSessionExpired
	default:
	}

	frame, err := DecodeFrame(s.r)
	if err != nil {
		return nil, err
	}

	s.Metrics.AddBytesReceived(int64(len(frame.Payload)))
	s.touch()
	return frame, nil
}

func (s *Session) WriteFrame(f *Frame) error {
	select {
	case <-s.closed:
		return ErrSessionExpired
	default:
	}

	s.writeMu.Lock()
	defer s.writeMu.Unlock()

	select {
	case <-s.closed:
		return ErrSessionExpired
	default:
	}

	f.Version = ProtocolVersion1
	s.Metrics.AddBytesSent(int64(len(f.Payload)))
	return f.Encode(s.w)
}

func (s *Session) ProcessHandshake(frame *Frame) error {
	if s.state != StateInit {
		return ErrHandshakeRequired
	}

	hs, err := DecodeHandshake(frame.Payload)
	if err != nil {
		return err
	}

	s.Role = hs.Role
	s.Capabilities = hs.Capabilities
	s.ExposeAddr = hs.ExposeAddr
	s.state = StateHandshaken
	return nil
}

func (s *Session) ProcessAuth(frame *Frame) error {
	if s.state != StateHandshaken {
		return ErrAuthRequired
	}

	auth, err := DecodeAuth(frame.Payload)
	if err != nil {
		return err
	}

	if !ValidateToken(auth.Token) {
		return ErrAuthFailed
	}

	s.state = StateAuthenticated
	s.StartHeartbeat()
	return nil
}

func (s *Session) Streams() *StreamManager {
	return s.streams
}

func (s *Session) StartHeartbeat() {
	s.touch()
	go s.sendHeartbeatLoop()
	go s.watchdogLoop()
}

func (s *Session) sendHeartbeatLoop() {
	ticker := time.NewTicker(HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.WriteFrame(&Frame{
				Type: MsgHeartbeat,
			}); err != nil {
				return
			}
		case <-s.closed:
			return
		}
	}
}

func (s *Session) watchdogLoop() {
	ticker := time.NewTicker(HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.mu.Lock()
			expired := time.Since(s.lastSeen) > HeartbeatTimeout
			s.mu.Unlock()

			if expired {
				s.Close()
				return
			}
		case <-s.closed:
			return
		}
	}
}

func (s *Session) touch() {
	s.mu.Lock()
	s.lastSeen = time.Now()
	s.mu.Unlock()
}

func (s *Session) IsAuthenticated() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.state == StateAuthenticated
}

func (s *Session) State() SessionState {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.state
}

func (s *Session) Close() error {
	s.once.Do(func() {
		close(s.closed)

		time.Sleep(100 * time.Millisecond)

		log.Println("│ INFO  │ Shutting down session...")

		streamIDs := s.streams.GetAllStreamIDs()
		for _, streamID := range streamIDs {
			s.streams.Close(streamID)
		}

		log.Println("│ INFO  │ Session closed gracefully")
	})
	return nil
}

func (s *Session) IsClosed() bool {
	select {
	case <-s.closed:
		return true
	default:
		return false
	}
}
