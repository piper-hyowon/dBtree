package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server  ServerConfig
	OTP     OTPConfig
	Session SessionConfig
	SMTP    SMTPConfig
}

type ServerConfig struct {
	Port                int
	ReadTimeoutSeconds  int
	WriteTimeoutSeconds int
	IdleTimeoutSeconds  int
}

type OTPConfig struct {
	ExpirationMinutes int
	MaxSendCount      int
	ResendInterval    time.Duration
}

type SessionConfig struct {
	ExpirationHours      int
	CleanupIntervalHours int
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

func NewConfig() (*Config, error) {
	port, err := getEnvInt("SERVER_PORT", 8080)
	if err != nil {
		return nil, err
	}

	readTimeout, err := getEnvInt("SERVER_READ_TIMEOUT", 10)
	if err != nil {
		return nil, err
	}

	writeTimeout, err := getEnvInt("SERVER_WRITE_TIMEOUT", 10)
	if err != nil {
		return nil, err
	}

	idleTimeout, err := getEnvInt("SERVER_IDLE_TIMEOUT", 60)
	if err != nil {
		return nil, err
	}

	otpExpiration, err := getEnvInt("OTP_EXPIRATION_MINUTES", 10)
	if err != nil {
		return nil, err
	}

	otpMaxSendCount, err := getEnvInt("OTP_MAX_SEND_COUNT", 5)
	if err != nil {
		return nil, err
	}

	otpResendInterval, err := getEnvInt("OTP_RESEND_INTERVAL_SECONDS", 60)
	if err != nil {
		return nil, err
	}

	sessionExpiration, err := getEnvInt("SESSION_EXPIRATION_HOURS", 24)
	if err != nil {
		return nil, err
	}

	sessionCleanup, err := getEnvInt("SESSION_CLEANUP_INTERVAL_HOURS", 1)
	if err != nil {
		return nil, err
	}

	smtpPort, err := getEnvInt("SMTP_PORT", 587)
	if err != nil {
		return nil, err
	}

	smtpHost := getEnvString("SMTP_HOST", "")
	smtpUsername := getEnvString("SMTP_USERNAME", "")
	smtpPassword := getEnvString("SMTP_PASSWORD", "")
	smtpFrom := getEnvString("SMTP_FROM", "")

	if smtpHost == "" || smtpUsername == "" || smtpPassword == "" {
		return nil, fmt.Errorf("SMTP 환경 변수 확인")
	}

	return &Config{
		Server: ServerConfig{
			Port:                port,
			ReadTimeoutSeconds:  readTimeout,
			WriteTimeoutSeconds: writeTimeout,
			IdleTimeoutSeconds:  idleTimeout,
		},
		OTP: OTPConfig{
			ExpirationMinutes: otpExpiration,
			MaxSendCount:      otpMaxSendCount,
			ResendInterval:    time.Duration(otpResendInterval) * time.Second,
		},
		Session: SessionConfig{
			ExpirationHours:      sessionExpiration,
			CleanupIntervalHours: sessionCleanup,
		},
		SMTP: SMTPConfig{
			Host:     smtpHost,
			Port:     smtpPort,
			Username: smtpUsername,
			Password: smtpPassword,
			From:     smtpFrom,
		},
	}, nil
}

func getEnvString(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) (int, error) {
	if value, exists := os.LookupEnv(key); exists {
		intValue, err := strconv.Atoi(value)
		if err != nil {
			return 0, fmt.Errorf("%s 환경 변수가 정수가 아님: %w", key, err)
		}
		return intValue, nil
	}
	return defaultValue, nil
}
