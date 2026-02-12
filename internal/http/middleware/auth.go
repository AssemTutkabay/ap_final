package middleware

import (
	"net/http"
	"os"
	"strings"

	"ap_final/internal/store"
)

type AuthMiddleware struct {
	users store.UserStore
}

func NewAuthMiddleware(users store.UserStore) *AuthMiddleware {
	return &AuthMiddleware{users: users}
}

func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secret := []byte(os.Getenv("AUTH_SECRET"))
		if len(secret) == 0 {
			writeErr(w, http.StatusInternalServerError, "AUTH_SECRET not set")
			return
		}

		h := r.Header.Get("Authorization")
		if h == "" || !strings.HasPrefix(h, "Bearer ") {
			writeErr(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		rawTok := strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
		userID, err := ParseToken(secret, rawTok)
		if err != nil {
			writeErr(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		u, ok, err := m.users.GetByID(r.Context(), userID)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "bad request")
			return
		}
		if !ok || !u.IsActive {
			writeErr(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		ctx := WithUser(r.Context(), u)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
