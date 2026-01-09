package protocol

import "time"

const (
	HeartbeatInterval = 10 * time.Second
	HeartbeatTimeout  = 30 * time.Second
)
