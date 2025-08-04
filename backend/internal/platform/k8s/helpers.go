package k8s

import (
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
	CPU    int
	Memory int
	Disk   int
}

type BackupSpec struct {
	Enabled       bool
	Schedule      string
	RetentionDays int
}

func BuildDBInstanceSpec(params DBInstanceParams) map[string]interface{} {
	spec := map[string]interface{}{
		"name": params.Name,
		"type": params.Type,
		"size": params.Size,
		"mode": params.Mode,
		"secretRef": map[string]interface{}{
			"name": params.SecretRef,
		},
		"resources": map[string]interface{}{
			"cpu":    params.Resources.CPU,
			"memory": params.Resources.Memory,
			"disk":   params.Resources.Disk,
		},
		"userId": params.UserID,
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
