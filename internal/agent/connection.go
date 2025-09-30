package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/tkcrm/modules/pkg/retry"
)

// ConnectionState represents the state of the connection.
type ConnectionState uint8

const (
	ConnectionStateDisconnected ConnectionState = iota // No active connection
	ConnectionStateConnected                           // Connected
)

// String returns the string representation of the ConnectionState.
func (cs ConnectionState) String() string {
	switch cs {
	case ConnectionStateDisconnected:
		return "disconnected"
	case ConnectionStateConnected:
		return "connected"
	default:
		return "unknown"
	}
}

// initConnection is a goroutine that manages the connection lifecycle.
func (s *Agent) initConnection(ctx context.Context) {
	// immediately try to connect
	go func() {
		s.changeStateCh <- ConnectionStateDisconnected
	}()

	// main connection loop
	for {
		select {
		case <-ctx.Done():
			return
		case newState := <-s.changeStateCh:
			s.setState(newState)

			s.logger.Infoln("connection state changed to:", newState.String())

			if newState == ConnectionStateDisconnected {
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

	if state := s.getState(); state == ConnectionStateConnected {
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

	services, err := s.client.FetchServices(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch services: %w", err)
	}

	s.setServices(services)

	s.logger.Infoln("services fetched successfully", "count:", len(services))

	s.changeStateCh <- ConnectionStateConnected

	// subscribe to service updates
	go func() {
		if err := s.client.SubscribeServices(ctx, func(eventType, serviceID string, service *models.Service) {
			s.logger.Infof("received event: %s, service ID: %s", eventType, serviceID)
		}); err != nil {
			s.logger.Errorln("error while subscribing to services:", err)
		}
	}()

	return nil
}
