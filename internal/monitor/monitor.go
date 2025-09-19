package monitor

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/sxwebdev/sentinel/internal/config"
	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/sxwebdev/sentinel/internal/notifier"
	"github.com/sxwebdev/sentinel/internal/receiver"
	"github.com/sxwebdev/sentinel/internal/services/baseservices"
	"github.com/sxwebdev/sentinel/internal/services/incidents"
	"github.com/sxwebdev/sentinel/internal/services/servicestate"
	"github.com/sxwebdev/sentinel/internal/store"
	"github.com/sxwebdev/sentinel/internal/store/repos/repo_service_states"
	"github.com/sxwebdev/sentinel/internal/store/storecmn"
	"github.com/sxwebdev/sentinel/internal/utils"
	"github.com/tkcrm/modules/pkg/db/dbutils"
	"github.com/tkcrm/mx/logger"
)

// MonitorService handles service monitoring
type MonitorService struct {
	logger       logger.Logger
	store        *store.Store
	config       *config.ConfigHub
	notifier     *notifier.Notifier
	receiver     *receiver.Receiver
	baseservices *baseservices.BaseServices
}

// NewMonitorService creates a new monitor service
func NewMonitorService(
	logger logger.Logger,
	store *store.Store,
	config *config.ConfigHub,
	notifier *notifier.Notifier,
	receiver *receiver.Receiver,
	baseservices *baseservices.BaseServices,
) *MonitorService {
	return &MonitorService{
		logger:       logger,
		store:        store,
		config:       config,
		notifier:     notifier,
		receiver:     receiver,
		baseservices: baseservices,
	}
}

// RecordSuccess records a successful check for a service
func (m *MonitorService) RecordSuccess(ctx context.Context, serviceID string, responseTime time.Duration) error {
	// Get current service state
	serviceState, err := m.baseservices.ServiceStates().GetByServiceID(ctx, serviceID)
	if err != nil {
		return fmt.Errorf("failed to get service state: %w", err)
	}

	// Update state
	now := time.Now()
	updateParams := servicestate.UpdateParams{
		ServiceState: models.ServiceState{
			Status:             models.StatusUp,
			LastCheck:          &now,
			ResponseTime:       utils.Pointer(responseTime.Milliseconds()),
			ConsecutiveFails:   0,
			ConsecutiveSuccess: serviceState.ConsecutiveSuccess + 1,
			TotalChecks:        serviceState.TotalChecks + 1,
			LastError:          nil,
		},
		FieldMask: dbutils.FieldMask[repo_service_states.ColumnName]{
			repo_service_states.ColumnNameServiceStatesStatus,
			repo_service_states.ColumnNameServiceStatesLastCheck,
			repo_service_states.ColumnNameServiceStatesResponseTime,
			repo_service_states.ColumnNameServiceStatesConsecutiveFails,
			repo_service_states.ColumnNameServiceStatesConsecutiveSuccess,
			repo_service_states.ColumnNameServiceStatesTotalChecks,
			repo_service_states.ColumnNameServiceStatesLastError,
		},
	}

	// Save to database
	if _, err := m.baseservices.ServiceStates().Update(ctx, serviceState.ID, updateParams); err != nil {
		return fmt.Errorf("failed to update service state: %w", err)
	}

	// Resolve any active incidents
	if err := m.resolveActiveIncidents(ctx, serviceID); err != nil {
		return err
	}

	return nil
}

// RecordFailure records a failed check for a service
func (m *MonitorService) RecordFailure(ctx context.Context, serviceID string, checkErr error, responseTime time.Duration) error {
	// Get current service from database
	service, err := m.baseservices.Services().GetViewByID(ctx, serviceID)
	if err != nil {
		return fmt.Errorf("service %s not found in database: %w", serviceID, err)
	}

	// Get current service state
	serviceState, err := m.baseservices.ServiceStates().GetByServiceID(ctx, serviceID)
	if err != nil {
		return fmt.Errorf("failed to get service state: %w", err)
	}

	// Update state
	now := time.Now()
	wasUp := serviceState.Status == models.StatusUp || serviceState.Status == models.StatusUnknown

	updateParams := servicestate.UpdateParams{
		ServiceState: models.ServiceState{
			Status:             models.StatusDown,
			LastCheck:          &now,
			ResponseTime:       utils.Pointer(responseTime.Milliseconds()),
			ConsecutiveFails:   serviceState.ConsecutiveFails + 1,
			ConsecutiveSuccess: 0,
			TotalChecks:        serviceState.TotalChecks + 1,
			LastError:          utils.Pointer(checkErr.Error()),
		},
		FieldMask: dbutils.FieldMask[repo_service_states.ColumnName]{
			repo_service_states.ColumnNameServiceStatesStatus,
			repo_service_states.ColumnNameServiceStatesLastCheck,
			repo_service_states.ColumnNameServiceStatesResponseTime,
			repo_service_states.ColumnNameServiceStatesConsecutiveFails,
			repo_service_states.ColumnNameServiceStatesConsecutiveSuccess,
			repo_service_states.ColumnNameServiceStatesTotalChecks,
			repo_service_states.ColumnNameServiceStatesLastError,
		},
	}

	// Save to database
	if _, err := m.baseservices.ServiceStates().Update(ctx, serviceState.ID, updateParams); err != nil {
		return fmt.Errorf("failed to update service state for %s: %w", service.Name, err)
	}

	// Create incident if service was up before
	if wasUp {
		if err := m.createIncident(ctx, service, checkErr); err != nil {
			return fmt.Errorf("failed to create incident: %w", err)
		}
	}

	return nil
}

// createIncident creates a new incident when a service goes down
func (m *MonitorService) createIncident(ctx context.Context, svc *models.ServiceFullView, err error) error {
	// Save incident to storage
	incident, err := m.baseservices.Incidents().Create(ctx, incidents.CreateParams{
		ServiceID: svc.ID,
		Error:     err.Error(),
	})
	if err != nil {
		return fmt.Errorf("failed to save incident for %s: %w", svc.Name, err)
	}

	// Send alert notification
	if m.notifier != nil {
		if err := m.notifier.SendAlert(svc, incident); err != nil {
			err := fmt.Errorf("failed to send alert notification for %s: %w", svc.Name, err)
			log.Println(err)
			return nil
		}
	}

	return nil
}

// resolveActiveIncidents resolves the active incident when a service recovers
func (m *MonitorService) resolveActiveIncidents(ctx context.Context, serviceID string) error {
	// Get service
	svc, err := m.baseservices.Services().GetViewByID(ctx, serviceID)
	if err != nil {
		return fmt.Errorf("failed to get service: %w", err)
	}

	// Resolve all active incidents for this service
	incidents, err := m.baseservices.Incidents().GetAllUnresolvedByServiceID(ctx, serviceID)
	if err != nil {
		return fmt.Errorf("failed to resolve incidents: %w", err)
	}

	err = storecmn.WrapTx(ctx, m.store.SQLite(), func(txCtx *sql.Tx) error {
		for _, incident := range incidents {
			resolverIncident, err := m.baseservices.Incidents().ResolveByID(ctx, incident.ID)
			if err != nil {
				return fmt.Errorf("failed to resolve incident %s: %w", incident.ID, err)
			}

			if m.notifier != nil {
				if err := m.notifier.SendRecovery(svc, resolverIncident); err != nil {
					m.logger.Errorf("failed to send recovery notification for %s: %v", svc.Name, err)
					return nil
				}
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to resolve incidents in transaction: %w", err)
	}

	return nil
}
