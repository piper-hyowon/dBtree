package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Server              ServerConfig
	OTP                 OTPConfig
	Session             SessionConfig
	SMTP                SMTPConfig
	CORS                CORSConfig
	DebugLogging        bool
	Postgres            PostgresConfig
	UseLocalMemoryStore bool
}

type CORSConfig struct {
	AllowedOrigins   []string
	AllowCredentials bool
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

type PostgresConfig struct {
	Host                   string
	Port                   int
	Username               string
	Password               string
	DBName                 string
	SSLMode                string
	MaxOpenConns           int
	MaxIdleConns           int
	ConnMaxLifetimeMinutes int
}

func NewConfig() (*Config, error) {
	debugLogging := getEnvString("DEBUG_LOGGING", "false") == "true"
	useLocalMemoryStore := getEnvString("USE_LOCAL_MEMORY_STORE", "true") == "true"
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

	corsAllowedOrigins := strings.Split(getEnvString("CORS_ALLOWED_ORIGINS", "http://localhost:3000"), ",")
	corsAllowCredentials := getEnvString("CORS_ALLOW_CREDENTIALS", "true") == "true"

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

	postgresHost := getEnvString("POSTGRES_HOST", "localhost")
	postgresPort, err := getEnvInt("POSTGRES_PORT", 5432)
	if err != nil {
		return nil, err
	}
	postgresUser := getEnvString("POSTGRES_USER", "")
	postgresPassword := getEnvString("POSTGRES_PASSWORD", "")
	postgresDBName := getEnvString("POSTGRES_DB", "postgres")
	postgresSSLMode := getEnvString("POSTGRES_SSL_MODE", "disable")
	maxOpenConns, err := getEnvInt("POSTGRES_MAX_OPEN_CONNS", 30)
	if err != nil {
		return nil, err
	}
	maxIdleConns, err := getEnvInt("POSTGRES_MAX_IDLE_CONNS", 20)
	if err != nil {
		return nil, err
	}
	connMaxLifetimeMinutes, err := getEnvInt("POSTGRES_CONN_MAX_LIFETIME", 0)
	if err != nil {
		return nil, err
	}

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
		CORS: CORSConfig{
			AllowedOrigins:   corsAllowedOrigins,
			AllowCredentials: corsAllowCredentials,
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
		DebugLogging:        debugLogging,
		UseLocalMemoryStore: useLocalMemoryStore,
		Postgres: PostgresConfig{
			Host:                   postgresHost,
			Port:                   postgresPort,
			Username:               postgresUser,
			Password:               postgresPassword,
			DBName:                 postgresDBName,
			SSLMode:                postgresSSLMode,
			MaxOpenConns:           maxOpenConns,
			MaxIdleConns:           maxIdleConns,
			ConnMaxLifetimeMinutes: connMaxLifetimeMinutes,
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
