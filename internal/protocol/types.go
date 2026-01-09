package protocol

type ProtocolVersion uint8
type MessageType uint8
type PeerRole uint8
type Capability uint64

const (
	ProtocolVersion1 ProtocolVersion = 1
)

const (
	MsgHandshake MessageType = iota + 1
	MsgHandshakeAck

	MsgAuth
	MsgAuthOK
	MsgAuthErr

	MsgBindOK

	MsgStreamOpen
	MsgStreamData
	MsgStreamClose

	MsgHeartbeat
	MsgError
)

const (
	RoleClient PeerRole = 1
	RoleServer PeerRole = 2
)

const (
	CapHeartbeat Capability = 1 << iota
	CapCompression
	CapReconnect
	CapMetrics
)
