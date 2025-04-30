package dbservice

import (
	"time"
)

type DatabaseType string

const (
	MongoDB DatabaseType = "mongodb"
	Redis   DatabaseType = "redis"
)

var DefaultResourceSpecs = map[DatabaseType]map[string]ResourceSpec{
	MongoDB: {
		"small":  {CPU: 1, Memory: 1024, Disk: 10},
		"medium": {CPU: 2, Memory: 2048, Disk: 20},
		"large":  {CPU: 4, Memory: 4096, Disk: 40},
	},
	Redis: {
		"small":  {CPU: 1, Memory: 512, Disk: 5},
		"medium": {CPU: 2, Memory: 1024, Disk: 10},
		"large":  {CPU: 2, Memory: 2048, Disk: 20},
	},
}

type ResourceSpec struct {
	CPU    int
	Memory int // MB
	Disk   int // GB
}

type LemonCost struct {
	CreationCost  float64
	HourlyLemons  float64
	MinimumLemons float64
}

var DefaultLemonCosts = map[DatabaseType]map[string]LemonCost{
	MongoDB: {
		"small":  {CreationCost: 10, HourlyLemons: 0.5, MinimumLemons: 5},
		"medium": {CreationCost: 20, HourlyLemons: 1.0, MinimumLemons: 10},
		"large":  {CreationCost: 30, HourlyLemons: 2.0, MinimumLemons: 20},
	},
	Redis: {
		"small":  {CreationCost: 5, HourlyLemons: 0.3, MinimumLemons: 3},
		"medium": {CreationCost: 15, HourlyLemons: 0.8, MinimumLemons: 8},
		"large":  {CreationCost: 25, HourlyLemons: 1.5, MinimumLemons: 15},
	},
}

type DatabaseInstance struct {
	ID              string
	Name            string
	Type            DatabaseType
	Mode            DatabaseMode
	Version         string
	Status          InstanceStatus
	StatusReason    string
	Config          map[string]interface{}
	OwnedBy         string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Resources       ResourceSpec
	LemonCost       LemonCost
	LemonsBalance   float64
	LastLemonUpdate time.Time
	LastHarvest     time.Time

	// Connection
	Endpoint  string
	Port      int
	SecretRef string

	PauseAfter  time.Time
	DeleteAfter time.Time
}

type InstanceStatus string

const (
	StatusProvisioning InstanceStatus = "provisioning"
	StatusRunning      InstanceStatus = "running"
	StatusStopped      InstanceStatus = "stopped"
	StatusPaused       InstanceStatus = "paused" // 레몬 부족시
	StatusError        InstanceStatus = "error"
	StatusDeleting     InstanceStatus = "deleting"
	StatusMaintenance  InstanceStatus = "maintenance"
	StatusBackingUp    InstanceStatus = "backing_up"
	StatusRestoring    InstanceStatus = "restoring"
	StatusUpgrading    InstanceStatus = "upgrading"
)

type DatabaseMode string

const (
	Basic   DatabaseMode = "basic"
	Replica DatabaseMode = "replica"
	Cluster DatabaseMode = "cluster"
)
