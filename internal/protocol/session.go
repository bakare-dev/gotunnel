package protocol

import (
	"io"
	"sync"
	"time"
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

	lastSeen time.Time
	mu       sync.Mutex
	closed   chan struct{}
}

func NewSession(r io.Reader, w io.Writer) *Session {
	return &Session{
		r:       r,
		w:       w,
		state:   StateInit,
		streams: NewStreamManager(),
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

	s.touch()
	return frame, nil
}

func (s *Session) WriteFrame(f *Frame) error {
	f.Version = ProtocolVersion1
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
			_ = s.WriteFrame(&Frame{
				Type: MsgHeartbeat,
			})
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
				close(s.closed)
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
