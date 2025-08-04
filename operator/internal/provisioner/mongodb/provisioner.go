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
			Type:     corev1.ServiceTypeClusterIP,
			Selector: p.getLabels(instance),
			Ports: []corev1.ServicePort{
				{
					Name:       "mongodb",
					Port:       mongoDBPort,
					TargetPort: intstr.FromInt32(mongoDBPort),
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
									Exec: &corev1.ExecAction{
										Command: []string{
											"mongosh",
											"--eval",
											"db.adminCommand('ping')",
											"--quiet",
										},
									},
								},
								InitialDelaySeconds: 30,
								PeriodSeconds:       10,
								TimeoutSeconds:      5,
								FailureThreshold:    3,
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									Exec: &corev1.ExecAction{
										Command: []string{
											"mongosh",
											"--eval",
											"db.adminCommand('ping')",
											"--quiet",
										},
									},
								},
								InitialDelaySeconds: 10,
								PeriodSeconds:       5,
								TimeoutSeconds:      5,
								FailureThreshold:    3,
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
			},
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
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
			},
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

func (p *MongoDBProvisioner) getLabels(instance *dbtreev1.DBInstance) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":      "mongodb",
		"app.kubernetes.io/instance":  instance.Name,
		"app.kubernetes.io/component": "database",
		"app.kubernetes.io/part-of":   "dbtree",
		"dbtree.cloud/db-type":        string(instance.Spec.Type),
		"dbtree.cloud/db-size":        string(instance.Spec.Size),
		"dbtree.cloud/instance-uid":   string(instance.UID), // 추가!
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
	return corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(fmt.Sprintf("%dm", instance.Spec.Resources.CPU*1000)),
			corev1.ResourceMemory: resource.MustParse(fmt.Sprintf("%dMi", instance.Spec.Resources.Memory)),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(fmt.Sprintf("%dm", instance.Spec.Resources.CPU*1000)),
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
storage:
  dbPath: /data/db
net:
  port: 27017
  bindIp: 0.0.0.0
`

	// Add security if auth is enabled
	if config == nil || config.AuthEnabled {
		// TODO:
		mongoConf += `# security:
  # authorization: enabled
`
	}

	// Add WiredTiger cache size if specified
	if config != nil && config.WiredTigerCache > 0 {
		mongoConf += fmt.Sprintf(`  wiredTiger:
    engineConfig:
      cacheSizeGB: %d
`, config.WiredTigerCache)
	}

	// Add replication config if needed
	if instance.Spec.Mode == dbtreev1.DBModeReplicaSet {
		mongoConf += `replication:
  replSetName: rs0
setParameter:
  enableMajorityReadConcern: true
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
