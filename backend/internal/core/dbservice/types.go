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
	StatusStopped      InstanceStatus = "stopped" // 유저가 수동 중지
	// StatusPaused 시스템에 의해 일시정지(레몬 부족), 다음 과금때도 없으면 자동 삭제 대상, 사용자가 레몬을 충전하면 자동으로 재시작 가능
	StatusPaused      InstanceStatus = "paused"
	StatusError       InstanceStatus = "error"
	StatusDeleting    InstanceStatus = "deleting"
	StatusMaintenance InstanceStatus = "maintenance"
	StatusBackingUp   InstanceStatus = "backing_up"
	StatusRestoring   InstanceStatus = "restoring"
	StatusUpgrading   InstanceStatus = "upgrading"
)

type ResourceSpec struct {
	CPU    int `json:"cpu" validate:"required,min=1,max=16"`
	Memory int `json:"memory" validate:"required,min=128,max=65536"` // MB
	Disk   int `json:"disk" validate:"required,min=1,max=1000"`      // GB
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
	CreationCost int
	HourlyLemons int
}

func (c *LemonCost) ToResponse() CostResponse {
	return CostResponse{
		CreationCost:  c.CreationCost,
		HourlyLemons:  c.HourlyLemons,
		DailyLemons:   c.HourlyLemons * 24,
		MonthlyLemons: c.HourlyLemons * 24 * 30,
	}
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
	Endpoint     string
	Port         int // 내부 포트 (27017, 6379 등)
	ExternalPort int // NodePort (30000-32767)

	Config       map[string]interface{}
	BackupConfig BackupConfig

	CreatedAt    time.Time
	UpdatedAt    time.Time
	LastBilledAt *time.Time
	PausedAt     *time.Time
	DeletedAt    *time.Time
}

func (d *DBInstance) ToResponse() *InstanceResponse {
	return &InstanceResponse{
		ID:                d.ExternalID,
		Name:              d.Name,
		Type:              d.Type,
		Size:              d.Size,
		Mode:              d.Mode,
		Status:            d.Status,
		StatusReason:      d.StatusReason,
		Resources:         d.Resources,
		Cost:              d.Cost.ToResponse(),
		Endpoint:          d.Endpoint,
		Port:              d.Port,
		ExternalPort:      d.ExternalPort,
		BackupEnabled:     d.BackupConfig.Enabled,
		Config:            d.Config,
		CreatedAt:         d.CreatedAt,
		UpdatedAt:         d.UpdatedAt,
		CreatedFromPreset: d.CreatedFromPreset,
		PausedAt:          d.PausedAt,
	}
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

func (p *DBPreset) ToResponse() PresetResponse {
	return PresetResponse{
		ID:                  p.ID,
		Type:                p.Type,
		Size:                p.Size,
		Mode:                p.Mode,
		Name:                p.Name,
		Icon:                p.Icon,
		Description:         p.Description,
		FriendlyDescription: p.FriendlyDescription,
		TechnicalTerms:      p.TechnicalTerms,
		UseCases:            p.UseCases,
		Resources:           p.Resources,
		Cost:                p.Cost.ToResponse(),
		DefaultConfig:       p.DefaultConfig,
		SortOrder:           p.SortOrder,
	}
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
		CreationCost: base * 10,
		HourlyLemons: base,
	}
}

func (d *DBInstance) ToCreateResponse(credentials *Credentials) *CreateInstanceResponse {
	return &CreateInstanceResponse{
		ID:          d.ExternalID,
		Name:        d.Name,
		Type:        d.Type,
		Status:      string(d.Status),
		Resources:   d.Resources,
		Cost:        d.Cost,
		Credentials: credentials,
		CreatedAt:   d.CreatedAt,
	}
}

type UserInstanceSummary struct {
	ID   string `json:"id"` // external_id
	Name string `json:"name"`
}
