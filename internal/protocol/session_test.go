package protocol

import (
	"bytes"
	"testing"
)

func TestSessionRequiresHandshake(t *testing.T) {
	buf := new(bytes.Buffer)
	sess := NewSession(buf, buf)

	frame := &Frame{
		Version: ProtocolVersion1,
		Type:    MsgStreamData,
		Payload: []byte("data"),
	}

	if err := sess.WriteFrame(frame); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	_, err := sess.ReadFrame()
	if err != ErrHandshakeRequired {
		t.Fatalf("expected ErrHandshakeRequired, got %v", err)
	}
}

func TestSessionHandshakeSuccess(t *testing.T) {
	buf := new(bytes.Buffer)
	sess := NewSession(buf, buf)

	hs := &Handshake{
		Role:         RoleClient,
		Capabilities: CapHeartbeat,
	}

	payload, _ := hs.Encode()

	frame := &Frame{
		Version: ProtocolVersion1,
		Type:    MsgHandshake,
		Payload: payload,
	}

	if err := sess.WriteFrame(frame); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	_, err := sess.ReadFrame()
	if err != nil {
		t.Fatalf("handshake failed: %v", err)
	}

	if sess.state != StateHandshaken {
		t.Fatalf("expected session to be in StateHandshaken, got %v", sess.state)
	}
}

func TestSessionAuthRequired(t *testing.T) {
	buf := new(bytes.Buffer)
	sess := NewSession(buf, buf)

	hs := &Handshake{
		Role:         RoleClient,
		Capabilities: CapHeartbeat,
	}

	payload, _ := hs.Encode()

	_ = sess.WriteFrame(&Frame{
		Type:    MsgHandshake,
		Payload: payload,
	})

	_, _ = sess.ReadFrame()

	_ = sess.WriteFrame(&Frame{
		Type:    MsgStreamData,
		Payload: []byte("data"),
	})

	_, err := sess.ReadFrame()
	if err != ErrAuthRequired {
		t.Fatalf("expected ErrAuthRequired, got %v", err)
	}
}
