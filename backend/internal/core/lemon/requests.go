package lemon

import "time"

type HarvestRequest struct {
	PositionID *int `json:"position_id"`
}

type HarvestResponse struct {
	HarvestAmount   int           `json:"harvest_amount"`    // 수확한 레몬 수
	NewBalance      int           `json:"new_balance"`       // 수확 후 잔액
	NextHarvestTime time.Duration `json:"next_harvest_time"` // 다음 수확까지 남은 시간
	TransactionID   string        `json:"transaction_id"`    // 생성된 트랜잭션 ID
}

type TreeStatusResponse struct {
	AvailablePositions []int      `json:"available_positions"`          // 현재 수확 가능한 레몬 위치 ID 목록
	TotalHarvested     int        `json:"total_harvested"`              // 모든 사용자가 수확한 총 레몬 수
	NextRegrowthTime   *time.Time `json:"next_regrowth_time,omitempty"` // 다음 레몬이 자라는 시간
}

type CanHarvestResponse = HarvestAvailability
