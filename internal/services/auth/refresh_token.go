package auth

import (
	"context"
)

func (s *Service[TUser]) RefreshToken(ctx context.Context, refreshToken string) (*RefreshTokenResponse, error) {
	if refreshToken == "" {
		return nil, ErrEmptyRefreshToken
	}

	return s.manager.RefreshToken(ctx, refreshToken)
}
