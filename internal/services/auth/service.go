package auth

import (
	"time"

	"github.com/sxwebdev/tokenmanager"
)

type Service[U IUser] struct {
	logger logger

	cache tokenmanager.ITokenStore

	userStore IUserStore[U]

	manager *Manager
}

func New[U IUser](
	l logger,
	cache tokenmanager.ITokenStore,
	userStore IUserStore[U],
	accessTokenSecretKey string,
	accessTokenDuration time.Duration,
	refreshTokenSecretKey string,
	refreshTokenDuration time.Duration,
) *Service[U] {
	return &Service[U]{
		logger:    l,
		cache:     cache,
		userStore: userStore,
		manager:   NewManager(accessTokenSecretKey, refreshTokenSecretKey, accessTokenDuration, refreshTokenDuration, cache),
	}
}

// Manager.
func (s *Service[U]) Manager() *Manager { return s.manager }
