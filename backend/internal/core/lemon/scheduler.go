package lemon

import "context"

type Scheduler interface {
	Start() error
	Stop() error
	IsRunning() bool
	RunNow(ctx context.Context) (int, error)
}
