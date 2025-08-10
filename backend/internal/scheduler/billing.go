package scheduler

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/piper-hyowon/dBtree/internal/core/dbservice"
	"github.com/piper-hyowon/dBtree/internal/core/errors"
	"github.com/piper-hyowon/dBtree/internal/core/lemon"
	"github.com/piper-hyowon/dBtree/internal/platform/k8s"
)

type BillingScheduler struct {
	dbiStore     dbservice.DBInstanceStore
	lemonStore   lemon.Store
	lemonService lemon.Service
	k8sClient    k8s.Client
	logger       *log.Logger

	ticker    *time.Ticker
	done      chan bool
	mutex     sync.Mutex
	isRunning bool
	interval  time.Duration
}

var _ ManualRunScheduler = (*BillingScheduler)(nil)

func NewBillingScheduler(
	dbiStore dbservice.DBInstanceStore,
	lemonStore lemon.Store,
	lemonService lemon.Service,
	k8sClient k8s.Client,
	logger *log.Logger,
	interval time.Duration,
) *BillingScheduler {
	if interval <= 0 {
		interval = 1 * time.Hour // 기본값: 1시간
	}

	return &BillingScheduler{
		dbiStore:     dbiStore,
		lemonStore:   lemonStore,
		lemonService: lemonService,
		k8sClient:    k8sClient,
		logger:       logger,
		interval:     interval,
		done:         make(chan bool),
	}
}

func (s *BillingScheduler) Start() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.isRunning {
		s.logger.Println("과금 스케줄러가 이미 실행 중입니다")
		return nil
	}

	s.ticker = time.NewTicker(s.interval)
	s.done = make(chan bool)
	s.isRunning = true

	go s.run()
	s.logger.Println("과금 스케줄러가 시작되었습니다")
	return nil
}

func (s *BillingScheduler) Stop() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.isRunning {
		s.logger.Println("과금 스케줄러가 이미 중지됨")
		return nil
	}

	s.ticker.Stop()
	s.done <- true
	s.isRunning = false
	s.logger.Println("과금 스케줄러가 중지되었습니다")
	return nil
}

func (s *BillingScheduler) IsRunning() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.isRunning
}

// RunNow 과금 처리 즉시 실행 (테스트/관리용)
func (s *BillingScheduler) RunNow(ctx context.Context) error {
	s.processBilling()
	return nil
}

func (s *BillingScheduler) run() {
	// 시작할 때 한번 실행
	s.processBilling()

	for {
		select {
		case <-s.ticker.C:
			s.processBilling()
		case <-s.done:
			return
		}
	}
}

func (s *BillingScheduler) processBilling() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	s.logger.Println("과금 처리 시작")

	// 1. 실행 중인 인스턴스 과금 처리
	instances, err := s.dbiStore.ListRunning(ctx)
	if err != nil {
		s.logger.Printf("과금 대상 인스턴스 조회 실패: %v", err)
		return
	}

	s.logger.Printf("실행 중인 인스턴스: %d개", len(instances))

	successCount := 0
	failCount := 0

	for _, instance := range instances {
		if err := s.processInstanceBilling(ctx, instance); err != nil {
			s.logger.Printf("인스턴스 %s 과금 처리 실패: %v", instance.ExternalID, err)
			failCount++
			continue
		}
		successCount++
	}

	s.logger.Printf("과금 처리 완료 - 성공: %d, 실패: %d", successCount, failCount)

	// 2. Paused 상태 인스턴스 확인
	s.checkPausedInstances(ctx)
}

func (s *BillingScheduler) processInstanceBilling(ctx context.Context, instance *dbservice.DBInstance) error {
	// 시간당 비용 계산

	hourlyCost := instance.Cost.HourlyLemons
	if hourlyCost == 0 {
		s.logger.Printf("인스턴스 %s는 무료 인스턴스", instance.ExternalID)
		return nil
	}

	// 사용자 잔액 확인
	userBalance, err := s.lemonStore.UserBalance(ctx, instance.UserID)
	if err != nil {
		return errors.Wrap(err)
	}

	// 잔액이 부족한 경우 즉시 처리 (중복 과금 체크 무시)
	if userBalance < hourlyCost {
		s.logger.Printf("인스턴스 %s: 잔액 부족 (%d < %d), 즉시 일시정지 처리",
			instance.ExternalID, userBalance, hourlyCost)
		return s.pauseInstanceDueToInsufficientFunds(ctx, instance)
	}

	// 마지막 과금 시간 확인 (잔액이 충분한 경우에만)
	if instance.LastBilledAt != nil {
		timeSinceLastBilling := time.Since(*instance.LastBilledAt)
		if timeSinceLastBilling < 50*time.Minute { // 50분 미만이면 스킵
			s.logger.Printf("인스턴스 %s는 최근에 과금됨 (%v 전)",
				instance.ExternalID, timeSinceLastBilling.Round(time.Minute))
			return nil
		}
	}

	// 레몬 차감 시도
	err = s.lemonService.ProcessInstanceFee(
		ctx,
		instance.UserID,
		instance.ExternalID,
		hourlyCost,
		lemon.ActionInstanceMaintain,
		&instance.ID,
	)

	if err != nil {
		// 잔액 부족 에러인지 확인
		var domainErr errors.DomainError
		if errors.As(err, &domainErr) && domainErr.Code() == errors.ErrInsufficientLemons {
			s.logger.Printf("인스턴스 %s: 레몬 부족으로 일시정지 처리", instance.ExternalID)

			// 인스턴스 일시정지
			now := time.Now()
			updateErr := s.dbiStore.UpdateStatus(
				ctx,
				instance.ID,
				dbservice.StatusPaused,
				"레몬 부족으로 일시정지됨",
			)

			if updateErr != nil {
				return errors.Wrapf(updateErr, "상태 업데이트 실패")
			}

			// K8s CRD 상태도 업데이트
			if instance.K8sNamespace != "" && instance.K8sResourceName != "" {
				if err := s.k8sClient.PatchDBInstanceStatus(
					ctx,
					instance.K8sNamespace,
					instance.K8sResourceName,
					string(dbservice.StatusPaused),
					"Insufficient lemons for billing",
				); err != nil {
					s.logger.Printf("K8s 상태 업데이트 실패: %v", err)
					// K8s 업데이트 실패는 치명적이지 않으므로 계속 진행
				}
			}

			// paused_at 시간도 업데이트
			instance.PausedAt = &now
			if err := s.dbiStore.Update(ctx, instance); err != nil {
				s.logger.Printf("paused_at 업데이트 실패: %v", err)
			}

			// TODO: 이메일 알림
			// s.emailService.SendInsufficientLemonsNotification(ctx, instance.UserID, instance.Name)

			return nil
		}
		return err
	}

	// 과금 시간 업데이트
	if err := s.dbiStore.UpdateBillingTime(ctx, instance.ID, time.Now()); err != nil {
		return errors.Wrapf(err, "과금 시간 업데이트 실패")
	}

	s.logger.Printf("인스턴스 %s 과금 성공: %d 레몬", instance.ExternalID, hourlyCost)
	return nil
}

func (s *BillingScheduler) checkPausedInstances(ctx context.Context) {
	// 모든 Paused 상태인 인스턴스 조회
	pausedInstances, err := s.dbiStore.ListByStatus(ctx, dbservice.StatusPaused)
	if err != nil {
		s.logger.Printf("일시정지 인스턴스 조회 실패: %v", err)
		return
	}

	if len(pausedInstances) == 0 {
		s.logger.Println("일시정지된 인스턴스 없음")
		return
	}

	s.logger.Printf("일시정지된 인스턴스: %d개", len(pausedInstances))

	// 1시간 이상 경과한 것만 처리
	//oneHourAgo := time.Now().Add(-1 * time.Hour)
	oneHourAgo := time.Now().Add(-2 * time.Minute) // TODO:
	now := time.Now()

	for _, instance := range pausedInstances {
		// pausedAt이 nil이면 스킵 (방어 코드)
		if instance.PausedAt == nil {
			s.logger.Printf("인스턴스 %s의 pausedAt이 nil입니다", instance.ExternalID)
			continue
		}

		// 디버깅을 위한 상세 로그
		elapsedTime := now.Sub(*instance.PausedAt)
		s.logger.Printf("인스턴스 %s - pausedAt: %v, 현재: %v, 경과: %v",
			instance.ExternalID,
			instance.PausedAt.Format("15:04:05"),
			now.Format("15:04:05"),
			elapsedTime.Round(time.Second))

		// 1시간 이상 경과했는지 확인
		if instance.PausedAt.Before(oneHourAgo) {
			s.logger.Printf("인스턴스 %s: 1시간 이상 경과, 삭제 처리 시작", instance.ExternalID)
			s.handlePausedInstance(ctx, instance)
		} else {
			remainingTime := time.Hour - elapsedTime
			s.logger.Printf("인스턴스 %s: 아직 %v 남음", instance.ExternalID, remainingTime.Round(time.Second))
		}
	}
}

func (s *BillingScheduler) handlePausedInstance(ctx context.Context, instance *dbservice.DBInstance) {
	// 사용자의 현재 레몬 확인
	userBalance, err := s.lemonStore.UserBalance(ctx, instance.UserID)
	if err != nil {
		s.logger.Printf("사용자 %s 잔액 조회 실패: %v", instance.UserID, err)
		return
	}

	hourlyCost := instance.Cost.HourlyLemons
	s.logger.Printf("인스턴스 %s: 잔액=%d, 시간당비용=%d", instance.ExternalID, userBalance, hourlyCost)

	// 여전히 잔액이 부족한 경우
	if userBalance < hourlyCost {
		s.logger.Printf("인스턴스 %s: 1시간 경과 후에도 잔액 부족 (%d < %d), 삭제 처리",
			instance.ExternalID, userBalance, hourlyCost)

		// 인스턴스 삭제 상태로 변경
		if err := s.dbiStore.UpdateStatus(
			ctx,
			instance.ID,
			dbservice.StatusDeleting,
			"1시간 이상 레몬 부족으로 자동 삭제",
		); err != nil {
			s.logger.Printf("삭제 상태 변경 실패: %v", err)
			return
		}

		// TODO: 삭제 됐다는 알림?
		// s.emailService.SendInstanceDeletionNotification(ctx, instance.UserID, instance.Name)

		// 실제 삭제는 dbservice가 처리하거나 별도 cleanup 프로세스에서 처리
		err = s.k8sClient.DeleteDBInstance(ctx, instance.K8sNamespace, instance.K8sResourceName)
		if err != nil {
			// TODO: ?
		}
		return
	}

	// 잔액이 충분해진 경우 - 다시 Running으로 전환
	s.logger.Printf("인스턴스 %s: 잔액 충전됨, 재시작 시도", instance.ExternalID)

	// 먼저 과금 처리
	if err := s.lemonService.ProcessInstanceFee(
		ctx,
		instance.UserID,
		instance.ExternalID,
		hourlyCost,
		lemon.ActionInstanceMaintain,
		&instance.ID,
	); err != nil {
		s.logger.Printf("재시작 과금 실패: %v", err)
		return
	}

	// 상태를 Running으로 변경
	if err := s.dbiStore.UpdateStatus(
		ctx,
		instance.ID,
		dbservice.StatusRunning,
		"레몬 충전으로 재시작됨",
	); err != nil {
		s.logger.Printf("재시작 상태 변경 실패: %v", err)
		// TODO: 과금은 됐는데 상태 변경 실패한 경우 처리
		return
	}

	// 과금 시간 업데이트
	if err := s.dbiStore.UpdateBillingTime(ctx, instance.ID, time.Now()); err != nil {
		s.logger.Printf("과금 시간 업데이트 실패: %v", err)
	}

	s.logger.Printf("인스턴스 %s 재시작 완료", instance.ExternalID)
}

// pauseInstanceDueToInsufficientFunds 레몬 부족으로 인스턴스 일시정지
func (s *BillingScheduler) pauseInstanceDueToInsufficientFunds(ctx context.Context, instance *dbservice.DBInstance) error {
	s.logger.Printf("인스턴스 %s: 레몬 부족으로 일시정지 처리", instance.ExternalID)

	// 인스턴스 일시정지
	now := time.Now()

	// DB 상태 업데이트
	updateErr := s.dbiStore.UpdateStatus(
		ctx,
		instance.ID,
		dbservice.StatusPaused,
		"레몬 부족으로 일시정지됨",
	)

	if updateErr != nil {
		return errors.Wrapf(updateErr, "상태 업데이트 실패")
	}

	// paused_at 시간도 업데이트
	instance.PausedAt = &now
	instance.Status = dbservice.StatusPaused
	if err := s.dbiStore.Update(ctx, instance); err != nil {
		s.logger.Printf("paused_at 업데이트 실패: %v", err)
	}

	// K8s CRD 상태도 업데이트
	if instance.K8sNamespace != "" && instance.K8sResourceName != "" {
		if err := s.k8sClient.PatchDBInstanceStatus(
			ctx,
			instance.K8sNamespace,
			instance.K8sResourceName,
			string(dbservice.StatusPaused),
			"Insufficient lemons for billing",
		); err != nil {
			s.logger.Printf("K8s 상태 업데이트 실패: %v", err)
		}
	}

	return nil
}
