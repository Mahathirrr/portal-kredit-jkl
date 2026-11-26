package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/Mahathirrr/kredit-jkl/internal/domain"
	"github.com/Mahathirrr/kredit-jkl/internal/repository"
	"github.com/Mahathirrr/kredit-jkl/internal/service"
)

type contextKey string

const UserContextKey contextKey = "current_user"

type AuthMiddleware struct {
	Auth  service.AuthService
	Users repository.UserRepository
}

func (m AuthMiddleware) Require(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		token = strings.TrimSpace(token)
		if token == "" {
			http.Error(w, `{"error":"butuh token"}`, http.StatusUnauthorized)
			return
		}
		claims, err := m.Auth.ParseToken(token)
		if err != nil {
			http.Error(w, `{"error":"token tidak valid"}`, http.StatusUnauthorized)
			return
		}
		user, err := m.Users.FindUserByID(r.Context(), claims.UserID)
		if err != nil || user == nil {
			http.Error(w, `{"error":"user tidak ditemukan"}`, http.StatusUnauthorized)
			return
		}
		user.PasswordHash = ""
		ctx := context.WithValue(r.Context(), UserContextKey, *user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func CurrentUser(ctx context.Context) (domain.User, bool) {
	user, ok := ctx.Value(UserContextKey).(domain.User)
	return user, ok
}

func RequireRoles(user domain.User, roles ...string) bool {
	for _, role := range roles {
		if user.Role == role {
			return true
		}
	}
	return false
}