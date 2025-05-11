package lemon

import "context"

type Scheduler interface {
	Start() error
	Stop() error
	IsRunning() bool
	RunNow(ctx context.Context) ([]int, error)
	InitializeLemons(ctx context.Context) error
}

// TODO: IsRunning, RunNow 호출하는 관리자용 API
