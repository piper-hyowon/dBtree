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

package redis

import (
	"context"
	"fmt"

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
	defaultRedisImage = "redis:7.2"
	redisPort         = 6379
)

// RedisProvisioner implements the Provisioner interface for Redis
type RedisProvisioner struct {
	client client.Client
	scheme *runtime.Scheme
}

// NewProvisioner creates a new Redis provisioner
func NewProvisioner(client client.Client, scheme *runtime.Scheme) provisioner.Provisioner {
	return &RedisProvisioner{
		client: client,
		scheme: scheme,
	}
}

// Provision creates all Redis resources
func (p *RedisProvisioner) Provision(ctx context.Context, instance *dbtreev1.DBInstance) error {
	namespace := instance.GetUserNamespace()

	// Create Secret
	if err := p.createSecret(ctx, instance, namespace); err != nil {
		return fmt.Errorf("failed to create secret: %w", err)
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

// Delete removes all Redis resources
func (p *RedisProvisioner) Delete(ctx context.Context, instance *dbtreev1.DBInstance) error {
	// Resources will be deleted automatically due to owner references
	return nil
}

// Update modifies existing Redis resources
func (p *RedisProvisioner) Update(ctx context.Context, instance *dbtreev1.DBInstance) error {
	// TODO: Implement update logic
	return nil
}

// GetStatus retrieves the current status of Redis instance
func (p *RedisProvisioner) GetStatus(ctx context.Context, instance *dbtreev1.DBInstance) (*dbtreev1.DBInstanceStatus, error) {
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

// createSecret creates the Redis password secret
func (p *RedisProvisioner) createSecret(ctx context.Context, instance *dbtreev1.DBInstance, namespace string) error {
	// Check if secret already exists
	existingSecret := &corev1.Secret{}
	err := p.client.Get(ctx, types.NamespacedName{
		Name:      instance.GetSecretName(),
		Namespace: namespace,
	}, existingSecret)

	if err == nil {
		// Secret already exists, don't regenerate password
		return nil
	}

	// Generate secure password
	password, err := utils.GeneratePassword()
	if err != nil {
		return fmt.Errorf("failed to generate password: %w", err)
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.GetSecretName(),
			Namespace: namespace,
		},
		Type: corev1.SecretTypeOpaque,
		StringData: map[string]string{
			"REDIS_PASSWORD": password,
			"password":       password,
			"connection-string": fmt.Sprintf("redis://:%s@%s:%d",
				password, instance.GetServiceName(), redisPort),
		},
	}

	// Set owner reference
	if err := controllerutil.SetControllerReference(instance, secret, p.scheme); err != nil {
		return err
	}

	// Create secret
	return p.client.Create(ctx, secret)
}

// createConfigMap creates the Redis configuration
func (p *RedisProvisioner) createConfigMap(ctx context.Context, instance *dbtreev1.DBInstance, namespace string) error {
	config := p.generateRedisConfig(instance)

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.GetConfigMapName(),
			Namespace: namespace,
		},
		Data: map[string]string{
			"redis.conf": config,
		},
	}

	// Set owner reference
	if err := controllerutil.SetControllerReference(instance, cm, p.scheme); err != nil {
		return err
	}

	// Create or update
	_, err := controllerutil.CreateOrUpdate(ctx, p.client, cm, func() error {
		cm.Data["redis.conf"] = config
		return nil
	})

	return err
}

// createService creates the Redis service
func (p *RedisProvisioner) createService(ctx context.Context, instance *dbtreev1.DBInstance, namespace string) error {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.GetServiceName(),
			Namespace: namespace,
			Labels:    p.getLabels(instance),
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: p.getLabels(instance),
			Ports: []corev1.ServicePort{
				{
					Name:       "redis",
					Port:       redisPort,
					TargetPort: intstr.FromInt(redisPort),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}

	// For sentinel/cluster mode, might need headless service
	if instance.Spec.Mode == dbtreev1.DBModeSentinel || instance.Spec.Mode == dbtreev1.DBModeCluster {
		svc.Spec.ClusterIP = corev1.ClusterIPNone
	}

	// Set owner reference
	if err := controllerutil.SetControllerReference(instance, svc, p.scheme); err != nil {
		return err
	}

	// Create or update
	_, err := controllerutil.CreateOrUpdate(ctx, p.client, svc, func() error {
		svc.Spec.Selector = p.getLabels(instance)
		return nil
	})

	return err
}

// createStatefulSet creates the Redis StatefulSet
func (p *RedisProvisioner) createStatefulSet(ctx context.Context, instance *dbtreev1.DBInstance, namespace string) error {
	replicas := p.getReplicas(instance)

	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.GetStatefulSetName(),
			Namespace: namespace,
			Labels:    p.getLabels(instance),
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName: instance.GetServiceName(),
			Replicas:    &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: p.getLabels(instance),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: p.getLabels(instance),
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "redis",
							Image: p.getImage(instance),
							Ports: []corev1.ContainerPort{
								{
									Name:          "redis",
									ContainerPort: redisPort,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							Resources: p.getResourceRequirements(instance),
							Env: []corev1.EnvVar{
								{
									Name: "REDIS_PASSWORD",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: instance.GetSecretName(),
											},
											Key: "REDIS_PASSWORD",
										},
									},
								},
							},
							VolumeMounts: p.getVolumeMounts(instance),
							Command:      p.getCommand(instance),
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									Exec: &corev1.ExecAction{
										Command: []string{
											"redis-cli",
											"ping",
										},
									},
								},
								InitialDelaySeconds: 30,
								PeriodSeconds:       10,
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									Exec: &corev1.ExecAction{
										Command: []string{
											"redis-cli",
											"ping",
										},
									},
								},
								InitialDelaySeconds: 5,
								PeriodSeconds:       5,
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
								},
							},
						},
					},
				},
			},
			VolumeClaimTemplates: p.getVolumeClaimTemplates(instance),
		},
	}

	// Set owner reference
	if err := controllerutil.SetControllerReference(instance, sts, p.scheme); err != nil {
		return err
	}

	// Create or update
	_, err := controllerutil.CreateOrUpdate(ctx, p.client, sts, func() error {
		// Update mutable fields
		sts.Spec.Template.Spec.Containers[0].Resources = p.getResourceRequirements(instance)
		return nil
	})

	return err
}

// Helper functions

func (p *RedisProvisioner) getLabels(instance *dbtreev1.DBInstance) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":      "redis",
		"app.kubernetes.io/instance":  instance.Name,
		"app.kubernetes.io/component": "database",
		"app.kubernetes.io/part-of":   "dbtree",
		"dbtree.cloud/db-type":        string(instance.Spec.Type),
		"dbtree.cloud/db-size":        string(instance.Spec.Size),
	}
}

func (p *RedisProvisioner) getReplicas(instance *dbtreev1.DBInstance) int32 {
	switch instance.Spec.Mode {
	case dbtreev1.DBModeBasic:
		return 1
	case dbtreev1.DBModeSentinel:
		// 3 Redis + 3 Sentinel (simplified for now)
		return 3
	case dbtreev1.DBModeCluster:
		// Minimum 6 for cluster (3 masters + 3 replicas)
		return 6
	default:
		return 1
	}
}

func (p *RedisProvisioner) getImage(instance *dbtreev1.DBInstance) string {
	// Parse config to get version
	config, _ := utils.ParseRedisConfig(instance.Spec.Config)
	if config != nil && config.Version != "" {
		return fmt.Sprintf("redis:%s", config.Version)
	}
	return defaultRedisImage
}

func (p *RedisProvisioner) getResourceRequirements(instance *dbtreev1.DBInstance) corev1.ResourceRequirements {
	// Calculate memory with some overhead for Redis
	memoryMi := instance.Spec.Resources.Memory

	return corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(fmt.Sprintf("%dm", instance.Spec.Resources.CPU*1000)),
			corev1.ResourceMemory: resource.MustParse(fmt.Sprintf("%dMi", memoryMi)),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(fmt.Sprintf("%dm", instance.Spec.Resources.CPU*1000)),
			corev1.ResourceMemory: resource.MustParse(fmt.Sprintf("%dMi", memoryMi)),
		},
	}
}

func (p *RedisProvisioner) getVolumeMounts(instance *dbtreev1.DBInstance) []corev1.VolumeMount {
	mounts := []corev1.VolumeMount{
		{
			Name:      "config",
			MountPath: "/etc/redis",
		},
	}

	// Add data volume if persistence is enabled
	if p.isPersistenceEnabled(instance) {
		mounts = append(mounts, corev1.VolumeMount{
			Name:      "data",
			MountPath: "/data",
		})
	}

	return mounts
}

func (p *RedisProvisioner) getVolumeClaimTemplates(instance *dbtreev1.DBInstance) []corev1.PersistentVolumeClaim {
	if !p.isPersistenceEnabled(instance) {
		return nil
	}

	return []corev1.PersistentVolumeClaim{
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

func (p *RedisProvisioner) getCommand(instance *dbtreev1.DBInstance) []string {
	cmd := []string{
		"redis-server",
		"/etc/redis/redis.conf",
	}

	// Add password authentication
	cmd = append(cmd, "--requirepass", "$(REDIS_PASSWORD)")

	// Add persistence options if enabled
	if p.isPersistenceEnabled(instance) {
		cmd = append(cmd, "--dir", "/data")
	}

	return cmd
}

func (p *RedisProvisioner) generateRedisConfig(instance *dbtreev1.DBInstance) string {
	// Parse custom config
	redisConfig, _ := utils.ParseRedisConfig(instance.Spec.Config)

	config := `# Redis configuration
port 6379
bind 0.0.0.0
protected-mode yes
tcp-backlog 511
timeout 0
tcp-keepalive 300
`

	// Add persistence configuration
	if redisConfig != nil && redisConfig.Persistence {
		if redisConfig.PersistenceType == "AOF" {
			config += `
# AOF Persistence
appendonly yes
appendfilename "appendonly.aof"
appendfsync everysec
no-appendfsync-on-rewrite no
auto-aof-rewrite-percentage 100
auto-aof-rewrite-min-size 64mb
`
		} else {
			// RDB persistence (default)
			config += `
# RDB Persistence
save 900 1
save 300 10
save 60 10000
stop-writes-on-bgsave-error yes
rdbcompression yes
rdbchecksum yes
dbfilename dump.rdb
`
		}
	}

	// Add max memory policy
	config += fmt.Sprintf(`
# Memory management
maxmemory %dmb
`, instance.Spec.Resources.Memory)

	if redisConfig != nil && redisConfig.MaxMemoryPolicy != "" {
		config += fmt.Sprintf("maxmemory-policy %s\n", redisConfig.MaxMemoryPolicy)
	} else {
		config += "maxmemory-policy allkeys-lru\n"
	}

	// Add cluster/sentinel specific configs
	switch instance.Spec.Mode {
	case dbtreev1.DBModeCluster:
		config += `
# Cluster
cluster-enabled yes
cluster-config-file nodes.conf
cluster-node-timeout 5000
`
	}

	return config
}

func (p *RedisProvisioner) isPersistenceEnabled(instance *dbtreev1.DBInstance) bool {
	config, _ := utils.ParseRedisConfig(instance.Spec.Config)
	if config != nil {
		return config.Persistence
	}
	return true // Default to enabled
}
