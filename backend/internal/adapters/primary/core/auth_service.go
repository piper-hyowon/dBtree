package core

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/piper-hyowon/dBtree/internal/constants"
	"github.com/piper-hyowon/dBtree/internal/domain/errors"
	"github.com/piper-hyowon/dBtree/internal/domain/model"
	"github.com/piper-hyowon/dBtree/internal/domain/ports/secondary"
)

var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
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
	if !isValidEmail(email) {
		return false, errors.ErrInvalidEmail
	}

	user, err := s.userRepo.FindByEmail(ctx, email)
	isNewUser := err != nil || user == nil

	otpCode, err := generateOTP(constants.OTPLength)
	if err != nil {
		s.logger.Printf("OTP 생성 실패: %v", err)
		return isNewUser, fmt.Errorf("%w: %v", errors.ErrInternal, err)
	}

	now := time.Now().UTC()
	expiresAt := now.Add(time.Minute * constants.OTPExpirationMinutes)

	otp := &model.OTP{
		Code:      otpCode,
		CreatedAt: now,
		ExpiresAt: expiresAt,
	}

	session, err := s.sessionRepo.Get(ctx, email)
	if err != nil {
		session = model.NewAuthSession(email, otp)
	} else {
		session.OTP = otp
		session.Status = model.AuthPending
		session.UpdatedAt = now
	}

	if err := s.sessionRepo.Save(ctx, session); err != nil {
		s.logger.Printf("세션 저장 실패: %v", err)
		return isNewUser, fmt.Errorf("%w: %v", errors.ErrInternal, err)
	}

	if err := s.emailService.SendOTP(ctx, email, otpCode); err != nil {
		s.logger.Printf("OTP 이메일 발송 실패: %v", err)
		return isNewUser, fmt.Errorf("%w: %v", errors.ErrInternal, err)
	}

	s.logger.Printf("인증 시작: 이메일=%s, 신규사용자=%v", email, isNewUser)
	return isNewUser, nil
}

func (s *AuthService) GetSession(ctx context.Context, email string) (*model.AuthSession, error) {
	if !isValidEmail(email) {
		return nil, errors.ErrInvalidEmail
	}

	session, err := s.sessionRepo.Get(ctx, email)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (s *AuthService) ResendOTP(ctx context.Context, email string) error {
	if !isValidEmail(email) {
		return errors.ErrInvalidEmail
	}

	session, err := s.sessionRepo.Get(ctx, email)
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
		s.logger.Printf("OTP 생성 실패: %v", err)
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
		s.logger.Printf("세션 업데이트 실패: %v", err)
		return fmt.Errorf("%w: %v", errors.ErrInternal, err)
	}

	if err := s.emailService.SendOTP(ctx, email, otpCode); err != nil {
		s.logger.Printf("OTP 이메일 재발송 실패: %v", err)
		return fmt.Errorf("%w: %v", errors.ErrInternal, err)
	}

	s.logger.Printf("OTP 재전송: 이메일=%s, 횟수=%d", email, session.ResendCount)
	return nil
}

func (s *AuthService) VerifyOTP(ctx context.Context, email string, code string) (*model.User, error) {
	if !isValidEmail(email) {
		return nil, errors.ErrInvalidEmail
	}

	if code == "" || len(code) != constants.OTPLength {
		return nil, errors.ErrInvalidOTP
	}

	session, err := s.sessionRepo.Get(ctx, email)
	if err != nil {
		return nil, errors.ErrSessionNotFound
	}

	if session.OTP == nil {
		return nil, errors.ErrInvalidOTP
	}

	if time.Now().UTC().After(session.OTP.ExpiresAt) {
		return nil, errors.ErrExpiredOTP
	}

	if session.OTP.Code != code {
		return nil, errors.ErrInvalidOTP
	}

	now := time.Now().UTC()
	session.Status = model.AuthVerified
	session.UpdatedAt = now

	if err := s.sessionRepo.Save(ctx, session); err != nil {
		s.logger.Printf("인증 완료 후 세션 업데이트 실패: %v", err)
		return nil, fmt.Errorf("%w: %v", errors.ErrInternal, err)
	}

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil || user == nil {
		if err := s.userRepo.Create(ctx, email); err != nil {
			s.logger.Printf("유저 생성 실패: %v", err)
			return nil, fmt.Errorf("%w: %v", errors.ErrInternal, err)
		}

		user, err = s.userRepo.FindByEmail(ctx, email)
		if err != nil {
			s.logger.Printf("신규 유저 조회 실패: %v", err)
			return nil, fmt.Errorf("%w: %v", errors.ErrInternal, err)
		}

		s.emailService.SendWelcome(ctx, email)
		s.logger.Printf("인증 완료, 유저 생성: 이메일=%s", email)
	} else {
		s.logger.Printf("인증 완료: 이메일=%s", email)
	}

	return user, nil
}

// 헬퍼
func isValidEmail(email string) bool {
	return email != "" && emailRegex.MatchString(email)
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
