package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/sxwebdev/sentinel/internal/store/storecmn"
)

type AuthResponse[TUser IUser] struct { //nolint:revive
	AuthorizationResponse

	User TUser
	Role string
}

func (s *Service[TUser]) Authorization(ctx context.Context, email, password string, additionalData SessionData) (
	*AuthResponse[TUser], error,
) {
	if email == "" {
		return nil, ErrEmptyEmail
	}

	if password == "" {
		return nil, ErrEmptyPassword
	}

	// get user
	user, err := s.userStore.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, storecmn.ErrNotFound) {
			return nil, ErrIncorrectEmailOrPassword
		}
		return nil, fmt.Errorf("get user by email error: %w", err)
	}

	if !IsCorrectPassword(user.GetPassword(), password) {
		return nil, ErrIncorrectEmailOrPassword
	}

	return s.authorization(ctx, user, additionalData)
}

func (s *Service[TUser]) authorization(ctx context.Context, user TUser, additionalData SessionData) (
	*AuthResponse[TUser], error,
) {
	res := &AuthResponse[TUser]{ //nolint:forcetypeassert
		User: any(user).(TUser), //nolint:nolintlint,forcetypeassert
		Role: user.GetRole(),
	}

	data, err := s.manager.Authorization(ctx, user.GetID(), additionalData)
	if err != nil {
		return nil, fmt.Errorf("failed to authorize: %w", err)
	}

	res.AuthorizationResponse = *data

	return res, nil
}
