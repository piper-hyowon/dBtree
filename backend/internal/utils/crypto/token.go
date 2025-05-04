package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/piper-hyowon/dBtree/internal/core/errors"
	"io"
	"runtime/debug"
)

func GenerateRandomToken(byteLength int) (string, error) {
	tokenBytes := make([]byte, byteLength)
	_, err := io.ReadFull(rand.Reader, tokenBytes)
	if err != nil {
		return "", errors.NewInternalErrorWithStack(err, string(debug.Stack()))
	}
	return base64.URLEncoding.EncodeToString(tokenBytes), nil
}
