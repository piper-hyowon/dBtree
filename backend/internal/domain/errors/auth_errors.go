package errors

import (
	"errors"
)

var (
	ErrInvalidEmail    = errors.New("invalid email")
	ErrTooManyResends  = errors.New("OTP 요청 횟수 초과, 잠시후 재시도")
	ErrTooEarlyResend  = errors.New("OTP 재전송은 1분 후 가능")
	ErrInvalidOTP      = errors.New("invalid OTP")
	ErrExpiredOTP      = errors.New("expired OTP")
	ErrSessionNotFound = errors.New("session 404")
	ErrInternal        = errors.New("500")
	ErrTokenExpired    = errors.New("token expired")
	ErrInvalidToken    = errors.New("invalid token")
	ErrUnauthorized    = errors.New("unauthorized")
)
