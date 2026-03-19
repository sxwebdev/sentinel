package auth

import "errors"

var (
	ErrEmptyAccessToken         = errors.New("empty access token")
	ErrEmptyRefreshToken        = errors.New("empty refresh token")
	ErrEmptyUserID              = errors.New("empty user id")
	ErrEmptyOrganizationID      = errors.New("empty organization id")
	ErrEmptyEmail               = errors.New("empty email")
	ErrEmptyPassword            = errors.New("empty password")
	ErrIncorrectEmailOrPassword = errors.New("incorrect email or password")
	ErrEmptyHash                = errors.New("empty hash")
	ErrEmptyCode                = errors.New("empty code")
)
