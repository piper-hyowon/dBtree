package lemon

import (
	"context"
	"time"
)

// Store 레몬 크레딧 시스템(유저 잔액 & 레몬 트랜잭션)
type Store interface {
	CreateTransaction(ctx context.Context, tx *Transaction) error
	TransactionByID(ctx context.Context, id string) (*Transaction, error)
	ByPositionID(ctx context.Context, positionID int) (*Lemon, error)
	TransactionListByUserID(ctx context.Context, userID string, limit, offset int) ([]*Transaction, error)
	TransactionListByInstanceID(ctx context.Context, instanceID string, limit, offset int) ([]*Transaction, error)

	UserBalance(ctx context.Context, userID string) (int, error)
	UserLastHarvestTime(ctx context.Context, userID string) (*time.Time, error)

	AvailablePositions(ctx context.Context) ([]int, error)
	TotalHarvestedCount(ctx context.Context) (int, error) // 총 수확량 반환
	UserTotalHarvestedCount(ctx context.Context, userID string) (int, error)
	HarvestWithTransaction(ctx context.Context, positionID int, userID string, amount int, now time.Time) (string, error)
	RegrowLemons(ctx context.Context, now time.Time) (int, error) // 수확후 일정시간이 지난 재생성된 레몬 수 반환
	NextRegrowthTime(ctx context.Context) (*time.Time, error)
}
