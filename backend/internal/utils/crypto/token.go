package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/piper-hyowon/dBtree/internal/core/errors"
	"io"
)

func GenerateRandomToken(byteLength int) (string, error) {
	tokenBytes := make([]byte, byteLength)
	_, err := io.ReadFull(rand.Reader, tokenBytes)
	if err != nil {
		return "", errors.Wrap(err)
	}
	return base64.URLEncoding.EncodeToString(tokenBytes), nil
}
