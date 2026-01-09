package protocol

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestFrameEncodeDecode(t *testing.T) {
	original := Frame{
		Version: ProtocolVersion1,
		Type:    MsgStreamData,
		Payload: []byte("hello tunnel"),
	}

	buf := new(bytes.Buffer)

	if err := original.Encode(buf); err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	decoded, err := DecodeFrame(buf)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if decoded.Version != original.Version {
		t.Fatalf("version mismatch")
	}

	if decoded.Type != original.Type {
		t.Fatalf("type mismatch")
	}

	if !bytes.Equal(decoded.Payload, original.Payload) {
		t.Fatalf("payload mismatch")
	}
}

func TestDecodeFrameShortHeader(t *testing.T) {
	buf := bytes.NewBuffer([]byte{0x01, 0x01})

	_, err := DecodeFrame(buf)
	if err == nil {
		t.Fatalf("expected error for short header")
	}
}

func TestDecodeFrameInvalidLength(t *testing.T) {
	buf := new(bytes.Buffer)

	buf.WriteByte(byte(ProtocolVersion1))
	buf.WriteByte(byte(MsgStreamData))
	buf.Write([]byte{0, 0, 0, 10})
	buf.Write([]byte("abc"))

	_, err := DecodeFrame(buf)
	if err == nil {
		t.Fatalf("expected error for invalid payload length")
	}
}
func TestDecodeFramePayloadTooLarge(t *testing.T) {
	buf := new(bytes.Buffer)

	buf.WriteByte(byte(ProtocolVersion1))

	buf.WriteByte(byte(MsgStreamData))

	if err := binary.Write(buf, binary.BigEndian, uint32(1)); err != nil {
		t.Fatalf("failed to write stream id: %v", err)
	}

	oversize := uint32(MaxPayloadSize + 1)
	if err := binary.Write(buf, binary.BigEndian, oversize); err != nil {
		t.Fatalf("failed to write length: %v", err)
	}

	_, err := DecodeFrame(buf)
	if err != ErrPayloadTooLarge {
		t.Fatalf("expected ErrPayloadTooLarge, got %v", err)
	}
}

func TestDecodeFrameUnsupportedVersion(t *testing.T) {
	buf := new(bytes.Buffer)

	buf.WriteByte(0xFF)
	buf.WriteByte(byte(MsgStreamData))

	if err := binary.Write(buf, binary.BigEndian, uint32(0)); err != nil {
		t.Fatalf("failed to write length: %v", err)
	}

	_, err := DecodeFrame(buf)
	if err != ErrUnsupportedProto {
		t.Fatalf("expected ErrUnsupportedProto, got %v", err)
	}
}
