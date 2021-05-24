package context

import (
	"context"

	"github.com/wmolicki/bookler/models"
)

const userKey privateKey = "user"

type privateKey string

// WithUser returns new context with user added into it
func WithUser(ctx context.Context, user *models.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

// User returns user value from context (nicely type casted)
func User(ctx context.Context) *models.User {
	if tmp := ctx.Value(userKey); tmp != nil {
		if user, ok := tmp.(*models.User); ok == true {
			return user
		}
	}
	return nil
}
