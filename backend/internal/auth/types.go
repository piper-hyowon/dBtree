package auth

import (
	"time"
)

type SessionStatus string

const (
	Pending  SessionStatus = "pending"
	Verified SessionStatus = "verified"
)

type Session struct {
	Email          string
	Status         SessionStatus
	OTP            *OTP
	Token          string
	TokenExpiresAt time.Time
	ResendCount    int
	LastResendAt   *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func NewSession(email string, otp *OTP) *Session {
	now := time.Now().UTC()
	return &Session{
		Email:        email,
		Status:       Pending,
		OTP:          otp,
		ResendCount:  0,
		LastResendAt: nil,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

type OTP struct {
	Email      string
	Code       string
	CreatedAt  time.Time
	ExpiresAt  time.Time
	SendCount  int
	LastSentAt time.Time
	VerifiedAt time.Time
	IsVerified bool
}

func (o *OTP) IsExpired() bool {
	return time.Now().After(o.ExpiresAt)
}

func (o *OTP) CanResend() bool {
	return o.SendCount < 5 && time.Since(o.LastSentAt) >= time.Minute
}

// OTP 생성자
func NewOTP(email string, code string, expiryMinutes int) *OTP {
	return &OTP{
		Email:      email,
		Code:       code,
		ExpiresAt:  time.Now().UTC().Add(time.Duration(expiryMinutes) * time.Minute),
		IsVerified: false,
		CreatedAt:  time.Now().UTC(),
	}
}

func (o *OTP) Verify(code string) bool {
	if time.Now().UTC().After(o.ExpiresAt) {
		return false
	}

	if o.Code != code {
		return false
	}

	o.IsVerified = true
	return true
}
