package auth

import (
	"crypto/rand"
	"encoding/hex"
)

func MakeRefreshToken() (token string) {
	key := make([]byte, 32)

	rand.Read(key)

	token = hex.EncodeToString(key)

	return
}
