package protocol

func ValidateToken(token string) bool {
	return token == "dev-token"
}
