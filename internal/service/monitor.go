package service

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/sxwebdev/sentinel/internal/config"
	"github.com/sxwebdev/sentinel/internal/notifier"
	"github.com/sxwebdev/sentinel/internal/storage"
)

// MonitorService manages the monitoring state and business logic
type MonitorService struct {
	storage  storage.Storage
	notifier notifier.Notifier
	states   map[string]*config.ServiceState
	mu       sync.RWMutex
}

// NewMonitorService creates a new monitor service
func NewMonitorService(storage storage.Storage, notifier notifier.Notifier) *MonitorService {
	return &MonitorService{
		storage:  storage,
		notifier: notifier,
		states:   make(map[string]*config.ServiceState),
	}
}

// InitializeService initializes monitoring state for a service
func (m *MonitorService) InitializeService(cfg config.ServiceConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()

	state := &config.ServiceState{
		Name:               cfg.Name,
		Protocol:           cfg.Protocol,
		Endpoint:           cfg.Endpoint,
		Status:             config.StatusUnknown,
		LastCheck:          time.Time{},
		NextCheck:          time.Now().Add(cfg.Interval),
		ConsecutiveFails:   0,
		ConsecutiveSuccess: 0,
		TotalChecks:        0,
		Tags:               cfg.Tags,
	}

	m.states[cfg.Name] = state
	log.Printf("Initialized monitoring for service %s", cfg.Name)
}

// InitializeWithActiveIncidents initializes the service and checks for active incidents
func (m *MonitorService) InitializeWithActiveIncidents(ctx context.Context, cfg config.ServiceConfig) {
	// First initialize the service normally
	m.InitializeService(cfg)

	// Check for active incidents for this service
	activeIncidents, err := m.storage.GetActiveIncidents(ctx)
	if err != nil {
		log.Printf("Failed to get active incidents during initialization for %s: %v", cfg.Name, err)
		return
	}

	log.Println("Active incidents: ", len(activeIncidents))

	// Check if there's an active incident for this service
	hasActiveIncident := false
	for _, incident := range activeIncidents {
		if incident.ServiceName == cfg.Name && !incident.Resolved {
			hasActiveIncident = true
			break
		}
	}

	// If there's an active incident, set status to down initially
	// The actual check will determine if the service is really down or if the incident should be resolved
	if hasActiveIncident {
		m.mu.Lock()
		if state, exists := m.states[cfg.Name]; exists {
			state.Status = config.StatusDown
			log.Printf("Service %s has active incident, status set to DOWN initially", cfg.Name)
		}
		m.mu.Unlock()
	}
}

// RecordSuccess records a successful health check
func (m *MonitorService) RecordSuccess(ctx context.Context, serviceName string, responseTime time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, exists := m.states[serviceName]
	if !exists {
		log.Printf("Service %s not found in states", serviceName)
		return
	}

	now := time.Now()
	wasDown := state.Status == config.StatusDown

	// Update state
	state.Status = config.StatusUp
	state.LastCheck = now
	state.ResponseTime = responseTime
	state.ConsecutiveFails = 0
	state.ConsecutiveSuccess++
	state.TotalChecks++
	state.LastError = ""

	log.Printf("Service %s: SUCCESS (response time: %v)", serviceName, responseTime)

	// If service was down and now up, resolve all active incidents for this service
	if wasDown {
		m.resolveAllActiveIncidents(ctx, serviceName)
	} else {
		// Even if service was already up, check for any lingering active incidents and resolve them
		m.resolveAllActiveIncidents(ctx, serviceName)
	}
}

// RecordFailure records a failed health check
func (m *MonitorService) RecordFailure(ctx context.Context, serviceName string, err error, responseTime time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, exists := m.states[serviceName]
	if !exists {
		log.Printf("Service %s not found in states", serviceName)
		return
	}

	now := time.Now()
	wasUp := state.Status == config.StatusUp || state.Status == config.StatusUnknown

	// Update state
	state.Status = config.StatusDown
	state.LastCheck = now
	state.ResponseTime = responseTime
	state.ConsecutiveFails++
	state.ConsecutiveSuccess = 0
	state.TotalChecks++
	state.LastError = err.Error()

	log.Printf("Service %s: FAILED (attempt %d) - %v", serviceName, state.ConsecutiveFails, err)

	// If service was up and now down, create incident and send alert
	if wasUp {
		m.createIncident(ctx, serviceName, err)
	}
}

// createIncident creates a new incident when a service goes down
func (m *MonitorService) createIncident(ctx context.Context, serviceName string, err error) {
	incident := &config.Incident{
		ID:          generateIncidentID(),
		ServiceName: serviceName,
		StartTime:   time.Now(),
		Error:       err.Error(),
		Resolved:    false,
	}

	// Save incident to storage
	if err := m.storage.SaveIncident(ctx, incident); err != nil {
		log.Printf("Failed to save incident for %s: %v", serviceName, err)
	}

	// Send alert notification
	if m.notifier != nil {
		if err := m.notifier.SendAlert(serviceName, incident); err != nil {
			log.Printf("Failed to send alert for %s: %v", serviceName, err)
		}
	}

	log.Printf("Created incident %s for service %s", incident.ID, serviceName)
}

// resolveActiveIncident resolves the active incident when a service recovers
func (m *MonitorService) resolveActiveIncident(ctx context.Context, serviceName string) {
	// Get active incidents for the service
	incidents, err := m.storage.GetActiveIncidents(ctx)
	if err != nil {
		log.Printf("Failed to get active incidents: %v", err)
		return
	}

	// Find and resolve all active incidents for this service
	resolvedCount := 0
	for _, incident := range incidents {
		if incident.ServiceName == serviceName && !incident.Resolved {
			// Resolve the incident
			now := time.Now()
			duration := now.Sub(incident.StartTime)

			incident.EndTime = &now
			incident.Duration = &duration
			incident.Resolved = true

			// Update incident in storage
			if err := m.storage.UpdateIncident(ctx, incident); err != nil {
				log.Printf("Failed to update incident %s: %v", incident.ID, err)
				continue
			}

			// Send recovery notification
			if m.notifier != nil {
				if err := m.notifier.SendRecovery(serviceName, incident); err != nil {
					log.Printf("Failed to send recovery notification for %s: %v", serviceName, err)
				}
			}

			log.Printf("Resolved incident %s for service %s (downtime: %v)", incident.ID, serviceName, duration)
			resolvedCount++
		}
	}

	if resolvedCount > 0 {
		log.Printf("Resolved %d active incidents for service %s", resolvedCount, serviceName)
	}
}

// GetServiceState returns the current state of a service
func (m *MonitorService) GetServiceState(serviceName string) (*config.ServiceState, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	state, exists := m.states[serviceName]
	if !exists {
		return nil, fmt.Errorf("service %s not found", serviceName)
	}

	// Return a copy to avoid race conditions
	stateCopy := *state
	return &stateCopy, nil
}

// GetAllServiceStates returns the current state of all services
func (m *MonitorService) GetAllServiceStates() map[string]*config.ServiceState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	states := make(map[string]*config.ServiceState)
	for name, state := range m.states {
		// Return copies to avoid race conditions
		stateCopy := *state
		states[name] = &stateCopy
	}

	return states
}

// GetServiceIncidents returns incidents for a specific service
func (m *MonitorService) GetServiceIncidents(ctx context.Context, serviceName string) ([]*config.Incident, error) {
	return m.storage.GetIncidentsByService(ctx, serviceName)
}

// GetRecentIncidents returns recent incidents across all services
func (m *MonitorService) GetRecentIncidents(ctx context.Context, limit int) ([]*config.Incident, error) {
	return m.storage.GetRecentIncidents(ctx, limit)
}

// GetServiceStats returns statistics for a service
func (m *MonitorService) GetServiceStats(ctx context.Context, serviceName string, since time.Time) (*storage.ServiceStats, error) {
	return m.storage.GetServiceStats(ctx, serviceName, since)
}

// generateIncidentID generates a unique incident ID
func generateIncidentID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// GetAllServices returns all services (alias for GetAllServiceStates for web compatibility)
func (m *MonitorService) GetAllServices() (map[string]*config.ServiceState, error) {
	states := m.GetAllServiceStates()
	return states, nil
}

// TriggerCheck manually triggers a health check for a service
func (m *MonitorService) TriggerCheck(ctx context.Context, serviceName string) error {
	// Get service state to check if it exists
	state, err := m.GetServiceState(serviceName)
	if err != nil {
		return fmt.Errorf("service %s not found", serviceName)
	}

	log.Printf("Manual check triggered for service %s (current status: %s)", serviceName, state.Status)

	// Note: In a real implementation, you would want to trigger an immediate check
	// For now, we just log the request. The actual check will happen on the next scheduled interval.
	// To implement immediate checks, you would need to add a channel or method to communicate
	// with the scheduler to trigger an immediate check.

	return nil
}

// resolveAllActiveIncidents resolves all active incidents for a service
func (m *MonitorService) resolveAllActiveIncidents(ctx context.Context, serviceName string) {
	// Get active incidents for the service
	incidents, err := m.storage.GetActiveIncidents(ctx)
	if err != nil {
		log.Printf("Failed to get active incidents: %v", err)
		return
	}

	// Find active incidents for this service
	for _, incident := range incidents {
		if incident.ServiceName == serviceName && !incident.Resolved {
			// Resolve the incident
			m.resolveActiveIncident(ctx, serviceName)
		}
	}
}

// ForceResolveIncidents forcibly resolves all active incidents for a service
func (m *MonitorService) ForceResolveIncidents(ctx context.Context, serviceName string) error {
	// Get service state to check if it exists
	state, err := m.GetServiceState(serviceName)
	if err != nil {
		return fmt.Errorf("service %s not found", serviceName)
	}

	log.Printf("Force resolving incidents for service %s (current status: %s)", serviceName, state.Status)

	// Resolve all active incidents for this service
	m.resolveActiveIncident(ctx, serviceName)

	return nil
}
