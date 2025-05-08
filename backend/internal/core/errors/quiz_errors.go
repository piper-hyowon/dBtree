package errors

import (
	"fmt"
	"time"
)

func NewQuizInProgressError() DomainError {
	return NewError(
		ErrQuizInProgress,
		"이미 진행중인 퀴즈가 있음",
		nil,
		nil,
	)
}

func NewNoQuizPassedError() DomainError {
	return NewError(
		ErrNoQuizPassed,
		"통과된 퀴즈 내역 없음",
		nil,
		nil,
	)
}

func NewQuizTimeExpiredError(submitTime time.Time, limitTime time.Time) DomainError {
	return NewError(
		ErrQuizTimeExpired,
		fmt.Sprintf("퀴즈 제한 시간 만료: %v(제출), %v(제한)", submitTime, limitTime),
		nil,
		nil,
	)
}

func NewClickCircleTimeExpiredError(clickTime time.Time, limitTime time.Time) DomainError {
	return NewError(
		ErrClickCircleTimeExpired,
		fmt.Sprintf("클릭 시간 만료: %v(제출), %v(제한)", clickTime, limitTime),
		nil,
		nil,
	)
}
