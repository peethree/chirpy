package auth

import (
	"crypto/rand"
	"encoding/hex"
)

func MakeRefreshToken() string, error {
	// 32 bytes of random data
	n := 32
	randomData := make([]byte, n)
	_, err := rand.Read(randomData)
	if err != nil {
		return "", err
	}

	// convert to string
	token := hex.EncodeToString(randomData)

	return token, nil
}