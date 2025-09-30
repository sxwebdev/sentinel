package agent

import (
	"context"
	"fmt"
	"time"
)

// ConnectionState represents the state of the connection.
type ConnectionState uint8

const (
	ConnectionStateDisconnected ConnectionState = iota // No active connection
	ConnectionStateConnected                           // Connected
)

// isConnected checks if the connection is active.
func (cm *Agent) isConnected() bool {
	return cm.state == ConnectionStateConnected
}

// initConnection is a goroutine that manages the connection lifecycle.
func (s *Agent) initConnection(ctx context.Context) {
	for ctx.Err() == nil {
		if s.isConnected() {
			if err := s.client.Ping(ctx); err != nil {
				s.logger.Errorln("lost connection to hub server:", err)
				s.state = ConnectionStateDisconnected
			} else {
				time.Sleep(3 * time.Second)
				continue
			}
			time.Sleep(2 * time.Second)
			continue
		}

		s.logger.Infoln("try to connecting to hub server:", s.config.HubServer.Addr)

		if err := s.connect(ctx); err != nil {
			s.logger.Errorln("failed to connect to hub server:", err)
			time.Sleep(3 * time.Second)
			continue
		}
	}
}

// connect establishes a connection to the hub server.
func (s *Agent) connect(ctx context.Context) error {
	if err := s.client.Authenticate(ctx); err != nil {
		return fmt.Errorf("failed to authenticate on hub: %w", err)
	}

	s.state = ConnectionStateConnected

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

	// subscribe to service updates
	// go func() {
	// 	if err := s.client.SubscribeServices(ctx, func(service *Service) {
	// 		s.logger.Infoln("received service update:", service.ID)
	// 		// handle service update
	// 	}); err != nil {
	// 		s.logger.Errorln("error while subscribing to services:", err)
	// 		s.state = ConnectionStateDisconnected
	// 	}
	// }()

	return nil
}
