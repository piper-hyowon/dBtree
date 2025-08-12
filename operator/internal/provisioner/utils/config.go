package utils

import (
	"encoding/json"
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
)

// MongoDBConfig represents MongoDB-specific configuration
type MongoDBConfig struct {
	Version         string  `json:"version,omitempty"`
	ReplicaCount    int32   `json:"replicaCount,omitempty"`
	ShardCount      int32   `json:"shardCount,omitempty"`
	WiredTigerCache float64 `json:"wiredTigerCacheSizeGB,omitempty"`
	AuthEnabled     bool    `json:"authEnabled,omitempty"`
	JournalEnabled  bool    `json:"journalEnabled,omitempty"`
}

// ParseMongoDBConfig parses the raw config into MongoDBConfig (Operator용)
func ParseMongoDBConfig(raw *runtime.RawExtension) (*MongoDBConfig, error) {
	if raw == nil || len(raw.Raw) == 0 {
		// Return default config
		return &MongoDBConfig{
			Version:        "7.0",
			AuthEnabled:    true,
			JournalEnabled: true,
			ReplicaCount:   1,
		}, nil
	}

	var config MongoDBConfig
	if err := json.Unmarshal(raw.Raw, &config); err != nil {
		return nil, fmt.Errorf("failed to parse MongoDB config: %w", err)
	}

	// Set defaults if not specified
	if config.Version == "" {
		config.Version = "7.0"
	}
	if config.ReplicaCount == 0 {
		config.ReplicaCount = 1
	}
	// JournalEnabled의 기본값은 true (zero value가 false이므로)
	// AuthEnabled의 기본값도 true

	return &config, nil
}

// RedisConfig represents Redis-specific configuration
type RedisConfig struct {
	Version         string `json:"version,omitempty"`
	MaxMemory       int    `json:"maxMemoryMB,omitempty"`
	MaxMemoryPolicy string `json:"maxMemoryPolicy,omitempty"`
	PersistenceMode string `json:"persistenceMode,omitempty"` // "rdb", "aof", "both", "none"
	SaveSeconds     int    `json:"saveSeconds,omitempty"`
	ReplicaCount    int32  `json:"replicaCount,omitempty"`
}

// ParseRedisConfig parses the raw config into RedisConfig
func ParseRedisConfig(raw *runtime.RawExtension) (*RedisConfig, error) {
	if raw == nil || len(raw.Raw) == 0 {
		// Return default config
		return &RedisConfig{
			Version:         "7.2",
			MaxMemoryPolicy: "allkeys-lru",
			PersistenceMode: "rdb",
			SaveSeconds:     900,
			ReplicaCount:    1,
		}, nil
	}

	var config RedisConfig
	if err := json.Unmarshal(raw.Raw, &config); err != nil {
		return nil, fmt.Errorf("failed to parse Redis config: %w", err)
	}

	// Set defaults if not specified
	if config.Version == "" {
		config.Version = "7.2"
	}
	if config.MaxMemoryPolicy == "" {
		config.MaxMemoryPolicy = "allkeys-lru"
	}
	if config.PersistenceMode == "" {
		config.PersistenceMode = "rdb"
	}
	if config.SaveSeconds == 0 {
		config.SaveSeconds = 900
	}
	if config.ReplicaCount == 0 {
		config.ReplicaCount = 1
	}

	return &config, nil
}
