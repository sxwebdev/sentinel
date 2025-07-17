package web

import (
	"context"
	"fmt"

	"github.com/sxwebdev/sentinel/internal/monitors"
	"github.com/sxwebdev/sentinel/internal/storage"
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
