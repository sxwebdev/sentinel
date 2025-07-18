package web

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/sxwebdev/sentinel/internal/receiver"
)

type websocketEvent struct {
	Type      string `json:"type"`
	Timestamp int64  `json:"timestamp"`
	Data      any    `json:"data"`
}

// handleWebSocket handles WebSocket connections
func (s *Server) handleWebSocket(c *websocket.Conn) {
	fmt.Printf("WebSocket: new connection from %s\n", c.RemoteAddr())

	// Add connection to the map
	s.wsMutex.Lock()
	s.wsConnections[c] = true
	connectionCount := len(s.wsConnections)
	s.wsMutex.Unlock()

	fmt.Printf("WebSocket: total connections: %d\n", connectionCount)

	// Remove connection when it closes
	defer func() {
		s.wsMutex.Lock()
		delete(s.wsConnections, c)
		remainingConnections := len(s.wsConnections)
		s.wsMutex.Unlock()
		c.Close()
		fmt.Printf("WebSocket: connection closed, remaining connections: %d\n", remainingConnections)
	}()

	// Keep connection alive and handle messages
	for {
		_, _, err := c.ReadMessage()
		if err != nil {
			if !strings.Contains(err.Error(), "close 1001") {
				fmt.Printf("WebSocket: read error: %v\n", err)
			}
			break
		}
	}
}

// BroadcastServiceUpdate sends service updates to all connected WebSocket clients
func (s *Server) broadcastServiceTriggered(data receiver.TriggerServiceData) error {
	if s.storage == nil {
		return nil
	}

	svc := ServiceDTO{
		ID: data.Svc.ID,
	}

	var err error
	if data.EventType != receiver.TriggerServiceEventTypeDeleted {
		svc, err = convertServiceToDTO(data.Svc)
		if err != nil {
			return fmt.Errorf("failed to convert service to DTO: %w", err)
		}
	}

	update := websocketEvent{
		Type:      "service_" + data.EventType.String(),
		Data:      svc,
		Timestamp: time.Now().Unix(),
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

	return nil
}

// broadcastStatsUpdate sends dashboard statistics updates to all connected WebSocket clients
func (s *Server) broadcastStatsUpdate(ctx context.Context) error {
	if s.storage == nil {
		return nil
	}

	stats, err := s.getDashboardStats(ctx)
	if err != nil {
		return fmt.Errorf("failed to get dashboard stats: %w", err)
	}

	update := websocketEvent{
		Type:      "stats_update",
		Data:      stats,
		Timestamp: time.Now().Unix(),
	}

	s.wsMutex.Lock()
	defer s.wsMutex.Unlock()

	// Send to all connections and handle errors
	activeConnections := 0
	for conn := range s.wsConnections {
		if err := conn.WriteJSON(update); err != nil {
			fmt.Printf("WebSocket broadcast error: failed to send stats update: %v\n", err)
			delete(s.wsConnections, conn)
			conn.Close()
		} else {
			activeConnections++
		}
	}

	return nil
}

func (s *Server) subscribeEvents(ctx context.Context) error {
	broker := s.receiver.TriggerService()
	sub := broker.Subscribe()
	defer broker.Unsubscribe(sub)

	if sub == nil {
		return fmt.Errorf("failed to subscribe to service updates broker")
	}

	for ctx.Err() == nil {
		select {
		case data := <-sub:
			if s.storage == nil ||
				data.EventType == receiver.TriggerServiceEventTypeCheck ||
				data.EventType == receiver.TriggerServiceEventTypeUnknown {
				continue
			}

			// Broadcast service triggered event
			if err := s.broadcastServiceTriggered(data); err != nil {
				fmt.Printf("WebSocket broadcast error: %v\n", err)
			}

			// Broadcast stats update
			if err := s.broadcastStatsUpdate(ctx); err != nil {
				fmt.Printf("WebSocket broadcast error: %v\n", err)
			}

		case <-ctx.Done():
			return nil
		}
	}

	return nil
}
