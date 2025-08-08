package stats

type GlobalStats struct {
	TotalHarvestedLemons  int `json:"totalHarvestedLemons"`
	TotalCreatedInstances int `json:"totalCreatedInstances"`
	TotalUsers            int `json:"totalUsers"`
}

type MiniLeaderboard struct {
	LemonRichUsers []UserRank `json:"lemonRichUsers"`
	QuizMasters    []UserRank `json:"quizMasters"`
}

type UserRank struct {
	MaskedEmail string `json:"maskedEmail"`
	Score       int    `json:"score"`
	Rank        int    `json:"rank"`
}
