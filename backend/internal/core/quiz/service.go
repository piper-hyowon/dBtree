package quiz

import "context"

type Service interface {
	StartQuiz(ctx context.Context, positionID int, userID string, userEmail string) (*StartQuizWithPositionResponse, error)
	SubmitAnswer(ctx context.Context, userEmail string, attemptID int, optionIdx int) (*SubmitAnswerResponse, error)
}
