package core

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"time"

	"github.com/piper-hyowon/dBtree/internal/constants"
	"github.com/piper-hyowon/dBtree/internal/domain/errors"
	"github.com/piper-hyowon/dBtree/internal/domain/model"
	"github.com/piper-hyowon/dBtree/internal/domain/ports/secondary"
	"github.com/piper-hyowon/dBtree/internal/utils/crypto"
)

type AuthService struct {
	sessionRepo  secondary.SessionRepo
	emailService secondary.EmailService
	userRepo     secondary.UserRepo
	logger       *log.Logger
}

func NewAuthService(
	sessionRepo secondary.SessionRepo,
	emailService secondary.EmailService,
	userRepo secondary.UserRepo,
	logger *log.Logger,
) *AuthService {
	return &AuthService{
		sessionRepo:  sessionRepo,
		emailService: emailService,
		userRepo:     userRepo,
		logger:       logger,
	}
}

func (s *AuthService) StartAuth(ctx context.Context, email string) (bool, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	isNewUser := err != nil || user == nil

	otpCode, err := generateOTP(constants.OTPLength)
	if err != nil {
		return isNewUser, fmt.Errorf("%w: %v", errors.ErrInternal, err)
	}

	now := time.Now().UTC()
	expiresAt := now.Add(time.Minute * constants.OTPExpirationMinutes)

	otp := &model.OTP{
		Code:      otpCode,
		CreatedAt: now,
		ExpiresAt: expiresAt,
	}

	session, err := s.sessionRepo.GetByEmail(ctx, email)
	if err != nil {
		session = model.NewAuthSession(email, otp)
	} else {
		session.OTP = otp
		session.Status = model.AuthPending
		session.UpdatedAt = now
	}

	if err := s.sessionRepo.Save(ctx, session); err != nil {
		return isNewUser, fmt.Errorf("%w: %v", errors.ErrInternal, err)
	}

	if err := s.emailService.SendOTP(ctx, email, otpCode); err != nil {
		return isNewUser, fmt.Errorf("%w: %v", errors.ErrInternal, err)
	}

	return isNewUser, nil
}

func (s *AuthService) GetSession(ctx context.Context, email string) (*model.AuthSession, error) {
	return s.sessionRepo.GetByEmail(ctx, email)
}

func (s *AuthService) ResendOTP(ctx context.Context, email string) error {
	session, err := s.sessionRepo.GetByEmail(ctx, email)
	if err != nil {
		return errors.ErrSessionNotFound
	}

	// 재전송 횟수 제한
	if session.ResendCount >= constants.MaxResendAttempts-1 {
		return errors.ErrTooManyResends
	}

	now := time.Now().UTC()

	// 첫 재발송이면 CreatedAt(첫 발송시간)기준으로 체크
	var lastSentTime time.Time
	if session.LastResendAt != nil {
		lastSentTime = *session.LastResendAt
	} else if session.OTP != nil {
		lastSentTime = session.OTP.CreatedAt
	}

	waitTime := time.Duration(constants.ResendWaitSeconds) * time.Second
	nextResendTime := lastSentTime.Add(waitTime)
	if now.Before(nextResendTime) {
		return errors.ErrTooEarlyResend
	}

	otpCode, err := generateOTP(constants.OTPLength)
	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrInternal, err)
	}

	expiresAt := now.Add(time.Minute * constants.OTPExpirationMinutes)

	session.OTP = &model.OTP{
		Code:      otpCode,
		CreatedAt: now,
		ExpiresAt: expiresAt,
	}
	session.ResendCount++
	session.LastResendAt = &now
	session.UpdatedAt = now

	if err := s.sessionRepo.Save(ctx, session); err != nil {
		return fmt.Errorf("%w: %v", errors.ErrInternal, err)
	}

	if err := s.emailService.SendOTP(ctx, email, otpCode); err != nil {
		return fmt.Errorf("%w: %v", errors.ErrInternal, err)
	}

	return nil
}

func (s *AuthService) VerifyOTP(ctx context.Context, email string, code string) (*model.User, string, error) {
	if code == "" || len(code) != constants.OTPLength {
		return nil, "", errors.ErrInvalidOTP
	}

	session, err := s.sessionRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, "", errors.ErrSessionNotFound
	}

	if session.OTP == nil {
		return nil, "", errors.ErrInvalidOTP
	}

	now := time.Now().UTC()

	if now.After(session.OTP.ExpiresAt) {
		return nil, "", errors.ErrExpiredOTP
	}

	if session.OTP.Code != code {
		return nil, "", errors.ErrInvalidOTP
	}

	token, err := crypto.GenerateRandomToken(32)
	if err != nil {
		return nil, "", fmt.Errorf("%w", errors.ErrInternal)
	}

	session.Token = token
	session.TokenExpiresAt = now.Add(time.Hour * constants.TokenExpirationHours)
	session.Status = model.AuthVerified
	session.UpdatedAt = time.Now().UTC()

	if err := s.sessionRepo.Save(ctx, session); err != nil {
		return nil, "", fmt.Errorf("%w: %v", errors.ErrInternal, err)
	}

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil || user == nil {
		if err := s.userRepo.Create(ctx, email); err != nil {
			return nil, "", fmt.Errorf("%w: %v", errors.ErrInternal, err)
		}

		user, err = s.userRepo.FindByEmail(ctx, email)
		if err != nil {
			return nil, "", fmt.Errorf("%w: %v", errors.ErrInternal, err)
		}

		s.emailService.SendWelcome(ctx, email)
	}
	return user, token, nil
}

func (s *AuthService) ValidateSession(ctx context.Context, token string) (*model.User, error) {
	if token == "" {
		return nil, errors.ErrInvalidToken
	}

	session, err := s.sessionRepo.GetByToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("세션검증실패: %w", err)
	}

	if session.Status != model.AuthVerified {
		return nil, errors.ErrUnauthorized
	}

	user, err := s.userRepo.FindByEmail(ctx, session.Email)
	if err != nil {
		return nil, fmt.Errorf("유저반환실패: %w", errors.ErrInternal)
	}

	return user, nil
}

func (s *AuthService) Logout(ctx context.Context, token string) error {
	if token == "" {
		return errors.ErrInvalidToken
	}

	session, err := s.sessionRepo.GetByToken(ctx, token)
	if err != nil {
		return errors.ErrSessionNotFound
	}

	email := session.Email

	err = s.sessionRepo.Delete(ctx, email)
	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrInternal, err)
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
