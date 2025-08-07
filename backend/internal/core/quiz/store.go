package quiz

import (
	"context"
	"time"
)

type Store interface {
	/* ---------- 진행 중인 퀴즈 ---------- */

	// InProgress 유저가 진행 중인 퀴즈 조회(없으면 nil)
	InProgress(ctx context.Context, userEmail string) (*StatusInfo, error)

	// CreateInProgress 진행 중인 퀴즈 저장
	// 이미 존재하면 false (Redis SETNX)
	CreateInProgress(ctx context.Context, userEmail string, info *StatusInfo, timeLimit int) (bool, error)

	UpdateInProgress(ctx context.Context, userEmail string, attemptID int) error

	DeleteInProgress(ctx context.Context, userEmail string) error

	/* ---------- 퀴즈 관련 ---------- */

	// ByPositionID positionID로 퀴즈 조회
	ByPositionID(ctx context.Context, positionID int) (*Quiz, error)

	Random(ctx context.Context) (*Quiz, error) // 레몬 재생성시 사용
	AssignToLemon(ctx context.Context, quizID int, positionID int) error
	Deactivate(ctx context.Context, quizID int) error

	/* ---------- 퀴즈 시도 관련 ---------- */

	CreateAttempt(ctx context.Context, userID string, quizID int, positionID int, startTime time.Time) (int, error)
	AttemptByID(ctx context.Context, id int) (*Attempt, error)
	UpdateAttemptStatus(ctx context.Context, attemptID int, status Status, isCorrect *bool, selectedOption *int, submitTime time.Time) error
	UpdateAttemptHarvestStatus(ctx context.Context, attemptID int, harvestStatus HarvestStatus, clickTime time.Time) error
	DeleteAttempt(ctx context.Context, attemptID int) error

	/* ---------- 통계 관련 ---------- */

	// TodayQuizMasters 오늘 가장 많이 맞춘 사람들(상위 limit 명)
	TodayQuizMasters(ctx context.Context, limit int) ([]Master, error)
}
