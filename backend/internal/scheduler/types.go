package scheduler

import "context"

type Scheduler interface {
	Start() error
	Stop() error
	IsRunning() bool
}

// ManualRunScheduler 수동 실행 가능한 스케줄러
type ManualRunScheduler interface {
	Scheduler
	RunNow(ctx context.Context) error
}
