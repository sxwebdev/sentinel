package scheduler

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/puzpuzpuz/xsync/v3"
	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/sxwebdev/sentinel/internal/monitors"
	"github.com/sxwebdev/sentinel/internal/notifier"
	"github.com/sxwebdev/sentinel/internal/receiver"
	"github.com/sxwebdev/sentinel/internal/services/baseservices"
	"github.com/sxwebdev/sentinel/internal/services/incidents"
	"github.com/sxwebdev/sentinel/internal/services/service"
	"github.com/sxwebdev/sentinel/internal/services/servicestate"
	"github.com/sxwebdev/sentinel/internal/store"
	"github.com/sxwebdev/sentinel/internal/store/repos/repo_service_states"
	"github.com/sxwebdev/sentinel/internal/utils"
	"github.com/tkcrm/modules/pkg/db/dbutils"
	"github.com/tkcrm/mx/logger"
)

// ErrServiceNotFound is returned when a service is not found
var ErrServiceNotFound = fmt.Errorf("service not found")

// Scheduler manages the monitoring of multiple services
type Scheduler struct {
	logger logger.Logger

	store    *store.Store
	notifier *notifier.Notifier

	receiver     *receiver.Receiver
	baseservices *baseservices.BaseServices

	jobs *xsync.MapOf[string, *job]
	wg   sync.WaitGroup
}

// New creates a new scheduler
func New(
	l logger.Logger,
	store *store.Store,
	notifier *notifier.Notifier,
	receiver *receiver.Receiver,
	baseServices *baseservices.BaseServices,
) *Scheduler {
	return &Scheduler{
		logger:       l,
		store:        store,
		notifier:     notifier,
		receiver:     receiver,
		baseservices: baseServices,
		jobs:         xsync.NewMapOf[string, *job](),
	}
}

// Name returns the name of the scheduler
func (s *Scheduler) Name() string { return "scheduler" }

// Start begins monitoring all configured services
func (s *Scheduler) Start(ctx context.Context) error {
	// Load enabled services from storage
	isEnabled := true
	services, err := s.baseservices.Services().FindView(ctx, service.FindParams{
		IsEnabled: &isEnabled,
	})
	if err != nil {
		return fmt.Errorf("failed to load services: %w", err)
	}

	s.logger.Infof("starting scheduler with %d enabled services", len(services.Items))

	// Get all services under read lock
	for _, svc := range services.Items {
		s.addService(ctx, svc)
	}

	errChan := make(chan error, 1)
	go func() {
		errChan <- s.subscribeEvents(ctx)
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
	}

	return nil
}

func (s *Scheduler) Stop(ctx context.Context) error {
	s.stopAll()
	s.wg.Wait()
	return nil
}

// stopAll stops monitoring for all services
func (s *Scheduler) stopAll() {
	s.jobs.Range(func(key string, value *job) bool {
		// Cancel any ongoing checks
		if value.checkCancel != nil {
			value.checkCancel()
		}

		select {
		case <-value.stopChan:
			// Channel already closed
		default:
			close(value.stopChan)
		}
		if value.ticker != nil {
			value.ticker.Stop()
		}
		return true
	})
}

// addService adds a service to be monitored
func (s *Scheduler) addService(ctx context.Context, svc *models.ServiceFullView) {
	// Only add enabled services to monitoring
	if !svc.IsEnabled {
		s.logger.Warnf("Skipping disabled service: %s (ID: %s)", svc.Name, svc.ID)
		return
	}

	// Create new service job with minimal info
	checkCtx, checkCancel := context.WithCancel(ctx)
	job := &job{
		serviceID:   svc.ID,
		serviceName: svc.Name,
		interval:    svc.Interval,
		timeout:     svc.Timeout,
		retries:     svc.Retries,
		stopChan:    make(chan struct{}),
		checkCtx:    checkCtx,
		checkCancel: checkCancel,
	}

	s.addJob(ctx, job)
}

// addJob adds a new job to the scheduler
func (s *Scheduler) addJob(ctx context.Context, job *job) {
	// Check if job already exists
	if existingJob, exists := s.jobs.Load(job.serviceID); exists {
		// Cancel any ongoing checks for the existing job
		if existingJob.checkCancel != nil {
			existingJob.checkCancel()
		}

		// Stop existing job gracefully
		select {
		case <-existingJob.stopChan:
			// Channel already closed
		default:
			close(existingJob.stopChan)
		}
		if existingJob.ticker != nil {
			existingJob.ticker.Stop()
		}
	}

	// Store the new job
	s.jobs.Store(job.serviceID, job)

	// Start monitoring in a new goroutine
	s.wg.Add(1)
	go s.monitorService(ctx, job)
}

// monitorService runs the monitoring loop for a single service
func (s *Scheduler) monitorService(ctx context.Context, job *job) {
	defer s.wg.Done()

	// Create ticker for regular checks
	job.ticker = time.NewTicker(job.interval)
	defer job.ticker.Stop()

	// Perform initial check
	if err := s.performCheck(job); err != nil {
		s.logger.Errorf("Error performing initial check for service %s: %v", job.serviceName, err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-job.stopChan:
			return
		case <-job.ticker.C:
			if err := s.performCheck(job); err != nil && !errors.Is(err, context.Canceled) {
				s.logger.Errorf("error performing check for service %s: %v", job.serviceName, err)
				continue
			}
		}
	}
}

// performCheck executes a health check for a service
func (s *Scheduler) performCheck(job *job) error {
	if !job.inProgress.CompareAndSwap(false, true) {
		// Another check is already in progress
		return nil
	}
	defer job.inProgress.Store(false)

	// Check if job context is cancelled (handles job updates/deletions)
	if job.checkCtx.Err() != nil {
		return job.checkCtx.Err()
	}

	serviceName := job.serviceName

	// Get current service configuration from database
	service, err := s.baseservices.Services().GetByID(job.checkCtx, job.serviceID)
	if err != nil {
		return fmt.Errorf("failed to get service %s: %w", serviceName, err)
	}

	// Create monitor for this check
	monitor, err := monitors.NewMonitor(service)
	if err != nil {
		return fmt.Errorf("failed to create monitor for %s: %w", serviceName, err)
	}

	// Ensure monitor resources are cleaned up
	defer func() {
		if closer, ok := monitor.(interface{ Close() error }); ok {
			if err := closer.Close(); err != nil {
				s.logger.Errorf("Error closing monitor for %s: %v", serviceName, err)
			}
		}
	}()

	// Perform the check with retries
	var lastErr error
	var lastAttemptResponseTime time.Duration

	for attempt := int64(1); attempt <= job.retries; attempt++ {
		// Create context with timeout for this specific check
		attemptCtx, cancel := context.WithTimeout(job.checkCtx, job.timeout)

		// Measure time for this specific attempt
		attemptStartTime := time.Now()
		err := monitor.Check(attemptCtx)
		attemptResponseTime := time.Since(attemptStartTime)
		lastAttemptResponseTime = attemptResponseTime

		// Cancel context immediately after use to avoid memory leak
		cancel()

		if err == nil {
			// Success - record the time of this successful attempt
			if err := s.recordSuccess(job.checkCtx, job.serviceID, attemptResponseTime); err != nil {
				return fmt.Errorf("failed to record success for %s: %w", serviceName, err)
			}

			if attempt == 1 {
				s.logger.Debugf("service %s check successful in %v", serviceName, attemptResponseTime)
			} else {
				s.logger.Debugf("service %s check successful (attempt %d/%d) in %v", serviceName, attempt, job.retries, attemptResponseTime)
			}

			svc, err := s.baseservices.Services().GetViewByID(job.checkCtx, job.serviceID)
			if err != nil {
				return fmt.Errorf("failed to get service view for %s: %w", serviceName, err)
			}

			// Publish update to receiver
			s.receiver.TriggerService().Publish(*receiver.NewTriggerServiceData(
				receiver.TriggerServiceEventTypeUpdatedState,
				svc,
			))

			return nil
		}

		lastErr = err

		// If not the last attempt, wait a bit before retrying
		if attempt < job.retries {
			// Check if job context is cancelled before retrying
			select {
			case <-job.checkCtx.Done():
				// Job context cancelled, stop retrying
				return job.checkCtx.Err()
			case <-time.After(time.Millisecond * 500 * time.Duration(attempt)):
				// Exponential backoff - continue to next attempt
			}
		}

		s.logger.Debugf("service %s check failed (attempt %d/%d): %s", serviceName, attempt, job.retries, err)
	}

	// All attempts failed - record the time of the last attempt
	if err := s.recordFailure(job.checkCtx, job.serviceID, lastErr, lastAttemptResponseTime); err != nil {
		return fmt.Errorf("failed to record failure for %s: %w", serviceName, err)
	}

	svc, err := s.baseservices.Services().GetViewByID(job.checkCtx, job.serviceID)
	if err != nil {
		return fmt.Errorf("failed to get service view for %s: %w", serviceName, err)
	}

	// Publish update to receiver
	s.receiver.TriggerService().Publish(*receiver.NewTriggerServiceData(
		receiver.TriggerServiceEventTypeUpdatedState,
		svc,
	))

	return nil
}

// checkService manually triggers a check for a specific service
func (s *Scheduler) checkService(serviceID string) error {
	job, exists := s.jobs.Load(serviceID)
	if !exists {
		return ErrServiceNotFound
	}

	return s.performCheck(job)
}

// removeJob removes a service dynamically (for runtime removals)
func (s *Scheduler) removeJob(serviceID string) error {
	job, exists := s.jobs.Load(serviceID)
	if !exists {
		return ErrServiceNotFound
	}

	// Cancel any ongoing checks for this job
	if job.checkCancel != nil {
		job.checkCancel()
	}

	// Stop the monitoring gracefully
	select {
	case <-job.stopChan:
		// Channel already closed
	default:
		close(job.stopChan)
	}
	if job.ticker != nil {
		job.ticker.Stop()
	}

	// Remove from services map
	s.jobs.Delete(serviceID)

	return nil
}

// updateJob updates a service configuration dynamically
func (s *Scheduler) updateJob(ctx context.Context, svc *models.ServiceFullView) error {
	s.addService(ctx, svc)

	return nil
}

func (s *Scheduler) subscribeEvents(ctx context.Context) error {
	broker := s.receiver.TriggerService()
	sub := broker.Subscribe()
	defer broker.Unsubscribe(sub)

	for ctx.Err() == nil {
		select {
		case item := <-sub:
			switch item.EventType {
			case receiver.TriggerServiceEventTypeCheck:
				if err := s.checkService(item.Svc.ID); err != nil {
					s.logger.Errorf("check service error: %v", err)
				}
			case receiver.TriggerServiceEventTypeCreated:
				s.addService(ctx, item.Svc)
			case receiver.TriggerServiceEventTypeUpdated:
				// Check if service was disabled
				if !item.Svc.IsEnabled {
					// Remove from monitoring if service is now disabled
					if err := s.removeJob(item.Svc.ID); err != nil {
						s.logger.Errorf("remove disabled service error: %v", err)
					}
				} else {
					// Update or add to monitoring if service is enabled
					if err := s.updateJob(ctx, item.Svc); err != nil {
						s.logger.Errorf("update service error: %v", err)
					}
				}
			case receiver.TriggerServiceEventTypeDeleted:
				if err := s.removeJob(item.Svc.ID); err != nil {
					s.logger.Errorf("remove service error: %v", err)
				}
			}

		case <-ctx.Done():
			return nil
		}
	}

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
func (m *Scheduler) recordFailure(ctx context.Context, serviceID string, checkErr error, responseTime time.Duration) error {
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
func (m *Scheduler) createIncident(ctx context.Context, svc *models.ServiceFullView, err error) error {
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

		if m.notifier != nil {
			if err := m.notifier.SendRecovery(svc, resolverIncident); err != nil {
				m.logger.Errorf("failed to send recovery notification for %s: %v", svc.Name, err)
				return nil
			}
		}
	}

	return nil
}
