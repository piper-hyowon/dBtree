package model

import (
	"time"
)

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
