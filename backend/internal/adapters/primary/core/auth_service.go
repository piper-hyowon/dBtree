package core

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/piper-hyowon/dBtree/internal/domain/model"
	"github.com/piper-hyowon/dBtree/internal/domain/ports/secondary"
)

var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
)

// 에러 정의
var (
	ErrInvalidEmail    = errors.New("invalid email")
	ErrTooManyResends  = errors.New("OTP 요청 횟수 초과, 잠시후 재시도")
	ErrTooEarlyResend  = errors.New("OTP 재전송은 1분 후 가능")
	ErrInvalidOTP      = errors.New("invalid OTP")
	ErrExpiredOTP      = errors.New("expired OTP")
	ErrSessionNotFound = errors.New("session 404")
	ErrInternal        = errors.New("500")
)

const (
	otpLength            = 6
	maxResendAttempts    = 5
	resendWaitSeconds    = 60
	otpExpirationMinutes = 10
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
		return false, ErrInvalidEmail
	}

	user, err := s.userRepo.FindByEmail(ctx, email)
	isNewUser := err != nil || user == nil

	otpCode, err := generateOTP(otpLength)
	if err != nil {
		s.logger.Printf("OTP 생성 실패: %v", err)
		return isNewUser, fmt.Errorf("%w: %v", ErrInternal, err)
	}

	now := time.Now().UTC()
	expiresAt := now.Add(time.Minute * otpExpirationMinutes)

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
		return isNewUser, fmt.Errorf("%w: %v", ErrInternal, err)
	}

	if err := s.emailService.SendOTP(ctx, email, otpCode); err != nil {
		s.logger.Printf("OTP 이메일 발송 실패: %v", err)
		return isNewUser, fmt.Errorf("%w: %v", ErrInternal, err)
	}

	s.logger.Printf("인증 시작: 이메일=%s, 신규사용자=%v", email, isNewUser)
	return isNewUser, nil
}

func (s *AuthService) ResendOTP(ctx context.Context, email string) error {
	if !isValidEmail(email) {
		return ErrInvalidEmail
	}

	session, err := s.sessionRepo.Get(ctx, email)
	if err != nil {
		return ErrSessionNotFound
	}

	if session.ResendCount >= maxResendAttempts {
		return ErrTooManyResends
	}

	if session.LastResendAt != nil {
		waitTime := time.Duration(resendWaitSeconds) * time.Second
		nextResendTime := session.LastResendAt.Add(waitTime)
		if time.Now().UTC().Before(nextResendTime) {
			return ErrTooEarlyResend
		}
	}

	otpCode, err := generateOTP(otpLength)
	if err != nil {
		s.logger.Printf("OTP 생성 실패: %v", err)
		return fmt.Errorf("%w: %v", ErrInternal, err)
	}

	now := time.Now().UTC()
	expiresAt := now.Add(time.Minute * otpExpirationMinutes)

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
		return fmt.Errorf("%w: %v", ErrInternal, err)
	}

	if err := s.emailService.SendOTP(ctx, email, otpCode); err != nil {
		s.logger.Printf("OTP 이메일 재발송 실패: %v", err)
		return fmt.Errorf("%w: %v", ErrInternal, err)
	}

	s.logger.Printf("OTP 재전송: 이메일=%s, 횟수=%d", email, session.ResendCount)
	return nil
}

func (s *AuthService) VerifyOTP(ctx context.Context, email string, code string) (*model.User, error) {
	if !isValidEmail(email) {
		return nil, ErrInvalidEmail
	}

	if code == "" || len(code) != otpLength {
		return nil, ErrInvalidOTP
	}

	session, err := s.sessionRepo.Get(ctx, email)
	if err != nil {
		return nil, ErrSessionNotFound
	}

	if session.OTP == nil {
		return nil, ErrInvalidOTP
	}

	if time.Now().UTC().After(session.OTP.ExpiresAt) {
		return nil, ErrExpiredOTP
	}

	if session.OTP.Code != code {
		return nil, ErrInvalidOTP
	}

	now := time.Now().UTC()
	session.Status = model.AuthVerified
	session.UpdatedAt = now

	if err := s.sessionRepo.Save(ctx, session); err != nil {
		s.logger.Printf("인증 완료 후 세션 업데이트 실패: %v", err)
		return nil, fmt.Errorf("%w: %v", ErrInternal, err)
	}

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil || user == nil {
		if err := s.userRepo.Create(ctx, email); err != nil {
			s.logger.Printf("유저 생성 실패: %v", err)
			return nil, fmt.Errorf("%w: %v", ErrInternal, err)
		}

		user, err = s.userRepo.FindByEmail(ctx, email)
		if err != nil {
			s.logger.Printf("신규 유저 조회 실패: %v", err)
			return nil, fmt.Errorf("%w: %v", ErrInternal, err)
		}

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
