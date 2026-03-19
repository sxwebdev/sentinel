package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/sxwebdev/sentinel/internal/store/storecmn"
)

type AuthenticateResponse[TUser IUser] struct {
	User TUser
	Role string
}

func (s *Service[TUser]) Authenticate(ctx context.Context, accessToken string) (*AuthenticateResponse[TUser], error) {
	if accessToken == "" {
		return nil, ErrEmptyAccessToken
	}

	tokenData, err := s.manager.Authenticate(ctx, accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate: %w", err)
	}

	// get user
	user, err := s.userStore.GetByID(ctx, tokenData.UserID)
	if err != nil {
		if errors.Is(err, storecmn.ErrNotFound) {
			return nil, storecmn.ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by email error: %w", err)
	}

	res := &AuthenticateResponse[TUser]{ //nolint:nolintlint,forcetypeassert
		User: any(user).(TUser),
		Role: user.GetRole(),
	}

	return res, nil
}
