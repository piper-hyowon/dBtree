package quiz

import (
	"context"
	"time"
)

type Store interface {
	// CheckInProgress 진행 중인 퀴즈 조회(없으면 nil)
	CheckInProgress(ctx context.Context, userEmail string) (*StatusInfo, error)

	// SaveInProgress 진행 중인 퀴즈 저장
	// 이미 존재하면 false (SETNX)
	SaveInProgress(ctx context.Context, userEmail string, info *StatusInfo, ttl time.Duration) (bool, error)

	DeleteInProgress(ctx context.Context, userEmail string) error

	SavePassed(ctx context.Context, userEmail string, positionID int, passedTime int64, ttl time.Duration) error
	PassedTime(ctx context.Context, userEmail string, positionID int) (int64, error) // 통과 이력 없으면 0, nil
	DeletePassed(ctx context.Context, userEmail string, positionID int) error
}
