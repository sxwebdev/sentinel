package web

import (
	"context"
	"fmt"
	"time"

	"github.com/sxwebdev/sentinel/internal/monitors"
	"github.com/sxwebdev/sentinel/internal/storage"
	"github.com/sxwebdev/sentinel/internal/utils"
)

// getServiceWithState gets a service with its current state
func (s *Server) getServiceWithState(ctx context.Context, service *storage.Service) (*ServiceWithState, error) {
	// Get service state
	serviceState, err := s.storage.GetServiceState(ctx, service.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get service state: %w", err)
	}

	res, err := convertServiceToDTO(service)
	if err != nil {
		return nil, fmt.Errorf("failed to convert service to DTO: %w", err)
	}

	return &ServiceWithState{
		Service: res,
		State:   serviceState,
	}, nil
}

// convertServiceToDTO converts a storage.Service to ServiceDTO
func convertServiceToDTO(service *storage.Service) (ServiceDTO, error) {
	config := monitors.Config{}
	if service.Config != nil {
		var err error
		config, err = monitors.ConvertFromMap(service.Config)
		if err != nil {
			return ServiceDTO{}, fmt.Errorf("failed to convert service config: %w", err)
		}
	}

	return ServiceDTO{
		ID:              service.ID,
		Name:            service.Name,
		Protocol:        service.Protocol,
		Interval:        uint32(service.Interval.Milliseconds()),
		Timeout:         uint32(service.Timeout.Milliseconds()),
		Retries:         service.Retries,
		Tags:            service.Tags,
		Config:          config,
		IsEnabled:       service.IsEnabled,
		ActiveIncidents: service.ActiveIncidents,
		TotalIncidents:  service.TotalIncidents,
	}, nil
}

// getDashboardStats calculates dashboard statistics
func (s *Server) getDashboardStats(ctx context.Context) (*DashboardStats, error) {
	// Get all services with their states
	services, err := s.monitorService.FindServices(ctx, storage.FindServicesParams{})
	if err != nil {
		return nil, err
	}

	// Get recent incidents
	activeIncidentsCount, err := s.storage.IncidentsCount(ctx, storage.FindIncidentsParams{
		Resolved: utils.Pointer(false),
	})
	if err != nil {
		return nil, err
	}

	// Get all service states
	serviceStates, err := s.storage.GetAllServiceStates(ctx)
	if err != nil {
		return nil, err
	}

	// Create a map for quick lookup of service states by service ID
	stateMap := make(map[string]*storage.ServiceStateRecord)
	for _, state := range serviceStates {
		stateMap[state.ServiceID] = state
	}

	// Initialize stats
	stats := DashboardStats{
		TotalServices:    int(services.Count),
		ServicesUp:       0,
		ServicesDown:     0,
		ServicesUnknown:  0,
		UptimePercentage: 0.0,
		AvgResponseTime:  0,
		TotalChecks:      0,
		ActiveIncidents:  0,
		LastCheckTime:    nil,
		ChecksPerMinute:  0,
		Protocols:        make(map[storage.ServiceProtocolType]int),
	}

	// Calculate statistics
	totalChecks := 0
	upServices := 0
	var lastCheckTime *time.Time
	var totalResponseTimeMs int64
	var responseTimeCount int64

	for _, service := range services.Items {
		// Get service state
		serviceState := stateMap[service.ID]

		// Count by status
		if serviceState != nil {
			switch serviceState.Status {
			case storage.StatusUp:
				stats.ServicesUp++
				upServices++
			case storage.StatusDown:
				stats.ServicesDown++
			case storage.StatusUnknown:
				stats.ServicesUnknown++
			}

			// Add response time to total (only from services that have response time data)
			if serviceState.ResponseTimeNS != nil && *serviceState.ResponseTimeNS > 0 {
				totalResponseTimeMs += *serviceState.ResponseTimeNS / 1000000 // Convert to milliseconds
				responseTimeCount++
			}
			totalChecks += serviceState.TotalChecks

			// Track last check time
			if serviceState.LastCheck != nil {
				if lastCheckTime == nil || serviceState.LastCheck.After(*lastCheckTime) {
					lastCheckTime = serviceState.LastCheck
				}
			}
		}

		// Count by protocol
		protocol := service.Protocol
		if protocol == "" {
			protocol = "unknown"
		}
		stats.Protocols[protocol]++
	}

	// Calculate averages
	if upServices > 0 {
		stats.UptimePercentage = float64(upServices) / float64(len(services.Items)) * 100
	}
	if responseTimeCount > 0 {
		stats.AvgResponseTime = totalResponseTimeMs / responseTimeCount
	}
	stats.TotalChecks = totalChecks

	// Count active incidents
	stats.ActiveIncidents = int(activeIncidentsCount)

	// Set last check time
	stats.LastCheckTime = lastCheckTime

	// Calculate checks per minute (estimate based on intervals)
	checksPerMinute := 0
	for _, service := range services.Items {
		if service.Interval > 0 {
			checksPerMinute += int(time.Minute / service.Interval)
		}
	}
	stats.ChecksPerMinute = checksPerMinute

	return &stats, nil
}
