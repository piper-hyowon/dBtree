package user

import (
	"context"
	"github.com/piper-hyowon/dBtree/internal/core/auth"
	"github.com/piper-hyowon/dBtree/internal/core/email"
	"github.com/piper-hyowon/dBtree/internal/core/errors"
	"github.com/piper-hyowon/dBtree/internal/core/user"
	"github.com/piper-hyowon/dBtree/internal/platform/rest"
	"log"
	"runtime/debug"
)

type service struct {
	emailService email.Service
	userStore    user.Store
	sessionStore auth.SessionStore
	logger       *log.Logger
}

var _ user.Service = (*service)(nil)

func NewService(
	emailService email.Service,
	userStore user.Store,
	sessionStore auth.SessionStore,
	logger *log.Logger,
) user.Service {
	return &service{
		emailService: emailService,
		userStore:    userStore,
		sessionStore: sessionStore,
		logger:       logger,
	}
}

func (s *service) Delete(ctx context.Context, userID string, userEmail string) error {
	// 세션 1시간마다 정리되므로 세션 삭제 생략 (config.SessionConfig.CleanupIntervalHours)

	// 유저 테이블에서 소프트 딜리트
	err := s.userStore.Delete(ctx, userID)
	if err != nil {
		return errors.NewInternalErrorWithStack(err, string(debug.Stack()))
	}

	// 유저 남은 세션 삭제
	err = s.sessionStore.Delete(ctx, userEmail)
	if err != nil {
		s.logger.Printf("사용자 탈퇴 중 세션 삭제 실패: %v", err)
	}

	// 탈퇴 이메일 발송
	u := rest.GetUserFromContext(ctx)
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
