package lemon

import "time"

type HarvestRequest struct {
	PositionID *int `json:"positionId"`
}

type HarvestResponse struct {
	HarvestAmount   int           `json:"harvestAmount"`   // 수확한 레몬 수
	NewBalance      int           `json:"newBalance"`      // 수확 후 잔액
	NextHarvestTime time.Duration `json:"nextHarvestTime"` // 다음 수확까지 남은 시간
	TransactionID   string        `json:"transactionId"`   // 생성된 트랜잭션 ID
}

type TreeStatusResponse struct {
	AvailablePositions []int      `json:"availablePositions"`         // 현재 수확 가능한 레몬 위치 ID 목록
	TotalHarvested     int        `json:"totalHarvested"`             // 모든 사용자가 수확한 총 레몬 수
	NextRegrowthTime   *time.Time `json:"nextRegrowthTime,omitempty"` // 다음 레몬이 자라는 시간
}

type CanHarvestResponse = HarvestAvailability
