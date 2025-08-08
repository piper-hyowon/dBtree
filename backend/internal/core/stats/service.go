package stats

import (
	"context"
	"github.com/piper-hyowon/dBtree/internal/core/dbservice"
	"github.com/piper-hyowon/dBtree/internal/core/lemon"
)

type Service interface {
	GetGlobalStats(ctx context.Context) (*GlobalStats, error)
	GetMiniLeaderboard(ctx context.Context) (*MiniLeaderboard, error) // TOP 3
	GetUserDailyHarvest(ctx context.Context, userID string, req *DailyHarvestRequest) ([]*lemon.DailyHarvest, error)
	GetUserTransactions(ctx context.Context, userID string, req *TransactionsRequest) (*TransactionsResponse, error)
	GetUserInstances(ctx context.Context, userID string) ([]*dbservice.UserInstanceSummary, error)
}
