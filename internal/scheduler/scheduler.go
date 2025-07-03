package scheduler

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/sxwebdev/sentinel/internal/config"
	"github.com/sxwebdev/sentinel/internal/monitors"
	"github.com/sxwebdev/sentinel/internal/service"
)

// ErrServiceNotFound is returned when a service is not found
var ErrServiceNotFound = fmt.Errorf("service not found")

// Scheduler manages the monitoring of multiple services
type Scheduler struct {
	services   map[string]*ServiceJob
	monitorSvc *service.MonitorService
	mu         sync.RWMutex
	wg         sync.WaitGroup
}

// ServiceJob represents a scheduled monitoring job for a service
type ServiceJob struct {
	Config   config.ServiceConfig
	Monitor  monitors.ServiceMonitor
	Ticker   *time.Ticker
	StopChan chan struct{}
}

// NewScheduler creates a new scheduler
func NewScheduler(monitorService *service.MonitorService) *Scheduler {
	return &Scheduler{
		services:   make(map[string]*ServiceJob),
		monitorSvc: monitorService,
	}
}

// AddService adds a service to be monitored
func (s *Scheduler) AddService(cfg config.ServiceConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create monitor for the service
	monitor, err := monitors.NewMonitor(cfg)
	if err != nil {
		return err
	}

	// Check if service already exists
	if existingJob, exists := s.services[cfg.Name]; exists {
		// Stop existing job
		close(existingJob.StopChan)
		existingJob.Ticker.Stop()
	}

	// Create new service job
	job := &ServiceJob{
		Config:   cfg,
		Monitor:  monitor,
		StopChan: make(chan struct{}),
	}

	s.services[cfg.Name] = job
	log.Printf("Added service %s with %s protocol", cfg.Name, cfg.Protocol)
	return nil
}

// RemoveService removes a service from monitoring
func (s *Scheduler) RemoveService(serviceName string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if job, exists := s.services[serviceName]; exists {
		close(job.StopChan)
		if job.Ticker != nil {
			job.Ticker.Stop()
		}
		delete(s.services, serviceName)
		log.Printf("Removed service %s", serviceName)
	}
}

// Start begins monitoring all configured services
func (s *Scheduler) Start(ctx context.Context) {
	// Get all services under read lock
	s.mu.RLock()
	services := make([]*ServiceJob, 0, len(s.services))
	for _, job := range s.services {
		services = append(services, job)
	}
	s.mu.RUnlock()

	log.Printf("Starting scheduler with %d services", len(services))

	// Start monitoring for all services
	for _, job := range services {
		s.wg.Add(1)
		go s.monitorService(ctx, job)
	}

	// Wait for context cancellation
	<-ctx.Done()
	log.Println("Scheduler received shutdown signal")

	// Stop all services
	s.stopAll()
	s.wg.Wait()
	log.Println("Scheduler stopped")
}

// StartWithInitialCheck begins monitoring with an immediate check of all services
func (s *Scheduler) StartWithInitialCheck(ctx context.Context) {
	// Get all services under read lock
	s.mu.RLock()
	services := make([]*ServiceJob, 0, len(s.services))
	for _, job := range s.services {
		services = append(services, job)
	}
	s.mu.RUnlock()

	log.Printf("Starting scheduler with %d services and initial check", len(services))

	// Start monitoring for all services
	for _, job := range services {
		s.wg.Add(1)
		go s.monitorService(ctx, job)
	}

	// Perform immediate check for all services
	log.Println("Performing initial check for all services...")
	for _, job := range services {
		s.performCheck(ctx, job)
	}

	// Wait for context cancellation
	<-ctx.Done()
	log.Println("Scheduler received shutdown signal")

	// Stop all services
	s.stopAll()
	s.wg.Wait()
	log.Println("Scheduler stopped")
}

// monitorService runs the monitoring loop for a single service
func (s *Scheduler) monitorService(ctx context.Context, job *ServiceJob) {
	defer s.wg.Done()

	serviceName := job.Config.Name
	log.Printf("Starting monitoring for service %s", serviceName)

	// Initialize service state with active incidents check
	s.monitorSvc.InitializeWithActiveIncidents(ctx, job.Config)

	// Create ticker for regular checks
	job.Ticker = time.NewTicker(job.Config.Interval)
	defer job.Ticker.Stop()

	// Perform initial check
	s.performCheck(ctx, job)

	for {
		select {
		case <-ctx.Done():
			log.Printf("Stopping monitoring for service %s (context cancelled)", serviceName)
			return
		case <-job.StopChan:
			log.Printf("Stopping monitoring for service %s (stop signal)", serviceName)
			return
		case <-job.Ticker.C:
			s.performCheck(ctx, job)
		}
	}
}

// performCheck executes a health check for a service
func (s *Scheduler) performCheck(ctx context.Context, job *ServiceJob) {
	serviceName := job.Config.Name

	// Create context with timeout for this specific check
	checkCtx, cancel := context.WithTimeout(ctx, job.Config.Timeout)
	defer cancel()

	startTime := time.Now()

	// Perform the check with retries
	var lastErr error
	for attempt := 1; attempt <= job.Config.Retries; attempt++ {
		err := job.Monitor.Check(checkCtx)
		if err == nil {
			// Success
			responseTime := time.Since(startTime)
			s.monitorSvc.RecordSuccess(ctx, serviceName, responseTime)
			return
		}

		lastErr = err

		// If not the last attempt, wait a bit before retrying
		if attempt < job.Config.Retries {
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
	s.monitorSvc.RecordFailure(ctx, serviceName, lastErr, responseTime)
}

// stopAll stops monitoring for all services
func (s *Scheduler) stopAll() {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Printf("Stopping %d services", len(s.services))
	for name, job := range s.services {
		log.Printf("Stopping service: %s", name)
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
	for name := range s.services {
		services = append(services, name)
	}
	return services
}

// CheckService manually triggers a check for a specific service
func (s *Scheduler) CheckService(ctx context.Context, serviceName string) error {
	s.mu.RLock()
	job, exists := s.services[serviceName]
	s.mu.RUnlock()

	if !exists {
		return ErrServiceNotFound
	}

	log.Printf("Manual check triggered for service %s", serviceName)
	s.performCheck(ctx, job)
	return nil
}

// TriggerCheck is an alias for CheckService for consistency
func (s *Scheduler) TriggerCheck(ctx context.Context, serviceName string) error {
	return s.CheckService(ctx, serviceName)
}
