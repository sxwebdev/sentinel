package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sxwebdev/sentinel/internal/notifier"
	"github.com/sxwebdev/sentinel/internal/storage"
)

// MonitorService manages the monitoring state and business logic
type MonitorService struct {
	store     storage.Storage
	notifier  notifier.Notifier
	scheduler interface{} // Will be set to *scheduler.Scheduler
	mu        sync.RWMutex
}

// NewMonitorService creates a new monitor service
func NewMonitorService(store storage.Storage, notifier notifier.Notifier) *MonitorService {
	return &MonitorService{
		store:    store,
		notifier: notifier,
	}
}

// SetScheduler sets the scheduler instance (called from main.go)
func (m *MonitorService) SetScheduler(scheduler interface{}) {
	m.scheduler = scheduler
}

// LoadServicesFromStorage loads all services from storage and initializes monitoring
func (m *MonitorService) LoadServicesFromStorage(ctx context.Context) error {
	services, err := m.store.GetAllServices(ctx)
	if err != nil {
		return fmt.Errorf("failed to load services from storage: %w", err)
	}

	for _, service := range services {
		// Initialize service state if it doesn't exist
		if service.State == nil {
			if err := m.InitializeService(ctx, *service); err != nil {
				return fmt.Errorf("failed to initialize service %s: %w", service.Name, err)
			}
		}

		// Add to scheduler if available
		if m.scheduler != nil {
			if addServiceMethod, ok := m.scheduler.(interface {
				AddService(storage.Service) error
			}); ok {
				if err := addServiceMethod.AddService(*service); err != nil {
					return fmt.Errorf("failed to add service %s to scheduler: %w", service.Name, err)
				}
			}
		}
	}

	return nil
}

// AddService adds a new service and starts monitoring it
func (m *MonitorService) AddService(ctx context.Context, service *storage.Service) error {
	// Generate ID if not provided
	if service.ID == "" {
		service.ID = storage.GenerateULID()
	}

	// Initialize state if not provided
	if service.State == nil {
		nextCheck := time.Now().Add(service.Interval)
		service.State = &storage.ServiceState{
			Status:             storage.StatusUnknown,
			LastCheck:          nil,
			NextCheck:          &nextCheck,
			ConsecutiveFails:   0,
			ConsecutiveSuccess: 0,
			TotalChecks:        0,
		}
	}

	// Save to storage
	if err := m.store.SaveService(ctx, service); err != nil {
		return fmt.Errorf("failed to save service: %w", err)
	}

	// Add to scheduler if available
	if m.scheduler != nil {
		if addServiceMethod, ok := m.scheduler.(interface {
			AddServiceDynamic(context.Context, storage.Service) error
		}); ok {
			if err := addServiceMethod.AddServiceDynamic(ctx, *service); err != nil {
				return fmt.Errorf("failed to add service to scheduler: %w", err)
			}
		}
	}

	return nil
}

// UpdateService updates an existing service
func (m *MonitorService) UpdateService(ctx context.Context, service *storage.Service) error {
	// Update in storage
	if err := m.store.UpdateService(ctx, service); err != nil {
		return fmt.Errorf("failed to update service: %w", err)
	}

	// Update in scheduler if available
	if m.scheduler != nil {
		if updateServiceMethod, ok := m.scheduler.(interface {
			UpdateServiceDynamic(context.Context, storage.Service) error
		}); ok {
			if err := updateServiceMethod.UpdateServiceDynamic(ctx, *service); err != nil {
				return fmt.Errorf("failed to update service in scheduler: %w", err)
			}
		}
	}

	return nil
}

// DeleteService removes a service and stops monitoring it
func (m *MonitorService) DeleteService(ctx context.Context, id string) error {
	// Get service to find name for scheduler cleanup
	service, err := m.store.GetService(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get service: %w", err)
	}

	// Remove from scheduler if available
	if m.scheduler != nil {
		if removeServiceMethod, ok := m.scheduler.(interface {
			RemoveServiceDynamic(string) error
		}); ok {
			if err := removeServiceMethod.RemoveServiceDynamic(service.ID); err != nil {
				return fmt.Errorf("failed to remove service from scheduler: %w", err)
			}
		}
	}

	// Delete from storage
	if err := m.store.DeleteService(ctx, id); err != nil {
		return fmt.Errorf("failed to delete service: %w", err)
	}

	return nil
}

// GetServiceByID gets a service by ID
func (m *MonitorService) GetServiceByID(ctx context.Context, id string) (*storage.Service, error) {
	return m.store.GetService(ctx, id)
}

// GetAllServiceConfigs gets all service configurations
func (m *MonitorService) GetAllServiceConfigs(ctx context.Context) ([]*storage.Service, error) {
	return m.store.GetAllServices(ctx)
}

// InitializeService initializes monitoring state for a service
func (m *MonitorService) InitializeService(ctx context.Context, cfg storage.Service) error {
	// Get current service from database
	service, err := m.store.GetService(ctx, cfg.ID)
	if err != nil {
		return fmt.Errorf("failed to get service: %w", err)
	}

	// Initialize state
	nextCheck := time.Now().Add(cfg.Interval)
	service.State = &storage.ServiceState{
		Status:    storage.StatusUnknown,
		NextCheck: &nextCheck,
	}

	// Update service with new state
	if err := m.store.UpdateService(ctx, service); err != nil {
		return fmt.Errorf("failed to update service state for %s: %w", cfg.Name, err)
	}
	return nil
}

// InitializeWithActiveIncidents initializes the service and checks for active incidents
func (m *MonitorService) InitializeWithActiveIncidents(ctx context.Context, cfg storage.Service) error {
	// First initialize the service normally
	if err := m.InitializeService(ctx, cfg); err != nil {
		return fmt.Errorf("failed to initialize service: %w", err)
	}

	// Check for active incidents for this service
	activeIncidents, err := m.store.GetActiveIncidents(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active incidents during initialization for %s: %w", cfg.Name, err)
	}

	// Check if there's an active incident for this service
	hasActiveIncident := false
	for _, incident := range activeIncidents {
		if incident.ServiceID == cfg.ID && !incident.Resolved {
			hasActiveIncident = true
			break
		}
	}

	// If there's an active incident, set status to down initially
	if hasActiveIncident {
		service, err := m.store.GetService(ctx, cfg.ID)
		if err != nil {
			return fmt.Errorf("failed to get service: %w", err)
		}

		if service.State == nil {
			service.State = &storage.ServiceState{}
		}
		service.State.Status = storage.StatusDown

		if err := m.store.UpdateService(ctx, service); err != nil {
			return fmt.Errorf("failed to update service state for %s: %w", cfg.Name, err)
		}
	}

	return nil
}

// RecordSuccess records a successful health check
func (m *MonitorService) RecordSuccess(ctx context.Context, serviceID string, responseTime time.Duration) error {
	// Get current service from database
	service, err := m.store.GetService(ctx, serviceID)
	if err != nil {
		return fmt.Errorf("service %s not found in database: %w", serviceID, err)
	}

	// Initialize state if nil
	if service.State == nil {
		service.State = &storage.ServiceState{}
	}

	now := time.Now()
	wasDown := service.State.Status == storage.StatusDown

	// Update state
	service.State.Status = storage.StatusUp
	service.State.LastCheck = &now
	service.State.ResponseTime = responseTime
	service.State.ConsecutiveFails = 0
	service.State.ConsecutiveSuccess++
	service.State.TotalChecks++
	service.State.LastError = ""

	// Save to database
	if err := m.store.UpdateService(ctx, service); err != nil {
		return fmt.Errorf("failed to update service state for %s: %w", service.Name, err)
	}

	// If service was down and now up, resolve all active incidents for this service
	if wasDown {
		if err := m.resolveAllActiveIncidents(ctx, serviceID, service.Name); err != nil {
			return fmt.Errorf("failed to resolve active incidents: %w", err)
		}
	} else {
		// Even if service was already up, check for any lingering active incidents and resolve them
		if err := m.resolveAllActiveIncidents(ctx, serviceID, service.Name); err != nil {
			return fmt.Errorf("failed to resolve lingering incidents: %w", err)
		}
	}

	return nil
}

// RecordFailure records a failed health check
func (m *MonitorService) RecordFailure(ctx context.Context, serviceID string, checkErr error, responseTime time.Duration) error {
	// Get current service from database
	service, err := m.store.GetService(ctx, serviceID)
	if err != nil {
		return fmt.Errorf("service %s not found in database: %w", serviceID, err)
	}

	// Initialize state if nil
	if service.State == nil {
		service.State = &storage.ServiceState{}
	}

	now := time.Now()
	wasUp := service.State.Status == storage.StatusUp || service.State.Status == storage.StatusUnknown

	// Update state
	service.State.Status = storage.StatusDown
	service.State.LastCheck = &now
	service.State.ResponseTime = responseTime
	service.State.ConsecutiveFails++
	service.State.ConsecutiveSuccess = 0
	service.State.TotalChecks++
	service.State.LastError = checkErr.Error()

	// Save to database
	if err := m.store.UpdateService(ctx, service); err != nil {
		return fmt.Errorf("failed to update service state for %s: %w", service.Name, err)
	}

	// If service was up and now down, create incident and send alert
	if wasUp {
		if err := m.createIncident(ctx, serviceID, service.Name, checkErr); err != nil {
			return fmt.Errorf("failed to create incident: %w", err)
		}
	}

	return nil
}

// createIncident creates a new incident when a service goes down
func (m *MonitorService) createIncident(ctx context.Context, serviceID, serviceName string, err error) error {
	incident := &storage.Incident{
		ID:        storage.GenerateULID(),
		ServiceID: serviceID,
		StartTime: time.Now(),
		Error:     err.Error(),
		Resolved:  false,
	}

	// Save incident to storage
	if err := m.store.SaveIncident(ctx, incident); err != nil {
		return fmt.Errorf("failed to save incident for %s: %w", serviceName, err)
	}

	// Send alert notification
	if m.notifier != nil {
		if err := m.notifier.SendAlert(serviceName, incident); err != nil {
			return fmt.Errorf("failed to send alert for %s: %w", serviceName, err)
		}
	}

	return nil
}

// resolveActiveIncident resolves the active incident when a service recovers
func (m *MonitorService) resolveActiveIncident(ctx context.Context, serviceID string, serviceName string) error {
	// Get active incidents for the service
	incidents, err := m.store.GetActiveIncidents(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active incidents: %w", err)
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
			if err := m.store.UpdateIncident(ctx, incident); err != nil {
				return fmt.Errorf("failed to update incident: %w", err)
			}

			// Send recovery notification
			if m.notifier != nil {
				if err := m.notifier.SendRecovery(serviceName, incident); err != nil {
					return fmt.Errorf("failed to send recovery notification: %w", err)
				}
			}

			resolvedCount++
		}
	}

	return nil
}

// GetServiceIncidents gets incidents for a specific service
func (m *MonitorService) GetServiceIncidents(ctx context.Context, serviceID string) ([]*storage.Incident, error) {
	return m.store.GetIncidentsByService(ctx, serviceID)
}

// GetRecentIncidents gets recent incidents across all services
func (m *MonitorService) GetRecentIncidents(ctx context.Context, limit int) ([]*storage.Incident, error) {
	return m.store.GetRecentIncidents(ctx, limit)
}

// GetServiceStats gets statistics for a service
func (m *MonitorService) GetServiceStats(ctx context.Context, serviceID string, since time.Time) (*storage.ServiceStats, error) {
	return m.store.GetServiceStats(ctx, serviceID, since)
}

// TriggerCheck triggers a manual check for a service
func (m *MonitorService) TriggerCheck(ctx context.Context, serviceID string) error {
	// Get service to check if it exists
	_, err := m.GetServiceByID(ctx, serviceID)
	if err != nil {
		return fmt.Errorf("service not found: %w", err)
	}

	// Trigger check via scheduler if available
	if m.scheduler != nil {
		if triggerMethod, ok := m.scheduler.(interface {
			TriggerCheck(context.Context, string) error
		}); ok {
			return triggerMethod.TriggerCheck(ctx, serviceID)
		}
	}

	return fmt.Errorf("scheduler not available")
}

// resolveAllActiveIncidents resolves all active incidents for a service
func (m *MonitorService) resolveAllActiveIncidents(ctx context.Context, serviceID string, serviceName string) error {
	return m.resolveActiveIncident(ctx, serviceID, serviceName)
}

// ForceResolveIncidents manually resolves all active incidents for a service
func (m *MonitorService) ForceResolveIncidents(ctx context.Context, serviceID string, serviceName string) error {
	return m.resolveAllActiveIncidents(ctx, serviceID, serviceName)
}
