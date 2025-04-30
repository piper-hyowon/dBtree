package core

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidEmail           = errors.New("유효하지 않은 이메일 주소")
	ErrTooManyResends         = errors.New("OTP 요청 횟수 초과, 잠시후 재시도")
	ErrTooEarlyResend         = errors.New("OTP 재전송은 1분 후 가능")
	ErrInvalidOTP             = errors.New("invalid OTP")
	ErrExpiredOTP             = errors.New("expired OTP")
	ErrSessionNotFound        = errors.New("session 404")
	ErrSessionExpired         = errors.New("session expired, 재인증 하시오")
	ErrSessionAlreadyVerified = errors.New("session already verified")

	ErrInternal     = errors.New("500")
	ErrTokenExpired = errors.New("token expired")
	ErrInvalidToken = errors.New("invalid token")
	ErrUnauthorized = errors.New("unauthorized")

	ErrUserNotFound = errors.New("user not found")
)

func NewEmailValidationError(msg string) error {
	return fmt.Errorf("%w: %s", ErrInvalidEmail, msg)
}
