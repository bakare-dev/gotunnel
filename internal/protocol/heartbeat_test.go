package protocol

import (
	"bytes"
	"testing"
	"time"
)

func TestSessionExpiresWithoutHeartbeat(t *testing.T) {
	buf := new(bytes.Buffer)
	sess := NewSession(buf, buf)

	sess.state = StateAuthenticated
	sess.StartHeartbeat()

	time.Sleep(HeartbeatTimeout + 1*time.Second)

	_, err := sess.ReadFrame()
	if err != ErrSessionExpired {
		t.Fatalf("expected ErrSessionExpired, got %v", err)
	}
}
