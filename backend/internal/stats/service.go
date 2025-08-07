package stats

import (
	"context"
	"github.com/piper-hyowon/dBtree/internal/core/stats"
	"log"
	"sync"

	coredbservice "github.com/piper-hyowon/dBtree/internal/core/dbservice"
	corelemon "github.com/piper-hyowon/dBtree/internal/core/lemon"
	corequiz "github.com/piper-hyowon/dBtree/internal/core/quiz"
	coreuser "github.com/piper-hyowon/dBtree/internal/core/user"
)

type service struct {
	lemonStore corelemon.Store
	userStore  coreuser.Store
	dbStore    coredbservice.DBInstanceStore
	quizStore  corequiz.Store
	logger     *log.Logger
}

func NewService(
	lemonStore corelemon.Store,
	userStore coreuser.Store,
	dbStore coredbservice.DBInstanceStore,
	quizStore corequiz.Store,
	logger *log.Logger,
) stats.Service {
	return &service{
		lemonStore: lemonStore,
		userStore:  userStore,
		dbStore:    dbStore,
		quizStore:  quizStore,
		logger:     logger,
	}
}

func (s *service) GetGlobalStats(ctx context.Context) (*stats.GlobalStats, error) {
	type result struct {
		totalLemons    int
		totalInstances int
		totalUsers     int
		err            error
	}

	ch := make(chan result, 3)

	// 병렬로 쿼리 실행
	go func() {
		count, err := s.lemonStore.TotalHarvestedCount(ctx)
		ch <- result{totalLemons: count, err: err}
	}()

	go func() {
		count, err := s.dbStore.TotalCreated(ctx)
		ch <- result{totalInstances: count, err: err}
	}()

	go func() {
		count, err := s.userStore.TotalUserCount(ctx)
		ch <- result{totalUsers: count, err: err}
	}()

	var stats stats.GlobalStats
	for i := 0; i < 3; i++ {
		r := <-ch
		if r.err != nil {
			return nil, r.err
		}
		if r.totalLemons > 0 {
			stats.TotalHarvestedLemons = r.totalLemons
		}
		if r.totalInstances > 0 {
			stats.TotalCreatedInstances = r.totalInstances
		}
		if r.totalUsers > 0 {
			stats.TotalUsers = r.totalUsers
		}
	}

	return &stats, nil
}

func (s *service) GetMiniLeaderboard(ctx context.Context) (*stats.MiniLeaderboard, error) {
	var (
		lemonRich   []stats.UserRank
		quizMasters []stats.UserRank
		wg          sync.WaitGroup
		mu          sync.Mutex
		errs        []error
	)

	wg.Add(2)

	// 레몬 부자 TOP 3
	go func() {
		defer wg.Done()
		users, err := s.userStore.TopLemonHolders(ctx, 3)
		if err != nil {
			mu.Lock()
			errs = append(errs, err)
			mu.Unlock()
			return
		}

		for i, u := range users {
			lemonRich = append(lemonRich, stats.UserRank{
				MaskedEmail: maskEmailForLeaderboard(u.Email),
				Score:       u.LemonBalance,
				Rank:        i + 1,
			})
		}
	}()

	// 오늘의 퀴즈 마스터 TOP 3
	go func() {
		defer wg.Done()
		masters, err := s.quizStore.TodayQuizMasters(ctx, 3)
		if err != nil {
			mu.Lock()
			errs = append(errs, err)
			mu.Unlock()
			return
		}

		for i, m := range masters {
			quizMasters = append(quizMasters, stats.UserRank{
				MaskedEmail: maskEmailForLeaderboard(m.Email),
				Score:       m.CorrectCount,
				Rank:        i + 1,
			})
		}
	}()

	wg.Wait()

	if len(errs) > 0 {
		return nil, errs[0]
	}

	return &stats.MiniLeaderboard{
		LemonRichUsers: lemonRich,
		QuizMasters:    quizMasters,
	}, nil
}
