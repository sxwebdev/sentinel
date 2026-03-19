package apiserver

import (
	"context"
	"errors"

	"connectrpc.com/authn"
)

func getUserDataContext(ctx context.Context) (*UserDataContext, error) {
	u, ok := authn.GetInfo(ctx).(*UserDataContext)
	if !ok || u == nil || u.User == nil {
		return nil, errors.New("no user data in context")
	}

	return u, nil
}
