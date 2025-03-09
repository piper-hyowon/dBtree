package model

import "time"

type User struct {
	ID             string
	Username       string
	Email          string
	LemonBalance   float64
	TotalHarvested float64
	LastHarvest    time.Time
	JoinedAt       time.Time
	Instances      []string // 인스턴스 IDs
}
