package lemon

import "time"

type HarvestRequest struct {
	PositionID int `json:"position_id"`
}

type HarvestResponse struct {
	NewBalance    int    `json:"new_balance"`
	TransactionID string `json:"transaction_id"`
}

type TreeStatus struct {
	AvailablePositions []int      `json:"available_positions"`          // 현재 수확 가능한 레몬 위치 ID 목록
	TotalHarvested     int        `json:"total_harvested"`              // 모든 사용자가 수확한 총 레몬 수
	NextRegrowthTime   *time.Time `json:"next_regrowth_time,omitempty"` // 다음 레몬이 자라는 시간
}

type RegrowthRules struct {
	RegrowthPeriod time.Duration `json:"regrowth_period"`
	MaxPositions   int           `json:"max_positions"`
}

var DefaultRegrowthRules = RegrowthRules{
	RegrowthPeriod: 1 * time.Hour, // 1시간마다 재생성
	MaxPositions:   10,
}
