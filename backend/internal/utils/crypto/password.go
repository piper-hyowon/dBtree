package crypto

import (
	"crypto/rand"
	"github.com/piper-hyowon/dBtree/internal/core/errors"
)

const (
	passwordLength = 16
	// 0, O, l, I 제외
	passwordChars = "abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ123456789!@#$%^&*"
)

func GenerateSecurePassword() (string, error) {
	bytes := make([]byte, passwordLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", errors.Wrap(err)
	}

	for i, b := range bytes {
		bytes[i] = passwordChars[b%byte(len(passwordChars))]
	}

	return string(bytes), nil
}

func GenerateSecurePasswordWithLength(length int) (string, error) {
	if length < 8 {
		return "", errors.NewInternalError(errors.New("password length must be at least 8"))
	}

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", errors.Wrap(err)
	}

	for i, b := range bytes {
		bytes[i] = passwordChars[b%byte(len(passwordChars))]
	}

	return string(bytes), nil
}
