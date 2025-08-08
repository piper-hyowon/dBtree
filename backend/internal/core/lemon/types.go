package lemon

import "time"

type Lemon struct {
	PositionID      int
	IsAvailable     bool
	LastHarvestedAt time.Time
	NextAvailableAt time.Time
}

const (
	WelcomeBonusAmount = 15
)

// ActionType 레몬 잔액에 영향을 주는 모든 활동
type ActionType string

const (
	ActionWelcomeBonus         ActionType = "welcome_bonus"          // 회원가입 보너스
	ActionHarvest              ActionType = "harvest"                // 레몬 수확
	ActionInstanceCreate       ActionType = "instance_create"        // DB 인스턴스 생성
	ActionInstanceMaintain     ActionType = "instance_maintain"      // 인스턴스 유지 비용
	ActionInstanceCreateRefund ActionType = "instance_create_refund" // 인스턴스 생성 실패 환불
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
	InstanceID int
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

type HarvestAvailability struct {
	CanHarvest bool          `json:"canHarvest"` // 수확 가능 여부
	WaitTime   time.Duration `json:"waitTime"`   // 기다려야 하는 시간
}

type RegrowthRules struct {
	RegrowthPeriod time.Duration `json:"regrowthPeriod"`
	MaxPositions   int           `json:"maxPositions"`
}

var DefaultRegrowthRules = RegrowthRules{
	RegrowthPeriod: 1 * time.Hour, // 1시간마다 재생성
	MaxPositions:   10,
}

type DailyHarvest struct {
	Date   time.Time `json:"date"`
	Amount int       `json:"amount"`
}

type TransactionWithInstance struct {
	ID           string     `json:"id"`
	InstanceName *string    `json:"instanceName"`
	ActionType   ActionType `json:"actionType"`
	Status       Status     `json:"status"`
	Amount       int        `json:"amount"`
	Balance      int        `json:"balance"`
	CreatedAt    time.Time  `json:"createdAt"`
	Note         string     `json:"note"`
}
