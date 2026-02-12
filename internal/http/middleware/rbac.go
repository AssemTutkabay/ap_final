package middleware

import (
	"net/http"

	"ap_final/internal/domain"
)

func RequireRoles(roles ...domain.UserRole) func(http.Handler) http.Handler {
	allowed := map[domain.UserRole]bool{}
	for _, r := range roles {
		allowed[r] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, ok := UserFrom(r.Context())
			if !ok {
				writeErr(w, http.StatusUnauthorized, "unauthorized")
				return
			}
			if !allowed[u.Role] {
				writeErr(w, http.StatusForbidden, "forbidden")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
