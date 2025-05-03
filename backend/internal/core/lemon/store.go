package lemon

import (
	"context"
	"time"
)

// Store 레몬 크레딧 시스템(유저 잔액 & 레몬 트랜잭션)
type Store interface {
	CreateTransaction(ctx context.Context, tx *Transaction) error
	FindTransactionByID(ctx context.Context, id string) (*Transaction, error)
	FindTransactionsByUserID(ctx context.Context, userID string, limit, offset int) ([]*Transaction, error)
	FindTransactionsByInstanceID(ctx context.Context, instanceID string, limit, offset int) ([]*Transaction, error)
	GetUserBalance(ctx context.Context, userID string) (int, error)
	GetUserLastHarvestTime(ctx context.Context, userID string) (*time.Time, error)
	UpdateUserLastHarvestTime(ctx context.Context, userID string, lastHarvestTime time.Time) error
}
