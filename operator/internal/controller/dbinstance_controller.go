/*
Copyright 2025 piper-hyowon.

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

package controller

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	dbtreev1 "github.com/piper-hyowon/dBtree/operator/api/v1"
	"github.com/piper-hyowon/dBtree/operator/internal/provisioner"
	"github.com/piper-hyowon/dBtree/operator/internal/provisioner/mongodb"
	"github.com/piper-hyowon/dBtree/operator/internal/provisioner/redis"
)

const (
	// Finalizer for cleanup
	dbInstanceFinalizer = "dbinstance.dbtree.cloud/finalizer"

	// Condition types
	ConditionTypeProvisioned = "Provisioned"
	ConditionTypeReady       = "Ready"
	ConditionTypeError       = "Error"

	// Annotations
	AnnotationBackendID = "dbtree.cloud/backend-id"
)

var (
	protocolUDP = corev1.ProtocolUDP
	protocolTCP = corev1.ProtocolTCP
)

// DBInstanceReconciler reconciles a DBInstance object
type DBInstanceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=dbtree.cloud,resources=dbinstances,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=dbtree.cloud,resources=dbinstances/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=dbtree.cloud,resources=dbinstances/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch,resources=cronjobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.k8s.io,resources=networkpolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch;create
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch

// internal/controller/dbinstance_controller.go - 주요 변경 부분

func (r *DBInstanceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch the DBInstance
	instance := &dbtreev1.DBInstance{}
	if err := r.Get(ctx, req.NamespacedName, instance); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get DBInstance")
		return ctrl.Result{}, err
	}

	// Check if the instance is being deleted
	if !instance.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.handleDeletion(ctx, instance)
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(instance, dbInstanceFinalizer) {
		controllerutil.AddFinalizer(instance, dbInstanceFinalizer)
		if err := r.Update(ctx, instance); err != nil {
			return ctrl.Result{}, err
		}
	}

	// 네임스페이스는 백엔드에서 이미 생성함!

	// Get provisioner based on database type
	prov := r.getProvisioner(instance.Spec.Type)
	if prov == nil {
		return r.setErrorCondition(ctx, instance, "InvalidDatabaseType",
			fmt.Sprintf("Unsupported database type: %s", instance.Spec.Type))
	}

	// Handle based on current state
	switch instance.Status.State {
	case "", dbtreev1.StatusProvisioning:
		return r.handleProvisioning(ctx, instance, prov)
	case dbtreev1.StatusRunning:
		return r.handleRunning(ctx, instance, prov)
	case dbtreev1.StatusPaused:
		return r.handlePaused(ctx, instance, prov)
	case dbtreev1.StatusStopped:
		return r.handleStopped(ctx, instance, prov)
	case dbtreev1.StatusError:
		return r.handleError(ctx, instance, prov)
	case dbtreev1.StatusDeleting:
		return r.handleDeletion(ctx, instance)
	default:
		log.Info("Unknown state, setting to provisioning", "state", instance.Status.State)
		instance.Status.State = dbtreev1.StatusProvisioning
		if err := r.updateStatus(ctx, instance); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}
}

// handleProvisioning creates all required resources
func (r *DBInstanceReconciler) handleProvisioning(ctx context.Context, instance *dbtreev1.DBInstance, prov provisioner.Provisioner) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("Handling provisioning state")

	// Update state if not set
	if instance.Status.State == "" {
		instance.Status.State = dbtreev1.StatusProvisioning
		instance.Status.StatusReason = "Starting provisioning"
		if err := r.updateStatus(ctx, instance); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Create resources
	if err := prov.Provision(ctx, instance); err != nil {
		log.Error(err, "Failed to provision resources")
		return r.setErrorCondition(ctx, instance, "ProvisioningFailed", err.Error())
	}

	// Wait for pods to be ready
	sts := &appsv1.StatefulSet{}
	if err := r.Get(ctx, types.NamespacedName{
		Name:      instance.GetStatefulSetName(),
		Namespace: instance.GetUserNamespace(),
	}, sts); err != nil {
		log.Error(err, "Failed to get StatefulSet")
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	// Check if pods are ready
	if sts.Status.ReadyReplicas != *sts.Spec.Replicas {
		log.Info("Waiting for pods to be ready",
			"ready", sts.Status.ReadyReplicas,
			"desired", *sts.Spec.Replicas)
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	// Update status to running
	instance.Status.State = dbtreev1.StatusRunning
	instance.Status.StatusReason = "Provisioning completed successfully"
	instance.Status.K8sNamespace = instance.Namespace
	instance.Status.K8sResourceName = instance.GetStatefulSetName()

	// Set endpoint and port
	instance.Status.Endpoint = instance.GetServiceName() + "." + instance.Namespace + ".svc.cluster.local"
	instance.Status.Port = instance.GetDefaultPort()
	instance.Status.SecretRef = instance.Spec.SecretRef.Name

	// Set conditions
	instance.SetCondition(ConditionTypeProvisioned, metav1.ConditionTrue,
		"ProvisioningSucceeded", "All resources created successfully")
	instance.SetCondition(ConditionTypeReady, metav1.ConditionTrue,
		"InstanceReady", "Database instance is ready")

	if err := r.updateStatus(ctx, instance); err != nil {
		return ctrl.Result{}, err
	}

	log.Info("Instance provisioned successfully",
		"endpoint", instance.Status.Endpoint,
		"port", instance.Status.Port)

	return ctrl.Result{RequeueAfter: time.Minute}, nil
}

// handleRunning monitors the running instance
func (r *DBInstanceReconciler) handleRunning(ctx context.Context, instance *dbtreev1.DBInstance, prov provisioner.Provisioner) (ctrl.Result, error) {
	// Check if StatefulSet is ready
	sts := &appsv1.StatefulSet{}
	if err := r.Get(ctx, types.NamespacedName{
		Name:      instance.GetStatefulSetName(),
		Namespace: instance.GetUserNamespace(),
	}, sts); err != nil {
		if apierrors.IsNotFound(err) {
			// StatefulSet not found, transition back to provisioning
			instance.Status.State = dbtreev1.StatusProvisioning
			instance.Status.StatusReason = "StatefulSet not found, reprovisioning"
			if err := r.updateStatus(ctx, instance); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Check readiness
	if sts.Status.ReadyReplicas != *sts.Spec.Replicas {
		instance.SetCondition(ConditionTypeReady, metav1.ConditionFalse,
			"PodsNotReady", fmt.Sprintf("Only %d/%d pods are ready",
				sts.Status.ReadyReplicas, *sts.Spec.Replicas))
		if err := r.updateStatus(ctx, instance); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	// All good, ensure Ready condition is True
	instance.SetCondition(ConditionTypeReady, metav1.ConditionTrue,
		"AllPodsReady", "All pods are ready")

	// Requeue after 1 minute to check again
	return ctrl.Result{RequeueAfter: 10 * time.Minute}, r.updateStatus(ctx, instance)
}

// handlePaused scales down the instance
func (r *DBInstanceReconciler) handlePaused(ctx context.Context, instance *dbtreev1.DBInstance, prov provisioner.Provisioner) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("Handling paused state")

	// Scale down StatefulSet to 0
	sts := &appsv1.StatefulSet{}
	if err := r.Get(ctx, types.NamespacedName{
		Name:      instance.GetStatefulSetName(),
		Namespace: instance.GetUserNamespace(),
	}, sts); err != nil {
		if !apierrors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
		// Already scaled down
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}

	// Scale to 0
	replicas := int32(0)
	sts.Spec.Replicas = &replicas
	if err := r.Update(ctx, sts); err != nil {
		return ctrl.Result{}, err
	}

	// Update PausedAt if not set
	if instance.Status.PausedAt == nil {
		now := metav1.Now()
		instance.Status.PausedAt = &now
	}

	instance.SetCondition(ConditionTypeReady, metav1.ConditionFalse,
		"InstancePaused", "Instance is paused to save resources")

	if err := r.updateStatus(ctx, instance); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

// handleStopped handles stopped state
func (r *DBInstanceReconciler) handleStopped(ctx context.Context, instance *dbtreev1.DBInstance, prov provisioner.Provisioner) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("Handling stopped state")

	// Similar to paused but without the deletion timer
	return r.handlePaused(ctx, instance, prov)
}

// handleError tries to recover from error state
func (r *DBInstanceReconciler) handleError(ctx context.Context, instance *dbtreev1.DBInstance, prov provisioner.Provisioner) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("Handling error state", "reason", instance.Status.StatusReason)

	// For now, just log and wait for manual intervention
	// In production, we might want to implement retry logic
	return ctrl.Result{RequeueAfter: 5 * time.Minute}, nil
}

// handleDeletion cleans up resources
func (r *DBInstanceReconciler) handleDeletion(ctx context.Context, instance *dbtreev1.DBInstance) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	if controllerutil.ContainsFinalizer(instance, dbInstanceFinalizer) {
		log.Info("Handling deletion")

		namespace := instance.GetUserNamespace()

		// Delete PVC
		pvc := &corev1.PersistentVolumeClaim{}
		pvcName := instance.GetPVCName()
		if err := r.Get(ctx, types.NamespacedName{
			Name:      pvcName,
			Namespace: namespace,
		}, pvc); err == nil {
			if err := r.Delete(ctx, pvc); err != nil && !apierrors.IsNotFound(err) {
				log.Error(err, "Failed to delete PVC", "name", pvcName)
			}
		}

		// Finalizer 제거
		controllerutil.RemoveFinalizer(instance, dbInstanceFinalizer)
		if err := r.Update(ctx, instance); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// ensureNamespace creates user namespace if it doesn't exist
func (r *DBInstanceReconciler) ensureNamespace(ctx context.Context, name, userID string) error {
	ns := &corev1.Namespace{}
	err := r.Get(ctx, types.NamespacedName{Name: name}, ns)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}

		// Create namespace
		ns = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
				Labels: map[string]string{
					"dbtree.cloud/user-id": userID,
					"dbtree.cloud/managed": "true",
				},
			},
		}

		if err := r.Create(ctx, ns); err != nil && !apierrors.IsAlreadyExists(err) {
			return err
		}
	}

	return nil
}

// createNetworkPolicy creates a NetworkPolicy for the instance
func (r *DBInstanceReconciler) createNetworkPolicy(ctx context.Context, instance *dbtreev1.DBInstance) error {
	labels := map[string]string{
		"app.kubernetes.io/name":      string(instance.Spec.Type),
		"app.kubernetes.io/instance":  instance.Name,
		"app.kubernetes.io/component": "database",
		"app.kubernetes.io/part-of":   "dbtree",
		"dbtree.cloud/instance-uid":   string(instance.UID),
	}

	np := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.GetNetworkPolicyName(),
			Namespace: instance.GetUserNamespace(),
			Labels:    labels,
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: labels,
			},
			PolicyTypes: []networkingv1.PolicyType{
				networkingv1.PolicyTypeIngress,
				networkingv1.PolicyTypeEgress,
			},
			Ingress: []networkingv1.NetworkPolicyIngressRule{
				{
					From: []networkingv1.NetworkPolicyPeer{
						{
							// Allow from same namespace
							PodSelector: &metav1.LabelSelector{},
						},
					},
					Ports: []networkingv1.NetworkPolicyPort{
						{
							Port: &intstr.IntOrString{
								Type:   intstr.Int,
								IntVal: int32(instance.GetDefaultPort()),
							},
							Protocol: &protocolTCP,
						},
					},
				},
			},
			Egress: []networkingv1.NetworkPolicyEgressRule{
				{
					// Allow DNS
					Ports: []networkingv1.NetworkPolicyPort{
						{
							Port: &intstr.IntOrString{
								Type:   intstr.Int,
								IntVal: 53,
							},
							Protocol: &protocolUDP,
						},
						{
							Port: &intstr.IntOrString{
								Type:   intstr.Int,
								IntVal: 53,
							},
							Protocol: &protocolTCP,
						},
					},
				},
				{
					// Allow to same namespace
					To: []networkingv1.NetworkPolicyPeer{
						{
							PodSelector: &metav1.LabelSelector{},
						},
					},
				},
			},
		},
	}

	// Create or update
	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, np, func() error {
		// Update spec if needed
		np.Spec.PodSelector.MatchLabels = labels
		return nil
	})

	return err
}

// getProvisioner returns the appropriate provisioner for the database type
func (r *DBInstanceReconciler) getProvisioner(dbType dbtreev1.DBType) provisioner.Provisioner {
	switch dbType {
	case dbtreev1.DBTypeMongoDB:
		return mongodb.NewProvisioner(r.Client, r.Scheme)
	case dbtreev1.DBTypeRedis:
		return redis.NewProvisioner(r.Client, r.Scheme)
	default:
		return nil
	}
}

// getBackupImage returns the backup container image for the database type
func (r *DBInstanceReconciler) getBackupImage(dbType dbtreev1.DBType) string {
	switch dbType {
	case dbtreev1.DBTypeMongoDB:
		return "mongo:7.0"
	case dbtreev1.DBTypeRedis:
		return "redis:7.2"
	default:
		return ""
	}
}

// getBackupCommand returns the backup command for the database type
func (r *DBInstanceReconciler) getBackupCommand(dbType dbtreev1.DBType) []string {
	switch dbType {
	case dbtreev1.DBTypeMongoDB:
		return []string{"mongodump", "--host", "$(DB_HOST)", "--port", "$(DB_PORT)"}
	case dbtreev1.DBTypeRedis:
		return []string{"redis-cli", "-h", "$(DB_HOST)", "-p", "$(DB_PORT)", "BGSAVE"}
	default:
		return []string{}
	}
}

// updateStatus updates the instance status
func (r *DBInstanceReconciler) updateStatus(ctx context.Context, instance *dbtreev1.DBInstance) error {
	instance.Status.ObservedGeneration = instance.Generation
	return r.Status().Update(ctx, instance)
}

// setErrorCondition sets error condition and updates status
func (r *DBInstanceReconciler) setErrorCondition(ctx context.Context, instance *dbtreev1.DBInstance, reason, message string) (ctrl.Result, error) {
	instance.Status.State = dbtreev1.StatusError
	instance.Status.StatusReason = message
	instance.SetCondition(ConditionTypeError, metav1.ConditionTrue, reason, message)

	return ctrl.Result{RequeueAfter: time.Minute}, r.updateStatus(ctx, instance)
}

// SetupWithManager sets up the controller with the Manager
func (r *DBInstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dbtreev1.DBInstance{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.Secret{}).
		Owns(&corev1.ConfigMap{}).
		Named("dbinstance").
		Complete(r)
}
