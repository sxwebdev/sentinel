package apiserver

import (
	"context"
	"errors"

	"connectrpc.com/authn"
	"github.com/sxwebdev/sentinel/internal/apiserver/interceptor"
)

func getUserDataContext(ctx context.Context) (*interceptor.UserDataContext, error) {
	u, ok := authn.GetInfo(ctx).(*interceptor.UserDataContext)
	if !ok || u == nil || u.User == nil {
		return nil, errors.New("no user data in context")
	}

	return u, nil
}
