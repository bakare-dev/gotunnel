package protocol

import (
	"encoding/binary"
	"io"
)

const (
	MaxPayloadSize = 16 * 1024 * 1024
)

type Frame struct {
	Version  ProtocolVersion
	Type     MessageType
	StreamID uint32
	Payload  []byte
}

func NewFrame(typ MessageType, payload []byte) *Frame {
	return &Frame{
		Version:  ProtocolVersion1,
		Type:     typ,
		StreamID: 0,
		Payload:  payload,
	}
}

func NewStreamFrame(typ MessageType, streamID uint32, payload []byte) *Frame {
	return &Frame{
		Version:  ProtocolVersion1,
		Type:     typ,
		StreamID: streamID,
		Payload:  payload,
	}
}

func (f *Frame) Encode(w io.Writer) error {
	if err := binary.Write(w, binary.BigEndian, uint8(f.Version)); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, uint8(f.Type)); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, f.StreamID); err != nil {
		return err
	}

	length := uint32(len(f.Payload))
	if err := binary.Write(w, binary.BigEndian, length); err != nil {
		return err
	}

	_, err := w.Write(f.Payload)
	return err
}

func DecodeFrame(r io.Reader) (*Frame, error) {
	var version uint8
	if err := binary.Read(r, binary.BigEndian, &version); err != nil {
		return nil, ErrShortHeader
	}

	if ProtocolVersion(version) != ProtocolVersion1 {
		return nil, ErrUnsupportedProto
	}

	var typ uint8
	if err := binary.Read(r, binary.BigEndian, &typ); err != nil {
		return nil, ErrShortHeader
	}

	var streamID uint32
	if err := binary.Read(r, binary.BigEndian, &streamID); err != nil {
		return nil, ErrShortHeader
	}

	var length uint32
	if err := binary.Read(r, binary.BigEndian, &length); err != nil {
		return nil, ErrShortHeader
	}

	if length > MaxPayloadSize {
		return nil, ErrPayloadTooLarge
	}

	payload := make([]byte, length)
	if _, err := io.ReadFull(r, payload); err != nil {
		return nil, ErrInvalidLength
	}

	return &Frame{
		Version:  ProtocolVersion(version),
		Type:     MessageType(typ),
		StreamID: streamID,
		Payload:  payload,
	}, nil
}
