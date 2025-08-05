package k8s

import (
	"context"
	"fmt"
	"github.com/piper-hyowon/dBtree/internal/core/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type MongoDBStatus struct {
	Phase        string
	Ready        bool
	Message      string
	Members      int
	ReadyMembers int
	Endpoint     string
	Port         int32
}

func (c *client) GetMongoDBStatus(ctx context.Context, namespace, name string) (*MongoDBStatus, error) {
	resource, err := c.dynamic.Resource(dbInstanceGVR).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "DBInstance 리소스 조회 실패")
	}

	// DBInstance CRD에서 상태 파싱
	status := parseDBInstanceStatus(resource)

	// StatefulSet 상태 확인
	stsName := fmt.Sprintf("%s-sts", name)
	sts, err := c.clientset.AppsV1().StatefulSets(namespace).Get(ctx, stsName, metav1.GetOptions{})
	if err == nil {
		status.Members = int(*sts.Spec.Replicas)
		status.ReadyMembers = int(sts.Status.ReadyReplicas)
		status.Ready = sts.Status.ReadyReplicas == *sts.Spec.Replicas && sts.Status.ReadyReplicas > 0
	}

	return status, nil
}

func parseDBInstanceStatus(resource *unstructured.Unstructured) *MongoDBStatus {
	result := &MongoDBStatus{
		Phase:   "Unknown",
		Message: "Status not available",
	}

	status, found, err := unstructured.NestedMap(resource.Object, "status")
	if err != nil || !found {
		return result
	}

	// state 필드 읽기
	if state, found, err := unstructured.NestedString(status, "state"); err == nil && found {
		// state를 그대로 사용
		result.Phase = state
	}

	// Ready condition 확인
	if conditions, found, err := unstructured.NestedSlice(status, "conditions"); err == nil && found {
		for _, condition := range conditions {
			if condMap, ok := condition.(map[string]interface{}); ok {
				if condType, _ := condMap["type"].(string); condType == "Ready" {
					if condStatus, _ := condMap["status"].(string); condStatus == "True" {
						result.Ready = true
					}
				}
			}
		}
	}

	// 나머지
	if reason, found, err := unstructured.NestedString(status, "statusReason"); err == nil && found {
		result.Message = reason
	}
	if endpoint, found, err := unstructured.NestedString(status, "endpoint"); err == nil && found {
		result.Endpoint = endpoint
	}
	if port, found, err := unstructured.NestedInt64(status, "port"); err == nil && found {
		result.Port = int32(port)
	}

	return result
}
