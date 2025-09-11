package dbservice

import (
	"context"
	"fmt"
	"github.com/piper-hyowon/dBtree/internal/core/resource"
	"github.com/piper-hyowon/dBtree/internal/utils/crypto"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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
	publicDBHost    string
	dbiStore        dbservice.DBInstanceStore
	lemonService    lemon.Service
	presetStore     dbservice.PresetStore
	userStore       user.Store
	k8sClient       k8s.Client
	portStore       dbservice.PortStore
	resourceManager resource.Manager
	logger          *log.Logger
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

	// K8s와 상태 동기화 (MongoDB만 지원하므로 타입 체크)
	if instance.K8sNamespace != "" && instance.K8sResourceName != "" && instance.Type == dbservice.MongoDB {
		// DBInstance CRD 가져오기
		crd, err := s.k8sClient.DBInstance(ctx, instance.K8sNamespace, instance.K8sResourceName)
		if err != nil {
			s.logger.Printf("Failed to get DBInstance CRD: %v", err)
		} else if crd != nil {
			// CRD의 status.state 확인
			status, found, err := unstructured.NestedString(crd.Object, "status", "state")
			if err == nil && found && status != "" {
				k8sStatus := dbservice.InstanceStatus(status)
				s.logger.Printf("K8s CRD status: %s", k8sStatus)

				// 상태가 다르면 DB 업데이트
				if k8sStatus != instance.Status {
					s.logger.Printf("Status mismatch - DB: %s, K8s: %s. Updating DB...",
						instance.Status, k8sStatus)

					reason, _, _ := unstructured.NestedString(crd.Object, "status", "statusReason")

					if err := s.dbiStore.UpdateStatus(ctx, instance.ID, k8sStatus, reason); err != nil {
						s.logger.Printf("Failed to update status in DB: %v", err)
					} else {
						instance.Status = k8sStatus
						instance.StatusReason = reason
					}
				}
			}
		}

		// MongoDB 상태 확인 (provisioning 등)
		if instance.Status == dbservice.StatusProvisioning {
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
		// 이미 없는 경우는 성공으로 처리
		if errors.Is(errors.NewResourceNotFoundError("instance", instanceID), err) {
			s.logger.Printf("Instance %s already deleted", instanceID)
			return nil
		}
		return errors.Wrap(err)
	}

	return nil
}

func (s *service) StartInstance(ctx context.Context, userID, instanceID string) error {
	// 1. 인스턴스 조회 및 권한 확인
	instance, err := s.dbiStore.Find(ctx, instanceID)
	if err != nil {
		return errors.Wrap(err)
	}
	if instance == nil || instance.UserID != userID {
		return errors.NewResourceNotFoundError("instance", instanceID)
	}

	// 2. 시작 가능한 상태인지 확인
	if !instance.CanStart() {
		return errors.NewInvalidStatusTransitionError(string(instance.Status), string(dbservice.StatusRunning))
	}

	// 3. 레몬 잔액 확인 (시작시 바로 과금)
	usr, err := s.userStore.FindById(ctx, userID)
	if err != nil {
		return errors.Wrap(err)
	}
	balance := usr.LemonBalance

	hourlyCost := instance.Cost.HourlyLemons
	if balance < hourlyCost {
		return errors.NewInsufficientLemonsError(hourlyCost, hourlyCost-balance)
	}

	// 4. 상태 변경 (DB)
	if err := s.dbiStore.UpdateStatus(ctx, instance.ID, dbservice.StatusRunning, "Started by user"); err != nil {
		return errors.Wrap(err)
	}

	// 5. K8s CRD 상태 변경
	if instance.K8sNamespace != "" && instance.K8sResourceName != "" {
		if err := s.k8sClient.PatchDBInstanceStatus(
			ctx,
			instance.K8sNamespace,
			instance.K8sResourceName,
			string(dbservice.StatusRunning),
			"Started by user request",
		); err != nil {
			s.logger.Printf("K8s 상태 업데이트 실패: %v", err)
			// 롤백
			_ = s.dbiStore.UpdateStatus(ctx, instance.ID, instance.Status, "K8s update failed")
			return errors.Wrap(err)
		}
	}

	// 6. 즉시 과금
	if err := s.lemonService.ProcessInstanceFee(
		ctx,
		userID,
		instance.ExternalID,
		hourlyCost,
		lemon.ActionInstanceMaintain,
		&instance.ID,
	); err != nil {
		s.logger.Printf("시작 과금 실패: %v", err)
		// 과금 실패시 다시 중지? 아니면 그냥 로그만?
	}

	// 7. 과금 시간 업데이트
	_ = s.dbiStore.UpdateBillingTime(ctx, instance.ID, time.Now())

	s.logger.Printf("인스턴스 %s 시작됨", instanceID)
	return nil
}

func (s *service) StopInstance(ctx context.Context, userID, instanceID string) error {
	// 1. 인스턴스 조회 및 권한 확인
	instance, err := s.dbiStore.Find(ctx, instanceID)
	if err != nil {
		return errors.Wrap(err)
	}
	if instance == nil || instance.UserID != userID {
		return errors.NewResourceNotFoundError("instance", instanceID)
	}

	// 2. 중지 가능한 상태인지 확인
	if !instance.CanStop() {
		return errors.NewInvalidStatusTransitionError(string(instance.Status), string(dbservice.StatusStopped))
	}

	// 3. 상태 변경 (DB)
	if err := s.dbiStore.UpdateStatus(ctx, instance.ID, dbservice.StatusStopped, "Stopped by user"); err != nil {
		return errors.Wrap(err)
	}

	// 4. K8s CRD 상태 변경
	if instance.K8sNamespace != "" && instance.K8sResourceName != "" {
		if err := s.k8sClient.PatchDBInstanceStatus(
			ctx,
			instance.K8sNamespace,
			instance.K8sResourceName,
			string(dbservice.StatusStopped),
			"Stopped by user request",
		); err != nil {
			s.logger.Printf("K8s 상태 업데이트 실패: %v", err)
			// 롤백
			_ = s.dbiStore.UpdateStatus(ctx, instance.ID, instance.Status, "K8s update failed")
			return errors.Wrap(err)
		}
	}

	s.logger.Printf("인스턴스 %s 중지됨", instanceID)
	return nil
}

func (s *service) RestartInstance(ctx context.Context, userID, instanceID string) error {
	if err := s.StopInstance(ctx, userID, instanceID); err != nil {
		return errors.Wrap(err)
	}

	time.Sleep(2 * time.Second)

	if err := s.StartInstance(ctx, userID, instanceID); err != nil {
		return errors.Wrap(err)
	}

	return nil
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

var _ dbservice.Service = (*service)(nil)

func NewService(
	publicDBHost string,
	dbiStore dbservice.DBInstanceStore,
	presetStore dbservice.PresetStore,
	lemonService lemon.Service,
	userStore user.Store,
	k8sClient k8s.Client,
	portStore dbservice.PortStore,
	resourceManager resource.Manager,
	logger *log.Logger,
) dbservice.Service {
	return &service{
		publicDBHost:    publicDBHost,
		dbiStore:        dbiStore,
		presetStore:     presetStore,
		lemonService:    lemonService,
		userStore:       userStore,
		k8sClient:       k8sClient,
		portStore:       portStore,
		resourceManager: resourceManager,
		logger:          logger,
	}
}

func (s *service) CreateInstance(ctx context.Context, userID string, userLemon int, req *dbservice.CreateInstanceRequest) (*dbservice.CreateInstanceResponse, error) {
	// 최대 보유 가능 개수 추가 확인
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
		if !preset.Available {
			return nil, errors.NewInvalidParameterError("preset", preset.UnavailableReason)
		}

		// 리소스 체크
		sysResource := resource.SystemResourceSpec{
			CPU:    preset.Resources.CPU,
			Memory: preset.Resources.Memory,
		}
		canAllocate, reason, err := s.resourceManager.CanAllocate(ctx, sysResource)
		if err != nil {
			return nil, errors.Wrap(err)
		}
		if !canAllocate {
			return nil, errors.NewSystemCapacityError(reason)
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

		// 리소스 체크
		sysResource := resource.SystemResourceSpec{
			CPU:    req.Resources.CPU,
			Memory: req.Resources.Memory,
		}
		canAllocate, reason, err := s.resourceManager.CanAllocate(ctx, sysResource)
		if err != nil {
			return nil, errors.Wrap(err)
		}
		if !canAllocate {
			return nil, errors.NewSystemCapacityError(reason)
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

	// 1. 먼저 DB에 인스턴스 저장 (ID 생성됨)
	if err := s.dbiStore.Create(ctx, instance); err != nil {
		return nil, errors.Wrap(err)
	}

	// 환불 보장을 위한 defer
	lemonDeducted := false
	portAllocated := false
	defer func() {
		if err != nil {
			if lemonDeducted {
				refundErr := s.lemonService.AddLemons(ctx, userID, instance.Cost.CreationCost,
					lemon.ActionInstanceCreateRefund, fmt.Sprintf("실패: %v", err), &instance.ID)
				if refundErr != nil {
					s.logger.Printf("CRITICAL: 환불 실패 - userID: %s, instanceID: %d, amount: %d, error: %v",
						userID, instance.ID, instance.Cost.CreationCost, refundErr)
					// TODO: 환불 실패 기록 테이블에 저장
				}
			}
			if portAllocated {
				_ = s.portStore.ReleasePort(ctx, instance.ExternalID)
			}
		}
	}()

	// 2. 레몬 차감
	if err := s.lemonService.DeductLemons(ctx, userID, instance.Cost.CreationCost,
		lemon.ActionInstanceCreate, fmt.Sprintf("인스턴스 %s 생성", instance.Name), &instance.ID); err != nil {
		_ = s.dbiStore.Delete(ctx, instance.ExternalID)
		return nil, errors.Wrap(err)
	}
	lemonDeducted = true

	// 3. 포트 할당 (K8s 리소스 생성 전에!)
	if s.portStore != nil {
		port, err := s.portStore.AllocatePort(ctx, instance.ExternalID)
		if err != nil {
			s.logger.Printf("WARNING: 외부 포트 할당 실패: %v", err)
		} else {
			s.logger.Printf("DEBUG: Port %d allocated successfully", port)
			instance.ExternalPort = port
			portAllocated = true

			// DB 업데이트
			if err := s.dbiStore.Update(ctx, instance); err != nil {
				s.logger.Printf("Failed to update port info in DB: %v", err)
			}
		}
	}

	// 4. K8s 리소스 생성 (이제 instance.ExternalPort가 설정된 상태)
	secretData, err := s.provisionK8sResources(ctx, instance)
	if err != nil {
		_ = s.dbiStore.UpdateStatus(ctx, instance.ID, dbservice.StatusError, "K8s provisioning failed")
		return nil, errors.Wrapf(err, "failed to provision k8s resources")
	}
	username, password := string(secretData["username"]), string(secretData["password"])

	// 5. 외부 접근 설정
	credentials := &dbservice.Credentials{
		Username: username,
		Password: password,
	}

	if instance.ExternalPort > 0 {
		credentials.ExternalHost = s.publicDBHost
		credentials.ExternalPort = instance.ExternalPort
		if instance.Type == dbservice.MongoDB {
			credentials.ExternalURI = fmt.Sprintf("mongodb://%s:%s@%s:%d/%s?authSource=admin",
				username, password, s.publicDBHost, instance.ExternalPort, req.Name)
		} else {
			credentials.ExternalURI = fmt.Sprintf("%s://%s:%s@%s:%d/%s",
				instance.Type, username, password,
				s.publicDBHost, instance.ExternalPort, req.Name)
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

	if err := s.dbiStore.Update(ctx, instance); err != nil {
		s.logger.Printf("Failed to update K8s info in DB: %v", err)
		// 실패해도 계속 진행 (이미 K8s 리소스는 생성됨)
	}

	return secretData, nil
}

func (s *service) createDBInstanceCRD(ctx context.Context, instance *dbservice.DBInstance) error {
	s.logger.Printf("DEBUG: Creating CRD - instance.ExternalPort: %d", instance.ExternalPort)

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
		Config:       instance.Config,
		ExternalPort: int32(instance.ExternalPort),
	}

	s.logger.Printf("DEBUG: DBInstanceParams.ExternalPort: %d", params.ExternalPort)

	// CRD 생성
	spec := k8s.BuildDBInstanceSpec(params)
	s.logger.Printf("DEBUG: spec map: %+v", spec)

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
		"cpu":    k8s.ConvertCPUToString(instance.Resources.CPU),
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
