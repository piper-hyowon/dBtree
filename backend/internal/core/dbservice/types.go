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

type ResourceSpec struct {
	CPU    int `json:"cpu"`
	Memory int `json:"memory"` // MB
	Disk   int `json:"disk"`   // GB
}

var DefaultResourceSpecs = map[DBType]map[DBSize]ResourceSpec{
	MongoDB: {
		SizeSmall:  {CPU: 1, Memory: 1024, Disk: 10},
		SizeMedium: {CPU: 2, Memory: 2048, Disk: 20},
		SizeLarge:  {CPU: 4, Memory: 4096, Disk: 40},
	},
	Redis: {
		SizeSmall:  {CPU: 1, Memory: 512, Disk: 5},
		SizeMedium: {CPU: 2, Memory: 1024, Disk: 10},
		SizeLarge:  {CPU: 2, Memory: 2048, Disk: 20},
	},
}

type LemonCost struct {
	CreationCost  int // 초기 생성 비용
	HourlyLemons  int // 시간당 비용
	MinimumLemons int // 최소 필요 레몬 수
}

var DefaultLemonCosts = map[DBType]map[DBSize]LemonCost{
	MongoDB: {
		SizeSmall:  {CreationCost: 100, HourlyLemons: 5, MinimumLemons: 50},
		SizeMedium: {CreationCost: 200, HourlyLemons: 10, MinimumLemons: 100},
		SizeLarge:  {CreationCost: 300, HourlyLemons: 20, MinimumLemons: 200},
	},
	Redis: {
		SizeSmall:  {CreationCost: 50, HourlyLemons: 5, MinimumLemons: 30},
		SizeMedium: {CreationCost: 150, HourlyLemons: 8, MinimumLemons: 80},
		SizeLarge:  {CreationCost: 250, HourlyLemons: 15, MinimumLemons: 150},
	},
}

type MongoDBConfig struct {
	Version         string `json:"version"`         // MongoDB 버전
	ReplicaCount    int    `json:"replicaCount"`    // 레플리카셋 구성 시 레플리카 수
	ShardCount      int    `json:"shardCount"`      // 샤딩 시 샤드 수
	AuthEnabled     bool   `json:"authEnabled"`     // 인증 활성화 여부
	WiredTigerCache int    `json:"wiredTigerCache"` // 캐시 크기 (MB)
}

type RedisConfig struct {
	Version         string `json:"version"`         // Redis 버전
	ReplicaCount    int    `json:"replicaCount"`    // 복제본 수
	Password        bool   `json:"password"`        // 패스워드 활성화 여부
	Persistence     bool   `json:"persistence"`     // 영속성 활성화 여부
	PersistenceType string `json:"persistenceType"` // AOF 또는 RDB
	MaxMemoryPolicy string `json:"maxMemoryPolicy"` // 메모리 정책
}

type NetworkConfig struct {
	Private bool `json:"private"` // 프라이빗 네트워크 사용 여부
	Port    int  `json:"port"`    // 포트 번호 (default 0)
}

type BackupConfig struct {
	Enabled       bool   `json:"enabled"`       // 백업 활성화
	Schedule      string `json:"schedule"`      // cron format
	RetentionDays int    `json:"retentionDays"` // 보관 기간
}

type DBInstanceSpec struct {
	Name        string            `json:"name"`
	Type        DBType            `json:"type"`
	Size        DBSize            `json:"size"`
	Mode        DBMode            `json:"mode"`
	Resources   ResourceSpec      `json:"resources"`
	Network     NetworkConfig     `json:"network"`
	Backup      BackupConfig      `json:"backup"`
	MongoDBConf *MongoDBConfig    `json:"mongodbConf,omitempty"`
	RedisConf   *RedisConfig      `json:"redisConf,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`
}

type DBInstance struct {
	ID           uuid.UUID      `json:"id"`
	UserID       uuid.UUID      `json:"userId"`
	Spec         DBInstanceSpec `json:"spec"`
	Status       InstanceStatus `json:"status"`
	StatusReason string         `json:"statusReason"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	LemonCost    LemonCost      `json:"lemonCost"`

	Endpoint  string `json:"endpoint"`
	Port      int    `json:"port"`
	SecretRef string `json:"secretRef"`

	PauseAfter  time.Time `json:"pauseAfter,omitempty"`
	DeleteAfter time.Time `json:"deleteAfter,omitempty"`
}
