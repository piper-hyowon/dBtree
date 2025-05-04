package errors

import "fmt"

func NewInvalidOTPError() DomainError {
	return NewError(
		ErrInvalidOTP,
		"유효하지 않은 인증 코드입니다",
		nil,
		nil,
	)
}

func NewExpiredOTPError() DomainError {
	return NewError(
		ErrExpiredOTP,
		"만료된 인증 코드입니다",
		nil,
		nil,
	)
}

func NewSessionNotFoundError() DomainError {
	return NewError(
		ErrSessionNotFound,
		"세션을 찾을 수 없습니다",
		nil,
		nil,
	)
}

func NewInvalidEmailError(msg string) DomainError {
	var cause error
	if msg != "" {
		cause = New(msg)
	}

	return NewError(
		ErrInvalidEmail,
		"유효하지 않은 이메일 주소",
		nil,
		cause,
	)
}

func NewTooManyResendsError(maxResends int) DomainError {
	return NewError(
		ErrTooManyResends,
		fmt.Sprintf("OTP 전송 횟수 제한(%d회)에 도달했습니다", maxResends),
		map[string]int{"maxResends": maxResends},
		nil,
	)
}

func NewTooEarlyResendError(waitSeconds int) DomainError {
	return NewError(
		ErrTooEarlyResend,
		fmt.Sprintf("OTP 재전송은 %d초 후에 가능합니다", waitSeconds),
		map[string]int{"waitSeconds": waitSeconds},
		nil,
	)
}

func NewInvalidTokenError() DomainError {
	return NewError(
		ErrInvalidToken,
		"유효하지 않은 토큰입니다",
		nil,
		nil,
	)
}

func NewUnauthorizedError() DomainError {
	return NewError(
		ErrUnauthorized,
		"인증되지 않은 요청입니다",
		nil,
		nil,
	)
}

func NewSessionExpiredError() DomainError {
	return NewError(
		ErrSessionExpired,
		"세션이 만료되었습니다",
		nil,
		nil,
	)
}
