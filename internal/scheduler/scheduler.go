package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sxwebdev/sentinel/internal/monitors"
	"github.com/sxwebdev/sentinel/internal/storage"
)

// MonitorServiceInterface defines the interface for monitor service operations
type MonitorServiceInterface interface {
	InitializeService(ctx context.Context, cfg storage.Service) error
	InitializeWithActiveIncidents(ctx context.Context, cfg storage.Service) error
	RecordSuccess(ctx context.Context, serviceID string, responseTime time.Duration) error
	RecordFailure(ctx context.Context, serviceID string, err error, responseTime time.Duration) error
	GetServiceByID(ctx context.Context, id string) (*storage.Service, error)
	TriggerCheck(ctx context.Context, serviceID string) error
}

// ErrServiceNotFound is returned when a service is not found
var ErrServiceNotFound = fmt.Errorf("service not found")

// Scheduler manages the monitoring of multiple services
type Scheduler struct {
	services   map[string]*ServiceJob // key is service ID
	monitorSvc MonitorServiceInterface
	mu         sync.RWMutex
	wg         sync.WaitGroup
}

// ServiceJob represents a scheduled monitoring job for a service
type ServiceJob struct {
	ServiceID string
	Name      string
	Interval  time.Duration
	Timeout   time.Duration
	Retries   int
	Ticker    *time.Ticker
	StopChan  chan struct{}
}

// NewScheduler creates a new scheduler
func NewScheduler(monitorService MonitorServiceInterface) *Scheduler {
	return &Scheduler{
		services:   make(map[string]*ServiceJob),
		monitorSvc: monitorService,
	}
}

// AddService adds a service to be monitored
func (s *Scheduler) AddService(cfg storage.Service) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if service already exists
	if existingJob, exists := s.services[cfg.ID]; exists {
		// Stop existing job
		close(existingJob.StopChan)
		existingJob.Ticker.Stop()
	}

	// Create new service job with minimal info
	job := &ServiceJob{
		ServiceID: cfg.ID,
		Name:      cfg.Name,
		Interval:  cfg.Interval,
		Timeout:   cfg.Timeout,
		Retries:   cfg.Retries,
		StopChan:  make(chan struct{}),
	}

	s.services[cfg.ID] = job
	return nil
}

// RemoveService removes a service from monitoring
func (s *Scheduler) RemoveService(serviceID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if job, exists := s.services[serviceID]; exists {
		close(job.StopChan)
		if job.Ticker != nil {
			job.Ticker.Stop()
		}
		delete(s.services, serviceID)
	}
}

// Start begins monitoring all configured services
func (s *Scheduler) Start(ctx context.Context) error {
	// Get all services under read lock
	s.mu.RLock()
	services := make([]*ServiceJob, 0, len(s.services))
	for _, job := range s.services {
		services = append(services, job)
	}
	s.mu.RUnlock()

	// Start monitoring for all services
	for _, job := range services {
		s.wg.Add(1)
		go s.monitorService(ctx, job)
	}

	// Wait for context cancellation
	<-ctx.Done()

	// Stop all services
	s.stopAll()
	s.wg.Wait()
	return nil
}

// StartWithInitialCheck begins monitoring with an immediate check of all services
func (s *Scheduler) StartWithInitialCheck(ctx context.Context) error {
	// Get all services under read lock
	s.mu.RLock()
	services := make([]*ServiceJob, 0, len(s.services))
	for _, job := range s.services {
		services = append(services, job)
	}
	s.mu.RUnlock()

	// Start monitoring for all services
	for _, job := range services {
		s.wg.Add(1)
		go s.monitorService(ctx, job)
	}

	// Perform immediate check for all services
	for _, job := range services {
		if err := s.performCheck(ctx, job); err != nil {
			return fmt.Errorf("failed to perform initial check for service %s: %w", job.Name, err)
		}
	}

	// Wait for context cancellation
	<-ctx.Done()

	// Stop all services
	s.stopAll()
	s.wg.Wait()
	return nil
}

// monitorService runs the monitoring loop for a single service
func (s *Scheduler) monitorService(ctx context.Context, job *ServiceJob) {
	defer s.wg.Done()

	// Create ticker for regular checks
	job.Ticker = time.NewTicker(job.Interval)
	defer job.Ticker.Stop()

	// Perform initial check
	if err := s.performCheck(ctx, job); err != nil {
		// Log error but continue monitoring
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
				// Log error but continue monitoring
				continue
			}
		}
	}
}

// performCheck executes a health check for a service
func (s *Scheduler) performCheck(ctx context.Context, job *ServiceJob) error {
	serviceName := job.Name

	// Create context with timeout for this specific check
	checkCtx, cancel := context.WithTimeout(ctx, job.Timeout)
	defer cancel()

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
		err := monitor.Check(checkCtx)
		if err == nil {
			// Success
			responseTime := time.Since(startTime)
			if err := s.monitorSvc.RecordSuccess(ctx, job.ServiceID, responseTime); err != nil {
				return fmt.Errorf("failed to record success for %s: %w", serviceName, err)
			}
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
	}

	// All attempts failed
	responseTime := time.Since(startTime)
	if err := s.monitorSvc.RecordFailure(ctx, job.ServiceID, lastErr, responseTime); err != nil {
		return fmt.Errorf("failed to record failure for %s: %w", serviceName, err)
	}

	return nil
}

// stopAll stops monitoring for all services
func (s *Scheduler) stopAll() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, job := range s.services {
		close(job.StopChan)
		if job.Ticker != nil {
			job.Ticker.Stop()
		}
	}
}

// GetServices returns the list of currently monitored services
func (s *Scheduler) GetServices() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	services := make([]string, 0, len(s.services))
	for _, job := range s.services {
		services = append(services, job.Name)
	}
	return services
}

// CheckService manually triggers a check for a specific service
func (s *Scheduler) CheckService(ctx context.Context, serviceID string) error {
	s.mu.RLock()
	job, exists := s.services[serviceID]
	s.mu.RUnlock()

	if !exists {
		return ErrServiceNotFound
	}

	return s.performCheck(ctx, job)
}

// TriggerCheck triggers an immediate check for a specific service
func (s *Scheduler) TriggerCheck(ctx context.Context, serviceID string) error {
	s.mu.RLock()
	job, exists := s.services[serviceID]
	s.mu.RUnlock()

	if !exists {
		return ErrServiceNotFound
	}

	return s.performCheck(ctx, job)
}

// AddServiceDynamic adds a service dynamically (for runtime additions)
func (s *Scheduler) AddServiceDynamic(ctx context.Context, cfg storage.Service) error {
	// Add to scheduler
	if err := s.AddService(cfg); err != nil {
		return fmt.Errorf("failed to add service: %w", err)
	}

	// Start monitoring immediately
	s.mu.RLock()
	job, exists := s.services[cfg.ID]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("service was not properly added")
	}

	// Start monitoring in a new goroutine
	s.wg.Add(1)
	go s.monitorService(ctx, job)

	return nil
}

// RemoveServiceDynamic removes a service dynamically (for runtime removals)
func (s *Scheduler) RemoveServiceDynamic(serviceID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, exists := s.services[serviceID]
	if !exists {
		return ErrServiceNotFound
	}

	// Stop the monitoring
	close(job.StopChan)
	if job.Ticker != nil {
		job.Ticker.Stop()
	}

	// Remove from services map
	delete(s.services, serviceID)

	return nil
}

// UpdateServiceDynamic updates a service configuration dynamically
func (s *Scheduler) UpdateServiceDynamic(ctx context.Context, cfg storage.Service) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if service already exists by ID
	existingJob, exists := s.services[cfg.ID]
	if exists {
		// Stop existing job
		close(existingJob.StopChan)
		if existingJob.Ticker != nil {
			existingJob.Ticker.Stop()
		}
	}

	// Create new service job with updated configuration
	job := &ServiceJob{
		ServiceID: cfg.ID,
		Name:      cfg.Name,
		Interval:  cfg.Interval,
		Timeout:   cfg.Timeout,
		Retries:   cfg.Retries,
		StopChan:  make(chan struct{}),
	}

	s.services[cfg.ID] = job

	// Start monitoring in a new goroutine
	s.wg.Add(1)
	go s.monitorService(ctx, job)

	return nil
}
