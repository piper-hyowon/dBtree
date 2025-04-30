package middleware

import (
	"context"
	"github.com/piper-hyowon/dBtree/internal/core/auth"
	corecontext "github.com/piper-hyowon/dBtree/internal/core/context"
	"github.com/piper-hyowon/dBtree/internal/core/user"
	"log"
	"net/http"
	"strings"
)

type AuthMiddleware struct {
	authService auth.Service
	logger      *log.Logger
}

func NewAuthMiddleware(authService auth.Service, logger *log.Logger) *AuthMiddleware {
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
		u, err := m.authService.ValidateSession(r.Context(), token)
		if err != nil {
			m.logger.Printf("인증 실패: %v", err)
			http.Error(w, "인증 실패", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), corecontext.UserKey, u)
		ctx = context.WithValue(ctx, corecontext.TokenKey, token)

		next(w, r.WithContext(ctx))
	}
}

func GetUserFromContext(ctx context.Context) *user.User {
	u, ok := ctx.Value(corecontext.UserKey).(*user.User)
	if !ok {
		return nil
	}
	return u
}

func GetTokenFromContext(ctx context.Context) string {
	token, ok := ctx.Value(corecontext.TokenKey).(string)
	if !ok {
		return ""
	}
	return token
}
