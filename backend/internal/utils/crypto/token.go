package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

func GenerateRandomToken(byteLength int) (string, error) {
	tokenBytes := make([]byte, byteLength)
	_, err := io.ReadFull(rand.Reader, tokenBytes)
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}
	return base64.URLEncoding.EncodeToString(tokenBytes), nil
}
