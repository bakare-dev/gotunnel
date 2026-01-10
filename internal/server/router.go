package server

import (
	"errors"
	"log"
	"net"
	"sync"

	"github.com/bakare-dev/gotunnel/internal/protocol"
)

var ErrNoSessionForPort = errors.New("no session for port")

type Router struct {
	mu       sync.RWMutex
	sessions map[int]*protocol.Session
	nextPort int
}

func NewRouter(startPort int) *Router {
	return &Router{
		sessions: make(map[int]*protocol.Session),
		nextPort: startPort,
	}
}

func (r *Router) AllocatePort(sess *protocol.Session) int {
	r.mu.Lock()
	defer r.mu.Unlock()

	port := r.nextPort
	r.nextPort++

	r.sessions[port] = sess
	sess.PublicPort = port

	return port
}

func (r *Router) Get(port int) (*protocol.Session, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	sess, ok := r.sessions[port]
	return sess, ok
}

func (r *Router) Remove(port int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.sessions, port)
}

func (r *Router) CloseAll() {
	r.mu.Lock()
	defer r.mu.Unlock()

	log.Printf("│ INFO  │ Closing %d active sessions...", len(r.sessions))

	for port, sess := range r.sessions {
		sess.Close()
		log.Printf("│ INFO  │ Closed session on port %d", port)
	}

	r.sessions = make(map[int]*protocol.Session)
}

func ExtractLocalPort(conn net.Conn) int {
	addr := conn.LocalAddr().(*net.TCPAddr)
	return addr.Port
}
