package dbservice

import (
	"context"
	"fmt"
	"github.com/piper-hyowon/dBtree/internal/utils/crypto"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"log"
	"time"

	"github.com/google/uuid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/piper-hyowon/dBtree/internal/core/dbservice"
	"github.com/piper-hyowon/dBtree/internal/core/errors"
	"github.com/piper-hyowon/dBtree/internal/core/lemon"
	"github.com/piper-hyowon/dBtree/internal/core/user"
	"github.com/piper-hyowon/dBtree/internal/platform/k8s"
)

type service struct {
	publicHost   string
	dbiStore     dbservice.DBInstanceStore
	lemonService lemon.Service
	presetStore  dbservice.PresetStore
	userStore    user.Store
	k8sClient    k8s.Client
	portStore    dbservice.PortStore
	logger       *log.Logger
}

func (s *service) ListInstances(ctx context.Context, userID string) ([]*dbservice.DBInstance, error) {
	instances, err := s.dbiStore.List(ctx, userID)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	return instances, nil
}

func (s *service) mapInfraStatusToInstanceStatus(status *k8s.MongoDBStatus) dbservice.InstanceStatus {
	switch status.Phase {
	case "Running":
		if status.Ready {
			return dbservice.StatusRunning
		}
		return dbservice.StatusProvisioning
	case "Failed":
		return dbservice.StatusError
	case "Pending":
		return dbservice.StatusProvisioning
	default:
		return dbservice.StatusProvisioning
	}
}

func (s *service) GetInstanceWithSync(ctx context.Context, userID, id string) (*dbservice.DBInstance, error) {
	// DB에서 조회
	instance, err := s.dbiStore.Find(ctx, id)
	if err != nil {
		return nil, err
	}

	// 권한 확인
	if instance.UserID != userID {
		return nil, errors.NewResourceNotFoundError("instance", id)
	}

	s.logger.Printf("Instance from DB - Status: %s, K8sNamespace: %s, K8sResourceName: %s",
		instance.Status, instance.K8sNamespace, instance.K8sResourceName)

	// Provisioning 상태면 K8s에서 실제 상태 확인
	if instance.Status == dbservice.StatusProvisioning && instance.Type == dbservice.MongoDB {
		status, err := s.k8sClient.GetMongoDBStatus(ctx, instance.K8sNamespace, instance.K8sResourceName)
		if err != nil {
			s.logger.Printf("Failed to get MongoDB status: %v", err)
		} else {
			s.logger.Printf("K8s status - Phase: %s, Ready: %v", status.Phase, status.Ready)

			// 상태 업데이트
			newStatus := s.mapMongoDBStatus(status)
			s.logger.Printf("Mapped status: %s", newStatus)

			if newStatus != instance.Status {
				if err := s.dbiStore.UpdateStatus(ctx, instance.ID, newStatus, status.Message); err != nil {
					s.logger.Printf("Failed to update status in DB: %v", err)
				} else {
					instance.Status = newStatus
					instance.StatusReason = status.Message
				}
			}

			// 연결 정보 업데이트
			if status.Ready && instance.Endpoint == "" {
				instance.Endpoint = status.Endpoint
				instance.Port = int(status.Port)
				_ = s.dbiStore.Update(ctx, instance)
			}
		}
	}

	return instance, nil
}
func (s *service) ListPresets(ctx context.Context) ([]*dbservice.DBPreset, error) {
	presets, err := s.presetStore.ListByType(ctx, dbservice.MongoDB)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	return presets, nil
}

func (s *service) mapMongoDBStatus(status *k8s.MongoDBStatus) dbservice.InstanceStatus {
	switch status.Phase {
	case "running":
		if status.Ready {
			return dbservice.StatusRunning
		}
		return dbservice.StatusProvisioning
	case "failed", "error":
		return dbservice.StatusError
	case "pending", "provisioning":
		return dbservice.StatusProvisioning
	default:
		s.logger.Printf("Unknown phase: %s", status.Phase) // 디버깅용
		return dbservice.StatusProvisioning
	}
}
func (s *service) UpdateInstance(ctx context.Context, userID, instanceID string, req *dbservice.UpdateInstanceRequest) (*dbservice.DBInstance, error) {
	//TODO implement me
	panic("implement me")
}

func (s *service) DeleteInstance(ctx context.Context, userID, instanceID string) error {
	instance, err := s.dbiStore.Find(ctx, instanceID)
	if err != nil {
		return errors.Wrap(err)
	}
	if instance == nil || instance.UserID != userID {
		return errors.NewResourceNotFoundError("instance", instanceID)
	}

	if !instance.CanDelete() {
		return errors.NewInvalidStatusTransitionError(string(instance.Status), string(dbservice.StatusDeleting))
	}

	if err := s.dbiStore.UpdateStatus(ctx, instance.ID, dbservice.StatusDeleting, "Deletion requested"); err != nil {
		return errors.Wrap(err)
	}

	// 포트 해제
	if s.portStore != nil && instance.ExternalID != "" {
		if err := s.portStore.ReleasePort(ctx, instance.ExternalID); err != nil {
			s.logger.Printf("Failed to release port for instance %s: %v", instanceID, err)
			// 포트 해제 실패는 무시하고 계속 진행 (TODO: reporting)
		}
	}

	// K8s DBInstance CRD 삭제 (Operator가 나머지 리소스 정리)
	if instance.K8sNamespace != "" && instance.K8sResourceName != "" {
		if err := s.k8sClient.DeleteDBInstance(ctx, instance.K8sNamespace, instance.K8sResourceName); err != nil {
			s.logger.Printf("Failed to delete K8s resources for instance %s: %v", instanceID, err)
			// K8s 삭제 실패해도 DB는 삭제 처리
		}
	}

	if err := s.dbiStore.Delete(ctx, instanceID); err != nil {
		return errors.Wrap(err)
	}

	return nil
}

func (s *service) StartInstance(ctx context.Context, userID, instanceID string) error {
	//TODO implement me
	panic("implement me")
}

func (s *service) StopInstance(ctx context.Context, userID, instanceID string) error {
	//TODO implement me
	panic("implement me")
}

func (s *service) RestartInstance(ctx context.Context, userID, instanceID string) error {
	//TODO implement me
	panic("implement me")
}

func (s *service) CreateBackup(ctx context.Context, userID, instanceID string, name string) (*dbservice.BackupRecord, error) {
	//TODO implement me
	panic("implement me")
}

func (s *service) ListBackups(ctx context.Context, userID, instanceID string) ([]*dbservice.BackupRecord, error) {
	//TODO implement me
	panic("implement me")
}

func (s *service) RestoreFromBackup(ctx context.Context, userID, instanceID string, backupID string) error {
	//TODO implement me
	panic("implement me")
}

func (s *service) InstanceMetrics(ctx context.Context, instanceID string) (*dbservice.InstanceMetrics, error) {
	//TODO implement me
	panic("implement me")
}

func (s *service) EstimateCost(ctx context.Context, req *dbservice.EstimateCostRequest) (*dbservice.EstimateCostResponse, error) {
	//TODO implement me
	panic("implement me")
}

var _ dbservice.Service = (*service)(nil)

func NewService(
	publicHost string,
	dbiStore dbservice.DBInstanceStore,
	presetStore dbservice.PresetStore,
	lemonService lemon.Service,
	userStore user.Store,
	k8sClient k8s.Client,
	portStore dbservice.PortStore,
	logger *log.Logger,
) dbservice.Service {
	return &service{
		publicHost:   publicHost,
		dbiStore:     dbiStore,
		presetStore:  presetStore,
		lemonService: lemonService,
		userStore:    userStore,
		k8sClient:    k8sClient,
		portStore:    portStore,
		logger:       logger,
	}
}

func (s *service) CreateInstance(ctx context.Context, userID string, userLemon int, req *dbservice.CreateInstanceRequest) (*dbservice.CreateInstanceResponse, error) {
	existingInstances, err := s.dbiStore.CountActive(ctx, userID)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	if existingInstances >= dbservice.MaxInstancesPerUser {
		return nil, errors.NewLimitExceededError("instance", dbservice.MaxInstancesPerUser)
	}

	existing, err := s.dbiStore.FindByUserAndName(ctx, userID, req.Name)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	if existing != nil {
		return nil, errors.NewInstanceNameConflictError(req.Name)
	}

	backupCfg := dbservice.BackupConfig{
		Enabled:       req.BackupEnabled,
		Schedule:      req.BackupSchedule,
		RetentionDays: req.BackupRetentionDays,
		StorageSize:   "",
	}

	var instance *dbservice.DBInstance
	if req.PresetID != nil {
		// 프리셋 기반 생성
		preset, err := s.presetStore.Find(ctx, *req.PresetID)
		if err != nil {
			return nil, errors.Wrap(err)
		}
		if preset == nil {
			return nil, errors.NewResourceNotFoundError("preset", *req.PresetID)
		}

		instance = &dbservice.DBInstance{
			ExternalID:        uuid.New().String(),
			UserID:            userID,
			Name:              req.Name,
			Type:              preset.Type,
			Size:              preset.Size,
			Mode:              preset.Mode,
			CreatedFromPreset: &preset.ID,
			Resources:         preset.Resources,
			Cost:              preset.Cost,
			Config:            preset.DefaultConfig,
			BackupConfig:      backupCfg,
			Status:            dbservice.StatusProvisioning,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}
	} else {
		// 커스텀 스펙
		if req.Type == nil || req.Resources == nil {
			return nil, errors.NewInvalidParameterError("type,resources", "required")
		}

		// 모드 기본값 설정
		mode := *req.Mode
		if req.Mode == nil {
			switch *req.Type {
			case dbservice.MongoDB:
				mode = dbservice.ModeStandalone
			case dbservice.Redis:
				mode = dbservice.ModeBasic
			}
		}

		configValidator := dbservice.NewConfigValidator()

		// Config 처리: 기본값 + 사용자 입력
		finalConfig := configValidator.GetDefaultConfig(*req.Type, mode)
		if req.Config != nil {
			for k, v := range req.Config {
				finalConfig[k] = v
			}
		}

		// Config 검증
		if err := configValidator.ValidateConfig(*req.Type, mode, finalConfig, req.Resources); err != nil {
			return nil, errors.NewInvalidParameterError("config", err.Error())
		}

		// 크기 계산
		size := req.Resources.CalculateSize()
		cost := dbservice.CalculateCustomCost(*req.Type, *req.Resources)

		instance = &dbservice.DBInstance{
			ExternalID:   uuid.New().String(),
			UserID:       userID,
			Name:         req.Name,
			Type:         *req.Type,
			Size:         size,
			Mode:         mode,
			Resources:    *req.Resources,
			Cost:         cost,
			Config:       finalConfig,
			BackupConfig: backupCfg,
			Status:       dbservice.StatusProvisioning,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
	}

	// 레몬 잔액 확인
	if userLemon < instance.Cost.CreationCost {
		return nil, errors.NewInsufficientLemonsError(instance.Cost.CreationCost+1, instance.Cost.CreationCost-userLemon)
	}

	// 환불 보장을 위한 defer
	lemonDeducted := false
	defer func() {
		if err != nil && lemonDeducted {
			refundErr := s.lemonService.AddLemons(ctx, userID, instance.Cost.CreationCost,
				lemon.ActionInstanceCreateRefund, fmt.Sprintf("실패: %v", err))
			if refundErr != nil {
				// 환불 실패는 심각한 문제 - 반드시 기록
				s.logger.Printf("CRITICAL: 환불 실패 - userID: %s, amount: %d, error: %v",
					userID, instance.Cost.CreationCost, refundErr)

				// 환불 실패 기록 테이블에 저장 (수동 처리를 위해)
				//s.recordRefundFailure(userID, instance.Cost.CreationCost, refundErr)
			}
		}
	}()

	// 레몬 차감 (생성 비용)
	if err := s.lemonService.DeductLemons(ctx, userID, instance.Cost.CreationCost, lemon.ActionInstanceCreate, ""); err != nil {
		return nil, errors.Wrap(err)
	}

	// K8s 네임스페이스 및 리소스 생성
	secretData, err := s.provisionK8sResources(ctx, instance)
	if err != nil {
		// 실패 시 레몬 환불
		if refundErr := s.lemonService.AddLemons(ctx, userID, instance.Cost.CreationCost,
			lemon.ActionInstanceCreateRefund, fmt.Sprint(err)); refundErr != nil {
			s.logger.Printf("CRITICAL: 환불 실패 - userID: %s, amount: %d, error: %v",
				userID, instance.Cost.CreationCost, refundErr)
			// 알림 시스템에 전송하거나 별도 테이블에 기록
		}

		return nil, errors.Wrapf(err, "failed to provision k8s resources")
	}
	username, password := string(secretData["username"]), string(secretData["password"])

	// DB에 인스턴스 저장
	if err := s.dbiStore.Create(ctx, instance); err != nil {
		// K8s 리소스 정리
		_ = s.cleanupK8sResources(ctx, instance)
		// 레몬 환불
		if refundErr := s.lemonService.AddLemons(ctx, userID, instance.Cost.CreationCost,
			lemon.ActionInstanceCreateRefund, fmt.Sprint(err)); refundErr != nil {
			s.logger.Printf("CRITICAL: 환불 실패 - userID: %s, amount: %d, error: %v",
				userID, instance.Cost.CreationCost, refundErr)
			// 알림 시스템에 전송하거나 별도 테이블에 기록
		}
		return nil, errors.Wrap(err)
	}

	// 외부 접속 설정 (에러가 나도 인스턴스는 이미 생성됨)
	var externalPort int

	if s.portStore != nil {
		port, err := s.portStore.AllocatePort(ctx, instance.ExternalID)
		if err != nil {
			s.logger.Printf("WARNING: 외부 포트 할당 실패: %v", err)
		} else {
			// NodePort 서비스 생성
			selector := map[string]string{
				"app":                      instance.Name,
				"dbtree.cloud/instance-id": instance.ExternalID,
			}

			dbPort := int32(27017) // MongoDB default
			if instance.Type == dbservice.Redis {
				dbPort = 6379
			}

			err = s.k8sClient.CreateNodePortService(
				ctx,
				instance.K8sNamespace,
				instance.Name,
				dbPort,
				int32(port),
				selector,
			)

			if err != nil {
				s.logger.Printf("WARNING: NodePort 서비스 생성 실패: %v", err)
				// 포트 할당 롤백
				_ = s.portStore.ReleasePort(ctx, instance.ExternalID)
			} else {
				externalPort = port
			}
		}
	}

	credentials := &dbservice.Credentials{
		Username: username,
		Password: password,
	}

	// 외부 접속 정보 추가
	if externalPort > 0 {
		externalHost := s.publicHost
		credentials.ExternalHost = externalHost
		credentials.ExternalPort = externalPort
		if instance.Type == dbservice.MongoDB {
			credentials.ExternalURI = fmt.Sprintf("mongodb://%s:%s@%s:%d/%s?authSource=%s",
				username, password, externalHost, externalPort, req.Name, username)
		} else {
			credentials.ExternalURI = fmt.Sprintf("%s://%s:%s@%s:%d/%s",
				instance.Type, username, password,
				externalHost, externalPort, req.Name)
		}
	}

	return instance.ToCreateResponse(credentials), nil
}

func (s *service) provisionK8sResources(ctx context.Context, instance *dbservice.DBInstance) (map[string][]byte, error) {
	// 네임스페이스 생성
	namespace := fmt.Sprintf("user-%s", instance.UserID)
	if err := s.k8sClient.CreateNamespace(ctx, namespace); err != nil {
		return nil, err
	}
	instance.K8sNamespace = namespace

	// Secret 생성
	secretName := fmt.Sprintf("%s-secret", instance.Name)
	secretData := s.generateSecretData(instance)
	if err := s.k8sClient.CreateSecret(ctx, namespace, secretName, secretData); err != nil {
		return nil, err
	}
	instance.K8sSecretRef = secretName

	// DBInstance CRD 생성
	if err := s.createDBInstanceCRD(ctx, instance); err != nil {
		return nil, err
	}

	return secretData, nil
}

func (s *service) createDBInstanceCRD(ctx context.Context, instance *dbservice.DBInstance) error {
	params := k8s.DBInstanceParams{
		Name:              instance.Name,
		Type:              string(instance.Type),
		Size:              string(instance.Size),
		Mode:              string(instance.Mode),
		SecretRef:         instance.K8sSecretRef,
		UserID:            instance.UserID,
		ExternalID:        instance.ExternalID,
		CreatedFromPreset: instance.CreatedFromPreset,
		Resources: k8s.ResourceSpec{
			CPU:    instance.Resources.CPU,
			Memory: instance.Resources.Memory,
			Disk:   instance.Resources.Disk,
		},
		Backup: k8s.BackupSpec{
			Enabled:       instance.BackupConfig.Enabled,
			Schedule:      instance.BackupConfig.Schedule,
			RetentionDays: instance.BackupConfig.RetentionDays,
		},
		Config: instance.Config,
	}

	// CRD 생성
	spec := k8s.BuildDBInstanceSpec(params)
	labels := map[string]string{
		"app.kubernetes.io/managed-by": "dbtree",
		"dbtree.cloud/user-id":         instance.UserID,
		"dbtree.cloud/instance-id":     instance.ExternalID,
	}

	dbInstanceCRD := k8s.BuildDBInstanceCRD(instance.K8sNamespace, instance.Name, spec, labels)

	// Dynamic client를 사용해서 CRD 생성
	gvr := schema.GroupVersionResource{
		Group:    "dbtree.cloud",
		Version:  "v1",
		Resource: "dbinstances",
	}

	_, err := s.k8sClient.Dynamic().Resource(gvr).Namespace(instance.K8sNamespace).
		Create(ctx, dbInstanceCRD, metav1.CreateOptions{})

	if err != nil {
		if k8serrors.IsAlreadyExists(err) {
			// 이미 존재하는 경우 업데이트 시도
			s.logger.Printf("DBInstance CRD already exists, attempting update: %s/%s",
				instance.K8sNamespace, instance.Name)

			return s.k8sClient.UpdateDBInstance(ctx, instance.K8sNamespace,
				instance.Name, dbInstanceCRD)
		}
		return errors.Wrapf(err, "failed to create DBInstance CRD")
	}

	instance.K8sResourceName = instance.Name
	return nil
}

func (s *service) updateK8sResource(ctx context.Context, instance *dbservice.DBInstance) error {
	// DBInstance CRD 업데이트
	gvr := schema.GroupVersionResource{
		Group:    "dbtree.cloud",
		Version:  "v1",
		Resource: "dbinstances",
	}

	// 기존 리소스 가져오기
	existing, err := s.k8sClient.Dynamic().Resource(gvr).Namespace(instance.K8sNamespace).
		Get(ctx, instance.K8sResourceName, metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "failed to get existing DBInstance")
	}

	// 스펙 업데이트
	spec := existing.Object["spec"].(map[string]interface{})
	spec["resources"] = map[string]interface{}{
		"cpu":    instance.Resources.CPU,
		"memory": instance.Resources.Memory,
		"disk":   instance.Resources.Disk,
	}
	spec["backup"] = map[string]interface{}{
		"enabled":       instance.BackupConfig.Enabled,
		"schedule":      instance.BackupConfig.Schedule,
		"retentionDays": instance.BackupConfig.RetentionDays,
	}

	_, err = s.k8sClient.Dynamic().Resource(gvr).Namespace(instance.K8sNamespace).
		Update(ctx, existing, metav1.UpdateOptions{})

	return err
}

func (s *service) deleteK8sResource(ctx context.Context, instance *dbservice.DBInstance) error {
	gvr := schema.GroupVersionResource{
		Group:    "dbtree.cloud",
		Version:  "v1",
		Resource: "dbinstances",
	}

	err := s.k8sClient.Dynamic().Resource(gvr).Namespace(instance.K8sNamespace).
		Delete(ctx, instance.K8sResourceName, metav1.DeleteOptions{})

	if err != nil && !k8serrors.IsNotFound(err) {
		return errors.Wrapf(err, "failed to delete DBInstance CRD")
	}

	return nil
}

func (s *service) scaleK8sResource(ctx context.Context, instance *dbservice.DBInstance, replicas int) error {
	// StatefulSet 스케일링을 통한 시작/중지
	// DBInstance CRD가 있으면 Operator가 자동으로 처리할 것
	// 여기서는 상태만 변경하고 Operator가 reconcile하도록 함

	gvr := schema.GroupVersionResource{
		Group:    "dbtree.cloud",
		Version:  "v1",
		Resource: "dbinstances",
	}

	// 상태 변경을 위한 패치
	patch := []byte(fmt.Sprintf(`{"status":{"state":"%s"}}`, instance.Status))

	_, err := s.k8sClient.Dynamic().Resource(gvr).Namespace(instance.K8sNamespace).
		Patch(ctx, instance.K8sResourceName, types.MergePatchType, patch, metav1.PatchOptions{})

	return err
}

func (s *service) cleanupK8sResources(ctx context.Context, instance *dbservice.DBInstance) error {
	// 실패 시 생성된 리소스 정리
	if instance.K8sResourceName != "" {
		_ = s.deleteK8sResource(ctx, instance)
	}
	// Secret과 네임스페이스는 다른 인스턴스가 사용할 수 있으므로 삭제하지 않음
	return nil
}

func (s *service) generateSecretData(instance *dbservice.DBInstance) map[string][]byte {
	password, _ := crypto.GenerateSecurePassword()

	switch instance.Type {
	case dbservice.MongoDB:
		return map[string][]byte{
			"username":                   []byte("admin"),
			"password":                   []byte(password),
			"MONGO_INITDB_ROOT_USERNAME": []byte("admin"),
			"MONGO_INITDB_ROOT_PASSWORD": []byte(password),
			"MONGO_INITDB_DATABASE":      []byte("admin"),
		}
	case dbservice.Redis:
		return map[string][]byte{
			"password": []byte(password),
		}
	default:
		return map[string][]byte{
			"username": []byte("admin"),
			"password": []byte(password),
		}
	}
}
