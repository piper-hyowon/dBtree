/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utils

import (
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
)

// MongoDBConfig represents MongoDB specific configuration
type MongoDBConfig struct {
	Version         string `json:"version,omitempty"`
	ReplicaCount    int32  `json:"replicaCount,omitempty"`
	ShardCount      int32  `json:"shardCount,omitempty"`
	AuthEnabled     bool   `json:"authEnabled,omitempty"`
	WiredTigerCache int32  `json:"wiredTigerCache,omitempty"`
}

// RedisConfig represents Redis specific configuration
type RedisConfig struct {
	Version         string `json:"version,omitempty"`
	ReplicaCount    int32  `json:"replicaCount,omitempty"`
	Password        bool   `json:"password,omitempty"`
	Persistence     bool   `json:"persistence,omitempty"`
	PersistenceType string `json:"persistenceType,omitempty"`
	MaxMemoryPolicy string `json:"maxMemoryPolicy,omitempty"`
}

// ParseMongoDBConfig parses MongoDB configuration from RawExtension
func ParseMongoDBConfig(raw *runtime.RawExtension) (*MongoDBConfig, error) {
	if raw == nil || len(raw.Raw) == 0 {
		// Return default config
		return &MongoDBConfig{
			Version:     "7.0",
			AuthEnabled: true,
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

	return &config, nil
}

// ParseRedisConfig parses Redis configuration from RawExtension
func ParseRedisConfig(raw *runtime.RawExtension) (*RedisConfig, error) {
	if raw == nil || len(raw.Raw) == 0 {
		// Return default config
		return &RedisConfig{
			Version:         "7.2",
			Password:        true,
			Persistence:     true,
			PersistenceType: "RDB",
			MaxMemoryPolicy: "allkeys-lru",
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
	if config.PersistenceType == "" && config.Persistence {
		config.PersistenceType = "RDB"
	}
	if config.MaxMemoryPolicy == "" {
		config.MaxMemoryPolicy = "allkeys-lru"
	}

	return &config, nil
}

// GetStringFromConfig safely extracts a string value from config map
func GetStringFromConfig(config map[string]interface{}, key string, defaultValue string) string {
	if val, ok := config[key]; ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return defaultValue
}

// GetIntFromConfig safely extracts an int value from config map
func GetIntFromConfig(config map[string]interface{}, key string, defaultValue int) int {
	if val, ok := config[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case int32:
			return int(v)
		case int64:
			return int(v)
		case float64:
			return int(v)
		}
	}
	return defaultValue
}

// GetBoolFromConfig safely extracts a bool value from config map
func GetBoolFromConfig(config map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := config[key]; ok {
		if boolVal, ok := val.(bool); ok {
			return boolVal
		}
	}
	return defaultValue
}
