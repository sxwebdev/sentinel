package scheduler

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/puzpuzpuz/xsync/v3"
	"github.com/sxwebdev/sentinel/internal/monitors"
	"github.com/sxwebdev/sentinel/internal/receiver"
	"github.com/sxwebdev/sentinel/internal/service"
	"github.com/sxwebdev/sentinel/internal/storage"
)

// ErrServiceNotFound is returned when a service is not found
var ErrServiceNotFound = fmt.Errorf("service not found")

// Scheduler manages the monitoring of multiple services
type Scheduler struct {
	receiver   *receiver.Receiver
	monitorSvc *service.MonitorService

	jobs *xsync.MapOf[string, *job]
	wg   sync.WaitGroup
}

// job represents a scheduled monitoring job for a service
type job struct {
	ServiceID   string
	ServiceName string
	Interval    time.Duration
	Timeout     time.Duration
	Retries     int
	Ticker      *time.Ticker
	StopChan    chan struct{}
}

// New creates a new scheduler
func New(
	monitorService *service.MonitorService,
	receiver *receiver.Receiver,
) *Scheduler {
	return &Scheduler{
		monitorSvc: monitorService,
		receiver:   receiver,
		jobs:       xsync.NewMapOf[string, *job](),
	}
}

// Start begins monitoring all configured services
func (s *Scheduler) Start(ctx context.Context) error {
	// Load services from database
	services, err := s.monitorSvc.LoadServicesFromStorage(ctx)
	if err != nil {
		return fmt.Errorf("failed to load services: %w", err)
	}

	// Get all services under read lock
	for _, svc := range services {
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
		close(value.StopChan)
		if value.Ticker != nil {
			value.Ticker.Stop()
		}
		return true
	})
}

// addService adds a service to be monitored
func (s *Scheduler) addService(ctx context.Context, svc *storage.Service) {
	// Create new service job with minimal info
	job := &job{
		ServiceID:   svc.ID,
		ServiceName: svc.Name,
		Interval:    svc.Interval,
		Timeout:     svc.Timeout,
		Retries:     svc.Retries,
		StopChan:    make(chan struct{}),
	}

	s.addJob(ctx, job)
}

// addJob adds a new job to the scheduler
func (s *Scheduler) addJob(ctx context.Context, job *job) {
	// Check if job already exists
	if existingJob, exists := s.jobs.Load(job.ServiceID); exists {
		// Stop existing job
		close(existingJob.StopChan)
		if existingJob.Ticker != nil {
			existingJob.Ticker.Stop()
		}
	}

	// Store the new job
	s.jobs.Store(job.ServiceID, job)

	// Start monitoring in a new goroutine
	s.wg.Add(1)
	go s.monitorService(ctx, job)
}

// monitorService runs the monitoring loop for a single service
func (s *Scheduler) monitorService(ctx context.Context, job *job) {
	defer s.wg.Done()

	// Create ticker for regular checks
	job.Ticker = time.NewTicker(job.Interval)
	defer job.Ticker.Stop()

	// Perform initial check
	if err := s.performCheck(ctx, job); err != nil {
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-job.StopChan:
			return
		case <-job.Ticker.C:
			if err := s.performCheck(ctx, job); err != nil {
				continue
			}
		}
	}
}

// performCheck executes a health check for a service
func (s *Scheduler) performCheck(ctx context.Context, job *job) error {
	serviceName := job.ServiceName

	startTime := time.Now()

	// Get current service configuration from database
	service, err := s.monitorSvc.GetServiceByID(ctx, job.ServiceID)
	if err != nil {
		return fmt.Errorf("failed to get service config for %s: %w", serviceName, err)
	}

	// Create monitor for this check
	monitor, err := monitors.NewMonitor(*service)
	if err != nil {
		return fmt.Errorf("failed to create monitor for %s: %w", serviceName, err)
	}

	// Perform the check with retries
	var lastErr error
	for attempt := 1; attempt <= job.Retries; attempt++ {
		// Create context with timeout for this specific check
		checkCtx, cancel := context.WithTimeout(ctx, job.Timeout)
		defer cancel()

		err := monitor.Check(checkCtx)
		if err == nil {
			// Success
			responseTime := time.Since(startTime)
			if err := s.monitorSvc.RecordSuccess(ctx, job.ServiceID, responseTime); err != nil {
				return fmt.Errorf("failed to record success for %s: %w", serviceName, err)
			}

			log.Printf("Service %s check successful (attempt %d/%d)\n", serviceName, attempt, job.Retries)

			return nil
		}

		lastErr = err

		// If not the last attempt, wait a bit before retrying
		if attempt < job.Retries {
			select {
			case <-checkCtx.Done():
				// Context cancelled, don't retry
				break
			case <-time.After(time.Second * time.Duration(attempt)):
				// Exponential backoff
				continue
			}
		}

		log.Printf("Service %s check failed (attempt %d/%d): %s\n", serviceName, attempt, job.Retries, err)
	}

	// All attempts failed
	responseTime := time.Since(startTime)
	if err := s.monitorSvc.RecordFailure(ctx, job.ServiceID, lastErr, responseTime); err != nil {
		return fmt.Errorf("failed to record failure for %s: %w", serviceName, err)
	}

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

	// Stop the monitoring
	close(job.StopChan)
	if job.Ticker != nil {
		job.Ticker.Stop()
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

	errChan := make(chan error, 1)
	go func() {
		for ctx.Err() == nil {
			select {
			case item := <-sub:
				switch item.EventType {
				case receiver.TriggerServiceEventTypeCheck:
					if err := s.checkService(ctx, item.Svc.ID); err != nil {
						log.Println("check service error", err)
					}
				case receiver.TriggerServiceEventTypeCreated:
					s.addService(ctx, item.Svc)
				case receiver.TriggerServiceEventTypeUpdated:
					if err := s.updateJob(ctx, item.Svc); err != nil {
						log.Println("update service error", err)
					}
				case receiver.TriggerServiceEventTypeDeleted:
					if err := s.removeJob(item.Svc.ID); err != nil {
						log.Println("remove service error", err)
					}
				default:
					log.Println("undefined trigger svc event type", item.EventType)
				}

			case <-ctx.Done():
				return
			}
		}
	}()

	select {
	case <-ctx.Done():
	case err := <-errChan:
		return err
	}

	return nil
}
