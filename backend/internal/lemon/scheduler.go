package lemon

import (
	"context"
	"github.com/piper-hyowon/dBtree/internal/core/lemon"
	"log"
	"sync"
	"time"
)

type schedulerService struct {
	store         lemon.Store
	logger        *log.Logger
	ticker        *time.Ticker
	done          chan bool
	mutex         sync.Mutex
	isRunning     bool
	checkInterval time.Duration
}

func NewScheduler(store lemon.Store, logger *log.Logger, checkInterval time.Duration) lemon.Scheduler {
	if checkInterval <= 0 {
		checkInterval = 1 * time.Minute
	}

	return &schedulerService{
		store:         store,
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
func (s *schedulerService) RunNow(ctx context.Context) (int, error) {
	now := time.Now()
	return s.store.RegrowLemons(ctx, now)
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
	regrown, err := s.store.RegrowLemons(ctx, now)
	if err != nil {
		s.logger.Printf("레몬 재생성 중 오류 발생: %v\n", err)
		return
	}

	if regrown > 0 {
		s.logger.Printf("%d개의 레몬이 재생성되었습니다\n", regrown)
	}
}
