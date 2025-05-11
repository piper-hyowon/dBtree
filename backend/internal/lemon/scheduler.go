package lemon

import (
	"context"
	"github.com/piper-hyowon/dBtree/internal/core/errors"
	"github.com/piper-hyowon/dBtree/internal/core/lemon"
	"github.com/piper-hyowon/dBtree/internal/core/quiz"
	"log"
	"sync"
	"time"
)

type schedulerService struct {
	lemonStore    lemon.Store
	quizStore     quiz.Store
	logger        *log.Logger
	ticker        *time.Ticker
	done          chan bool
	mutex         sync.Mutex
	isRunning     bool
	checkInterval time.Duration
}

func NewScheduler(lemonStore lemon.Store, quizStore quiz.Store, logger *log.Logger, checkInterval time.Duration) lemon.Scheduler {
	if checkInterval <= 0 {
		checkInterval = 1 * time.Minute
	}

	return &schedulerService{
		lemonStore:    lemonStore,
		quizStore:     quizStore,
		logger:        logger,
		done:          make(chan bool),
		isRunning:     false,
		checkInterval: checkInterval,
	}
}

func (s *schedulerService) Start() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.isRunning {
		s.logger.Println("레몬 재생성 스케줄러가 이미 실행 중입니다")
		return nil
	}

	s.ticker = time.NewTicker(s.checkInterval)
	s.done = make(chan bool)
	s.isRunning = true

	go s.run()
	s.logger.Println("레몬 재생성 스케줄러가 시작되었습니다")
	return nil
}

func (s *schedulerService) Stop() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.isRunning {
		s.logger.Println("레몬 재생성 스케줄러가 이미 중지됨")
		return nil
	}

	s.ticker.Stop()
	s.done <- true
	s.isRunning = false
	s.logger.Println("레몬 재생성 스케줄러가 중지되었습니다")
	return nil
}

func (s *schedulerService) IsRunning() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.isRunning
}

// RunNow 레몬 재생성 메서드 즉시 실행
func (s *schedulerService) RunNow(ctx context.Context) ([]int, error) {
	now := time.Now()
	return s.lemonStore.RegrowLemons(ctx, now)
}

func (s *schedulerService) run() {
	s.regrowLemons()

	for {
		select {
		case <-s.ticker.C:
			s.regrowLemons()
		case <-s.done:
			return
		}
	}
}

func (s *schedulerService) regrowLemons() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	now := time.Now()
	positionIDs, err := s.lemonStore.RegrowLemons(ctx, now)
	if err != nil {
		s.logger.Printf("레몬 재생성 중 오류 발생: %v\n", err)
		return
	}

	// 재생성된 각 레몬에 새 퀴즈 할당
	for _, posID := range positionIDs {
		// 랜덤 퀴즈 가져오기
		randomQuiz, err := s.quizStore.Random(ctx)
		if err != nil {
			s.logger.Printf("레몬 위치 %d에 대한 랜덤 퀴즈 가져오기 실패: %v\n", posID, err)
			continue
		}

		// 퀴즈를 레몬에 할당
		err = s.quizStore.AssignToLemon(ctx, randomQuiz.ID, posID)
		if err != nil {
			s.logger.Printf("레몬 위치 %d에 퀴즈 할당 실패: %v\n", posID, err)
			continue
		}

		s.logger.Printf("레몬 위치 %d에 새로운 퀴즈(ID: %d) 할당 완료\n", posID, randomQuiz.ID)
	}

	if len(positionIDs) > 0 {
		s.logger.Printf("%d개의 레몬 재생성, 퀴즈 할당 완료\n", len(positionIDs))
	}
}

func (s *schedulerService) InitializeLemons(ctx context.Context) error {
	// 1. 사용 가능한 레몬 위치 가져오기
	availablePositions, err := s.lemonStore.AvailablePositions(ctx)
	if err != nil {
		return err
	}

	if len(availablePositions) == 0 {
		s.logger.Println("사용 가능한 레몬이 없습니다. 초기 데이터를 확인하세요.")
		return nil
	}

	// 2. 각 사용 가능한 레몬 위치에 대해:
	for _, posID := range availablePositions {
		// 현재 퀴즈가 할당되어 있는지 확인
		_, err := s.quizStore.ByPositionID(ctx, posID)

		// 에러가 없으면 이미 퀴즈가 할당된 것
		if err == nil {
			s.logger.Printf("레몬 위치 %d에 이미 퀴즈가 할당되어 있습니다.\n", posID)
			continue
		}

		var domainErr errors.DomainError
		if !errors.As(err, &domainErr) {
			s.logger.Printf("레몬 위치 %d의 퀴즈 확인 중 일반 에러 발생: %v\n", posID, err)
			continue
		}

		if domainErr.Code() != errors.ErrResourceNotFound {
			// ResourceNotFound가 아닌 다른 도메인 에러
			s.logger.Printf("레몬 위치 %d의 퀴즈 확인 중 도메인 에러 발생: %v\n", posID, err)
			continue
		}

		// 여기는 ResourceNotFound 에러인 경우만 있음! (퀴즈가 없는 경우)

		// 퀴즈가 없으면 새로 할당
		randomQuiz, err := s.quizStore.Random(ctx)
		if err != nil {
			s.logger.Printf("레몬 위치 %d에 대한 랜덤 퀴즈 가져오기 실패: %v\n", posID, err)
			continue
		}

		err = s.quizStore.AssignToLemon(ctx, randomQuiz.ID, posID)
		if err != nil {
			s.logger.Printf("레몬 위치 %d에 퀴즈 할당 실패: %v\n", posID, err)
			continue
		}

		s.logger.Printf("레몬 위치 %d에 새로운 퀴즈(ID: %d) 초기 할당 완료\n", posID, randomQuiz.ID)
	}

	s.logger.Printf("레몬 초기화 완료: %d개의 사용 가능한 레몬 확인됨\n", len(availablePositions))
	return nil
}
