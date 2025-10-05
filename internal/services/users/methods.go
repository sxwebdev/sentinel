package users

import (
	"context"

	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/sxwebdev/sentinel/internal/store/repos/repo_users"
)

// GetByID gets user by ID
func (s *Service) GetByID(ctx context.Context, id string) (*models.User, error) {
	return s.store.Users().GetByID(ctx, id)
}

// GetByEmail gets user by email
func (s *Service) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return s.store.Users().GetByEmail(ctx, email)
}

type CreateParams = repo_users.CreateParams

// Create creates new user
func (s *Service) Create(ctx context.Context, params CreateParams) (*models.User, error) {
	if err := s.validator.Struct(params); err != nil {
		return nil, err
	}

	if err := checkPassword(params.Password, params.Password); err != nil {
		return nil, err
	}

	hashedPassword, err := generateHashFromPassword(params.Password)
	if err != nil {
		return nil, err
	}

	params.Password = hashedPassword

	return s.store.Users().Create(ctx, params)
}

type UpdateParams = repo_users.UpdateParams

// Update updates user
func (s *Service) Update(ctx context.Context, params UpdateParams) (*models.User, error) {
	if err := s.validator.Struct(params); err != nil {
		return nil, err
	}
	return s.store.Users().Update(ctx, params)
}

// CheckRootUserExists checks if a root user exists
func (s *Service) CheckRootUserExists(ctx context.Context) (bool, error) {
	val, err := s.store.Users().CheckRootUserExists(ctx)
	if err != nil {
		return false, err
	}
	return val > 0, nil
}
