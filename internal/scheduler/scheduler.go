package scheduler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sxwebdev/sentinel/internal/checker"
	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/sxwebdev/sentinel/internal/receiver"
	"github.com/sxwebdev/sentinel/internal/services/baseservices"
	"github.com/sxwebdev/sentinel/internal/services/incidents"
	"github.com/sxwebdev/sentinel/internal/services/servicestate"
	"github.com/sxwebdev/sentinel/internal/store/repos/repo_service_states"
	"github.com/sxwebdev/sentinel/internal/utils"
	"github.com/tkcrm/modules/pkg/db/dbutils"
	"github.com/tkcrm/mx/logger"
)

// Scheduler manages the monitoring of multiple services
type Scheduler struct {
	logger logger.Logger

	checker *checker.Checker

	receiver     *receiver.Receiver
	baseservices *baseservices.BaseServices
}

// New creates a new scheduler
func New(
	l logger.Logger,
	receiver *receiver.Receiver,
	baseServices *baseservices.BaseServices,
) *Scheduler {
	s := &Scheduler{
		logger:       l,
		receiver:     receiver,
		baseservices: baseServices,
	}

	s.checker = checker.New(
		l,
		receiver,
		s.onSuccess,
		s.onFailure,
	)

	return s
}

// Name returns the name of the scheduler
func (s *Scheduler) Name() string { return "scheduler" }

// Start begins monitoring all configured services
func (s *Scheduler) Start(ctx context.Context) error {
	// Load enabled services from storage
	services, err := s.baseservices.Services().GetAllEnabledWithoutAgents(ctx)
	if err != nil {
		return fmt.Errorf("failed to load services: %w", err)
	}

	s.logger.Infof("starting scheduler with %d enabled services", len(services))

	// Get all services under read lock
	for _, svc := range services {
		s.checker.AddService(ctx, checker.AddServiceParams{
			ID:        svc.ID,
			Name:      svc.Name,
			Protocol:  svc.Protocol,
			IsEnabled: svc.IsEnabled,
			Interval:  time.Duration(svc.Interval) * time.Millisecond,
			Timeout:   time.Duration(svc.Timeout) * time.Millisecond,
			Retries:   svc.Retries,
			Config:    svc.Config.ConvertToMap(),
		})
	}

	return s.checker.Start(ctx)
}

func (s *Scheduler) Stop(ctx context.Context) error {
	return s.checker.Stop(ctx)
}

// onSuccess is called when a service check succeeds
func (s *Scheduler) onSuccess(ctx context.Context, serviceID string, responseTime time.Duration) error {
	// Success - record the time of this successful attempt
	if err := s.recordSuccess(ctx, serviceID, responseTime); err != nil {
		return fmt.Errorf("failed to record success: %w", err)
	}

	svc, err := s.baseservices.Services().GetViewByID(ctx, serviceID)
	if err != nil {
		return fmt.Errorf("failed to get service view: %w", err)
	}

	// Publish update to receiver
	s.receiver.TriggerService().Publish(*receiver.NewTriggerServiceData(
		receiver.TriggerServiceEventTypeUpdatedState,
		svc,
	))

	return nil
}

// onFailure is called when a service check fails
func (s *Scheduler) onFailure(ctx context.Context, serviceID string, checkErr error, responseTime time.Duration) error {
	// All attempts failed - record the time of the last attempt
	if err := s.recordFailure(ctx, serviceID, checkErr, responseTime); err != nil {
		return fmt.Errorf("failed to record failure: %w", err)
	}

	svc, err := s.baseservices.Services().GetViewByID(ctx, serviceID)
	if err != nil {
		return fmt.Errorf("failed to get service view: %w", err)
	}

	// Publish update to receiver
	s.receiver.TriggerService().Publish(*receiver.NewTriggerServiceData(
		receiver.TriggerServiceEventTypeUpdatedState,
		svc,
	))

	return nil
}

// RecordSuccess records a successful check for a service
func (m *Scheduler) recordSuccess(ctx context.Context, serviceID string, responseTime time.Duration) error {
	// Get current service state
	serviceState, err := m.baseservices.ServiceStates().GetByServiceID(ctx, serviceID)
	if err != nil {
		return fmt.Errorf("failed to get service state: %w", err)
	}

	// Update state
	now := time.Now()
	updateParams := servicestate.UpdateParams{
		ServiceState: models.ServiceState{
			Status:             models.ServiceStatusUp,
			LastCheck:          &now,
			AvgResponseTime:    utils.Pointer(responseTime.Milliseconds()),
			ConsecutiveFails:   0,
			ConsecutiveSuccess: serviceState.ConsecutiveSuccess + 1,
			TotalChecks:        serviceState.TotalChecks + 1,
			LastError:          nil,
		},
		FieldMask: dbutils.FieldMask[repo_service_states.ColumnName]{
			repo_service_states.ColumnNameServiceStatesStatus,
			repo_service_states.ColumnNameServiceStatesLastCheck,
			repo_service_states.ColumnNameServiceStatesAvgResponseTime,
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
func (m *Scheduler) recordFailure(ctx context.Context, serviceID string, checkErr error, responseTime time.Duration) error {
	// Get current service state
	serviceState, err := m.baseservices.ServiceStates().GetByServiceID(ctx, serviceID)
	if err != nil {
		return fmt.Errorf("failed to get service state: %w", err)
	}

	// Update state
	now := time.Now()
	wasUp := serviceState.Status == models.ServiceStatusUp || serviceState.Status == models.ServiceStatusUnknown

	updateParams := servicestate.UpdateParams{
		ServiceState: models.ServiceState{
			Status:             models.ServiceStatusDown,
			LastCheck:          &now,
			AvgResponseTime:    utils.Pointer(responseTime.Milliseconds()),
			ConsecutiveFails:   serviceState.ConsecutiveFails + 1,
			ConsecutiveSuccess: 0,
			TotalChecks:        serviceState.TotalChecks + 1,
			LastError:          utils.Pointer(checkErr.Error()),
		},
		FieldMask: dbutils.FieldMask[repo_service_states.ColumnName]{
			repo_service_states.ColumnNameServiceStatesStatus,
			repo_service_states.ColumnNameServiceStatesLastCheck,
			repo_service_states.ColumnNameServiceStatesAvgResponseTime,
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

	// Create incident if service was up before
	if wasUp {
		if err := m.createIncident(ctx, serviceID, checkErr); err != nil {
			return fmt.Errorf("failed to create incident: %w", err)
		}
	}

	return nil
}

// createIncident creates a new incident when a service goes down
func (m *Scheduler) createIncident(ctx context.Context, serviceID string, err error) error {
	// Save incident to storage
	incident, err := m.baseservices.Incidents().Create(ctx, incidents.CreateParams{
		ServiceID: serviceID,
		Error:     err.Error(),
	})
	if err != nil {
		return fmt.Errorf("failed to save incident: %w", err)
	}

	// Get current service from database
	svc, err := m.baseservices.Services().GetViewByID(ctx, serviceID)
	if err != nil {
		return fmt.Errorf("service %s not found in database: %w", serviceID, err)
	}

	// Send alert notification
	message := m.formatAlertMessage(svc, incident)
	if err := m.baseservices.Notifications().History().SendAlert(ctx, serviceID, incident.ID, message); err != nil {
		m.logger.Errorf("failed to send alert notification: %v", err)
	}

	return nil
}

// resolveActiveIncidents resolves the active incident when a service recovers
func (m *Scheduler) resolveActiveIncidents(ctx context.Context, serviceID string) error {
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

	if len(incidents) == 0 {
		// No active incidents to resolve
		return nil
	}

	for _, incident := range incidents {
		resolverIncident, err := m.baseservices.Incidents().ResolveByID(ctx, incident.ID)
		if err != nil {
			return fmt.Errorf("failed to resolve incident %s: %w", incident.ID, err)
		}

		message := m.formatRecoveryMessage(svc, resolverIncident)
		if err := m.baseservices.Notifications().History().SendAlert(ctx, svc.ID, resolverIncident.ID, message); err != nil {
			m.logger.Errorf("failed to send recovery notification for %s: %v", svc.Name, err)
		}
	}

	return nil
}

// formatAlertMessage formats an alert message
func (s *Scheduler) formatAlertMessage(service *models.ServiceFullView, incident *models.Incident) string {
	tags := "-"
	if len(service.Tags) > 0 {
		tags = strings.Join(service.Tags, ", ")
	}

	return fmt.Sprintf(
		"🔴 [ALERT] %s is DOWN\n\n"+
			"• Service: %s\n"+
			"• Tags: %s\n"+
			"• Error: %s\n"+
			"• Started: %s\n"+
			"• Incident ID: %s",
		service.Name,
		service.Name,
		tags,
		incident.Error,
		incident.CreatedAt.Format("2006-01-02 15:04:05"),
		incident.ID,
	)
}

// formatRecoveryMessage formats a recovery message
func (s *Scheduler) formatRecoveryMessage(service *models.ServiceFullView, incident *models.Incident) string {
	var duration string
	if incident.Duration != nil {
		duration = utils.FormatDuration(time.Duration(*incident.Duration) * time.Millisecond)
	} else {
		duration = utils.FormatDuration(time.Since(incident.CreatedAt))
	}

	var resolvedAt string
	if incident.ResolvedAt != nil {
		resolvedAt = incident.ResolvedAt.Format("2006-01-02 15:04:05")
	} else {
		resolvedAt = time.Now().Format("2006-01-02 15:04:05")
	}

	tags := "-"
	if len(service.Tags) > 0 {
		tags = strings.Join(service.Tags, ", ")
	}

	return fmt.Sprintf(
		"🟢 [RECOVERY] %s is UP\n\n"+
			"• Service: %s\n"+
			"• Tags: %s\n"+
			"• Downtime: %s\n"+
			"• Recovered: %s\n"+
			"• Incident ID: %s",
		service.Name,
		service.Name,
		tags,
		duration,
		resolvedAt,
		incident.ID,
	)
}
