package postgres

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	"github.com/piper-hyowon/dBtree/internal/core/errors"

	"github.com/piper-hyowon/dBtree/internal/core/auth"
	"time"
)

type SessionStore struct {
	db *sql.DB
}

var _ auth.SessionStore = (*SessionStore)(nil)

func NewSessionStore(db *sql.DB) auth.SessionStore {
	return &SessionStore{
		db: db,
	}
}

func (s *SessionStore) Save(ctx context.Context, session *auth.Session) error {
	exists, err := s.sessionExists(ctx, session.Email)
	if err != nil {
		return errors.Wrapf(err, "세션 확인 실패")
	}

	var otp sql.NullString
	var tokenExpiresAt, otpCreatedAt, otpExpiresAt sql.NullTime
	var lastResendAt sql.NullTime

	if session.OTP != nil {
		otp = sql.NullString{String: session.OTP.Code, Valid: true}
		otpCreatedAt = sql.NullTime{Time: session.OTP.CreatedAt, Valid: true}
		otpExpiresAt = sql.NullTime{Time: session.OTP.ExpiresAt, Valid: true}
	} else {
		otp = sql.NullString{Valid: false}
		otpCreatedAt = sql.NullTime{Valid: false}
		otpExpiresAt = sql.NullTime{Valid: false}
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
				otp = $2, 
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
			otp,
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
			(id, email, status, otp, otp_created_at, otp_expires_at, token, token_expires_at, resend_count, last_resend_at, created_at, updated_at) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		`

		_, err = s.db.ExecContext(ctx, query,
			uuid.New().String(),
			session.Email,
			string(session.Status),
			otp,
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
		return errors.Wrapf(err, "세션 저장 실패")
	}

	return nil
}

func (s *SessionStore) FindByEmail(ctx context.Context, email string) (*auth.Session, error) {
	query := `
		SELECT 
			email, status, otp, otp_created_at, otp_expires_at, 
			token, token_expires_at, resend_count, last_resend_at, 
			created_at, updated_at 
		FROM sessions 
		WHERE email = $1
	`

	var (
		status                                string
		otp, otpCreatedAtStr, otpExpiresAtStr sql.NullString
		token                                 sql.NullString
		tokenExpiresAt, lastResendAt          sql.NullTime
		resendCount                           int
		createdAt, updatedAt                  time.Time
	)

	err := s.db.QueryRowContext(ctx, query, email).Scan(
		&email,
		&status,
		&otp,
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
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "세션 조회 실패")
	}

	session := &auth.Session{
		Email:       email,
		Status:      auth.SessionStatus(status),
		ResendCount: resendCount,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}

	if otp.Valid && otpCreatedAtStr.Valid && otpExpiresAtStr.Valid {
		otpCreatedAt, _ := time.Parse(time.RFC3339, otpCreatedAtStr.String)
		otpExpiresAt, _ := time.Parse(time.RFC3339, otpExpiresAtStr.String)

		session.OTP = &auth.OTP{
			Email:     email,
			Code:      otp.String,
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

func (s *SessionStore) FindByToken(ctx context.Context, token string) (*auth.Session, error) {
	query := `
        SELECT 
            email, status, otp, otp_created_at, otp_expires_at, 
            token, token_expires_at, resend_count, last_resend_at, 
            created_at, updated_at 
        FROM sessions 
        WHERE token = $1 AND token_expires_at > $2
    `

	var (
		email                        string
		status                       string
		otp                          sql.NullString
		otpCreatedAt, otpExpiresAt   sql.NullTime
		tokenValue                   sql.NullString
		tokenExpiresAt, lastResendAt sql.NullTime
		resendCount                  int
		createdAt, updatedAt         time.Time
	)

	err := s.db.QueryRowContext(ctx, query, token, time.Now().UTC()).Scan(
		&email,
		&status,
		&otp,
		&otpCreatedAt,
		&otpExpiresAt,
		&tokenValue,
		&tokenExpiresAt,
		&resendCount,
		&lastResendAt,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "토큰으로 세션 조회 실패")
	}

	session := &auth.Session{
		Email:       email,
		Status:      auth.SessionStatus(status),
		Token:       token,
		ResendCount: resendCount,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}

	if otp.Valid && otpCreatedAt.Valid && otpExpiresAt.Valid {
		session.OTP = &auth.OTP{
			Email:     email,
			Code:      otp.String,
			CreatedAt: otpCreatedAt.Time,
			ExpiresAt: otpExpiresAt.Time,
		}
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
func (s *SessionStore) Delete(ctx context.Context, email string) error {
	query := `DELETE FROM sessions WHERE email = $1`

	result, err := s.db.ExecContext(ctx, query, email)
	if err != nil {
		return errors.Wrapf(err, "세션 삭제 실패")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrapf(err, "세션 삭제 결과 확인 실패")
	}

	if rowsAffected == 0 {
		//return errors.NewSessionNotFoundError()
	}

	return nil
}

func (s *SessionStore) Cleanup(ctx context.Context) error {
	now := time.Now().UTC()

	query := `
		DELETE FROM sessions 
		WHERE (otp_expires_at < $1 AND token IS NULL) 
		   OR (token_expires_at < $1)
	`

	_, err := s.db.ExecContext(ctx, query, now)
	if err != nil {
		return errors.Wrapf(err, "만료된 세션 정리 실패")
	}

	return nil
}

func (s *SessionStore) sessionExists(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM sessions WHERE email = $1)`

	var exists bool
	err := s.db.QueryRowContext(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}
