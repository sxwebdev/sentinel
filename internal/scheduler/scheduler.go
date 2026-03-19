package scheduler

import (
	"context"
	"time"

	"github.com/sxwebdev/sentinel/internal/alertresolver"
	"github.com/sxwebdev/sentinel/internal/checker"
	"github.com/sxwebdev/sentinel/internal/receiver"
	"github.com/sxwebdev/sentinel/internal/services/baseservices"
	"github.com/tkcrm/mx/logger"
)

// Scheduler manages the monitoring of multiple services
type Scheduler struct {
	logger logger.Logger

	checker *checker.Checker

	receiver     *receiver.Receiver
	baseservices *baseservices.BaseServices
	ar           *alertresolver.AlertResolver
}

// New creates a new scheduler
func New(
	l logger.Logger,
	receiver *receiver.Receiver,
	baseServices *baseservices.BaseServices,
	ar *alertresolver.AlertResolver,
) *Scheduler {
	s := &Scheduler{
		logger:       l,
		receiver:     receiver,
		baseservices: baseServices,
		ar:           ar,
	}

	s.checker = checker.New(
		l,
		receiver,
		s.onSuccess,
		s.onFailure,
	)

	s.checker.SetIsHub(true)

	return s
}

// Name returns the name of the scheduler
func (s *Scheduler) Name() string { return "scheduler" }

// Start begins monitoring all configured services
func (s *Scheduler) Start(ctx context.Context) error {
	// Load enabled services from storage
	// services, err := s.baseservices.Services().GetAllEnabledWithoutAgents(ctx)
	// if err != nil {
	// 	return fmt.Errorf("failed to load services: %w", err)
	// }

	// s.logger.Infof("starting scheduler with %d enabled services", len(services))

	// // Get all services under read lock
	// for _, svc := range services {
	// 	s.checker.AddService(ctx, checker.AddServiceParams{
	// 		ID:        svc.ID,
	// 		Name:      svc.Name,
	// 		Protocol:  svc.Protocol,
	// 		IsEnabled: svc.IsEnabled,
	// 		Interval:  time.Duration(svc.Interval) * time.Millisecond,
	// 		Timeout:   time.Duration(svc.Timeout) * time.Millisecond,
	// 		Retries:   svc.Retries,
	// 		Config:    svc.Config.ConvertToMap(),
	// 		FromHub:   true,
	// 	})
	// }

	// return s.checker.Start(ctx)

	return nil
}

func (s *Scheduler) Stop(ctx context.Context) error {
	return s.checker.Stop(ctx)
}

// onSuccess is called when a service check succeeds
func (s *Scheduler) onSuccess(ctx context.Context, serviceID string, responseTime time.Duration) error {
	// Success - record the time of this successful attempt
	// if err := s.ar.RecordSuccess(ctx, serviceID, responseTime); err != nil {
	// 	return fmt.Errorf("failed to record success: %w", err)
	// }

	// svc, err := s.baseservices.Services().GetViewByID(ctx, serviceID)
	// if err != nil {
	// 	return fmt.Errorf("failed to get service view: %w", err)
	// }

	// // Publish update to receiver
	// s.receiver.TriggerService().Publish(*receiver.NewTriggerServiceData(
	// 	receiver.TriggerServiceEventTypeUpdatedState,
	// 	svc,
	// ))

	return nil
}

// onFailure is called when a service check fails
func (s *Scheduler) onFailure(ctx context.Context, serviceID string, checkErr error, responseTime time.Duration) error {
	// All attempts failed - record the time of the last attempt
	// if err := s.ar.RecordFailure(ctx, serviceID, checkErr, responseTime); err != nil {
	// 	return fmt.Errorf("failed to record failure: %w", err)
	// }

	// svc, err := s.baseservices.Services().GetViewByID(ctx, serviceID)
	// if err != nil {
	// 	return fmt.Errorf("failed to get service view: %w", err)
	// }

	// // Publish update to receiver
	// s.receiver.TriggerService().Publish(*receiver.NewTriggerServiceData(
	// 	receiver.TriggerServiceEventTypeUpdatedState,
	// 	svc,
	// ))

	return nil
}
