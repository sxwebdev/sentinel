package checker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/puzpuzpuz/xsync/v3"
	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/sxwebdev/sentinel/internal/monitors"
	"github.com/sxwebdev/sentinel/internal/receiver"
	"github.com/tkcrm/mx/logger"
)

type Checker struct {
	logger logger.Logger

	receiver *receiver.Receiver

	// Job management, key is service ID
	jobs *xsync.MapOf[string, *job]
	wg   sync.WaitGroup

	onSuccessFn func(ctx context.Context, serviceID string, responseTime time.Duration) error
	onFailureFn func(ctx context.Context, serviceID string, checkErr error, responseTime time.Duration) error
}

func New(
	l logger.Logger,
	receiver *receiver.Receiver,
	onSuccessFn func(ctx context.Context, serviceID string, responseTime time.Duration) error,
	onFailureFn func(ctx context.Context, serviceID string, checkErr error, responseTime time.Duration) error,
) *Checker {
	return &Checker{
		logger:      l,
		receiver:    receiver,
		jobs:        xsync.NewMapOf[string, *job](),
		onSuccessFn: onSuccessFn,
		onFailureFn: onFailureFn,
	}
}

// Name returns the name of the service
func (c *Checker) Name() string {
	return "checker"
}

// Start begins monitoring all configured services
func (s *Checker) Start(ctx context.Context) error {
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

func (s *Checker) Stop(ctx context.Context) error {
	s.stopAll()
	s.wg.Wait()
	return nil
}

// stopAll stops monitoring for all services
func (s *Checker) stopAll() {
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

type AddServiceParams struct {
	ID        string
	Name      string
	Protocol  models.ServiceProtocolType
	IsEnabled bool
	Interval  time.Duration
	Timeout   time.Duration
	Retries   int64
	Config    map[string]any
}

// AddService adds a service to be monitored
func (s *Checker) AddService(ctx context.Context, svc AddServiceParams) {
	// Only add enabled services to monitoring
	if !svc.IsEnabled {
		s.logger.Warnf("Skipping disabled service: %s (ID: %s)", svc.Name, svc.ID)
		return
	}

	// Create new service job with minimal info
	checkCtx, checkCancel := context.WithCancel(ctx)
	job := &job{
		serviceID:       svc.ID,
		serviceName:     svc.Name,
		serviceProtocol: svc.Protocol,
		interval:        svc.Interval,
		timeout:         svc.Timeout,
		retries:         svc.Retries,
		serviceConfig:   svc.Config,
		stopChan:        make(chan struct{}),
		checkCtx:        checkCtx,
		checkCancel:     checkCancel,
	}

	s.addJob(ctx, job)
}

// addJob adds a new job to the scheduler
func (s *Checker) addJob(ctx context.Context, job *job) {
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
	s.wg.Go(func() {
		s.monitorService(ctx, job)
	})
}

// removeJob removes a service dynamically (for runtime removals)
func (s *Checker) removeJob(serviceID string) error {
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

// monitorService runs the monitoring loop for a single service
func (s *Checker) monitorService(ctx context.Context, job *job) {
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

// checkService manually triggers a check for a specific service
func (s *Checker) checkService(serviceID string) error {
	job, exists := s.jobs.Load(serviceID)
	if !exists {
		return ErrServiceNotFound
	}

	return s.performCheck(job)
}

func (s *Checker) subscribeEvents(ctx context.Context) error {
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
				s.AddService(ctx, AddServiceParams{
					ID:        item.Svc.ID,
					Name:      item.Svc.Name,
					Protocol:  item.Svc.Protocol,
					IsEnabled: item.Svc.IsEnabled,
					Interval:  time.Duration(item.Svc.Interval) * time.Millisecond,
					Timeout:   time.Duration(item.Svc.Timeout) * time.Millisecond,
					Retries:   item.Svc.Retries,
					Config:    item.Svc.Config,
				})
			case receiver.TriggerServiceEventTypeUpdated:
				// Check if service was disabled
				if !item.Svc.IsEnabled {
					// Remove from monitoring if service is now disabled
					if err := s.removeJob(item.Svc.ID); err != nil {
						s.logger.Errorf("remove disabled service error: %v", err)
					}
				} else {
					// Update or add to monitoring if service is enabled
					s.AddService(ctx, AddServiceParams{
						ID:        item.Svc.ID,
						Name:      item.Svc.Name,
						Protocol:  item.Svc.Protocol,
						IsEnabled: item.Svc.IsEnabled,
						Interval:  time.Duration(item.Svc.Interval) * time.Millisecond,
						Timeout:   time.Duration(item.Svc.Timeout) * time.Millisecond,
						Retries:   item.Svc.Retries,
						Config:    item.Svc.Config,
					})
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

// performCheck executes a health check for a service
func (s *Checker) performCheck(job *job) error {
	if !job.inProgress.CompareAndSwap(false, true) {
		// Another check is already in progress
		return nil
	}
	defer job.inProgress.Store(false)

	// Check if job context is cancelled (handles job updates/deletions)
	if job.checkCtx.Err() != nil {
		return job.checkCtx.Err()
	}

	// Create monitor for this check
	monitor, err := monitors.NewMonitor(monitors.MonitorParams{
		ServiceName: job.serviceName,
		Protocol:    job.serviceProtocol,
		Timeout:     job.timeout,
		Config:      job.serviceConfig,
	})
	if err != nil {
		return fmt.Errorf("failed to create monitor for %s: %w", job.serviceName, err)
	}

	// Ensure monitor resources are cleaned up
	defer func() {
		if closer, ok := monitor.(io.Closer); ok {
			if err := closer.Close(); err != nil {
				s.logger.Errorf("Error closing monitor for %s: %v", job.serviceName, err)
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
			if attempt == 1 {
				s.logger.Debugf("service %s check successful in %v", job.serviceName, attemptResponseTime)
			} else {
				s.logger.Debugf("service %s check successful (attempt %d/%d) in %v", job.serviceName, attempt, job.retries, attemptResponseTime)
			}

			if err := s.onSuccessFn(job.checkCtx, job.serviceID, attemptResponseTime); err != nil {
				return err
			}

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

		s.logger.Debugf("service %s check failed (attempt %d/%d): %s", job.serviceName, attempt, job.retries, err)
	}

	if err := s.onFailureFn(job.checkCtx, job.serviceID, lastErr, lastAttemptResponseTime); err != nil {
		return err
	}

	return nil
}
