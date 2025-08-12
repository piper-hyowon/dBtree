package user

import "time"

type User struct {
	ID                string
	Email             string
	IsDeleted         bool
	LemonBalance      int
	TotalEarnedLemons int64
	TotalSpentLemons  int64
	LastHarvestAt     *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
	WelcomeBonusGiven bool
}

func (u *User) ToProfileResponse() ProfileResponse {
	return ProfileResponse{
		ID:                u.ID,
		Email:             u.Email,
		LemonBalance:      u.LemonBalance,
		TotalEarnedLemons: u.TotalEarnedLemons,
		TotalSpentLemons:  u.TotalSpentLemons,
		LastHarvestAt:     u.LastHarvestAt,
		JoinedAt:          u.CreatedAt,
	}
}
