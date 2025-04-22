package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/piper-hyowon/dBtree/internal/adapters/primary/core"
	"github.com/piper-hyowon/dBtree/internal/domain/model"
)

type contextKey string

const UserKey contextKey = "user"

type AuthMiddleware struct {
	authService *core.AuthService
	logger      *log.Logger
}

func NewAuthMiddleware(authService *core.AuthService, logger *log.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
		logger:      logger,
	}
}

func (m *AuthMiddleware) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "인증 헤더 없음", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "인증 포맷 오류", http.StatusUnauthorized)
			return
		}
		token := parts[1]

		user, err := m.authService.ValidateSession(r.Context(), token)
		if err != nil {
			m.logger.Printf("인증 실패: %v", err)
			http.Error(w, "인증 실패", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserKey, user)

		next(w, r.WithContext(ctx))
	}
}

func GetUserFromContext(ctx context.Context) *model.User {
	user, ok := ctx.Value(UserKey).(*model.User)
	if !ok {
		return nil
	}
	return user
}
