package repo_services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/sxwebdev/sentinel/pkg/dbutils"
)

type UpdateServiceRequest struct {
	Name      string                     `json:"name" yaml:"name"`
	Protocol  models.ServiceProtocolType `json:"protocol" yaml:"protocol"`
	Interval  dbutils.Duration           `json:"interval" yaml:"interval" swaggertype:"primitive,integer"`
	Timeout   dbutils.Duration           `json:"timeout" yaml:"timeout" swaggertype:"primitive,integer"`
	Retries   int64                      `json:"retries" yaml:"retries"`
	Tags      dbutils.JSONField          `json:"tags" yaml:"tags"`
	Config    dbutils.JSONField          `json:"config" yaml:"config"`
	IsEnabled bool                       `json:"is_enabled" yaml:"is_enabled"`
}

func (s *CustomQueries) Update(ctx context.Context, id string, service UpdateServiceRequest) (*models.ServiceFullView, error) {
	ub := sqlbuilder.NewUpdateBuilder()
	ub.Update("services")

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
		ub.Assign("name", service.Name),
		ub.Assign("protocol", service.Protocol),
		ub.Assign("interval", service.Interval.String()),
		ub.Assign("timeout", service.Timeout.String()),
		ub.Assign("retries", service.Retries),
		ub.Assign("tags", string(tagsJSON)),
		ub.Assign("config", string(configJSON)),
		ub.Assign("is_enabled", service.IsEnabled),
		ub.Assign("updated_at", time.Now()),
	}

	// Set all assignments at once
	ub.Set(assignments...)
	ub.Where(ub.Equal("id", id))

	sql, args := ub.Build()
	result, err := s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update service: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return nil, fmt.Errorf("service not found")
	}

	return s.GetViewByID(ctx, id)
}
