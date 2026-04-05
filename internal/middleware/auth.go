package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/yertaypert/gym-booking-api/internal/auth"
	"github.com/yertaypert/gym-booking-api/internal/domain"
)

type contextKey string

const UserIDKey contextKey = "user_id"
const UserRoleKey contextKey = "role"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header is required", http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid auth format", http.StatusUnauthorized)
			return
		}

		claims, err := auth.ParseJWT(parts[1])
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UserRoleKey, domain.UserRole(claims.Role))

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
