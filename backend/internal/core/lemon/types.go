package lemon

import "time"

type Position struct {
	PositionID      int       `json:"position_id"` // 고정된 위치 ID(0~9)
	IsAvailable     bool      `json:"is_available"`
	LastHarvestedAt time.Time `json:"last_harvested_at"` // 마지막 수확 시간
	NextAvailableAt time.Time `json:"next_available_at"` // 다음 수확 가능 시간
}

const (
	WelcomeBonusAmount = 50
)

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
	BonusAmounts    map[int]int
}

var DefaultHarvestRules = HarvestRules{
	BaseAmount:      5,
	CooldownPeriod:  1 * time.Hour,
	MaxStoredLemons: 500,

	BonusAmounts: map[int]int{
		3:  5,  // 3일 연속 수확
		7:  15, // 7일
		30: 50, // 30일
	},
}

type HarvestResult struct {
	HarvestAmount   int           `json:"harvest_amount"`    // 수확한 레몬 수
	NewBalance      int           `json:"new_balance"`       // 수확 후 잔액
	NextHarvestTime time.Duration `json:"next_harvest_time"` // 다음 수확까지 남은 시간
	TransactionID   string        `json:"transaction_id"`    // 생성된 트랜잭션 ID
}

type HarvestAvailability struct {
	CanHarvest bool          `json:"can_harvest"` // 수확 가능 여부
	WaitTime   time.Duration `json:"wait_time"`   // 기다려야 하는 시간
}
