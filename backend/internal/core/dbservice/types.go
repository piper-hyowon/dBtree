package dbservice

import (
	"time"
)

type DBType string

const (
	MongoDB DBType = "mongodb"
	Redis   DBType = "redis"
)

type DBSize string

const (
	SizeSmall  DBSize = "small"
	SizeMedium DBSize = "medium"
	SizeLarge  DBSize = "large"
)

type DBMode string

const (
	// MongoDB
	ModeStandalone DBMode = "standalone"
	ModeReplicaSet DBMode = "replica_set"
	ModeSharded    DBMode = "sharded"

	// Redis
	ModeBasic    DBMode = "basic"
	ModeSentinel DBMode = "sentinel"
	ModeCluster  DBMode = "cluster"
)

type InstanceStatus string

const (
	StatusProvisioning InstanceStatus = "provisioning"
	StatusRunning      InstanceStatus = "running"
	StatusStopped      InstanceStatus = "stopped"
	StatusPaused       InstanceStatus = "paused"
	StatusError        InstanceStatus = "error"
	StatusDeleting     InstanceStatus = "deleting"
	StatusMaintenance  InstanceStatus = "maintenance"
	StatusBackingUp    InstanceStatus = "backing_up"
	StatusRestoring    InstanceStatus = "restoring"
	StatusUpgrading    InstanceStatus = "upgrading"
)

type ResourceSpec struct {
	CPU    int `json:"cpu"`
	Memory int `json:"memory"` // MB
	Disk   int `json:"disk"`   // GB
}

type LemonCost struct {
	CreationCost  int
	HourlyLemons  int
	MinimumLemons int
}

type BackupConfig struct {
	Enabled       bool
	Schedule      string // cron format
	RetentionDays int
}

type DBInstance struct {
	ID         int64
	ExternalID string
	UserID     string

	// 기본 정보
	Name              string
	Type              DBType
	Size              DBSize
	Mode              DBMode
	CreatedFromPreset *string

	// 스펙
	Resources ResourceSpec
	Cost      LemonCost

	Status       InstanceStatus
	StatusReason string

	// K8s
	K8sNamespace    string
	K8sResourceName string

	// Connection Info
	Endpoint string
	Port     int

	Config       map[string]interface{}
	BackupConfig BackupConfig

	CreatedAt    time.Time
	UpdatedAt    time.Time
	LastBilledAt *time.Time
	PausedAt     *time.Time
	DeletedAt    *time.Time
}

func (d *DBInstance) CanTransitionTo(target InstanceStatus) bool {
	transitions := map[InstanceStatus][]InstanceStatus{
		StatusProvisioning: {StatusRunning, StatusError},
		StatusRunning:      {StatusPaused, StatusStopped, StatusMaintenance, StatusBackingUp},
		StatusPaused:       {StatusRunning, StatusDeleting},
		StatusStopped:      {StatusRunning, StatusDeleting},
		StatusError:        {StatusDeleting},
		StatusMaintenance:  {StatusRunning},
		StatusBackingUp:    {StatusRunning},
		StatusRestoring:    {StatusRunning, StatusError},
	}

	allowed, ok := transitions[d.Status]
	if !ok {
		return false
	}

	for _, s := range allowed {
		if s == target {
			return true
		}
	}
	return false
}

func (d *DBInstance) CanStart() bool {
	return d.CanTransitionTo(StatusRunning)
}

func (d *DBInstance) CanStop() bool {
	return d.CanTransitionTo(StatusStopped)
}

func (d *DBInstance) CanDelete() bool {
	return d.CanTransitionTo(StatusDeleting)
}

func (d *DBInstance) CalculateHourlyCost() int {
	if d.Status != StatusRunning {
		return 0
	}
	return d.Cost.HourlyLemons
}

func (d *DBInstance) ShouldPause(userBalance int) bool {
	return d.Status == StatusRunning && userBalance < d.Cost.MinimumLemons
}

// 프리셋
type DBPreset struct {
	ID            string
	Type          DBType
	Size          DBSize
	Mode          DBMode
	Name          string
	Icon          string
	Description   string
	UseCases      []string
	Resources     ResourceSpec
	Cost          LemonCost
	DefaultConfig map[string]interface{}
	SortOrder     int
	IsActive      bool
}
