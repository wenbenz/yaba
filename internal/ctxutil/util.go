package ctxutil

import (
	"context"

	"github.com/google/uuid"
)

func WithUser(ctx context.Context, user uuid.UUID) context.Context {
	return context.WithValue(ctx, CTXUser, user)
}

func GetUser(ctx context.Context) uuid.UUID {
	u, ok := ctx.Value(CTXUser).(uuid.UUID)
	if !ok {
		return uuid.Nil
	}

	return u
}
