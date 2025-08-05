package dbservice

import (
	"encoding/json"
	"github.com/piper-hyowon/dBtree/internal/core/errors"
	"github.com/piper-hyowon/dBtree/internal/platform/validation"
)

type ConfigValidator interface {
	ValidateConfig(dbType DBType, mode DBMode, config map[string]interface{}, resources *ResourceSpec) error
	GetDefaultConfig(dbType DBType, mode DBMode) map[string]interface{}
	MergeWithDefaults(dbType DBType, mode DBMode, userConfig map[string]interface{}) map[string]interface{}
}

type MongoDBConfig struct {
	Version         string `json:"version" validate:"required,oneof=6.0 7.0"`
	WiredTigerCache *int32 `json:"wiredTigerCache,omitempty" validate:"omitempty,min=1"`
	ReplicaCount    *int32 `json:"replicaCount,omitempty" validate:"omitempty,oneof=3 5 7"`
	ShardCount      *int32 `json:"shardCount,omitempty" validate:"omitempty,min=2,max=10"`
}

type configValidator struct{}

func NewConfigValidator() ConfigValidator {
	return &configValidator{}
}

func (cv *configValidator) ValidateConfig(dbType DBType, mode DBMode, rawConfig map[string]interface{}, resources *ResourceSpec) error {
	if rawConfig == nil || len(rawConfig) == 0 {
		return nil
	}

	switch dbType {
	case MongoDB:
		return cv.validateMongoDBConfig(mode, rawConfig, resources)
	case Redis:
		// TODO: Redis 추가
		return errors.NewInvalidParameterError("type", "Redis는 아직 지원하지 않습니다")
	default:
		return errors.NewInvalidParameterError("type", "지원하지 않는 데이터베이스 타입입니다")
	}
}

func (cv *configValidator) validateMongoDBConfig(mode DBMode, rawConfig map[string]interface{}, resources *ResourceSpec) error {
	jsonBytes, err := json.Marshal(rawConfig)
	if err != nil {
		return errors.NewInvalidParameterError("config", "올바른 JSON 형식이 아닙니다")
	}

	var config MongoDBConfig
	if err := json.Unmarshal(jsonBytes, &config); err != nil {
		return errors.NewInvalidParameterError("config", "MongoDB 설정 구조가 올바르지 않습니다")
	}

	// 구조체 validation
	if err := validation.ValidateStruct(&config); err != nil {
		return err
	}

	// Mode별 필드 검증
	switch mode {
	case ModeStandalone:
		if config.ReplicaCount != nil {
			return errors.NewInvalidParameterError("config.replicaCount",
				"standalone 모드에서는 replicaCount를 설정할 수 없습니다")
		}
		if config.ShardCount != nil {
			return errors.NewInvalidParameterError("config.shardCount",
				"standalone 모드에서는 shardCount를 설정할 수 없습니다")
		}

	case ModeReplicaSet:
		if config.ShardCount != nil {
			return errors.NewInvalidParameterError("config.shardCount",
				"replica set 모드에서는 shardCount를 설정할 수 없습니다")
		}
		// replicaCount는 설정 가능 (기본값 3)

	case ModeSharded:
		if config.ReplicaCount != nil {
			return errors.NewInvalidParameterError("config.replicaCount",
				"sharded 모드에서는 replicaCount를 설정할 수 없습니다")
		}
		// shardCount는 설정 가능 (기본값 2)
	}

	// WiredTigerCache 검증 (메모리의 50% 이하)
	if config.WiredTigerCache != nil {
		maxCache := int32(resources.Memory) / 2 / 1024 // MB to GB
		if maxCache < 1 {
			maxCache = 1
		}
		if *config.WiredTigerCache > maxCache {
			return errors.NewInvalidParameterError("config.wiredTigerCache",
				"wiredTigerCache는 할당된 메모리의 50% 이하여야 합니다")
		}
	}

	return nil
}

func (cv *configValidator) GetDefaultConfig(dbType DBType, mode DBMode) map[string]interface{} {
	switch dbType {
	case MongoDB:
		config := map[string]interface{}{
			"version": "7.0",
		}

		// Mode별 기본값은 MongoDB Operator가 처리
		return config

	case Redis:
		// Redis는 아직 지원하지 않음
		return map[string]interface{}{}

	default:
		return map[string]interface{}{}
	}
}

func (cv *configValidator) MergeWithDefaults(dbType DBType, mode DBMode, userConfig map[string]interface{}) map[string]interface{} {
	defaults := cv.GetDefaultConfig(dbType, mode)

	if userConfig == nil {
		return defaults
	}

	// 사용자 설정으로 기본값 덮어쓰기
	result := make(map[string]interface{})
	for k, v := range defaults {
		result[k] = v
	}
	for k, v := range userConfig {
		result[k] = v
	}

	return result
}
