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

package mongodb

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	dbtreev1 "github.com/piper-hyowon/dBtree/operator/api/v1"
	"github.com/piper-hyowon/dBtree/operator/internal/provisioner"
	"github.com/piper-hyowon/dBtree/operator/internal/provisioner/utils"
)

const (
	defaultMongoDBImage = "mongo:7.0"
	mongoDBPort         = 27017
)

// MongoDBProvisioner implements the Provisioner interface for MongoDB
type MongoDBProvisioner struct {
	client client.Client
	scheme *runtime.Scheme
}

// NewProvisioner creates a new MongoDB provisioner
func NewProvisioner(client client.Client, scheme *runtime.Scheme) provisioner.Provisioner {
	return &MongoDBProvisioner{
		client: client,
		scheme: scheme,
	}
}

// Provision creates all MongoDB resources
func (p *MongoDBProvisioner) Provision(ctx context.Context, instance *dbtreev1.DBInstance) error {
	namespace := instance.GetUserNamespace()

	// Ensure Secret exists (created by backend)
	if err := p.ensureSecret(ctx, instance, namespace); err != nil {
		return fmt.Errorf("failed to ensure secret: %w", err)
	}

	// Create ConfigMap
	if err := p.createConfigMap(ctx, instance, namespace); err != nil {
		return fmt.Errorf("failed to create configmap: %w", err)
	}

	// Create Service
	if err := p.createService(ctx, instance, namespace); err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}

	// Create StatefulSet
	if err := p.createStatefulSet(ctx, instance, namespace); err != nil {
		return fmt.Errorf("failed to create statefulset: %w", err)
	}

	return nil
}

// Delete removes all MongoDB resources
func (p *MongoDBProvisioner) Delete(ctx context.Context, instance *dbtreev1.DBInstance) error {
	// Resources will be deleted automatically due to owner references
	return nil
}

// Update modifies existing MongoDB resources
func (p *MongoDBProvisioner) Update(ctx context.Context, instance *dbtreev1.DBInstance) error {
	namespace := instance.GetUserNamespace()

	// 1. Update StatefulSet (for resource changes)
	sts := &appsv1.StatefulSet{}
	if err := p.client.Get(ctx, types.NamespacedName{
		Name:      instance.GetStatefulSetName(),
		Namespace: namespace,
	}, sts); err != nil {
		return fmt.Errorf("failed to get statefulset: %w", err)
	}

	// Check if resources need update
	currentResources := &sts.Spec.Template.Spec.Containers[0].Resources
	desiredResources := p.getResourceRequirements(instance)

	resourcesChanged := false
	if !currentResources.Requests.Cpu().Equal(*desiredResources.Requests.Cpu()) ||
		!currentResources.Requests.Memory().Equal(*desiredResources.Requests.Memory()) {
		resourcesChanged = true
	}

	// Check if replicas need update (for scaling)
	replicasChanged := false
	desiredReplicas := p.getReplicas(instance)
	if *sts.Spec.Replicas != desiredReplicas {
		replicasChanged = true
		sts.Spec.Replicas = &desiredReplicas
	}

	// Update resources if changed
	if resourcesChanged {
		sts.Spec.Template.Spec.Containers[0].Resources = desiredResources
	}

	// Apply StatefulSet changes
	if resourcesChanged || replicasChanged {
		if err := p.client.Update(ctx, sts); err != nil {
			return fmt.Errorf("failed to update statefulset: %w", err)
		}
	}

	// 2. Update ConfigMap (for configuration changes)
	cm := &corev1.ConfigMap{}
	if err := p.client.Get(ctx, types.NamespacedName{
		Name:      instance.GetConfigMapName(),
		Namespace: namespace,
	}, cm); err != nil {
		return fmt.Errorf("failed to get configmap: %w", err)
	}

	// Generate new config
	newConfig := p.generateMongoConfig(instance)
	if cm.Data["mongod.conf"] != newConfig {
		cm.Data["mongod.conf"] = newConfig
		if err := p.client.Update(ctx, cm); err != nil {
			return fmt.Errorf("failed to update configmap: %w", err)
		}

		// Restart pods to apply config changes
		// This is done by updating an annotation
		if sts.Spec.Template.Annotations == nil {
			sts.Spec.Template.Annotations = make(map[string]string)
		}
		sts.Spec.Template.Annotations["dbtree.cloud/config-hash"] = fmt.Sprintf("%d", time.Now().Unix())
		if err := p.client.Update(ctx, sts); err != nil {
			return fmt.Errorf("failed to trigger pod restart: %w", err)
		}
	}

	// 3. Update Service if port changed
	if instance.Status.Port != 0 {
		svc := &corev1.Service{}
		if err := p.client.Get(ctx, types.NamespacedName{
			Name:      instance.GetServiceName(),
			Namespace: namespace,
		}, svc); err != nil {
			return fmt.Errorf("failed to get service: %w", err)
		}

		if svc.Spec.Ports[0].Port != instance.GetDefaultPort() {
			svc.Spec.Ports[0].Port = instance.GetDefaultPort()
			svc.Spec.Ports[0].TargetPort = intstr.FromInt32(int32(int(instance.GetDefaultPort())))
			if err := p.client.Update(ctx, svc); err != nil {
				return fmt.Errorf("failed to update service: %w", err)
			}
		}
	}

	return nil
}

// GetStatus retrieves the current status of MongoDB instance
func (p *MongoDBProvisioner) GetStatus(ctx context.Context, instance *dbtreev1.DBInstance) (*dbtreev1.DBInstanceStatus, error) {
	namespace := instance.GetUserNamespace()

	// Check StatefulSet status
	sts := &appsv1.StatefulSet{}
	if err := p.client.Get(ctx, types.NamespacedName{
		Name:      instance.GetStatefulSetName(),
		Namespace: namespace,
	}, sts); err != nil {
		return nil, err
	}

	status := &dbtreev1.DBInstanceStatus{
		State: dbtreev1.StatusRunning,
	}

	if sts.Status.ReadyReplicas != sts.Status.Replicas {
		status.State = dbtreev1.StatusProvisioning
		status.StatusReason = fmt.Sprintf("Waiting for pods: %d/%d ready",
			sts.Status.ReadyReplicas, sts.Status.Replicas)
	}

	return status, nil
}

// createSecret creates the MongoDB credentials secret
func (p *MongoDBProvisioner) createSecret(ctx context.Context, instance *dbtreev1.DBInstance, namespace string) error {
	tempSecretName := instance.Annotations["dbtree.cloud/temp-secret"]
	if tempSecretName == "" {
		return fmt.Errorf("temp secret name not provided")
	}

	tempSecret := &corev1.Secret{}
	err := p.client.Get(ctx, types.NamespacedName{
		Name:      tempSecretName,
		Namespace: namespace,
	}, tempSecret)
	if err != nil {
		return fmt.Errorf("failed to get temp secret: %w", err)
	}

	password := string(tempSecret.Data["password"])

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.GetSecretName(),
			Namespace: namespace,
		},
		StringData: map[string]string{
			"username":                   "admin",
			"password":                   password,
			"MONGO_INITDB_ROOT_USERNAME": "admin",
			"MONGO_INITDB_ROOT_PASSWORD": password,
			"MONGO_INITDB_DATABASE":      "admin",
			"connection-string": fmt.Sprintf("mongodb://admin:%s@%s:%d/admin",
				password, instance.GetServiceName(), mongoDBPort),
		},
	}

	// 임시 Secret 삭제
	if err := p.client.Delete(ctx, tempSecret); err != nil {
		fmt.Println(err, "Failed to delete temp secret", "name", tempSecretName)
		// 삭제 실패해도 계속 진행
	}

	// Annotation 제거
	delete(instance.Annotations, "dbtree.cloud/temp-secret")
	if err := p.client.Update(ctx, instance); err != nil {
		fmt.Println(err, "Failed to remove annotation")
	}
	return p.client.Create(ctx, secret)
}

// createConfigMap creates the MongoDB configuration
func (p *MongoDBProvisioner) createConfigMap(ctx context.Context, instance *dbtreev1.DBInstance, namespace string) error {
	config := p.generateMongoConfig(instance)

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.GetConfigMapName(),
			Namespace: namespace,
		},
		Data: map[string]string{
			"mongod.conf": config,
		},
	}

	// Set owner reference
	//if err := controllerutil.SetControllerReference(instance, cm, p.scheme); err != nil {
	//	return err
	//}

	// Create or update
	_, err := controllerutil.CreateOrUpdate(ctx, p.client, cm, func() error {
		cm.Data["mongod.conf"] = config
		return nil
	})

	return err
}

// createService creates the MongoDB service
func (p *MongoDBProvisioner) createService(ctx context.Context, instance *dbtreev1.DBInstance, namespace string) error {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.GetServiceName(),
			Namespace: namespace,
			Labels:    p.getLabels(instance),
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeNodePort,
			Selector: p.getLabels(instance),
			Ports: []corev1.ServicePort{
				{
					Name:       "mongodb",
					Port:       mongoDBPort,
					TargetPort: intstr.FromInt32(mongoDBPort),
					NodePort:   instance.Spec.ExternalPort,
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}

	// For replica set, create headless service
	if instance.Spec.Mode == dbtreev1.DBModeReplicaSet {
		svc.Spec.ClusterIP = corev1.ClusterIPNone
	}

	// Set owner reference
	//if err := controllerutil.SetControllerReference(instance, svc, p.scheme); err != nil {
	//	return err
	//}

	// Create or update
	_, err := controllerutil.CreateOrUpdate(ctx, p.client, svc, func() error {
		svc.Spec.Selector = p.getLabels(instance)
		return nil
	})

	return err
}

// createStatefulSet creates the MongoDB StatefulSet
func (p *MongoDBProvisioner) createStatefulSet(ctx context.Context, instance *dbtreev1.DBInstance, namespace string) error {
	replicas := p.getReplicas(instance)

	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.GetStatefulSetName(),
			Namespace: namespace,
		},
	}

	// Create or update
	_, err := controllerutil.CreateOrUpdate(ctx, p.client, sts, func() error {
		// Labels 설정
		sts.Labels = p.getLabels(instance)

		// Owner reference 설정
		if err := controllerutil.SetControllerReference(instance, sts, p.scheme); err != nil {
			return err
		}

		// StatefulSet의 불변 필드는 생성 시에만 설정
		if sts.CreationTimestamp.IsZero() {
			sts.Spec.ServiceName = instance.GetServiceName()
			sts.Spec.Selector = &metav1.LabelSelector{
				MatchLabels: p.getLabels(instance),
			}
			sts.Spec.VolumeClaimTemplates = []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "data",
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{
							corev1.ReadWriteOnce,
						},
						Resources: corev1.VolumeResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: resource.MustParse(fmt.Sprintf("%dGi", instance.Spec.Resources.Disk)),
							},
						},
					},
				},
			}
		}

		// 변경 가능한 필드들
		sts.Spec.Replicas = &replicas
		sts.Spec.Template = corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: p.getLabels(instance),
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "mongodb",
						Image: p.getImage(instance),
						Ports: []corev1.ContainerPort{
							{
								Name:          "mongodb",
								ContainerPort: mongoDBPort,
								Protocol:      corev1.ProtocolTCP,
							},
						},
						Resources: p.getResourceRequirements(instance),
						Env: []corev1.EnvVar{
							{
								Name: "MONGO_INITDB_ROOT_USERNAME",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: instance.Spec.SecretRef.Name,
										},
										Key: "username",
									},
								},
							},
							{
								Name: "MONGO_INITDB_ROOT_PASSWORD",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: instance.Spec.SecretRef.Name,
										},
										Key: "password",
									},
								},
							},
							{
								Name:  "MONGO_INITDB_DATABASE",
								Value: "admin",
							},
						},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "data",
								MountPath: "/data/db",
							},
							{
								Name:      "config",
								MountPath: "/etc/mongod",
							},
						},
						LivenessProbe: &corev1.Probe{
							ProbeHandler: corev1.ProbeHandler{
								TCPSocket: &corev1.TCPSocketAction{
									Port: intstr.FromInt32(mongoDBPort),
								},
							},
							InitialDelaySeconds: 40,
							PeriodSeconds:       10,
						},
						ReadinessProbe: &corev1.Probe{
							ProbeHandler: corev1.ProbeHandler{
								TCPSocket: &corev1.TCPSocketAction{
									Port: intstr.FromInt32(mongoDBPort),
								},
							},
							InitialDelaySeconds: 40,
							PeriodSeconds:       10,
						},
					},
				},
				Volumes: []corev1.Volume{
					{
						Name: "config",
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: instance.GetConfigMapName(),
								},
								Items: []corev1.KeyToPath{
									{
										Key:  "mongod.conf",
										Path: "mongod.conf",
									},
								},
							},
						},
					},
				},
			},
		}

		return nil
	})

	return err
}

// Helper functions

func (p *MongoDBProvisioner) getLabels(instance *dbtreev1.DBInstance) map[string]string {
	return map[string]string{
		"app":                         instance.Name,
		"dbtree.cloud/instance-id":    instance.Spec.ExternalID,
		"app.kubernetes.io/name":      "mongodb",
		"app.kubernetes.io/instance":  instance.Name,
		"app.kubernetes.io/component": "database",
		"app.kubernetes.io/part-of":   "dbtree",
		"dbtree.cloud/db-type":        string(instance.Spec.Type),
		"dbtree.cloud/db-size":        string(instance.Spec.Size),
		"dbtree.cloud/instance-uid":   string(instance.UID),
	}
}

func (p *MongoDBProvisioner) getReplicas(instance *dbtreev1.DBInstance) int32 {
	// Parse config for custom replica count
	config, _ := utils.ParseMongoDBConfig(instance.Spec.Config)

	switch instance.Spec.Mode {
	case dbtreev1.DBModeStandalone:
		return 1
	case dbtreev1.DBModeReplicaSet:
		if config != nil && config.ReplicaCount > 0 {
			return config.ReplicaCount
		}
		return 3 // Default to 3 for replica set
	case dbtreev1.DBModeSharded:
		if config != nil && config.ShardCount > 0 {
			// This would need more complex logic for sharded clusters
			return config.ShardCount
		}
		return 1
	default:
		return 1
	}
}

func (p *MongoDBProvisioner) getImage(instance *dbtreev1.DBInstance) string {
	// Parse config to get version
	config, _ := utils.ParseMongoDBConfig(instance.Spec.Config)
	if config != nil && config.Version != "" {
		return fmt.Sprintf("mongo:%s", config.Version)
	}
	return defaultMongoDBImage
}

func (p *MongoDBProvisioner) getResourceRequirements(instance *dbtreev1.DBInstance) corev1.ResourceRequirements {
	// CPU string을 Quantity로 파싱
	cpuQuantity := instance.Spec.Resources.GetCPUQuantity()

	// CPU limit은 request의 2배
	cpuLimitMillicores := cpuQuantity.MilliValue() * 2
	cpuLimit := resource.NewMilliQuantity(cpuLimitMillicores, resource.DecimalSI)

	memoryMi := instance.Spec.Resources.Memory
	if memoryMi < 512 {
		memoryMi = 512
	}

	return corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    cpuQuantity,
			corev1.ResourceMemory: resource.MustParse(fmt.Sprintf("%dMi", instance.Spec.Resources.Memory)),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    *cpuLimit,
			corev1.ResourceMemory: resource.MustParse(fmt.Sprintf("%dMi", instance.Spec.Resources.Memory)),
		},
	}
}

func (p *MongoDBProvisioner) getCommand(instance *dbtreev1.DBInstance) []string {
	switch instance.Spec.Mode {
	case dbtreev1.DBModeReplicaSet:
		return []string{
			"mongod",
			"--config", "/etc/mongod/mongod.conf",
			"--replSet", "rs0",
		}
	default:
		return []string{
			"mongod",
			"--config", "/etc/mongod/mongod.conf",
		}
	}
}

func (p *MongoDBProvisioner) generateMongoConfig(instance *dbtreev1.DBInstance) string {
	// Parse custom config
	config, _ := utils.ParseMongoDBConfig(instance.Spec.Config)

	// Basic configuration
	mongoConf := `# MongoDB configuration
systemLog:
  destination: file
  path: /var/log/mongodb/mongod.log
  logAppend: true
net:
  port: 27017
  bindIp: 0.0.0.0
`

	// WiredTiger cache size 설정
	var cacheSize float64

	// 1. Custom config에 캐시 설정이 있으면 우선 사용
	if config != nil && config.WiredTigerCache > 0 {
		cacheSize = float64(config.WiredTigerCache)
	} else {
		// 메모리의 40% 사용 (MongoDB 기본값은 50%)
		memoryGB := float64(instance.Spec.Resources.Memory) / 1024.0
		cacheSize = memoryGB * 0.4

		// 최소값 보장
		if cacheSize < 0.1 {
			cacheSize = 0.1
		}
	}

	// WiredTiger 설정 추가
	mongoConf += fmt.Sprintf(`storage:
  dbPath: /data/db
  journal:
	enabled: true
	commitIntervalMs: 100
  wiredTiger:
    engineConfig:
      cacheSizeGB: %.2f
    collectionConfig:
      blockCompressor: snappy
    indexConfig:
      prefixCompression: true
`, cacheSize)

	// Security 설정 (옵션)
	if config == nil || config.AuthEnabled {
		// 기본적으로 인증 활성화 (현재는 주석 처리)
		mongoConf += `
# Security (uncomment to enable)
# security:
#   authorization: enabled
`
	}

	// Replication 설정 (Replica Set 모드일 때)
	if instance.Spec.Mode == dbtreev1.DBModeReplicaSet {
		mongoConf += `
# Replication
replication:
  replSetName: rs0
  enableMajorityReadConcern: true
`
	}

	// Sharding 설정 (Sharded 모드일 때)
	if instance.Spec.Mode == dbtreev1.DBModeSharded {
		mongoConf += `
# Sharding
sharding:
  clusterRole: shardsvr
  archiveMovedChunks: false
`
	}

	// Operation Profiling (성능 모니터링)
	mongoConf += `
# Operation Profiling
operationProfiling:
  mode: off
  slowOpThresholdMs: 100
`

	// 사이즈별 추가 최적화
	switch instance.Spec.Size {
	case dbtreev1.DBSizeTiny:
		// Tiny: 매우 제한적인 리소스
		mongoConf += `
# Tiny size optimizations
setParameter:
  internalQueryExecMaxBlockingSortBytes: 33554432  # 32MB (기본값의 1/3)
  maxIndexBuildMemoryUsageMegabytes: 100           # 100MB
`
	case dbtreev1.DBSizeSmall:
		// Small: 약간의 최적화
		mongoConf += `
# Small size optimizations
setParameter:
  internalQueryExecMaxBlockingSortBytes: 67108864  # 64MB (기본값의 2/3)
  maxIndexBuildMemoryUsageMegabytes: 200           # 200MB
`
	case dbtreev1.DBSizeMedium, dbtreev1.DBSizeLarge:
		// Medium/Large: 기본값 사용
		mongoConf += `
# Standard settings
setParameter:
  internalQueryExecMaxBlockingSortBytes: 104857600  # 100MB (기본값)
  maxIndexBuildMemoryUsageMegabytes: 500            # 500MB
`
	}

	return mongoConf
}

func (p *MongoDBProvisioner) ensureSecret(ctx context.Context, instance *dbtreev1.DBInstance, namespace string) error {
	// Secret이 이미 backend에서 생성되었으므로, 존재 여부만 확인
	if instance.Spec.SecretRef == nil {
		return fmt.Errorf("secretRef is required but not provided")
	}

	secret := &corev1.Secret{}
	err := p.client.Get(ctx, types.NamespacedName{
		Name:      instance.Spec.SecretRef.Name,
		Namespace: namespace,
	}, secret)

	if err != nil {
		return fmt.Errorf("failed to get secret %s: %w", instance.Spec.SecretRef.Name, err)
	}

	// Validate secret has required fields
	requiredFields := []string{
		"username",
		"password",
		"MONGO_INITDB_ROOT_USERNAME",
		"MONGO_INITDB_ROOT_PASSWORD",
	}

	for _, field := range requiredFields {
		if _, ok := secret.Data[field]; !ok {
			return fmt.Errorf("secret %s is missing required field: %s", secret.Name, field)
		}
	}

	// Add connection string if not present
	if _, ok := secret.Data["connection-string"]; !ok {
		password := string(secret.Data["password"])
		username := string(secret.Data["username"])

		connectionString := fmt.Sprintf("mongodb://%s:%s@%s:%d/admin",
			username, password, instance.GetServiceName(), mongoDBPort)

		secret.Data["connection-string"] = []byte(connectionString)

		if err := p.client.Update(ctx, secret); err != nil {
			return fmt.Errorf("failed to update secret with connection string: %w", err)
		}
	}

	return nil
}
