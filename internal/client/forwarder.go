package client

import (
	"net"
	"sync"

	"github.com/bakare-dev/gotunnel/internal/protocol"
)

type Forwarder struct {
	sess       *protocol.Session
	targetAddr string

	mu    sync.Mutex
	conns map[uint32]net.Conn
}

func NewForwarder(sess *protocol.Session, targetAddr string) *Forwarder {
	return &Forwarder{
		sess:       sess,
		targetAddr: targetAddr,
		conns:      make(map[uint32]net.Conn),
	}
}
