package errors

import (
	"fmt"
	"time"
)

func NewQuizInProgressError() DomainError {
	return NewError(
		ErrQuizInProgress,
		"이미 진행중인 퀴즈가 있음(시작 불가)",
		nil,
		nil,
	)
}

func NewNoQuizInProgressError() DomainError {
	return NewError(
		ErrNoQuizInProgress,
		"진행중인 퀴즈가 없음(제출 불가)",
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

func NewHarvestAlreadyProcessedError() DomainError {
	return NewError(
		ErrHarvestAlreadyProcessed,
		"이미 처리된 수확 프로세스입니다",
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
