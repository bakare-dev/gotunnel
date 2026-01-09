package protocol

import "errors"

var (
	ErrShortHeader      = errors.New("protocol: short frame header")
	ErrInvalidLength    = errors.New("protocol: invalid payload length")
	ErrPayloadTooLarge  = errors.New("protocol: payload exceeds maximum allowed size")
	ErrUnsupportedProto = errors.New("protocol: unsupported protocol version")

	ErrHandshakeRequired = errors.New("protocol: handshake required")
	ErrAuthRequired      = errors.New("protocol: authentication required")
	ErrIncompatiblePeers = errors.New("protocol: incompatible peer capabilities")
	ErrAuthFailed        = errors.New("protocol: authentication failed")
	ErrSessionExpired    = errors.New("protocol: session expired (heartbeat timeout)")
)
