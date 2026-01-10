package client

import (
	"net"
	"sync"

	"github.com/bakare-dev/gotunnel/internal/protocol"
	"github.com/bakare-dev/gotunnel/internal/tunnel"
)

type Forwarder struct {
	sess       *protocol.Session
	targetAddr string

	mu       sync.Mutex
	conns    map[uint32]net.Conn
	httpLogs map[uint32]*tunnel.HTTPLog
}

func NewForwarder(sess *protocol.Session, targetAddr string) *Forwarder {
	return &Forwarder{
		sess:       sess,
		targetAddr: targetAddr,
		conns:      make(map[uint32]net.Conn),
		httpLogs:   make(map[uint32]*tunnel.HTTPLog),
	}
}
