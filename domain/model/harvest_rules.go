package model

import "time"

type HarvestRules struct {
	BaseAmount      float64
	CooldownPeriod  time.Duration
	MaxStoredLemons float64
}

var DefaultHarvestRules = HarvestRules{
	BaseAmount:      5.0,
	CooldownPeriod:  6 * time.Hour,
	MaxStoredLemons: 500.0,
}
