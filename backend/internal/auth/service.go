package auth

import (
	"context"
	"crypto/rand"
	"github.com/piper-hyowon/dBtree/internal/core/auth"
	"github.com/piper-hyowon/dBtree/internal/core/email"
	"github.com/piper-hyowon/dBtree/internal/core/errors"
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
	isNewUser := u == nil
	if err != nil {
		return isNewUser, errors.Wrap(err)
	}

	code, err := generateOTP(auth.OTPLength)

	if err != nil {
		return isNewUser, errors.Wrap(err)
	}

	otp := auth.NewOTP(email, code, auth.OTPExpirationMinutes)

	session, err := s.sessionStore.FindByEmail(ctx, email)
	if err != nil {
		return isNewUser, errors.Wrap(err)
	}
	if session == nil {
		session = auth.NewSession(email, otp)
	} else {
		session.OTP = otp
		session.Status = auth.Pending
		session.UpdatedAt = time.Now().UTC()
	}

	if err := s.sessionStore.Save(ctx, session); err != nil {
		return isNewUser, errors.Wrap(err)
	}

	if err := s.emailService.SendOTP(ctx, email, code); err != nil {
		if isEmailDeliveryError(err) {
			return isNewUser, errors.NewInvalidEmailError(err.Error())
		}
		return isNewUser, errors.Wrap(err)
	}

	return isNewUser, nil
}

func (s *service) GetSession(ctx context.Context, email string) (*auth.Session, error) {
	return s.sessionStore.FindByEmail(ctx, email)
}

func (s *service) ResendOTP(ctx context.Context, email string) error {
	session, err := s.sessionStore.FindByEmail(ctx, email)
	if err != nil {
		return err
	}
	if session == nil {
		return errors.NewSessionNotFoundError()
	}

	now := time.Now().UTC()

	// 인증된 세션이고 토큰이 만료된 경우, 상태를 Pending 으로 변경
	if session.Status == auth.Verified {
		// 토큰 유효 여부와 관계없이 재인증을 위한 OTP 발송 허용
		session.Status = auth.Pending
		session.Token = ""      // 기존 토큰 무효화
		session.ResendCount = 0 // 재발송 횟수 초기화
	}

	// 재전송 횟수 제한
	if session.ResendCount >= auth.MaxResendAttempts-1 {
		return errors.NewTooManyResendsError(auth.MaxResendAttempts)
	}

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
		return errors.NewTooEarlyResendError(auth.ResendWaitSeconds)
	}

	otp, err := generateOTP(auth.OTPLength)
	if err != nil {
		return errors.Wrap(err)
	}

	session.OTP = auth.NewOTP(email, otp, auth.OTPExpirationMinutes)
	session.ResendCount++
	session.LastResendAt = &now
	session.UpdatedAt = now

	if err := s.sessionStore.Save(ctx, session); err != nil {
		return errors.Wrap(err)
	}

	if err := s.emailService.SendOTP(ctx, email, otp); err != nil {
		return errors.Wrap(err)
	}

	return nil
}

func (s *service) VerifyOTP(ctx context.Context, email string, code string) (*user.User, string, error) {
	if len(code) != auth.OTPLength {
		return nil, "", errors.NewInvalidOTPError()
	}

	session, err := s.sessionStore.FindByEmail(ctx, email)
	if err != nil {
		return nil, "", errors.NewSessionNotFoundError()
	}

	// 이미 인증된 세션이 있으면 기존 토큰 반환
	if session.Status == auth.Verified {
		// 토큰이 만료시 재인증 요구
		now := time.Now().UTC()
		if session.TokenExpiresAt.Before(now) {
			return nil, "", errors.NewSessionExpiredError()
		}

		if session.OTP != nil {
			// 이미 인증된 OTP 로 중복 인증 시도
			if session.OTP.Code == code && !now.After(session.OTP.ExpiresAt) {
				// OTP를 재사용 방지
				session.OTP = nil
				session.UpdatedAt = now

				if err := s.sessionStore.Save(ctx, session); err != nil {
					return nil, "", errors.Wrap(err)
				}

				// 기존 토큰 반환
				u, err := s.userStore.FindByEmail(ctx, email)
				if err != nil {
					return nil, "", errors.Wrap(err)
				}
				return u, session.Token, nil
			}

			// OTP 불일치 or 만료
			return nil, "", errors.NewInvalidOTPError()
		}

		return nil, "", errors.NewSessionAlreadyVerifiedError()

	}

	if session.OTP == nil {
		return nil, "", errors.NewInvalidOTPError()
	}

	now := time.Now().UTC()

	if now.After(session.OTP.ExpiresAt) {
		return nil, "", errors.NewExpiredOTPError()
	}

	if session.OTP.Code != code {
		// TODO: 실패 횟수 기록 -> 유저 블락 처리?
		return nil, "", errors.NewInvalidOTPError()
	}

	token, err := crypto.GenerateRandomToken(32)
	if err != nil {
		return nil, "", errors.Wrap(err)
	}

	session.Token = token
	session.TokenExpiresAt = now.Add(time.Hour * auth.TokenExpirationHours)
	session.Status = auth.Verified
	session.UpdatedAt = now
	session.OTP = nil

	if err := s.sessionStore.Save(ctx, session); err != nil {
		return nil, "", errors.Wrap(err)
	}

	u, err := s.userStore.FindByEmail(ctx, email)
	if err != nil || u == nil {
		if err := s.userStore.CreateIfNotExists(ctx, email); err != nil {
			return nil, "", errors.Wrap(err)
		}

		u, err = s.userStore.FindByEmail(ctx, email)
		if err != nil {
			return nil, "", errors.Wrap(err)
		}

		_ = s.emailService.SendWelcome(ctx, email)
	}
	return u, token, nil
}

func (s *service) ValidateSession(ctx context.Context, token string) (*user.User, error) {
	if token == "" {
		return nil, errors.NewInvalidTokenError()
	}

	session, err := s.sessionStore.FindByToken(ctx, token)
	if err != nil {
		return nil, errors.NewInternalError(err)
	}

	if session == nil {
		return nil, errors.NewSessionNotFoundError()
	}

	if session.Status != auth.Verified {
		return nil, errors.NewUnauthorizedError()
	}

	u, err := s.userStore.FindByEmail(ctx, session.Email)
	if err != nil || u == nil {
		return nil, errors.Wrap(err)
	}

	return u, nil
}

func (s *service) Logout(ctx context.Context, token string) error {
	session, err := s.sessionStore.FindByToken(ctx, token)
	if err != nil {
		return errors.Wrap(err)
	}

	if session == nil {
		return nil
	}

	e := session.Email

	err = s.sessionStore.Delete(ctx, e)
	if err != nil {
		return errors.Wrap(err)
	}

	return nil
}

func generateOTP(length int) (string, error) {
	const otpChars = "0123456789"

	result := make([]byte, length)
	randomBytes := make([]byte, length)

	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", errors.Wrap(err)
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
