package quiz

type StartQuizWithPositionResponse struct {
	Question  string   `json:"question"`
	Options   []string `json:"options"`
	TimeLimit int      `json:"time_limit"` // 초 단위
}

type SubmitAnswerRequest struct {
	OptionIdx int `json:"option_idx"`
}
