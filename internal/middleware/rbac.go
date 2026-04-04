package middleware

import (
	"net/http"

	"github.com/yertaypert/gym-booking-api/internal/domain"
)

func RequireRoles(roles ...domain.UserRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := r.Context().Value(UserRoleKey).(domain.UserRole)
			if !ok {
				http.Error(w, "invalid user role", http.StatusForbidden)
				return
			}

			for _, allowedRole := range roles {
				if role == allowedRole {
					next.ServeHTTP(w, r)
					return
				}
			}

			http.Error(w, "forbidden", http.StatusForbidden)
		})
	}
}
