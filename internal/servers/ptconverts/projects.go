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

	return &projectsv1.Project{
		Id:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Settings:    ConvertProjectSettingsToProto(p.Settings),
		CreatedAt:   timestamppb.New(p.CreatedAt),
		UpdatedAt:   timestamppb.New(p.UpdatedAt),
	}
}

// ConvertProjectSettingsToProto converts a models.ProjectSettings to its protobuf representation
func ConvertProjectSettingsToProto(ps models.ProjectSettings) *projectsv1.ProjectSettings {
	return &projectsv1.ProjectSettings{
		MonitorDefaults: &projectsv1.ProjectMonitorDefaults{
			Interval: ps.MonitorDefaults.DefaultInterval,
			Timeout:  ps.MonitorDefaults.DefaultTimeout,
			Retries:  ps.MonitorDefaults.DefaultRetries,
		},
	}
}

// ConvertProjectSettingsFromProto converts a protobuf ProjectSettings to its models representation
func ConvertProjectSettingsFromProto(ps *projectsv1.ProjectSettings) models.ProjectSettings {
	if ps == nil {
		return models.ProjectSettings{}
	}

	monitorDefaults := models.ProjectMonitorDefaults{}
	if ps.MonitorDefaults != nil {
		monitorDefaults = models.ProjectMonitorDefaults{
			DefaultInterval: ps.MonitorDefaults.Interval,
			DefaultTimeout:  ps.MonitorDefaults.Timeout,
			DefaultRetries:  ps.MonitorDefaults.Retries,
		}
	}

	return models.ProjectSettings{
		MonitorDefaults: monitorDefaults,
	}
}
