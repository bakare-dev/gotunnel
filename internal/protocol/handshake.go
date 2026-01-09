package protocol

import (
	"encoding/binary"
)

type Handshake struct {
	Role         PeerRole
	Capabilities Capability
	ExposeAddr   string
}

func (h *Handshake) Encode() ([]byte, error) {
	expose := []byte(h.ExposeAddr)

	buf := make([]byte, 1+8+2+len(expose))
	buf[0] = byte(h.Role)
	binary.BigEndian.PutUint64(buf[1:], uint64(h.Capabilities))
	binary.BigEndian.PutUint16(buf[9:], uint16(len(expose)))
	copy(buf[11:], expose)

	return buf, nil
}

func DecodeHandshake(payload []byte) (*Handshake, error) {
	if len(payload) < 11 {
		return nil, ErrInvalidLength
	}

	role := PeerRole(payload[0])
	caps := Capability(binary.BigEndian.Uint64(payload[1:]))

	exposeLen := binary.BigEndian.Uint16(payload[9:])
	if len(payload) < int(11+exposeLen) {
		return nil, ErrInvalidLength
	}

	expose := string(payload[11 : 11+exposeLen])

	return &Handshake{
		Role:         role,
		Capabilities: caps,
		ExposeAddr:   expose,
	}, nil
}

func (h *Handshake) Has(cap Capability) bool {
	return h.Capabilities&cap != 0
}

func Negotiate(local, remote Capability) (Capability, error) {
	common := local & remote
	if common == 0 {
		return 0, ErrIncompatiblePeers
	}
	return common, nil
}
