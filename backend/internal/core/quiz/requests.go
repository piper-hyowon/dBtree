package quiz

import "time"

type StartQuizWithPositionResponse struct {
	Question  string   `json:"question"`
	Options   []string `json:"options"`
	TimeLimit int      `json:"timeLimit"` // 초 단위
	AttemptID int      `json:"attemptID"`
}

type SubmitAnswerRequest struct {
	OptionIdx *int `json:"optionIdx"`
	AttemptID *int `json:"attemptID"`
}

type SubmitAnswerResponse struct {
	IsCorrect        bool       `json:"isCorrect"`
	Status           Status     `json:"status"`
	CorrectOption    int        `json:"correctOption"`
	HarvestEnabled   bool       `json:"harvestEnabled,omitempty"`
	HarvestTimeoutAt *time.Time `json:"harvestTimeoutAt,omitempty"`
	AttemptID        int        `json:"attemptID"`
}
