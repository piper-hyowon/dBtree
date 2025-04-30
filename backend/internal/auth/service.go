package auth

import (
	"context"
	"crypto/rand"
	"fmt"
	"github.com/piper-hyowon/dBtree/internal/core"
	"github.com/piper-hyowon/dBtree/internal/core/auth"
	"github.com/piper-hyowon/dBtree/internal/core/email"
	"github.com/piper-hyowon/dBtree/internal/core/user"
	"log"
	"strings"
	"time"

	"github.com/piper-hyowon/dBtree/internal/utils/crypto"
)

type service struct {
	sessionStore auth.SessionStore
	emailService email.Service
	userStore    user.Store
	logger       *log.Logger
}

// 컴파일 타임에 인터페이스 구현 체크
var _ auth.Service = (*service)(nil)

func NewService(
	sessionStore auth.SessionStore,
	emailService email.Service,
	userStore user.Store,
	logger *log.Logger,
) auth.Service {
	return &service{
		sessionStore: sessionStore,
		emailService: emailService,
		userStore:    userStore,
		logger:       logger,
	}
}

func (s *service) StartAuth(ctx context.Context, email string) (bool, error) {
	u, err := s.userStore.FindByEmail(ctx, email)
	isNewUser := err != nil || u == nil

	otpCode, err := generateOTP(auth.OTPLength)

	if err != nil {
		return isNewUser, fmt.Errorf("%w: %v", core.ErrInternal, err)
	}

	otp := auth.NewOTP(email, otpCode, auth.OTPExpirationMinutes)

	session, err := s.sessionStore.GetByEmail(ctx, email)
	if err != nil {
		session = auth.NewSession(email, otp)
	} else {
		session.OTP = otp
		session.Status = auth.Pending
		session.UpdatedAt = time.Now().UTC()
	}

	if err := s.sessionStore.Save(ctx, session); err != nil {
		return isNewUser, fmt.Errorf("%w: %v", core.ErrInternal, err)
	}

	if err := s.emailService.SendOTP(ctx, email, otpCode); err != nil {
		if isEmailDeliveryError(err) {
			return isNewUser, fmt.Errorf("%w: %v", core.ErrInvalidEmail, err)
		}
		return isNewUser, fmt.Errorf("%w: %v", core.ErrInternal, err)
	}

	return isNewUser, nil
}

func (s *service) GetSession(ctx context.Context, email string) (*auth.Session, error) {
	return s.sessionStore.GetByEmail(ctx, email)
}

func (s *service) ResendOTP(ctx context.Context, email string) error {
	session, err := s.sessionStore.GetByEmail(ctx, email)
	if err != nil {
		return core.ErrSessionNotFound
	}

	// 재전송 횟수 제한
	if session.ResendCount >= auth.MaxResendAttempts-1 {
		return core.ErrTooManyResends
	}

	now := time.Now().UTC()

	// 첫 재발송이면 CreatedAt(첫 발송시간)기준으로 체크
	var lastSentTime time.Time
	if session.LastResendAt != nil {
		lastSentTime = *session.LastResendAt
	} else if session.OTP != nil {
		lastSentTime = session.OTP.CreatedAt
	}

	waitTime := time.Duration(auth.ResendWaitSeconds) * time.Second
	nextResendTime := lastSentTime.Add(waitTime)
	if now.Before(nextResendTime) {
		return core.ErrTooEarlyResend
	}

	otpCode, err := generateOTP(auth.OTPLength)
	if err != nil {
		return fmt.Errorf("%w: %v", core.ErrInternal, err)
	}

	session.OTP = auth.NewOTP(email, otpCode, auth.OTPExpirationMinutes)
	session.ResendCount++
	session.LastResendAt = &now
	session.UpdatedAt = now

	if err := s.sessionStore.Save(ctx, session); err != nil {
		return fmt.Errorf("%w: %v", core.ErrInternal, err)
	}

	if err := s.emailService.SendOTP(ctx, email, otpCode); err != nil {
		return fmt.Errorf("%w: %v", core.ErrInternal, err)
	}

	return nil
}

func (s *service) VerifyOTP(ctx context.Context, email string, code string) (*user.User, string, error) {
	if code == "" || len(code) != auth.OTPLength {
		return nil, "", core.ErrInvalidOTP
	}

	session, err := s.sessionStore.GetByEmail(ctx, email)
	if err != nil {
		return nil, "", core.ErrSessionNotFound
	}

	// 이미 인증된 세션 -> 기존 토큰 반환
	if session.Status == auth.Verified {
		// 토큰이 만료시 재인증 요구
		now := time.Now().UTC()
		if session.TokenExpiresAt.Before(now) {
			return nil, "", core.ErrSessionExpired
		}

		if session.OTP != nil {
			// 이미 인증된 OTP 로 중복 인증 시도
			if session.OTP.Code == code && !now.After(session.OTP.ExpiresAt) {
				// OTP를 재사용 방지
				session.OTP = nil
				session.UpdatedAt = now

				if err := s.sessionStore.Save(ctx, session); err != nil {
					return nil, "", fmt.Errorf("%w: %v", core.ErrInternal, err)
				}

				// 기존 토큰 반환
				u, err := s.userStore.FindByEmail(ctx, email)
				if err != nil {
					return nil, "", fmt.Errorf("%w: %v", core.ErrInternal, err)
				}
				return u, session.Token, nil
			}

			// OTP 불일치 or 만료
			return nil, "", core.ErrInvalidOTP
		}

		return nil, "", core.ErrSessionAlreadyVerified

	}

	if session.OTP == nil {
		return nil, "", core.ErrInvalidOTP
	}

	now := time.Now().UTC()

	if now.After(session.OTP.ExpiresAt) {
		return nil, "", core.ErrExpiredOTP
	}

	if session.OTP.Code != code {
		// TODO: 실패 횟수 기록 -> 유저 블락 처리?
		return nil, "", core.ErrInvalidOTP
	}

	token, err := crypto.GenerateRandomToken(32)
	if err != nil {
		return nil, "", fmt.Errorf("%w", core.ErrInternal)
	}

	session.Token = token
	session.TokenExpiresAt = now.Add(time.Hour * auth.TokenExpirationHours)
	session.Status = auth.Verified
	session.UpdatedAt = now
	session.OTP = nil

	if err := s.sessionStore.Save(ctx, session); err != nil {
		return nil, "", fmt.Errorf("%w: %v", core.ErrInternal, err)
	}

	u, err := s.userStore.FindByEmail(ctx, email)
	if err != nil || u == nil {
		if err := s.userStore.Create(ctx, email); err != nil {
			return nil, "", fmt.Errorf("%w: %v", core.ErrInternal, err)
		}

		u, err = s.userStore.FindByEmail(ctx, email)
		if err != nil {
			return nil, "", fmt.Errorf("%w: %v", core.ErrInternal, err)
		}

		s.emailService.SendWelcome(ctx, email)
	}
	return u, token, nil
}

func (s *service) ValidateSession(ctx context.Context, token string) (*user.User, error) {
	if token == "" {
		return nil, core.ErrInvalidToken
	}

	session, err := s.sessionStore.GetByToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("세션검증실패: %w", err)
	}

	if session.Status != auth.Verified {
		return nil, core.ErrUnauthorized
	}

	u, err := s.userStore.FindByEmail(ctx, session.Email)
	if err != nil {
		return nil, fmt.Errorf("유저반환실패: %w", core.ErrInternal)
	}

	return u, nil
}

func (s *service) Logout(ctx context.Context, token string) error {
	if token == "" {
		return core.ErrInvalidToken
	}

	session, err := s.sessionStore.GetByToken(ctx, token)
	if err != nil {
		return core.ErrSessionNotFound
	}

	e := session.Email

	err = s.sessionStore.Delete(ctx, e)
	if err != nil {
		return fmt.Errorf("%w: %v", core.ErrInternal, err)
	}

	return nil
}

func generateOTP(length int) (string, error) {
	const otpChars = "0123456789"

	result := make([]byte, length)
	randomBytes := make([]byte, length)

	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	for i, b := range randomBytes {
		result[i] = otpChars[b%byte(len(otpChars))]
	}

	return string(result), nil
}

func isEmailDeliveryError(err error) bool {
	errorMsg := err.Error()

	emailErrorKeywords := []string{
		"Email address is not verified",
		"Message rejected",
		"Invalid recipient",
		"Unknown user",
		"Mailbox unavailable",
		"No such user",
		"Recipient address rejected",
		"not exist",
		"does not exist",
		"Invalid email",
	}

	for _, keyword := range emailErrorKeywords {
		if strings.Contains(errorMsg, keyword) {
			return true
		}
	}

	return false
}
