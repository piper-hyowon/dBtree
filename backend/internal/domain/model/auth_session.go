package model

import (
	"time"
)

type AuthSessionStatus string

const (
	AuthPending  AuthSessionStatus = "pending"
	AuthVerified AuthSessionStatus = "verified"
)

type AuthSession struct {
	Email          string
	Status         AuthSessionStatus
	OTP            *OTP
	Token          string
	TokenExpiresAt time.Time
	ResendCount    int
	LastResendAt   *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// AuthSession 생성자
func NewAuthSession(email string, otp *OTP) *AuthSession {
	now := time.Now().UTC()
	return &AuthSession{
		Email:        email,
		Status:       AuthPending,
		OTP:          otp,
		ResendCount:  0,
		LastResendAt: nil,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}
