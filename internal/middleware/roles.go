package middleware

import (
	"net/http"

	"github.com/Woun1zoN/go-identity-service/internal/error_handling"
)

func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := r.Context().Value(RoleKey).(string)
			if !ok {
				errorhandling.Unauthorized(w, r, GetRequestID(r))
                return
			}

			for _, allowed := range roles {
				if role == allowed {
					next.ServeHTTP(w, r)
					return
				}
			}

			errorhandling.Forbidden(w, r, GetRequestID(r))
		})
	}
}