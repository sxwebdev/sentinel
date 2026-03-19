package auth

import (
	"context"
)

type IUserStore[T IUser] interface {
	GetByEmail(ctx context.Context, email string) (T, error)
	GetByID(ctx context.Context, id string) (T, error)
}
