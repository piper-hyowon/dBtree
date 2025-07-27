package errors

import (
	"fmt"
	"github.com/piper-hyowon/dBtree/internal/core/dbservice"
)

func NewInvalidStatusTransitionError(current, target dbservice.InstanceStatus) DomainError {
	return NewError(
		ErrInvalidStatusTransition,
		fmt.Sprintf("상태 %s -> %s 변경 불가", current, target),
		map[string]string{"current": string(current), "target": string(target)},
		nil,
	)
}

func NewInstanceNameConflictError(name string) DomainError {
	return NewError(
		ErrResourceConflict,
		fmt.Sprintf("'%s' 인스턴스 이름 중복", name),
		map[string]string{"name": name},
		nil,
	)
}

func NewInsufficientLemonsForInstanceError(required, current int) DomainError {
	return NewError(
		ErrInsufficientLemons,
		fmt.Sprintf("인스턴스 생성 레몬 부족: 필요: %d (현재: %d)", required, current),
		map[string]int{"required": required, "current": current},
		nil,
	)
}
