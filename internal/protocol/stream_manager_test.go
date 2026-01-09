package protocol

import (
	"testing"
)

func TestStreamManagerOpenGetClose(t *testing.T) {
	m := NewStreamManager()

	s1 := m.Open()
	s2 := m.Open()

	if s1.ID == s2.ID {
		t.Fatalf("stream IDs must be unique")
	}

	if _, ok := m.Get(s1.ID); !ok {
		t.Fatalf("stream 1 missing")
	}

	m.Close(s1.ID)

	if _, ok := m.Get(s1.ID); ok {
		t.Fatalf("stream 1 should be closed")
	}
}
