package service

// import (
// 	"context"
// 	"fmt"
// 	"time"
// )

// type Stats struct {
// 	ServiceID           string        `json:"service_id"`
// 	Period              time.Duration `json:"period" swaggertype:"primitive,integer"`
// 	TotalIncidents      int64         `json:"total_incidents"`
// 	TotalDowntime       int64         `json:"total_downtime" swaggertype:"primitive,integer"`
// 	UptimePercentage    float64       `json:"uptime_percentage"`
// 	AvgResponseTime     int64         `json:"avg_response_time" swaggertype:"primitive,integer"`
// 	ResolvedIncidents   int64         `json:"resolved_incidents"`
// 	UnresolvedIncidents int64         `json:"unresolved_incidents"`
// }

// func (s *Service) Stats(ctx context.Context, serviceID string, since time.Time) (*Stats, error) {
// 	if serviceID == "" || since.IsZero() {
// 		return nil, fmt.Errorf("service ID and start time are required for stats")
// 	}

// 	incidentsStats, err := s.store.Incidents().StatsByServiceID(ctx, serviceID, since)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get incidents stats: %w", err)
// 	}

// 	incidentsStatsDomain := incidentsStats.ToDomain()

// 	// Get average response time from service state
// 	var avgResponseTime int64
// 	serviceState, err := s.store.ServiceStates().GetByServiceID(ctx, serviceID)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get service state: %w", err)
// 	}

// 	if serviceState.AvgResponseTime != nil {
// 		avgResponseTime = *serviceState.AvgResponseTime
// 	}

// 	return &Stats{
// 		ServiceID:           serviceID,
// 		Period:              time.Since(since),
// 		TotalIncidents:      incidentsStatsDomain.TotalIncidents,
// 		TotalDowntime:       incidentsStatsDomain.TotalDowntime,
// 		UptimePercentage:    incidentsStats.UptimePercentage30d,
// 		AvgResponseTime:     avgResponseTime,
// 		ResolvedIncidents:   incidentsStatsDomain.ResolvedIncidents,
// 		UnresolvedIncidents: incidentsStatsDomain.UnresolvedIncidents,
// 	}, nil
// }
