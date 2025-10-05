//nolint:inamedparam
package auth

type logger interface {
	Debug(...any)
	Debugf(template string, args ...any)

	Info(...any)
	Infof(template string, args ...any)

	Warn(...any)
	Warnf(template string, args ...any)

	Error(...any)
	Errorf(template string, args ...any)
}

// UserClaims is a custom JWT claims that contains some user's information.
type UserClaims struct {
	ID string `json:"id"`
}

type IUser interface {
	GetID() string
	GetPassword() string
	GetEmail() string
	GetRole() string
}
