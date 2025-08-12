package k8s

import (
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type DBInstanceParams struct {
	Name              string
	Type              string
	Size              string
	Mode              string
	SecretRef         string
	UserID            string
	ExternalID        string
	CreatedFromPreset *string
	Resources         ResourceSpec
	Backup            BackupSpec
	Config            map[string]interface{}
}

type ResourceSpec struct {
	CPU    float64
	Memory int
	Disk   int
}

type BackupSpec struct {
	Enabled       bool
	Schedule      string
	RetentionDays int
}

// ConvertCPUToString converts float64 CPU to string for CRD
func ConvertCPUToString(cpu float64) string {
	if cpu < 1 {
		// 1 미만은 밀리코어로 표현 (0.25 → "250m")
		return fmt.Sprintf("%dm", int(cpu*1000))
	}
	// 1 이상은 소수점 포함 문자열로 (1.5 → "1.5")
	return fmt.Sprintf("%.2f", cpu)
}

func BuildDBInstanceSpec(params DBInstanceParams) map[string]interface{} {
	spec := map[string]interface{}{
		"name":       params.Name,
		"type":       params.Type,
		"size":       params.Size,
		"mode":       params.Mode,
		"externalId": params.ExternalID,
		"secretRef": map[string]interface{}{
			"name": params.SecretRef,
		},
		"resources": map[string]interface{}{
			"cpu":    ConvertCPUToString(params.Resources.CPU),
			"memory": params.Resources.Memory,
			"disk":   params.Resources.Disk,
		},
		"userId": params.UserID,
	}

	if params.CreatedFromPreset != nil {
		spec["createdFromPreset"] = *params.CreatedFromPreset
	}

	backupSpec := map[string]interface{}{
		"enabled": params.Backup.Enabled,
	}

	if params.Backup.Schedule != "" {
		backupSpec["schedule"] = params.Backup.Schedule
	}

	if params.Backup.RetentionDays > 0 {
		backupSpec["retentionDays"] = params.Backup.RetentionDays
	}

	spec["backup"] = backupSpec

	if params.Config != nil {
		spec["config"] = params.Config
	}

	return spec
}

func BuildDBInstanceCRD(namespace, name string, spec map[string]interface{}, labels map[string]string) *unstructured.Unstructured {
	metadata := map[string]interface{}{
		"name":      name,
		"namespace": namespace,
	}

	if labels != nil {
		metadata["labels"] = labels
	}

	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "dbtree.cloud/v1",
			"kind":       "DBInstance",
			"metadata":   metadata,
			"spec":       spec,
		},
	}
}
