package monitor

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/sxwebdev/sentinel/internal/config"
	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/sxwebdev/sentinel/internal/notifier"
	"github.com/sxwebdev/sentinel/internal/receiver"
	"github.com/sxwebdev/sentinel/internal/services/baseservices"
	"github.com/sxwebdev/sentinel/internal/services/servicestate"
	"github.com/sxwebdev/sentinel/internal/storage"
	"github.com/sxwebdev/sentinel/internal/store"
	"github.com/sxwebdev/sentinel/internal/store/repos/repo_service_states"
	"github.com/sxwebdev/sentinel/internal/utils"
	"github.com/tkcrm/modules/pkg/db/dbutils"
)

// MonitorService handles service monitoring
type MonitorService struct {
	store        *store.Store
	storage      *storage.Storage
	config       *config.ConfigHub
	notifier     *notifier.Notifier
	receiver     *receiver.Receiver
	baseservices *baseservices.BaseServices
}

// NewMonitorService creates a new monitor service
func NewMonitorService(
	store *store.Store,
	storage *storage.Storage,
	config *config.ConfigHub,
	notifier *notifier.Notifier,
	receiver *receiver.Receiver,
	baseservices *baseservices.BaseServices,
) *MonitorService {
	return &MonitorService{
		store:        store,
		storage:      storage,
		config:       config,
		notifier:     notifier,
		receiver:     receiver,
		baseservices: baseservices,
	}
}

// FindServices loads all enabled services from storage and initializes monitoring
// func (m *MonitorService) FindServices(ctx context.Context, params storage.FindServicesParams) (storecmn.FindResponseWithCount[*storage.Service], error) {
// 	return m.storage.FindServices(ctx, params)
// }

// CreateService adds a new service and starts monitoring it
// func (m *MonitorService) CreateService(ctx context.Context, params storage.CreateUpdateServiceRequest) (*storage.Service, error) {
// 	if len(params.Tags) > 0 {
// 		slices.Sort(params.Tags)
// 	}

// 	// Save to storage
// 	svc, err := m.storage.CreateService(ctx, params)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create service: %w", err)
// 	}

// 	m.receiver.TriggerService().Publish(*receiver.NewTriggerServiceData(
// 		receiver.TriggerServiceEventTypeCreated,
// 		svc,
// 	))

// 	return svc, nil
// }

// UpdateService updates an existing service
// func (m *MonitorService) UpdateService(ctx context.Context, id string, params storage.CreateUpdateServiceRequest) (*models.ServiceFullView, error) {
// 	if len(params.Tags) > 0 {
// 		slices.Sort(params.Tags)
// 	}

// 	// Update in storage
// 	var err error
// 	_, err = m.storage.UpdateService(ctx, id, params)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to update service: %w", err)
// 	}

// 	svc, err := m.store.Services().GetViewByID(ctx, id)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get service: %w", err)
// 	}

// 	m.receiver.TriggerService().Publish(*receiver.NewTriggerServiceData(
// 		receiver.TriggerServiceEventTypeUpdated,
// 		svc,
// 	))

// 	return svc, nil
// }

// DeleteService removes a service and stops monitoring it
// func (m *MonitorService) DeleteService(ctx context.Context, id string) error {
// 	// Get service to find name for scheduler cleanup
// 	svc, err := m.store.Services().GetViewByID(ctx, id)
// 	if err != nil {
// 		return fmt.Errorf("failed to get service: %w", err)
// 	}

// 	m.receiver.TriggerService().Publish(*receiver.NewTriggerServiceData(
// 		receiver.TriggerServiceEventTypeDeleted,
// 		svc,
// 	))

// 	// Delete from storage
// 	if err := m.storage.DeleteService(ctx, id); err != nil {
// 		return fmt.Errorf("failed to delete service: %w", err)
// 	}

// 	return nil
// }

// GetServiceByID gets a service by ID
// func (m *MonitorService) GetServiceByID(ctx context.Context, id string) (*storage.Service, error) {
// 	return m.storage.GetServiceByID(ctx, id)
// }

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
			ResponseTimeNs:     utils.Pointer(responseTime.Nanoseconds()),
			ConsecutiveFails:   0,
			ConsecutiveSuccess: serviceState.ConsecutiveSuccess + 1,
			TotalChecks:        serviceState.TotalChecks + 1,
			LastError:          nil,
		},
		FieldMask: dbutils.FieldMask[repo_service_states.ColumnName]{
			repo_service_states.ColumnNameServiceStatesStatus,
			repo_service_states.ColumnNameServiceStatesLastCheck,
			repo_service_states.ColumnNameServiceStatesResponseTimeNs,
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
	service, err := m.store.Services().GetViewByID(ctx, serviceID)
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
			ResponseTimeNs:     utils.Pointer(responseTime.Nanoseconds()),
			ConsecutiveFails:   serviceState.ConsecutiveFails + 1,
			ConsecutiveSuccess: 0,
			TotalChecks:        serviceState.TotalChecks + 1,
			LastError:          utils.Pointer(checkErr.Error()),
		},
		FieldMask: dbutils.FieldMask[repo_service_states.ColumnName]{
			repo_service_states.ColumnNameServiceStatesStatus,
			repo_service_states.ColumnNameServiceStatesLastCheck,
			repo_service_states.ColumnNameServiceStatesResponseTimeNs,
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
	incident := &storage.Incident{
		ID:        utils.GenerateULID(),
		ServiceID: svc.ID,
		StartTime: time.Now(),
		Error:     err.Error(),
		Resolved:  false,
	}

	// Save incident to storage
	if err := m.storage.SaveIncident(ctx, incident); err != nil {
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
	svc, err := m.store.Services().GetViewByID(ctx, serviceID)
	if err != nil {
		return fmt.Errorf("failed to get service: %w", err)
	}

	// Resolve all active incidents for this service
	incidents, err := m.storage.ResolveAllIncidents(ctx, serviceID)
	if err != nil {
		return fmt.Errorf("failed to resolve incidents: %w", err)
	}

	for _, incident := range incidents {
		// Send recovery notification
		if m.notifier != nil {
			if err := m.notifier.SendRecovery(svc, incident); err != nil {
				err := fmt.Errorf("failed to send recovery notification for %s: %w", svc.Name, err)
				log.Println(err)
				return nil
			}
		}
	}

	return nil
}

// DeleteIncident deletes a specific incident
func (m *MonitorService) DeleteIncident(ctx context.Context, serviceID, incidentID string) error {
	// Delete the incident
	if err := m.storage.DeleteIncident(ctx, incidentID); err != nil {
		return fmt.Errorf("failed to delete incident: %w", err)
	}

	return nil
}

// TriggerCheck triggers a manual check for a service
func (m *MonitorService) TriggerCheck(ctx context.Context, id string) error {
	// Get service to check if it exists
	svc, err := m.store.Services().GetViewByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get service: %w", err)
	}

	m.receiver.TriggerService().Publish(*receiver.NewTriggerServiceData(
		receiver.TriggerServiceEventTypeCheck,
		svc,
	))

	return nil
}

// resolveAllActiveIncidents resolves all active incidents for a service
func (m *MonitorService) resolveAllActiveIncidents(ctx context.Context, serviceID string) error {
	return m.resolveActiveIncidents(ctx, serviceID)
}

// ForceResolveIncidents manually resolves all active incidents for a service
func (m *MonitorService) ForceResolveIncidents(ctx context.Context, serviceID string) error {
	return m.resolveAllActiveIncidents(ctx, serviceID)
}

// CheckService performs a health check on a service
func (m *MonitorService) CheckService(ctx context.Context, service *storage.Service) error {
	// Get current service state
	serviceState, err := m.baseservices.ServiceStates().GetByServiceID(ctx, service.ID)
	if err != nil {
		return fmt.Errorf("failed to get service state: %w", err)
	}

	// Perform the check (simplified - just record success/failure)
	startTime := time.Now()
	responseTime := time.Since(startTime)
	now := time.Now()

	// For now, just record success (this should be replaced with actual check logic)
	wasDown := serviceState.Status == models.StatusDown

	updateParams := servicestate.UpdateParams{
		ServiceState: models.ServiceState{
			Status:             models.StatusUp,
			LastCheck:          &now,
			ResponseTimeNs:     utils.Pointer(responseTime.Nanoseconds()),
			ConsecutiveFails:   0,
			ConsecutiveSuccess: serviceState.ConsecutiveSuccess + 1,
			TotalChecks:        serviceState.TotalChecks + 1,
			LastError:          nil,
		},
		FieldMask: dbutils.FieldMask[repo_service_states.ColumnName]{
			repo_service_states.ColumnNameServiceStatesStatus,
			repo_service_states.ColumnNameServiceStatesLastCheck,
			repo_service_states.ColumnNameServiceStatesResponseTimeNs,
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

	// Resolve incident if service was down before
	if wasDown {
		if err := m.resolveActiveIncidents(ctx, service.ID); err != nil {
			return fmt.Errorf("failed to resolve incident: %w", err)
		}
	}

	// Update service state
	if _, err := m.baseservices.ServiceStates().Update(ctx, serviceState.ID, updateParams); err != nil {
		return fmt.Errorf("failed to update service state: %w", err)
	}

	return nil
}
