package stats

import "context"

type Service interface {
	GetGlobalStats(ctx context.Context) (*GlobalStats, error)
	GetMiniLeaderboard(ctx context.Context) (*MiniLeaderboard, error) // TOP 3
}
