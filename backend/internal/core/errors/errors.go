package errors

import (
	"errors"
	"fmt"
	"runtime"
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
	ErrEndpointNotFound ErrorCode = 1103

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

	ErrHarvestCooldown    ErrorCode = 1500
	ErrLemonStorageFull   ErrorCode = 1501
	ErrInsufficientLemons ErrorCode = 1502

	ErrQuizInProgress          ErrorCode = 1600
	ErrNoQuizPassed            ErrorCode = 1601
	ErrHarvestAlreadyProcessed ErrorCode = 1602 // 이미 수확 처리가 완료
	ErrQuizTimeExpired         ErrorCode = 1603
	ErrClickCircleTimeExpired  ErrorCode = 1604
	ErrNoQuizInProgress        ErrorCode = 1605
	ErrLemonAlreadyHarvested   ErrorCode = 1606

	ErrInvalidStatusTransition ErrorCode = 1700
	ErrInstanceQuotaExceeded   ErrorCode = 1701
	ErrInvalidInstanceName     ErrorCode = 1702
	ErrInvalidResourceSpec     ErrorCode = 1703
	ErrInstanceNotReady        ErrorCode = 1704
)

var errorStrings = map[ErrorCode]string{
	ErrUnknown:                 "unknown_error",
	ErrInternalServer:          "internal_server_error",
	ErrInvalidParameter:        "invalid_parameter",
	ErrMissingParameter:        "missing_parameter",
	ErrMethodNotAllowed:        "method_not_allowed",
	ErrEndpointNotFound:        "endpoint_not_found",
	ErrInvalidOTP:              "invalid_otp",
	ErrExpiredOTP:              "expired_otp",
	ErrSessionNotFound:         "session_not_found",
	ErrTooManyResends:          "too_many_resends",
	ErrTooEarlyResend:          "too_early_resend",
	ErrInvalidToken:            "invalid_token",
	ErrUnauthorized:            "unauthorized",
	ErrSessionExpired:          "session_expired",
	ErrAlreadyVerified:         "already_verified",
	ErrInvalidEmail:            "invalid_email",
	ErrResourceNotFound:        "resource_not_found",
	ErrResourceConflict:        "resource_conflict",
	ErrHarvestCooldown:         "harvest_cooldown",
	ErrLemonStorageFull:        "lemon_storage_full",
	ErrInsufficientLemons:      "insufficient_lemons",
	ErrQuizInProgress:          "quiz_in_progress",
	ErrNoQuizPassed:            "no_quiz_passed",
	ErrQuizTimeExpired:         "time_expired",
	ErrClickCircleTimeExpired:  "click_circle_time_expired",
	ErrNoQuizInProgress:        "no_quiz_in_progress",
	ErrLemonAlreadyHarvested:   "lemon_already_harvested",
	ErrInvalidStatusTransition: "invalid_status_transition",
	ErrInstanceQuotaExceeded:   "instance_quota_exceeded",
	ErrInvalidInstanceName:     "invalid_instance_name",
	ErrInvalidResourceSpec:     "invalid_resource_spec",
	ErrInstanceNotReady:        "instance_not_ready",
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

func (e *baseDomainError) Is(target error) bool {
	t, ok := target.(*baseDomainError)
	if !ok {
		return false
	}
	return e.code == t.code
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

// ===== Helper Functions =====

// Wrap 일반 에러를 내부 서버 에러로 래핑
// 호출 위치 스택 정보(파일:라인:함수명) 캡처
func Wrap(err error) error {
	if err == nil {
		return nil
	}

	pc, file, line, _ := runtime.Caller(1)
	fn := runtime.FuncForPC(pc)
	stack := fmt.Sprintf("%s:%d %s()", file, line, fn.Name())

	return NewInternalErrorWithStack(err, stack)
}

func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	message := fmt.Sprintf(format, args...)
	wrapped := fmt.Errorf("%s: %w", message, err)

	pc, file, line, _ := runtime.Caller(1)
	fn := runtime.FuncForPC(pc)
	stack := fmt.Sprintf("%s:%d %s()", file, line, fn.Name())

	return NewInternalErrorWithStack(wrapped, stack)
}
