package auth

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"github.com/piper-hyowon/dBtree/internal/common"
	"time"
)

type PostgresSessionStore struct {
	db *sql.DB
}

var _ SessionStore = (*PostgresSessionStore)(nil)

func NewPostgrestore(db *sql.DB) SessionStore {
	return &PostgresSessionStore{
		db: db,
	}
}

func (s *PostgresSessionStore) Save(ctx context.Context, session *Session) error {
	if session == nil || session.Email == "" {
		return fmt.Errorf("invalid session: %w", common.ErrInternal)
	}

	exists, err := s.sessionExists(ctx, session.Email)
	if err != nil {
		return fmt.Errorf("세션 확인 실패: %w", err)
	}

	var otpCode, otpCreatedAt, otpExpiresAt sql.NullString
	var tokenExpiresAt sql.NullTime
	var lastResendAt sql.NullTime

	if session.OTP != nil {
		otpCode = sql.NullString{String: session.OTP.Code, Valid: true}
		otpCreatedAt = sql.NullString{String: session.OTP.CreatedAt.Format(time.RFC3339), Valid: true}
		otpExpiresAt = sql.NullString{String: session.OTP.ExpiresAt.Format(time.RFC3339), Valid: true}
	}

	if !session.TokenExpiresAt.IsZero() {
		tokenExpiresAt = sql.NullTime{Time: session.TokenExpiresAt, Valid: true}
	}

	if session.LastResendAt != nil {
		lastResendAt = sql.NullTime{Time: *session.LastResendAt, Valid: true}
	}

	now := time.Now().UTC()

	if exists {
		query := `
			UPDATE "sessions" 
			SET status = $1, 
				otp_code = $2, 
				otp_created_at = $3, 
				otp_expires_at = $4, 
				token = $5, 
				token_expires_at = $6, 
				resend_count = $7, 
				last_resend_at = $8, 
				updated_at = $9 
			WHERE email = $10
		`

		_, err = s.db.ExecContext(ctx, query,
			string(session.Status),
			otpCode,
			otpCreatedAt,
			otpExpiresAt,
			session.Token,
			tokenExpiresAt,
			session.ResendCount,
			lastResendAt,
			now,
			session.Email,
		)
	} else {
		query := `
			INSERT INTO sessions 
			(id, email, status, otp_code, otp_created_at, otp_expires_at, token, token_expires_at, resend_count, last_resend_at, created_at, updated_at) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		`

		_, err = s.db.ExecContext(ctx, query,
			uuid.New().String(),
			session.Email,
			string(session.Status),
			otpCode,
			otpCreatedAt,
			otpExpiresAt,
			session.Token,
			tokenExpiresAt,
			session.ResendCount,
			lastResendAt,
			now,
			now,
		)
	}

	if err != nil {
		return fmt.Errorf("세션 저장 실패: %w", err)
	}

	return nil
}

func (s *PostgresSessionStore) GetByEmail(ctx context.Context, email string) (*Session, error) {
	if email == "" {
		return nil, fmt.Errorf("empty email: %w", common.ErrInternal)
	}

	query := `
		SELECT 
			email, status, otp_code, otp_created_at, otp_expires_at, 
			token, token_expires_at, resend_count, last_resend_at, 
			created_at, updated_at 
		FROM sessions 
		WHERE email = $1
	`

	var (
		status                                    string
		otpCode, otpCreatedAtStr, otpExpiresAtStr sql.NullString
		token                                     sql.NullString
		tokenExpiresAt, lastResendAt              sql.NullTime
		resendCount                               int
		createdAt, updatedAt                      time.Time
	)

	err := s.db.QueryRowContext(ctx, query, email).Scan(
		&email,
		&status,
		&otpCode,
		&otpCreatedAtStr,
		&otpExpiresAtStr,
		&token,
		&tokenExpiresAt,
		&resendCount,
		&lastResendAt,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, common.ErrSessionNotFound
		}
		return nil, fmt.Errorf("세션 조회 실패: %w", err)
	}

	session := &Session{
		Email:       email,
		Status:      SessionStatus(status),
		ResendCount: resendCount,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}

	if otpCode.Valid && otpCreatedAtStr.Valid && otpExpiresAtStr.Valid {
		otpCreatedAt, _ := time.Parse(time.RFC3339, otpCreatedAtStr.String)
		otpExpiresAt, _ := time.Parse(time.RFC3339, otpExpiresAtStr.String)

		session.OTP = &OTP{
			Email:     email,
			Code:      otpCode.String,
			CreatedAt: otpCreatedAt,
			ExpiresAt: otpExpiresAt,
		}
	}

	if token.Valid {
		session.Token = token.String
	}

	if tokenExpiresAt.Valid {
		session.TokenExpiresAt = tokenExpiresAt.Time
	}

	if lastResendAt.Valid {
		last := lastResendAt.Time
		session.LastResendAt = &last
	}

	return session, nil
}

func (s *PostgresSessionStore) GetByToken(ctx context.Context, token string) (*Session, error) {
	if token == "" {
		return nil, fmt.Errorf("empty token: %w", common.ErrInvalidToken)
	}

	query := `
		SELECT 
			email 
		FROM sessions 
		WHERE token = $1 AND token_expires_at > $2
	`

	var email string
	err := s.db.QueryRowContext(ctx, query, token, time.Now().UTC()).Scan(&email)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, common.ErrSessionNotFound
		}
		return nil, fmt.Errorf("토큰으로 세션 조회 실패: %w", err)
	}

	return s.GetByEmail(ctx, email)
}

func (s *PostgresSessionStore) Delete(ctx context.Context, email string) error {
	if email == "" {
		return fmt.Errorf("empty email: %w", common.ErrInternal)
	}

	query := `DELETE FROM sessions WHERE email = $1`

	result, err := s.db.ExecContext(ctx, query, email)
	if err != nil {
		return fmt.Errorf("세션 삭제 실패: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("세션 삭제 결과 확인 실패: %w", err)
	}

	if rowsAffected == 0 {
		return common.ErrSessionNotFound
	}

	return nil
}

func (s *PostgresSessionStore) Cleanup(ctx context.Context) error {
	now := time.Now().UTC()

	query := `
		DELETE FROM sessions 
		WHERE (otp_expires_at < $1 AND token IS NULL) 
		   OR (token_expires_at < $1)
	`

	_, err := s.db.ExecContext(ctx, query, now)
	if err != nil {
		return fmt.Errorf("만료된 세션 정리 실패: %w", err)
	}

	return nil
}

func (s *PostgresSessionStore) sessionExists(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM sessions WHERE email = $1)`

	var exists bool
	err := s.db.QueryRowContext(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}
