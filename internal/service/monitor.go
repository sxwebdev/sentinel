package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/sxwebdev/sentinel/internal/config"
	"github.com/sxwebdev/sentinel/internal/notifier"
	"github.com/sxwebdev/sentinel/internal/receiver"
	"github.com/sxwebdev/sentinel/internal/storage"
)

// MonitorService handles service monitoring
type MonitorService struct {
	storage  storage.Storage
	config   *config.Config
	notifier notifier.Notifier
	receiver *receiver.Receiver
}

// NewMonitorService creates a new monitor service
func NewMonitorService(storage storage.Storage, config *config.Config, notifier notifier.Notifier, receiver *receiver.Receiver) *MonitorService {
	return &MonitorService{
		storage:  storage,
		config:   config,
		notifier: notifier,
		receiver: receiver,
	}
}

// FindServices loads all enabled services from storage and initializes monitoring
func (m *MonitorService) FindServices(ctx context.Context, params storage.FindServicesParams) ([]*storage.Service, error) {
	return m.storage.FindServices(ctx, params)
}

// CreateService adds a new service and starts monitoring it
func (m *MonitorService) CreateService(ctx context.Context, service storage.CreateUpdateServiceRequest) (*storage.Service, error) {
	// Save to storage
	svc, err := m.storage.CreateService(ctx, service)
	if err != nil {
		return nil, fmt.Errorf("failed to create service: %w", err)
	}

	m.receiver.TriggerService().Publish(*receiver.NewTriggerServiceData(
		receiver.TriggerServiceEventTypeCreated,
		svc,
	))

	return svc, nil
}

// UpdateService updates an existing service
func (m *MonitorService) UpdateService(ctx context.Context, id string, params storage.CreateUpdateServiceRequest) (*storage.Service, error) {
	// Update in storage
	svc, err := m.storage.UpdateService(ctx, id, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update service: %w", err)
	}

	m.receiver.TriggerService().Publish(*receiver.NewTriggerServiceData(
		receiver.TriggerServiceEventTypeUpdated,
		svc,
	))

	return svc, nil
}

// DeleteService removes a service and stops monitoring it
func (m *MonitorService) DeleteService(ctx context.Context, id string) error {
	// Get service to find name for scheduler cleanup
	svc, err := m.storage.GetServiceByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get service: %w", err)
	}

	m.receiver.TriggerService().Publish(*receiver.NewTriggerServiceData(
		receiver.TriggerServiceEventTypeDeleted,
		svc,
	))

	// Delete from storage
	if err := m.storage.DeleteService(ctx, id); err != nil {
		return fmt.Errorf("failed to delete service: %w", err)
	}

	return nil
}

// GetServiceByID gets a service by ID
func (m *MonitorService) GetServiceByID(ctx context.Context, id string) (*storage.Service, error) {
	return m.storage.GetServiceByID(ctx, id)
}

// RecordSuccess records a successful check for a service
func (m *MonitorService) RecordSuccess(ctx context.Context, serviceID string, responseTime time.Duration) error {
	// Get current service from database
	service, err := m.storage.GetServiceByID(ctx, serviceID)
	if err != nil {
		return fmt.Errorf("service %s not found in database: %w", serviceID, err)
	}

	// Get current service state
	serviceState, err := m.storage.GetServiceState(ctx, serviceID)
	if err != nil {
		return fmt.Errorf("failed to get service state: %w", err)
	}

	// Update state
	now := time.Now()
	serviceState.Status = storage.StatusUp
	serviceState.LastCheck = &now
	serviceState.ResponseTimeNS = &[]int64{responseTime.Nanoseconds()}[0]
	serviceState.ConsecutiveFails = 0
	serviceState.ConsecutiveSuccess++
	serviceState.TotalChecks++
	serviceState.LastError = ""

	// Save to database
	if err := m.storage.UpdateServiceState(ctx, serviceState); err != nil {
		return fmt.Errorf("failed to update service state for %s: %w", service.Name, err)
	}

	// Resolve any active incidents
	if err := m.resolveActiveIncident(ctx, serviceID); err != nil {
		return fmt.Errorf("failed to resolve incident: %w", err)
	}

	return nil
}

// RecordFailure records a failed check for a service
func (m *MonitorService) RecordFailure(ctx context.Context, serviceID string, checkErr error, responseTime time.Duration) error {
	// Get current service from database
	service, err := m.storage.GetServiceByID(ctx, serviceID)
	if err != nil {
		return fmt.Errorf("service %s not found in database: %w", serviceID, err)
	}

	// Get current service state
	serviceState, err := m.storage.GetServiceState(ctx, serviceID)
	if err != nil {
		return fmt.Errorf("failed to get service state: %w", err)
	}

	// Update state
	now := time.Now()
	wasUp := serviceState.Status == storage.StatusUp || serviceState.Status == storage.StatusUnknown

	serviceState.Status = storage.StatusDown
	serviceState.LastCheck = &now
	serviceState.ResponseTimeNS = &[]int64{responseTime.Nanoseconds()}[0]
	serviceState.ConsecutiveFails++
	serviceState.ConsecutiveSuccess = 0
	serviceState.TotalChecks++
	serviceState.LastError = checkErr.Error()

	// Save to database
	if err := m.storage.UpdateServiceState(ctx, serviceState); err != nil {
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
func (m *MonitorService) createIncident(ctx context.Context, svc *storage.Service, err error) error {
	incident := &storage.Incident{
		ID:        storage.GenerateULID(),
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

// resolveActiveIncident resolves the active incident when a service recovers
func (m *MonitorService) resolveActiveIncident(ctx context.Context, serviceID string) error {
	// Get active incidents for the service
	incidents, err := m.storage.GetActiveIncidents(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active incidents: %w", err)
	}

	svc, err := m.storage.GetServiceByID(ctx, serviceID)
	if err != nil {
		return fmt.Errorf("failed to get service: %w", err)
	}

	// Find and resolve all active incidents for this service
	resolvedCount := 0
	for _, incident := range incidents {
		if incident.ServiceID == serviceID && !incident.Resolved {
			// Resolve the incident
			now := time.Now()
			duration := now.Sub(incident.StartTime)

			incident.EndTime = &now
			incident.Duration = &duration
			incident.Resolved = true

			// Update incident in storage
			if err := m.storage.UpdateIncident(ctx, incident); err != nil {
				return fmt.Errorf("failed to update incident: %w", err)
			}

			// Send recovery notification
			if m.notifier != nil {
				if err := m.notifier.SendRecovery(svc, incident); err != nil {
					err := fmt.Errorf("failed to send recovery notification for %s: %w", svc.Name, err)
					log.Println(err)
					return nil
				}
			}

			resolvedCount++
		}
	}

	return nil
}

// GetServiceIncidents gets incidents for a specific service
func (m *MonitorService) GetServiceIncidents(ctx context.Context, serviceID string) ([]*storage.Incident, error) {
	return m.storage.GetIncidentsByService(ctx, serviceID)
}

// GetRecentIncidents gets recent incidents across all services
func (m *MonitorService) GetRecentIncidents(ctx context.Context, limit int) ([]*storage.Incident, error) {
	return m.storage.GetRecentIncidents(ctx, limit)
}

// DeleteIncident deletes a specific incident
func (m *MonitorService) DeleteIncident(ctx context.Context, serviceID, incidentID string) error {
	// Delete the incident
	if err := m.storage.DeleteIncident(ctx, incidentID); err != nil {
		return fmt.Errorf("failed to delete incident: %w", err)
	}

	return nil
}

// GetServiceStats gets statistics for a service
func (m *MonitorService) GetServiceStats(ctx context.Context, serviceID string, since time.Time) (*storage.ServiceStats, error) {
	return m.storage.GetServiceStats(ctx, serviceID, since)
}

// TriggerCheck triggers a manual check for a service
func (m *MonitorService) TriggerCheck(ctx context.Context, serviceID string) error {
	// Get service to check if it exists
	svc, err := m.GetServiceByID(ctx, serviceID)
	if err != nil {
		return err
	}

	m.receiver.TriggerService().Publish(*receiver.NewTriggerServiceData(
		receiver.TriggerServiceEventTypeCheck,
		svc,
	))

	return nil
}

// resolveAllActiveIncidents resolves all active incidents for a service
func (m *MonitorService) resolveAllActiveIncidents(ctx context.Context, serviceID string) error {
	return m.resolveActiveIncident(ctx, serviceID)
}

// ForceResolveIncidents manually resolves all active incidents for a service
func (m *MonitorService) ForceResolveIncidents(ctx context.Context, serviceID string) error {
	return m.resolveAllActiveIncidents(ctx, serviceID)
}

// CheckService performs a health check on a service
func (m *MonitorService) CheckService(ctx context.Context, service *storage.Service) error {
	// Get current service state
	serviceState, err := m.storage.GetServiceState(ctx, service.ID)
	if err != nil {
		return fmt.Errorf("failed to get service state: %w", err)
	}

	// Initialize state if not exists
	if serviceState == nil {
		serviceState = &storage.ServiceStateRecord{
			ID:                 storage.GenerateULID(),
			ServiceID:          service.ID,
			Status:             storage.StatusUnknown,
			ConsecutiveFails:   0,
			ConsecutiveSuccess: 0,
			TotalChecks:        0,
		}
	}

	// Perform the check (simplified - just record success/failure)
	startTime := time.Now()
	responseTime := time.Since(startTime)
	now := time.Now()

	// For now, just record success (this should be replaced with actual check logic)
	wasDown := serviceState.Status == storage.StatusDown

	serviceState.Status = storage.StatusUp
	serviceState.LastCheck = &now
	serviceState.ResponseTimeNS = &[]int64{responseTime.Nanoseconds()}[0]
	serviceState.ConsecutiveFails = 0
	serviceState.ConsecutiveSuccess++
	serviceState.TotalChecks++
	serviceState.LastError = ""

	// Resolve incident if service was down before
	if wasDown {
		if err := m.resolveActiveIncident(ctx, service.ID); err != nil {
			return fmt.Errorf("failed to resolve incident: %w", err)
		}
	}

	// Update service state
	if err := m.storage.UpdateServiceState(ctx, serviceState); err != nil {
		return fmt.Errorf("failed to update service state: %w", err)
	}

	return nil
}
