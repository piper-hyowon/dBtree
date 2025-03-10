package model

import "time"

// 레몬 잔액에 영향을 주는 모든 활동
type ActionType string

const (
	ActionHarvest      ActionType = "harvest"
	ActionInstanceCost ActionType = "instance_cost"
	ActionUsageCost    ActionType = "usage_cost"
	ActionWelcomeBonus ActionType = "welcome_bonus"
)

// 레몬 잔액 변화 기록
type LemonTransaction struct {
	ID         string
	UserID     string
	InstanceID string
	ActionType ActionType
	Amount     float64
	Balance    float64
	Timestamp  time.Time
	Note       string
}
