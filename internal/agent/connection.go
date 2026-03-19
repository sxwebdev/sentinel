package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/tkcrm/modules/pkg/retry"
)

// connectionState represents the state of the connection.
type connectionState uint8

const (
	connectionStateDisconnected connectionState = iota // No active connection
	connectionStateConnected                           // Connected
)

// String returns the string representation of the ConnectionState.
func (cs connectionState) String() string {
	switch cs {
	case connectionStateDisconnected:
		return "disconnected"
	case connectionStateConnected:
		return "connected"
	default:
		return "unknown"
	}
}

// initConnection is a goroutine that manages the connection lifecycle.
func (s *Agent) initConnection(ctx context.Context) {
	// immediately try to connect
	go func() {
		s.changeStateCh <- connectionStateDisconnected
	}()

	// main connection loop
	for {
		select {
		case <-ctx.Done():
			return
		case newState := <-s.changeStateCh:
			s.setState(newState)

			s.logger.Infoln("connection state changed to:", newState.String())

			if newState == connectionStateDisconnected {
				if !s.isConnection.CompareAndSwap(false, true) {
					continue
				}

				if err := retry.New(
					retry.WithContext(ctx),
					retry.WithPolicy(retry.PolicyInfinite),
					retry.WithDelay(2*time.Second),
				).Do(func() error {
					s.logger.Infoln("try to connecting to hub server:", s.config.HubServer.Addr)
					err := s.connect(ctx)
					if err != nil {
						s.logger.Errorln("failed to connect to hub server:", err)
					}
					return err
				}); err != nil {
					s.logger.Errorf("unexpected error while connecting to hub server: %s", err)
				}
			}
		}
	}
}

// connect establishes a connection to the hub server.
func (s *Agent) connect(ctx context.Context) error {
	defer s.isConnection.Store(false)

	if state := s.getState(); state == connectionStateConnected {
		s.logger.Error("already connected to hub server")
		return nil
	}

	if err := s.client.Authenticate(ctx); err != nil {
		return fmt.Errorf("failed to authenticate on hub: %w", err)
	}

	s.logger.Infoln("successfully authenticated to hub server")

	s.logger.Infoln("reporting system info to hub server")

	if err := s.client.ReportSystemInfo(ctx); err != nil {
		return fmt.Errorf("failed to report system info: %w", err)
	}

	s.logger.Infoln("system info reported successfully")

	s.logger.Infoln("fetching services from hub server")

	// services, err := s.client.FetchServices(ctx)
	// if err != nil {
	// 	return fmt.Errorf("failed to fetch services: %w", err)
	// }

	// s.logger.Infoln("services fetched successfully count:", len(services))

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
	// 	})
	// }

	// if len(services) > 0 {
	// 	s.logger.Infoln("added services to checker count:", len(services))
	// }

	s.changeStateCh <- connectionStateConnected

	// subscribe to service updates
	// go func() {
	// 	if err := s.client.SubscribeServices(ctx, func(eventType subscribeServiceType, serviceID string, service *models.Service) {
	// 		switch eventType {
	// 		case subscribeServiceTypeUpsert:
	// 			s.checker.AddService(ctx, checker.AddServiceParams{
	// 				ID:        service.ID,
	// 				Name:      service.Name,
	// 				Protocol:  service.Protocol,
	// 				IsEnabled: service.IsEnabled,
	// 				Interval:  time.Duration(service.Interval) * time.Millisecond,
	// 				Timeout:   time.Duration(service.Timeout) * time.Millisecond,
	// 				Retries:   service.Retries,
	// 				Config:    service.Config.ConvertToMap(),
	// 			})
	// 		case subscribeServiceTypeDelete:
	// 			s.checker.DeleteService(serviceID)
	// 		}
	// 	}); err != nil {
	// 		s.logger.Errorf("error while subscribing to services: %v", err)
	// 	}
	// }()

	// stream check results
	go func() {
		if err := s.client.StreamChecks(ctx); err != nil {
			s.logger.Errorf("error while streaming checks: %v", err)
		}
	}()

	return nil
}
