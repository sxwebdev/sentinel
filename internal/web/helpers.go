package web

import (
	"fmt"

	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/sxwebdev/sentinel/internal/monitors"
)

// convertServiceToDTO converts a models.Service to ServiceDTO
func convertServiceToDTO(service *models.ServiceFullView) (ServiceDTO, error) {
	config := monitors.Config{}
	if service.Config != nil {
		var err error
		config, err = monitors.ConvertFromMap(service.Config)
		if err != nil {
			return ServiceDTO{}, fmt.Errorf("failed to convert service config: %w", err)
		}
	}

	dto := ServiceDTO{
		ID:                 service.ID,
		Name:               service.Name,
		Protocol:           service.Protocol,
		Interval:           service.Interval,
		Timeout:            service.Timeout,
		Retries:            service.Retries,
		Tags:               service.Tags,
		Config:             config,
		IsEnabled:          service.IsEnabled,
		ActiveIncidents:    service.ActiveIncidents,
		TotalIncidents:     service.TotalIncidents,
		Status:             service.Status,
		LastCheck:          service.LastCheck,
		LastError:          service.LastError,
		ConsecutiveFails:   service.ConsecutiveFails,
		ConsecutiveSuccess: service.ConsecutiveSuccess,
		TotalChecks:        service.TotalChecks,
		AvgResponseTime:    service.AvgResponseTime,
	}

	return dto, nil
}

// getDashboardStats calculates dashboard statistics
// func (s *Server) getDashboardStats(ctx context.Context) (*DashboardStats, error) {
// 	// Get all services with their states
// 	services, err := s.baseServices.Services().GetAllEnabled(ctx)
// 	if err != nil {
// 		return nil, err
// 	}

// 	incidentsStatsData, err := s.baseServices.Incidents().Stats(ctx)
// 	if err != nil {
// 		return nil, err
// 	}

// 	servicesStatsData, err := s.baseServices.ServiceStates().Stats(ctx)
// 	if err != nil {
// 		return nil, err
// 	}

// 	incidentsStats := incidentsStatsData.ToDomain()
// 	servicesStats := servicesStatsData.ToDomain()

// 	var totalUptimeAtomic atomic.Value
// 	totalUptimeAtomic.Store(float64(0))

// 	since := time.Now().AddDate(0, -1, 0)

// 	eg, gCtx := errgroup.WithContext(ctx)
// 	for _, svc := range services {
// 		eg.Go(func() error {
// 			svcIncidentsStatsData, err := s.baseServices.Incidents().StatsByServiceID(gCtx, svc.ID, since)
// 			if err != nil {
// 				return err
// 			}

// 			svcIncidentsStats := svcIncidentsStatsData.ToDomain()

// 			totalUptimeAtomic.Store(totalUptimeAtomic.Load().(float64) + svcIncidentsStats.UptimePercentage30d)

// 			return nil
// 		})
// 	}

// 	if err := eg.Wait(); err != nil {
// 		return nil, err
// 	}

// 	var avgUptime float64
// 	totalUptime := totalUptimeAtomic.Load().(float64)
// 	if totalUptime > 0 {
// 		avgUptime = totalUptime / float64(len(services))
// 	}

// 	// Initialize stats
// 	stats := DashboardStats{
// 		TotalServices:    servicesStats.TotalServices,
// 		ServicesUp:       servicesStats.ServicesUp,
// 		ServicesDown:     servicesStats.ServicesDown,
// 		ServicesUnknown:  servicesStats.ServicesUnknown,
// 		UptimePercentage: avgUptime,
// 		AvgResponseTime:  servicesStats.AvgResponseTime,
// 		TotalChecks:      servicesStats.TotalChecks,
// 		ActiveIncidents:  incidentsStats.UnresolvedIncidents,
// 		ChecksPerMinute:  0,
// 		Protocols:        make(map[models.ServiceProtocolType]int),
// 	}

// 	// Calculate statistics

// 	for _, service := range services {
// 		// Count by protocol
// 		protocol := service.Protocol
// 		if protocol == "" {
// 			protocol = "unknown"
// 		}

// 		stats.Protocols[protocol]++
// 	}

// 	// Calculate checks per minute (estimate based on intervals)
// 	var checksPerMinute int64
// 	for _, service := range services {
// 		if service.Interval > 0 {
// 			checksPerMinute += int64(60000 / service.Interval)
// 		}
// 	}
// 	stats.ChecksPerMinute = checksPerMinute

// 	return &stats, nil
// }
