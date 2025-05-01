package lemon

import "time"

// ActionType 레몬 잔액에 영향을 주는 모든 활동
type ActionType string

const (
	ActionWelcomeBonus     ActionType = "welcome_bonus"     // 회원가입 보너스
	ActionHarvest          ActionType = "harvest"           // 레몬 수확
	ActionInstanceCreate   ActionType = "instance_create"   // DB 인스턴스 생성
	ActionInstanceMaintain ActionType = "instance_maintain" // 인스턴스 유지 비용
)

type Status string

const (
	StatusInProgress Status = "in_progress"
	StatusSuccessful Status = "successful"
	StatusFailed     Status = "failed"
)

// Transaction 레몬 잔액 변화 기록
type Transaction struct {
	ID         string
	UserID     string
	InstanceID string
	ActionType ActionType
	Status     Status
	Amount     int
	Balance    int
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Note       string
}

type HarvestRules struct {
	BaseAmount      int
	CooldownPeriod  time.Duration
	MaxStoredLemons int
}

var DefaultHarvestRules = HarvestRules{
	BaseAmount:      5.0,
	CooldownPeriod:  6 * time.Hour,
	MaxStoredLemons: 500.0,
}
