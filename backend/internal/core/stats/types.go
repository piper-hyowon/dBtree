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

func (r *DailyHarvestRequest) SetDefaults() {
	if r.Days <= 0 {
		r.Days = 7 // 기본 7일
	}
}

func (r *TransactionsRequest) SetDefaults() {
	r.PaginationParams.SetDefaults(31) // 기본 31개 (한달)
}
