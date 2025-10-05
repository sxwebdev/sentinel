package interceptor

import (
	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/sxwebdev/sentinel/internal/services/auth"
	"github.com/sxwebdev/tokenmanager"
)

type (
	userRole        string
	UserDataContext struct {
		User        *models.User
		AccessToken string
		Claims      *tokenmanager.Data[auth.SessionData]
	}
)
