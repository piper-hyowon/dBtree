package user

import (
	"context"
	"fmt"
	"github.com/piper-hyowon/dBtree/internal/core"
	"github.com/piper-hyowon/dBtree/internal/core/email"
	"github.com/piper-hyowon/dBtree/internal/core/user"
	"github.com/piper-hyowon/dBtree/internal/platform/rest/middleware"
	"log"
)

type service struct {
	emailService email.Service
	userStore    user.Store
	logger       *log.Logger
}

var _ user.Service = (*service)(nil)

func NewService(
	emailService email.Service,
	userStore user.Store,
	logger *log.Logger,
) user.Service {
	return &service{
		emailService: emailService,
		userStore:    userStore,
		logger:       logger,
	}
}

func (s *service) Delete(ctx context.Context, userID string) error {
	if userID == "" {
		return core.ErrInvalidToken
	}

	// 세션 1시간마다 정리되므로 세션 삭제 생략 (config.SessionConfig.CleanupIntervalHours)

	// 유저 테이블에서 소프트 딜리트
	err := s.userStore.Delete(ctx, userID)
	if err != nil {
		return fmt.Errorf("%w: %v", core.ErrInternal, err)
	}

	// 탈퇴 이메일 발송
	u := middleware.GetUserFromContext(ctx)
	if u != nil && u.Email != "" {
		go func(email string) {
			sendCtx := context.Background()
			err := s.emailService.SendGoodbye(sendCtx, email)
			if err != nil {
				s.logger.Printf("탈퇴 확인 이메일 발송 실패: %v", err)
			}
		}(u.Email)
	}

	return nil
}
