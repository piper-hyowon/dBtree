package quiz

import (
	"fmt"
	"github.com/piper-hyowon/dBtree/internal/core/errors"
	"runtime/debug"
)

type StatusInfo struct {
	QuizID         int
	PositionID     int
	StartTimestamp int64
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
