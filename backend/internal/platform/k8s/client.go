package k8s

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"log"
	"path/filepath"
	"time"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"github.com/piper-hyowon/dBtree/internal/core/errors"
	"github.com/piper-hyowon/dBtree/internal/platform/config"
)

type Client interface {
	Clientset() kubernetes.Interface
	Dynamic() dynamic.Interface
	RESTConfig() *rest.Config

	CreateNamespace(ctx context.Context, name string) error
	CreateSecret(ctx context.Context, namespace, name string, data map[string][]byte) error
	DeleteNamespace(ctx context.Context, name string) error

	CreateDBInstance(ctx context.Context, namespace string, instance *unstructured.Unstructured) error
	UpdateDBInstance(ctx context.Context, namespace, name string, instance *unstructured.Unstructured) error
	DeleteDBInstance(ctx context.Context, namespace, name string) error
	DBInstance(ctx context.Context, namespace, name string) (*unstructured.Unstructured, error)

	PatchDBInstanceStatus(ctx context.Context, namespace, name string, state string, reason string) error

	CreateNodePortService(ctx context.Context, namespace, name string, targetPort, nodePort int32, selector map[string]string) error

	GetMongoDBStatus(ctx context.Context, namespace, name string) (*MongoDBStatus, error)
}

type client struct {
	clientset  kubernetes.Interface
	dynamic    dynamic.Interface
	restConfig *rest.Config
	logger     *log.Logger
}

var _ Client = (*client)(nil)

func NewClient(config config.K8sConfig, logger *log.Logger) (Client, error) {
	var restConfig *rest.Config
	var err error

	// 프로덕션
	if config.InCluster {
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			return nil, errors.Wrapf(err, "in-cluster config 로드 실패")
		}
	} else {
		// 로컬
		kubeconfig := config.KubeConfigPath
		if kubeconfig == "" {
			if home := homedir.HomeDir(); home != "" {
				kubeconfig = filepath.Join(home, ".kube", "config")
			}
		}

		restConfig, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, errors.Wrapf(err, "kubeconfig 로드 실패: %s", kubeconfig)
		}
	}

	restConfig.Timeout = 30 * time.Second

	// 클라이언트 생성
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "clientset 생성 실패")
	}

	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "dynamic client 생성 실패")
	}

	// 연결 테스트
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = clientset.Discovery().ServerVersion()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to connect to kubernetes cluster")
	}

	logger.Printf("K8s 클라이언트 초기화 완료")

	return &client{
		clientset:  clientset,
		dynamic:    dynamicClient,
		restConfig: restConfig,
		logger:     logger,
	}, nil
}

func (c *client) Clientset() kubernetes.Interface {
	return c.clientset
}

func (c *client) Dynamic() dynamic.Interface {
	return c.dynamic
}

func (c *client) RESTConfig() *rest.Config {
	return c.restConfig
}

func (c *client) CreateNamespace(ctx context.Context, name string) error {
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"dbtree.cloud/managed": "true",
				"dbtree.cloud/type":    "user",
			},
		},
	}

	_, err := c.clientset.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})
	if err != nil {
		if k8serrors.IsAlreadyExists(err) {
			c.logger.Printf("Namespace %s already exists", name)
			return nil
		}
		return errors.Wrapf(err, "failed to create namespace %s", name)
	}

	c.logger.Printf("Created namespace: %s", name)
	return nil
}

func (c *client) CreateSecret(ctx context.Context, namespace, name string, data map[string][]byte) error {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"dbtree.cloud/managed": "true",
			},
		},
		Type: corev1.SecretTypeOpaque,
		Data: data,
	}

	_, err := c.clientset.CoreV1().Secrets(namespace).Create(ctx, secret, metav1.CreateOptions{})
	if err != nil {
		if k8serrors.IsAlreadyExists(err) {
			// 이미 존재하면 업데이트
			existing, getErr := c.clientset.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
			if getErr != nil {
				return errors.Wrapf(getErr, "failed to get existing secret")
			}

			existing.Data = data
			_, updateErr := c.clientset.CoreV1().Secrets(namespace).Update(ctx, existing, metav1.UpdateOptions{})
			if updateErr != nil {
				return errors.Wrapf(updateErr, "failed to update secret")
			}

			c.logger.Printf("Updated existing secret: %s/%s", namespace, name)
			return nil
		}
		return errors.Wrapf(err, "failed to create secret %s/%s", namespace, name)
	}

	c.logger.Printf("Created secret: %s/%s", namespace, name)
	return nil
}

// DeleteNamespace deletes a namespace
func (c *client) DeleteNamespace(ctx context.Context, name string) error {
	err := c.clientset.CoreV1().Namespaces().Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			c.logger.Printf("Namespace %s not found", name)
			return nil
		}
		return errors.Wrapf(err, "failed to delete namespace %s", name)
	}

	c.logger.Printf("Deleted namespace: %s", name)
	return nil
}

// DBInstance CRD methods

var dbInstanceGVR = schema.GroupVersionResource{
	Group:    "dbtree.cloud",
	Version:  "v1",
	Resource: "dbinstances",
}

// CreateDBInstance creates a new DBInstance CRD
func (c *client) CreateDBInstance(ctx context.Context, namespace string, instance *unstructured.Unstructured) error {
	_, err := c.dynamic.Resource(dbInstanceGVR).Namespace(namespace).Create(ctx, instance, metav1.CreateOptions{})
	if err != nil {
		return errors.Wrapf(err, "failed to create DBInstance CRD")
	}

	c.logger.Printf("Created DBInstance: %s/%s", namespace, instance.GetName())
	return nil
}

// UpdateDBInstance updates an existing DBInstance CRD
func (c *client) UpdateDBInstance(ctx context.Context, namespace, name string, instance *unstructured.Unstructured) error {
	// 기존 리소스 가져오기
	existing, err := c.dynamic.Resource(dbInstanceGVR).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "failed to get existing DBInstance")
	}

	// ResourceVersion 설정 (optimistic concurrency control)
	instance.SetResourceVersion(existing.GetResourceVersion())

	_, err = c.dynamic.Resource(dbInstanceGVR).Namespace(namespace).Update(ctx, instance, metav1.UpdateOptions{})
	if err != nil {
		return errors.Wrapf(err, "failed to update DBInstance CRD")
	}

	c.logger.Printf("Updated DBInstance: %s/%s", namespace, name)
	return nil
}

// DeleteDBInstance deletes a DBInstance CRD
func (c *client) DeleteDBInstance(ctx context.Context, namespace, name string) error {
	err := c.dynamic.Resource(dbInstanceGVR).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			c.logger.Printf("DBInstance %s/%s not found", namespace, name)
			return nil
		}
		return errors.Wrapf(err, "failed to delete DBInstance CRD")
	}

	c.logger.Printf("Deleted DBInstance: %s/%s", namespace, name)
	return nil
}

// DBInstance gets a DBInstance CRD
func (c *client) DBInstance(ctx context.Context, namespace, name string) (*unstructured.Unstructured, error) {
	instance, err := c.dynamic.Resource(dbInstanceGVR).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "failed to get DBInstance CRD")
	}

	return instance, nil
}

func (c *client) CreateNodePortService(ctx context.Context, namespace, name string, targetPort, nodePort int32, selector map[string]string) error {
	if _, exists := selector["app.kubernetes.io/instance"]; !exists {
		selector["app.kubernetes.io/instance"] = name
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name + "-external",
			Namespace: namespace,
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeNodePort,
			Ports: []corev1.ServicePort{{
				Port:       targetPort,
				TargetPort: intstr.FromInt32(targetPort),
				NodePort:   nodePort,
			}},
			Selector: selector,
		},
	}

	_, err := c.clientset.CoreV1().Services(namespace).Create(ctx, service, metav1.CreateOptions{})
	return err
}

func (c *client) PatchDBInstanceStatus(ctx context.Context, namespace, name string, state string, reason string) error {
	// lastUpdated 필드 제거 - CRD에 정의되지 않은 필드
	patch := []byte(fmt.Sprintf(`{"status":{"state":"%s","statusReason":"%s"}}`, state, reason))

	_, err := c.dynamic.Resource(dbInstanceGVR).Namespace(namespace).
		Patch(ctx, name, types.MergePatchType, patch, metav1.PatchOptions{}, "status")

	if err != nil {
		return errors.Wrapf(err, "failed to patch DBInstance status")
	}

	c.logger.Printf("Patched DBInstance status: %s/%s to %s", namespace, name, state)
	return nil
}
