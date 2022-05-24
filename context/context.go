package context

import (
	"context"
	"go-web-dev/models"
)

const (
	userKey privateKey = "user"
)

// enables access to keys from this package from typing
type privateKey string

func WithUser(ctx context.Context, user *models.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

func User(ctx context.Context) *models.User {
	if temp, ok := ctx.Value(userKey).(*models.User); ok {
		return temp
	}
	return nil
}
