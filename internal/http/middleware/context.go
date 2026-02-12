package middleware

import (
	"context"

	"ap_final/internal/domain"
)

type ctxKey string

const userKey ctxKey = "auth_user"

func WithUser(ctx context.Context, u domain.User) context.Context {
	return context.WithValue(ctx, userKey, u)
}

func UserFrom(ctx context.Context) (domain.User, bool) {
	u, ok := ctx.Value(userKey).(domain.User)
	return u, ok
}
