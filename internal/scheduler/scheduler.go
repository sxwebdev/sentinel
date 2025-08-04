package scheduler

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/puzpuzpuz/xsync/v3"
	"github.com/sxwebdev/sentinel/internal/monitor"
	"github.com/sxwebdev/sentinel/internal/monitors"
	"github.com/sxwebdev/sentinel/internal/receiver"
	"github.com/sxwebdev/sentinel/internal/storage"
	"github.com/tkcrm/mx/logger"
)

// ErrServiceNotFound is returned when a service is not found
var ErrServiceNotFound = fmt.Errorf("service not found")

// Scheduler manages the monitoring of multiple services
type Scheduler struct {
	logger logger.Logger

	receiver   *receiver.Receiver
	monitorSvc *monitor.MonitorService

	jobs *xsync.MapOf[string, *job]
	wg   sync.WaitGroup
}

// job represents a scheduled monitoring job for a service
type job struct {
	serviceID   string
	serviceName string
	interval    time.Duration
	timeout     time.Duration
	retries     int
	ticker      *time.Ticker
	stopChan    chan struct{}
	inProgress  atomic.Bool
}

// New creates a new scheduler
func New(
	l logger.Logger,
	monitorService *monitor.MonitorService,
	receiver *receiver.Receiver,
) *Scheduler {
	return &Scheduler{
		logger:     l,
		monitorSvc: monitorService,
		receiver:   receiver,
		jobs:       xsync.NewMapOf[string, *job](),
	}
}

// Name returns the name of the scheduler
func (s *Scheduler) Name() string { return "scheduler" }

// Start begins monitoring all configured services
func (s *Scheduler) Start(ctx context.Context) error {
	// Load enabled services from storage
	isEnabled := true
	services, err := s.monitorSvc.FindServices(ctx, storage.FindServicesParams{
		IsEnabled: &isEnabled,
	})
	if err != nil {
		return fmt.Errorf("failed to load services: %w", err)
	}

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
func (s *Scheduler) addService(ctx context.Context, svc *storage.Service) {
	// Only add enabled services to monitoring
	if !svc.IsEnabled {
		s.logger.Warnf("Skipping disabled service: %s (ID: %s)", svc.Name, svc.ID)
		return
	}

	// Create new service job with minimal info
	job := &job{
		serviceID:   svc.ID,
		serviceName: svc.Name,
		interval:    svc.Interval,
		timeout:     svc.Timeout,
		retries:     svc.Retries,
		stopChan:    make(chan struct{}),
	}

	s.addJob(ctx, job)
}

// addJob adds a new job to the scheduler
func (s *Scheduler) addJob(ctx context.Context, job *job) {
	// Check if job already exists
	if existingJob, exists := s.jobs.Load(job.serviceID); exists {
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
	if err := s.performCheck(ctx, job); err != nil {
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
			if err := s.performCheck(ctx, job); err != nil && !errors.Is(err, context.Canceled) {
				s.logger.Errorf("error performing check for service %s: %v", job.serviceName, err)
				continue
			}
		}
	}
}

// performCheck executes a health check for a service
func (s *Scheduler) performCheck(ctx context.Context, job *job) error {
	if !job.inProgress.CompareAndSwap(false, true) {
		// Another check is already in progress
		return nil
	}
	defer job.inProgress.Store(false)

	serviceName := job.serviceName

	// Get current service configuration from database
	service, err := s.monitorSvc.GetServiceByID(ctx, job.serviceID)
	if err != nil {
		return fmt.Errorf("failed to get service config for %s: %w", serviceName, err)
	}

	// Create monitor for this check
	monitor, err := monitors.NewMonitor(*service)
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
	for attempt := 1; attempt <= job.retries; attempt++ {
		// Create context with timeout for this specific check
		checkCtx, cancel := context.WithTimeout(ctx, job.timeout)

		// Measure time for this specific attempt
		attemptStartTime := time.Now()
		err := monitor.Check(checkCtx)
		attemptResponseTime := time.Since(attemptStartTime)
		lastAttemptResponseTime = attemptResponseTime

		// Cancel context immediately after use to avoid memory leak
		cancel()

		if err == nil {
			// Success - record the time of this successful attempt
			if err := s.monitorSvc.RecordSuccess(ctx, job.serviceID, attemptResponseTime); err != nil {
				return fmt.Errorf("failed to record success for %s: %w", serviceName, err)
			}

			s.logger.Debugf("service %s check successful (attempt %d/%d) in %v", serviceName, attempt, job.retries, attemptResponseTime)

			service, err := s.monitorSvc.GetServiceByID(ctx, job.serviceID)
			if err != nil {
				return fmt.Errorf("failed to get service config for %s: %w", serviceName, err)
			}

			// Publish update to receiver
			s.receiver.TriggerService().Publish(*receiver.NewTriggerServiceData(
				receiver.TriggerServiceEventTypeUpdatedState,
				service,
			))

			return nil
		}

		lastErr = err

		// If not the last attempt, wait a bit before retrying
		if attempt < job.retries {
			select {
			case <-checkCtx.Done():
				// Context cancelled, don't retry
				break
			case <-time.After(time.Second * time.Duration(attempt)):
				// Exponential backoff
				continue
			}
		}

		s.logger.Debugf("service %s check failed (attempt %d/%d): %s", serviceName, attempt, job.retries, err)
	}

	// All attempts failed - record the time of the last attempt
	if err := s.monitorSvc.RecordFailure(ctx, job.serviceID, lastErr, lastAttemptResponseTime); err != nil {
		return fmt.Errorf("failed to record failure for %s: %w", serviceName, err)
	}

	service, err = s.monitorSvc.GetServiceByID(ctx, job.serviceID)
	if err != nil {
		return fmt.Errorf("failed to get service config for %s: %w", serviceName, err)
	}

	// Publish update to receiver
	s.receiver.TriggerService().Publish(*receiver.NewTriggerServiceData(
		receiver.TriggerServiceEventTypeUpdatedState,
		service,
	))

	return nil
}

// checkService manually triggers a check for a specific service
func (s *Scheduler) checkService(ctx context.Context, serviceID string) error {
	job, exists := s.jobs.Load(serviceID)
	if !exists {
		return ErrServiceNotFound
	}

	return s.performCheck(ctx, job)
}

// removeJob removes a service dynamically (for runtime removals)
func (s *Scheduler) removeJob(serviceID string) error {
	job, exists := s.jobs.Load(serviceID)
	if !exists {
		return ErrServiceNotFound
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
func (s *Scheduler) updateJob(ctx context.Context, svc *storage.Service) error {
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
				if err := s.checkService(ctx, item.Svc.ID); err != nil {
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
