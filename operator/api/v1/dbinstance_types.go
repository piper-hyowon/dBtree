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

package v1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DBType represents the database type
// +kubebuilder:validation:Enum=mongodb;redis
type DBType string

const (
	DBTypeMongoDB DBType = "mongodb"
	DBTypeRedis   DBType = "redis"
)

// DBSize represents the instance size
// +kubebuilder:validation:Enum=small;medium;large
type DBSize string

const (
	DBSizeSmall  DBSize = "small"
	DBSizeMedium DBSize = "medium"
	DBSizeLarge  DBSize = "large"
)

// DBMode represents the deployment mode
// Backend modes: standalone, replica_set, sharded (MongoDB) / basic, sentinel, cluster (Redis)
// +kubebuilder:validation:Enum=standalone;replica_set;sharded;basic;sentinel;cluster
type DBMode string

const (
	// MongoDB modes
	DBModeStandalone DBMode = "standalone"
	DBModeReplicaSet DBMode = "replica_set"
	DBModeSharded    DBMode = "sharded"

	// Redis modes
	DBModeBasic    DBMode = "basic"
	DBModeSentinel DBMode = "sentinel"
	DBModeCluster  DBMode = "cluster"
)

// InstanceStatus matches backend's InstanceStatus enum
// +kubebuilder:validation:Enum=provisioning;running;stopped;paused;error;deleting;maintenance;backing_up;restoring;upgrading
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

// ResourceSpec matches backend's ResourceSpec
type ResourceSpec struct {
	// CPU in cores (matches backend)
	// +kubebuilder:validation:Minimum=1
	CPU int32 `json:"cpu"`

	// Memory in MB (matches backend)
	// +kubebuilder:validation:Minimum=128
	Memory int32 `json:"memory"`

	// Disk in GB (matches backend)
	// +kubebuilder:validation:Minimum=1
	Disk int32 `json:"disk"`
}

// BackupConfig matches backend's BackupConfig
type BackupConfig struct {
	// Enable automatic backups
	Enabled bool `json:"enabled"`

	// Backup schedule in cron format
	// Standard cron format: minute hour day month weekday
	// Examples: "0 2 * * *" (daily at 2am), "*/30 * * * *" (every 30 minutes)
	// +optional
	Schedule string `json:"schedule,omitempty"`

	// Retention days for backups
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=90
	// +optional
	RetentionDays int32 `json:"retentionDays,omitempty"`

	// Storage size for backup PVC (e.g., "10Gi")
	// +optional
	// +kubebuilder:default="10Gi"
	StorageSize string `json:"storageSize,omitempty"`
}

func (d *DBInstance) GetBackupPVCName() string {
	return d.Name + "-backup-pvc"
}

// GetBackupStorageSize returns the backup storage size with default
func (d *DBInstance) GetBackupStorageSize() string {
	if d.Spec.Backup.StorageSize != "" {
		return d.Spec.Backup.StorageSize
	}
	return "10Gi" // 기본값
}

// DBInstanceSpec defines the desired state of DBInstance
// Maps to backend's CreateInstanceRequest
type DBInstanceSpec struct {
	// Instance name (3-50 chars as per backend validation)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=3
	// +kubebuilder:validation:MaxLength=50
	Name string `json:"name"`

	// Database type
	// +kubebuilder:validation:Required
	Type DBType `json:"type"`

	// Instance size
	// +kubebuilder:validation:Required
	Size DBSize `json:"size"`

	// Deployment mode
	// +kubebuilder:validation:Required
	Mode DBMode `json:"mode"`

	// Reference to the credentials secret
	// +kubebuilder:validation:Required
	SecretRef *corev1.LocalObjectReference `json:"secretRef"`

	// Created from preset ID (optional, matches backend)
	// +optional
	CreatedFromPreset *string `json:"createdFromPreset,omitempty"`

	// Compute resources
	// +kubebuilder:validation:Required
	Resources ResourceSpec `json:"resources"`

	// Backup configuration
	// +kubebuilder:validation:Required
	Backup BackupConfig `json:"backup"`

	// UserID is the owner (matches backend)
	// +kubebuilder:validation:Required
	UserID string `json:"userId"`

	// Configuration map (matches backend Config field)
	// +optional
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Type=object
	Config *runtime.RawExtension `json:"config,omitempty"`

	// ExternalID from backend (백엔드의 DBInstance.ExternalID)
	// +kubebuilder:validation:Required
	ExternalID string `json:"externalId"`
}

// InstanceMetrics matches backend's metrics fields
type InstanceMetrics struct {
	// CPU usage percentage (string to match CRD)
	// +optional
	CPUUsage string `json:"cpuUsage,omitempty"`

	// Memory usage percentage
	// +optional
	MemoryUsage string `json:"memoryUsage,omitempty"`

	// Disk usage percentage
	// +optional
	DiskUsage string `json:"diskUsage,omitempty"`

	// Active connections count
	// +optional
	Connections int32 `json:"connections,omitempty"`

	// Operations per second
	// +optional
	OperationsPerSecond int32 `json:"operationsPerSecond,omitempty"`
}

// DBInstanceStatus defines the observed state of DBInstance
// Maps to backend's DBInstance runtime fields
type DBInstanceStatus struct {
	// Current state
	State InstanceStatus `json:"state,omitempty"`

	// Reason for current state
	// +optional
	StatusReason string `json:"statusReason,omitempty"`

	// K8s namespace (backend: K8sNamespace)
	// +optional
	K8sNamespace string `json:"k8sNamespace,omitempty"`

	// K8s resource name (backend: K8sResourceName)
	// +optional
	K8sResourceName string `json:"k8sResourceName,omitempty"`

	// Connection endpoint
	// +optional
	Endpoint string `json:"endpoint,omitempty"`

	// Service port
	// +optional
	Port int32 `json:"port,omitempty"`

	// Reference to credentials secret
	// +optional
	SecretRef string `json:"secretRef,omitempty"`

	// Runtime metrics
	// +optional
	Metrics *InstanceMetrics `json:"metrics,omitempty"`

	// Last metrics update time
	// +optional
	LastMetricsUpdate *metav1.Time `json:"lastMetricsUpdate,omitempty"`

	// Last billing time (backend: LastBilledAt)
	// +optional
	LastBilledAt *metav1.Time `json:"lastBilledAt,omitempty"`

	// Paused timestamp (backend: PausedAt)
	// +optional
	PausedAt *metav1.Time `json:"pausedAt,omitempty"`

	// Standard K8s conditions
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`

	// ObservedGeneration for reconciliation optimization
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,shortName=dbi
// +kubebuilder:printcolumn:name="DB Type",type=string,JSONPath=`.spec.type`
// +kubebuilder:printcolumn:name="Size",type=string,JSONPath=`.spec.size`
// +kubebuilder:printcolumn:name="Mode",type=string,JSONPath=`.spec.mode`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.state`
// +kubebuilder:printcolumn:name="Endpoint",type=string,JSONPath=`.status.endpoint`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// DBInstance is the Schema for the dbinstances API
type DBInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DBInstanceSpec   `json:"spec,omitempty"`
	Status DBInstanceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DBInstanceList contains a list of DBInstance
type DBInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DBInstance `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DBInstance{}, &DBInstanceList{})
}

// Helper methods for Controller logic

// GetUserNamespace returns the namespace for this instance
func (d *DBInstance) GetUserNamespace() string {
	return "user-" + d.Spec.UserID
}

// GetExternalID returns the external ID matching backend format
func (d *DBInstance) GetExternalID() string {
	// Backend uses UUID, but we use namespace/name as external ID
	return string(d.UID)
}

// Resource naming helpers
func (d *DBInstance) GetSecretName() string {
	return d.Name + "-secret"
}

func (d *DBInstance) GetServiceName() string {
	return d.Name + "-svc"
}

func (d *DBInstance) GetStatefulSetName() string {
	return d.Name + "-sts"
}

func (d *DBInstance) GetConfigMapName() string {
	return d.Name + "-config"
}

func (d *DBInstance) GetPVCName() string {
	// PVC name pattern for StatefulSet volumeClaimTemplates
	return "data-" + d.GetStatefulSetName() + "-0"
}

func (d *DBInstance) GetBackupCronJobName() string {
	return d.Name + "-backup"
}

func (d *DBInstance) GetNetworkPolicyName() string {
	return d.Name + "-netpol"
}

// State checks
func (d *DBInstance) IsReady() bool {
	return d.Status.State == StatusRunning
}

func (d *DBInstance) IsPaused() bool {
	return d.Status.State == StatusPaused
}

func (d *DBInstance) NeedsBackup() bool {
	return d.Spec.Backup.Enabled && d.Spec.Backup.Schedule != ""
}

// CanTransitionTo validates state transitions (matches backend logic)
func (d *DBInstance) CanTransitionTo(target InstanceStatus) bool {
	current := d.Status.State

	// From backend: transitions map[InstanceStatus][]InstanceStatus
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

	allowed, ok := transitions[current]
	if !ok {
		return false
	}

	for _, state := range allowed {
		if state == target {
			return true
		}
	}

	return false
}

// Mode validation
func (d *DBInstance) IsValidMode() bool {
	switch d.Spec.Type {
	case DBTypeMongoDB:
		return d.Spec.Mode == DBModeStandalone ||
			d.Spec.Mode == DBModeReplicaSet ||
			d.Spec.Mode == DBModeSharded
	case DBTypeRedis:
		return d.Spec.Mode == DBModeBasic ||
			d.Spec.Mode == DBModeSentinel ||
			d.Spec.Mode == DBModeCluster
	default:
		return false
	}
}

// Condition helpers
func (d *DBInstance) SetCondition(conditionType string, status metav1.ConditionStatus, reason, message string) {
	meta.SetStatusCondition(&d.Status.Conditions, metav1.Condition{
		Type:               conditionType,
		Status:             status,
		LastTransitionTime: metav1.Now(),
		Reason:             reason,
		Message:            message,
		ObservedGeneration: d.Generation,
	})
}

func (d *DBInstance) GetCondition(conditionType string) *metav1.Condition {
	return meta.FindStatusCondition(d.Status.Conditions, conditionType)
}

// GetDefaultPort returns default port based on DB type
func (d *DBInstance) GetDefaultPort() int32 {
	switch d.Spec.Type {
	case DBTypeMongoDB:
		return 27017
	case DBTypeRedis:
		return 6379
	default:
		return 0
	}
}

// Billing helpers
func (d *DBInstance) ShouldBeBilled() bool {
	// Bill if running and last billed more than 1 hour ago
	if d.Status.State != StatusRunning {
		return false
	}

	if d.Status.LastBilledAt == nil {
		return true
	}

	return metav1.Now().Time.Sub(d.Status.LastBilledAt.Time).Hours() >= 1
}

func (d *DBInstance) ShouldBeDeleted() bool {
	// Delete if paused for more than 1 hour
	if d.Status.State != StatusPaused || d.Status.PausedAt == nil {
		return false
	}

	return metav1.Now().Time.Sub(d.Status.PausedAt.Time).Hours() >= 1
}
