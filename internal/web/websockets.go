package web

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sxwebdev/sentinel/internal/storage"
)

// BroadcastServiceUpdate sends service updates to all connected WebSocket clients
func (s *Server) broadcastServiceUpdate(ctx context.Context) {
	// Проверяем, не закрыта ли база данных
	if s.storage == nil {
		fmt.Println("WebSocket broadcast: storage is nil, skipping update")
		return
	}

	services, err := s.monitorService.GetAllServices(ctx)
	if err != nil {
		fmt.Printf("WebSocket broadcast error: failed to get services: %v\n", err)
		return
	}

	// Get incident statistics
	incidentStats, err := s.monitorService.GetAllServicesIncidentStats(ctx)
	if err != nil {
		fmt.Printf("WebSocket broadcast error: failed to get incident stats: %v\n", err)
		return
	}

	// Quick lookup map for incident stats by service ID
	statsMap := make(map[string]*storage.ServiceIncidentStats)
	for _, stats := range incidentStats {
		statsMap[stats.ServiceID] = stats
	}

	// Get services with their states
	servicesWithState := []*ServiceWithState{}
	for _, service := range services {
		serviceWithState, err := s.getServiceWithState(ctx, service)
		if err != nil {
			// Log error but continue with other services
			fmt.Printf("WebSocket broadcast error: failed to get state for service %s: %v\n", service.ID, err)
			continue
		}

		// Add incident statistics to the service
		if stats, exists := statsMap[service.ID]; exists {
			// Add incident stats to the service object
			serviceWithState.Service.ActiveIncidents = stats.ActiveIncidents
			serviceWithState.Service.TotalIncidents = stats.TotalIncidents
		}

		servicesWithState = append(servicesWithState, serviceWithState)
	}

	update := fiber.Map{
		"type":      "service_update",
		"services":  servicesWithState,
		"timestamp": time.Now().Unix(),
	}

	s.wsMutex.Lock()
	defer s.wsMutex.Unlock()

	// Send to all connections and handle errors
	activeConnections := 0
	for conn := range s.wsConnections {
		if err := conn.WriteJSON(update); err != nil {
			fmt.Printf("WebSocket broadcast error: failed to send to connection: %v\n", err)
			delete(s.wsConnections, conn)
			conn.Close()
		} else {
			activeConnections++
		}
	}

	// if activeConnections > 0 {
	// 	fmt.Printf("WebSocket broadcast: sent update to %d connections\n", activeConnections)
	// }
}

func (s *Server) subscribeEvents(ctx context.Context) error {
	broker := s.receiver.ServiceUpdated()
	sub := broker.Subscribe()
	defer broker.Unsubscribe(sub)

	if sub == nil {
		return fmt.Errorf("failed to subscribe to service updates broker")
	}

	// Используем select для обработки событий с проверкой контекста
	for ctx.Err() == nil {
		select {
		case <-sub:
			// Проверяем, не закрыта ли база данных перед отправкой обновлений
			if s.storage == nil {
				fmt.Println("WebSocket: storage is nil, skipping broadcast")
				continue
			}
			s.broadcastServiceUpdate(ctx)

		case <-ctx.Done():
			return nil
		}
	}

	return nil
}
