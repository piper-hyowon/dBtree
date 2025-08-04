package dbservice

import (
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
)

type ConfigValidator interface {
	ValidateConfig(dbType DBType, mode DBMode, config map[string]interface{}, resources *ResourceSpec) error
	GetDefaultConfig(dbType DBType, mode DBMode) map[string]interface{}
}

type MongoDBConfig struct {
	Version         string `json:"version" validate:"required,oneof=6.0 7.0"`
	WiredTigerCache *int32 `json:"wiredTigerCache,omitempty" validate:"omitempty,min=1,max=128"`
	ReplicaCount    *int32 `json:"replicaCount,omitempty" validate:"omitempty,oneof=3 5 7"`
	ShardCount      *int32 `json:"shardCount,omitempty" validate:"omitempty,min=2,max=10"`
}

type configValidator struct {
	validator *validator.Validate
}

func NewConfigValidator() ConfigValidator {
	return &configValidator{
		validator: validator.New(),
	}
}

func (cv *configValidator) ValidateConfig(dbType DBType, mode DBMode, rawConfig map[string]interface{}, resources *ResourceSpec) error {
	if dbType != MongoDB {
		// TODO:
		return fmt.Errorf("redis support not implemented yet")
	}

	jsonBytes, err := json.Marshal(rawConfig)
	if err != nil {
		return fmt.Errorf("invalid config format: %w", err)
	}

	var config MongoDBConfig
	if err := json.Unmarshal(jsonBytes, &config); err != nil {
		return fmt.Errorf("invalid config structure: %w", err)
	}

	// Mode별 필드 검증
	switch mode {
	case ModeStandalone:
		if config.ReplicaCount != nil {
			return fmt.Errorf("replicaCount cannot be set for standalone mode")
		}
		if config.ShardCount != nil {
			return fmt.Errorf("shardCount cannot be set for standalone mode")
		}

	case ModeReplicaSet:
		if config.ShardCount != nil {
			return fmt.Errorf("shardCount cannot be set for replica set mode")
		}

	case ModeSharded:
		if config.ReplicaCount != nil {
			return fmt.Errorf("replicaCount cannot be set for sharded mode")
		}
	}

	if config.WiredTigerCache != nil {
		maxCache := int32(resources.Memory) / 2 / 1024
		if *config.WiredTigerCache > maxCache {
			return fmt.Errorf("wiredTigerCache cannot exceed 50%% of memory (%dGB max)", maxCache)
		}
	}

	return cv.validator.Struct(config)
}

func (cv *configValidator) GetDefaultConfig(dbType DBType, _ DBMode) map[string]interface{} {
	if dbType != MongoDB {
		return map[string]interface{}{}
	}

	config := map[string]interface{}{
		"version": "7.0",
	}

	// Mode별 기본값은 Operator가 처리
	return config
}
