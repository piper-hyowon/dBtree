package errors

import (
	"errors"
	"fmt"
)

// NewInternalErrorWithStack 내부 서버 오류 발생시 호출
// 이미 DomainError 타입인 경우 중복 래핑을 방지하고 원본 에러를 그대로 반환(스택 보존)
func NewInternalErrorWithStack(cause error, stack string) DomainError {
	var domainErr DomainError
	if errors.As(cause, &domainErr) {
		return domainErr
	}

	var message string
	if cause != nil {
		message = "서버 내부 오류가 발생했습니다: " + cause.Error()
	} else {
		message = "서버 내부 오류가 발생했습니다"
	}

	return NewErrorWithStack(
		ErrInternalServer,
		message,
		nil,
		cause,
		stack,
	)
}

func NewInternalError(cause error) DomainError {
	var message string
	if cause != nil {
		message = "서버 내부 오류가 발생했습니다: " + cause.Error()
	} else {
		message = "서버 내부 오류가 발생했습니다"
	}

	return NewError(
		ErrInternalServer,
		message,
		nil,
		cause,
	)
}

func NewInvalidParameterError(param string, reason string) DomainError {
	message := "유효하지 않은 파라미터입니다"
	if reason != "" {
		message = reason
	}

	return NewError(
		ErrInvalidParameter,
		message,
		map[string]string{"parameter": param},
		nil,
	)
}

func NewMissingParameterError(param string) DomainError {
	return NewError(
		ErrMissingParameter,
		fmt.Sprintf("필수 파라미터 누락: %s", param),
		map[string]string{"parameter": param},
		nil,
	)
}

func NewMethodNotAllowedError(allowedMethods string) DomainError {
	return NewError(
		ErrMethodNotAllowed,
		"허용되지 않는 메서드입니다",
		map[string]string{"allowed": allowedMethods},
		nil,
	)
}

func NewResourceNotFoundError(resourceName string, data string) DomainError {
	return NewError(
		ErrResourceNotFound,
		"존재하지 않는 데이터",
		map[string]string{"resourceName": data},
		nil,
	)
}
