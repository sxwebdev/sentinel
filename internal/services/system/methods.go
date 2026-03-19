package system

import (
	"context"
	"fmt"

	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/sxwebdev/sentinel/internal/store/repos/repo_users"
	"github.com/sxwebdev/sentinel/internal/utils"
)

// CheckIsInitialized checks if the system is initialized
func (s *Service) CheckIsInitialized(ctx context.Context) (bool, error) {
	return s.usersService.CheckRootUserExists(ctx)
}

// Initialize initializes the system by creating the root user
func (s *Service) Initialize(ctx context.Context, rootEmail, rootPassword string) error {
	exists, err := s.CheckIsInitialized(ctx)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("system already initialized")
	}

	// Create user
	_, err = s.usersService.Create(ctx, repo_users.CreateParams{
		ID:       utils.GenerateULID(),
		Email:    rootEmail,
		Password: rootPassword,
		Role:     models.UserRoleRoot,
	})
	if err != nil {
		return err
	}

	return nil
}
