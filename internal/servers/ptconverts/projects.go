package ptconverts

import (
	projectsv1 "github.com/sxwebdev/sentinel/internal/hub/hubserver/api/sentinel/projects/v1"
	"github.com/sxwebdev/sentinel/internal/models"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ConvertProjectToProto converts a models.Project to its protobuf representation
func ConvertProjectToProto(p *models.Project) *projectsv1.Project {
	if p == nil {
		return new(projectsv1.Project)
	}

	settings := &projectsv1.ProjectSettings{
		MonitorDefaults: &projectsv1.ProjectMonitorDefaults{
			Interval: p.Settings.MonitorDefaults.DefaultInterval,
			Timeout:  p.Settings.MonitorDefaults.DefaultTimeout,
			Retries:  p.Settings.MonitorDefaults.DefaultRetries,
		},
	}

	return &projectsv1.Project{
		Id:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Settings:    settings,
		CreatedAt:   timestamppb.New(p.CreatedAt),
		UpdatedAt:   timestamppb.New(p.UpdatedAt),
	}
}
