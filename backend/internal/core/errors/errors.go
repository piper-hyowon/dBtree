package errors

import (
	"errors"
	"fmt"
)

var (
	Is     = errors.Is
	As     = errors.As
	New    = errors.New
	Unwrap = errors.Unwrap
)

type ErrorCode int

const (
	ErrUnknown        ErrorCode = 1000
	ErrInternalServer ErrorCode = 1001

	ErrInvalidParameter ErrorCode = 1100
	ErrMissingParameter ErrorCode = 1101
	ErrMethodNotAllowed ErrorCode = 1102

	ErrInvalidOTP      ErrorCode = 1200
	ErrExpiredOTP      ErrorCode = 1201
	ErrSessionNotFound ErrorCode = 1202
	ErrTooManyResends  ErrorCode = 1203
	ErrTooEarlyResend  ErrorCode = 1204
	ErrInvalidToken    ErrorCode = 1205
	ErrUnauthorized    ErrorCode = 1206
	ErrSessionExpired  ErrorCode = 1207
	ErrAlreadyVerified ErrorCode = 1208

	ErrInvalidEmail ErrorCode = 1301

	ErrResourceNotFound ErrorCode = 1400
	ErrResourceConflict ErrorCode = 1401
)

var errorStrings = map[ErrorCode]string{
	ErrUnknown:          "unknown_error",
	ErrInternalServer:   "internal_server_error",
	ErrInvalidParameter: "invalid_parameter",
	ErrMissingParameter: "missing_parameter",
	ErrMethodNotAllowed: "method_not_allowed",
	ErrInvalidOTP:       "invalid_otp",
	ErrExpiredOTP:       "expired_otp",
	ErrSessionNotFound:  "session_not_found",
	ErrTooManyResends:   "too_many_resends",
	ErrTooEarlyResend:   "too_early_resend",
	ErrInvalidToken:     "invalid_token",
	ErrUnauthorized:     "unauthorized",
	ErrSessionExpired:   "session_expired",
	ErrAlreadyVerified:  "already_verified",
	ErrInvalidEmail:     "invalid_email",
	ErrResourceNotFound: "resource_not_found",
	ErrResourceConflict: "resource_conflict",
}

func (c ErrorCode) String() string {
	if s, ok := errorStrings[c]; ok {
		return s
	}
	return fmt.Sprintf("undefined_error(%d)", c)
}

func (c ErrorCode) IsValid() bool {
	_, ok := errorStrings[c]
	return ok
}

type DomainError interface {
	error
	Code() ErrorCode
	ErrorData() any
	Stack() string
}

type baseDomainError struct {
	code    ErrorCode
	message string
	data    any
	cause   error
	stack   string
}

func (e *baseDomainError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %v", e.message, e.cause)
	}
	return e.message
}

func (e *baseDomainError) Code() ErrorCode {
	return e.code
}

func (e *baseDomainError) ErrorData() any {
	return e.data
}

func (e *baseDomainError) Unwrap() error {
	return e.cause
}

func (e *baseDomainError) Stack() string {
	return e.stack
}

func NewError(code ErrorCode, message string, data any, cause error) DomainError {
	if !code.IsValid() {
		code = ErrUnknown
	}

	return &baseDomainError{
		code:    code,
		message: message,
		data:    data,
		cause:   cause,
		stack:   "",
	}
}

func NewErrorWithStack(code ErrorCode, message string, data any, cause error, stack string) DomainError {
	if !code.IsValid() {
		code = ErrUnknown
	}

	return &baseDomainError{
		code:    code,
		message: message,
		data:    data,
		cause:   cause,
		stack:   stack,
	}
}
