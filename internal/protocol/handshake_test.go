package protocol

import "testing"

func TestHandshakeEncodeDecode(t *testing.T) {
	h := &Handshake{
		Role:         RoleClient,
		Capabilities: CapHeartbeat | CapReconnect,
	}

	payload, err := h.Encode()
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	decoded, err := DecodeHandshake(payload)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if decoded.Role != h.Role {
		t.Fatalf("role mismatch")
	}

	if decoded.Capabilities != h.Capabilities {
		t.Fatalf("capability mismatch")
	}
}

func TestCapabilityNegotiation(t *testing.T) {
	clientCaps := CapHeartbeat | CapReconnect
	serverCaps := CapHeartbeat | CapMetrics

	common, err := Negotiate(clientCaps, serverCaps)
	if err != nil {
		t.Fatalf("negotiation failed")
	}

	if common != CapHeartbeat {
		t.Fatalf("unexpected negotiated capabilities")
	}
}

func TestCapabilityNegotiationFailure(t *testing.T) {
	_, err := Negotiate(CapReconnect, CapMetrics)
	if err == nil {
		t.Fatalf("expected negotiation failure")
	}
}
