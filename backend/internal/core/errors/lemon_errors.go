package errors

import (
	"fmt"
	"time"
)

func NewHarvestCooldownError(nextHarvestTime time.Duration) DomainError {
	return NewError(
		ErrHarvestCooldown,
		fmt.Sprintf("아직 레몬 수확 불가(%f 분 후에 가능)", nextHarvestTime.Minutes()),
		map[string]int{"nextHarvestTime": int(nextHarvestTime.Minutes())},
		nil,
	)
}

func NewLemonStorageFullError(maxLemon int) DomainError {
	return NewError(
		ErrLemonStorageFull,
		fmt.Sprintf("레몬 저장소가 가득 찼습니다(최대 보유 가능 레몬 %d 개)", maxLemon),
		nil,
		nil,
	)
}

func NewInsufficientLemonsError(required int, missing int) DomainError {
	return NewError(
		ErrInsufficientLemons,
		fmt.Sprintf(" 최소 %d 레몬이 필요합니다: %d 개 부족", required, missing),
		map[string]int{"required": required, "missing": missing},
		nil,
	)
}

func NewLemonAlreadyHarvestedError() DomainError {
	return NewError(
		ErrLemonAlreadyHarvested,
		"다른 사용자가 해당 레몬을 먼저 수확했습니다",
		nil,
		nil,
	)
}
