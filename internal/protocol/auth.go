package protocol

type Auth struct {
	Token string
}

func EncodeAuth(token string) []byte {
	return []byte(token)
}

func DecodeAuth(payload []byte) (*Auth, error) {
	if len(payload) == 0 {
		return nil, ErrInvalidLength
	}

	return &Auth{Token: string(payload)}, nil
}
