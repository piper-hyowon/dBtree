package dbservice

import (
	"github.com/google/uuid"
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

func (t DBType) DefaultMode() DBMode {
	switch t {
	case MongoDB:
		return ModeStandalone
	case Redis:
		return ModeBasic
	default:
		return ""
	}
}

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

func (r ResourceSpec) CalculateSize() DBSize {
	if r.Memory <= 512 && r.CPU <= 1 {
		return SizeSmall
	} else if r.Memory <= 2048 && r.CPU <= 2 {
		return SizeMedium
	}
	return SizeLarge
}

type LemonCost struct {
	CreationCost  int
	HourlyLemons  int
	MinimumLemons int
}

type BackupConfig struct {
	Enabled       bool   `json:"enabled"`
	Schedule      string `json:"schedule,omitempty"` // cron format
	RetentionDays int    `json:"retentionDays,omitempty"`
	StorageSize   string `json:"storageSize,omitempty"` // 10Gi
}

type DBInstance struct {
	ID         int64
	ExternalID string // UUID as string (DB에는 UUID)
	UserID     string // UUID as string

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
	K8sSecretRef    string // Secret 리소스 참조

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
		StatusRunning:      {StatusPaused, StatusStopped, StatusMaintenance, StatusBackingUp, StatusDeleting},
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
	return d.Status != StatusDeleting
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
	ID                  string
	Type                DBType
	Size                DBSize
	Mode                DBMode
	Name                string
	Icon                string
	Description         string
	FriendlyDescription string
	TechnicalTerms      map[string]interface{}
	UseCases            []string
	Resources           ResourceSpec
	Cost                LemonCost
	DefaultConfig       map[string]interface{}
	SortOrder           int
	IsActive            bool
}

// BackupRecord (백업 요청 메타데이터만, 실제 백업은 K8s)
type BackupRecord struct {
	ID           int64
	InstanceID   int64
	ExternalID   uuid.UUID
	Name         string
	Type         BackupType
	Status       BackupStatus
	K8sJobName   string // K8s Job/CronJob 참조
	SizeBytes    int64
	StoragePath  string // S3/PVC 경로
	CreatedAt    time.Time
	CompletedAt  *time.Time
	ExpiresAt    *time.Time // 자동 삭제 예정일
	ErrorMessage string     // 실패시 에러 메시지
}

type BackupType string

const (
	BackupTypeManual    BackupType = "manual"
	BackupTypeScheduled BackupType = "scheduled"
)

type BackupStatus string

const (
	BackupStatusPending   BackupStatus = "pending"
	BackupStatusRunning   BackupStatus = "running"
	BackupStatusCompleted BackupStatus = "completed"
	BackupStatusFailed    BackupStatus = "failed"
)

type InstanceMetrics struct {
	InstanceID          uuid.UUID `json:"instanceId"`
	CPUUsage            string    `json:"cpuUsage"`
	MemoryUsage         string    `json:"memoryUsage"`
	DiskUsage           string    `json:"diskUsage"`
	Connections         int       `json:"connections"`
	OperationsPerSecond int       `json:"operationsPerSecond,omitempty"`
	Timestamp           time.Time `json:"timestamp"`
}

// CalculateCustomCost 프리셋이 아닐경우 직접 계산
func CalculateCustomCost(dbType DBType, resources ResourceSpec) LemonCost {
	var base int

	// 메모리 기반 비용
	switch dbType {
	case Redis:
		base = resources.Memory / 512 // 512MB당 1레몬
	case MongoDB:
		base = resources.Memory / 1024 * 3 // 1GB당 3레몬
	}

	// CPU 추가 비용 (1 vCPU 초과분에 대해)
	if resources.CPU > 1 {
		base += (resources.CPU - 1) * 2
	}

	// 디스크 추가 비용 (10GB 초과분에 대해)
	if resources.Disk > 10 {
		base += (resources.Disk - 10) / 10
	}

	// 최소값 보장
	if base < 1 {
		base = 1
	}

	return LemonCost{
		CreationCost:  base * 10,
		HourlyLemons:  base,
		MinimumLemons: base * 24,
	}
}
