package user

import "time"

type ProfileResponse struct {
	Email             string     `json:"email"`
	LemonBalance      int        `json:"lemonBalance"`
	TotalEarnedLemons int64      `json:"totalEarnedLemons"`
	TotalSpentLemons  int64      `json:"totalSpentLemons"`
	LastHarvestAt     *time.Time `json:"lastHarvestAt"`
	JoinedAt          time.Time  `json:"joinedAt"`
}
