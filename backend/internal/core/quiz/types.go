package quiz

import (
	"fmt"
	"github.com/piper-hyowon/dBtree/internal/core/errors"
	"runtime/debug"
	"time"
)

const (
	TimeBufferSeconds  = 3 // 만큼 더해서 퀴즈 상태 Redis TTL 에 활용(퀴즈 제한 시간보다 여유롭게 삭제)
	HarvestTimeSeconds = 5 // 원 클릭 제한 시간
)

type StatusInfo struct {
	QuizID         int
	PositionID     int
	StartTimestamp int64
	AttemptID      int
}

type Difficulty string

const (
	DifficultyEasy   Difficulty = "easy"
	DifficultyNormal Difficulty = "normal"
)

func (d Difficulty) IsValid() bool {
	switch d {
	case DifficultyEasy, DifficultyNormal:
		return true
	}
	return false
}

type Category string

const (
	CategoryBasics Category = "basics"
	CategorySQL    Category = "sql"
	CategoryDesign Category = "design"
)

func (c Category) IsValid() bool {
	switch c {
	case CategoryBasics, CategorySQL, CategoryDesign:
		return true
	}
	return false
}

// Status 퀴즈 상태
type Status string

// HarvestStatus 수확 상태
type HarvestStatus string

const (
	StatusStarted           Status        = "started"     // 퀴즈 시작 / 정답 미제출
	StatusDone              Status        = "done"        // 제출 완료(정답 여부는 is_correct로 구분)
	StatusTimeout           Status        = "timeout"     // 제한 시간 초과
	HarvestStatusNone       HarvestStatus = "none"        // 아직 수확 단계 아님(Default)
	HarvestStatusInProgress HarvestStatus = "in_progress" // 원이 나타나서 클릭 대기 중
	HarvestStatusSuccess    HarvestStatus = "success"     // 레몬 수확 성공
	HarvestStatusTimeout    HarvestStatus = "timeout"     // 원 클릭 시간 초과
	HarvestStatusFailure    HarvestStatus = "failure"     // 수확 실패 (다른 사용자가 먼저 수확)
)

type Quiz struct {
	ID               int
	Question         string
	Options          []string
	CorrectOptionIdx int
	Difficulty       Difficulty
	Category         Category
	Explanation      string
	TimeLimit        int
	PositionID       int
	UsageCount       int
	IsActive         bool
}

func NewQuiz(question string,
	options []string,
	correctOptionIdx int,
	difficulty Difficulty,
	category Category,
	explanation string,
	timeLimit int) (*Quiz, error) {
	if !difficulty.IsValid() {
		return nil, errors.NewInternalErrorWithStack(fmt.Errorf("invalid difficulty: %s", difficulty), string(debug.Stack()))
	}

	if !category.IsValid() {
		return nil, errors.NewInternalErrorWithStack(fmt.Errorf("invalid category: %s", category), string(debug.Stack()))
	}

	if correctOptionIdx < 0 || correctOptionIdx >= len(options) {
		return nil, errors.NewInternalErrorWithStack(fmt.Errorf("invalid correctOptionIdx(index out of range): %s", correctOptionIdx), string(debug.Stack()))

	}

	return &Quiz{
		Question:         question,
		Options:          options,
		CorrectOptionIdx: correctOptionIdx,
		Difficulty:       difficulty,
		Category:         category,
		Explanation:      explanation,
		TimeLimit:        timeLimit,
		IsActive:         false,
		UsageCount:       0,
	}, nil
}

func (q *Quiz) CheckAnswer(optionIndex int) bool {
	return optionIndex == q.CorrectOptionIdx
}

type Attempt struct {
	ID              string `json:"id"`
	UserID          string `json:"userID"`
	QuizID          int    `json:"quizID"`
	LemonPositionID int    `json:"lemonPositionID"`
	IsCorrect       bool   `json:"isCorrect"`
	SelectedOption  int    `json:"selectedOption"`

	StartTime        time.Time `json:"startTime"`
	SubmitTime       time.Time `json:"submitTime"`
	TimeTaken        int       `json:"timeTaken"`        //   퀴즈 푸는데 걸린 시간(초)
	TimeTakenClicked int       `json:"timeTakenClicked"` // 원 클릭하는데 걸린 시간(초)

	Status        Status        `json:"status"`
	HarvestStatus HarvestStatus `json:"harvestStatus"`
}
