package ptconverts

import (
	usersv1 "github.com/sxwebdev/sentinel/internal/hub/hubserver/api/sentinel/users/v1"
	"github.com/sxwebdev/sentinel/internal/models"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ConvertUserToProto converts a models.User to its protobuf representation
func ConvertUserToProto(u *models.User) *usersv1.User {
	if u == nil {
		return new(usersv1.User)
	}

	return &usersv1.User{
		Id:        u.ID,
		Email:     u.Email,
		FullName:  u.FullName,
		Role:      u.Role,
		AvatarUrl: u.Avatar,
		CreatedAt: timestamppb.New(u.CreatedAt),
		UpdatedAt: timestamppb.New(u.UpdatedAt),
	}
}

// ConvertUserFromProto converts a protobuf User to its models representation
func ConvertUserFromProto(u *usersv1.User) *models.User {
	if u == nil {
		return nil
	}

	return &models.User{
		ID:        u.Id,
		Email:     u.Email,
		FullName:  u.FullName,
		Role:      models.UserRole(u.Role),
		Avatar:    u.AvatarUrl,
		CreatedAt: u.CreatedAt.AsTime(),
		UpdatedAt: u.UpdatedAt.AsTime(),
	}
}
