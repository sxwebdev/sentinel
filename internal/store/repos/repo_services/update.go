package repo_services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/huandu/go-sqlbuilder"
	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/sxwebdev/sentinel/internal/store/storecmn"
)

type UpdateServiceRequest struct {
	Name                   string                     `json:"name"`
	Protocol               models.ServiceProtocolType `json:"protocol"`
	Interval               int64                      `json:"interval"`
	Timeout                int64                      `json:"timeout"`
	Retries                int64                      `json:"retries"`
	Tags                   storecmn.JSONField         `json:"tags"`
	Config                 storecmn.JSONField         `json:"config"`
	IsEnabled              bool                       `json:"is_enabled"`
	IsNotificationsEnabled bool                       `json:"is_notifications_enabled"`
}

func (s *CustomQueries) Update(ctx context.Context, id string, service UpdateServiceRequest) (*models.ServiceFullView, error) {
	ub := sqlbuilder.NewUpdateBuilder()
	ub.Update(TableNameServices.String())

	tagsJSON, err := json.Marshal(service.Tags)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tags: %w", err)
	}

	configJSON, err := json.Marshal(service.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	// Prepare all fields for update
	assignments := []string{
		ub.Assign(ColumnNameServicesName.String(), service.Name),
		ub.Assign(ColumnNameServicesProtocol.String(), service.Protocol),
		ub.Assign(ColumnNameServicesInterval.String(), service.Interval),
		ub.Assign(ColumnNameServicesTimeout.String(), service.Timeout),
		ub.Assign(ColumnNameServicesRetries.String(), service.Retries),
		ub.Assign(ColumnNameServicesTags.String(), string(tagsJSON)),
		ub.Assign(ColumnNameServicesConfig.String(), string(configJSON)),
		ub.Assign(ColumnNameServicesIsEnabled.String(), service.IsEnabled),
		ub.Assign(ColumnNameServicesIsNotificationsEnabled.String(), service.IsNotificationsEnabled),
		"updated_at = CURRENT_TIMESTAMP",
	}

	// Set all assignments at once
	ub.Set(assignments...)
	ub.Where(ub.Equal("id", id))

	sql, args := ub.Build()
	if _, err := s.db.ExecContext(ctx, sql, args...); err != nil {
		return nil, fmt.Errorf("failed to update service: %w", err)
	}

	return s.GetViewByID(ctx, id)
}
