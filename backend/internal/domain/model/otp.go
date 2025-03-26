package model

import (
	"time"
)

type OTP struct {
	Email      string
	Code       string
	ExpiresAt  time.Time
	IsVerified bool
	CreatedAt  time.Time
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

func (o *OTP) IsExpired() bool {
	return time.Now().UTC().After(o.ExpiresAt)
}
