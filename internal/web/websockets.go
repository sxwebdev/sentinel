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
		// Recover from potential panic to ensure cleanup
		if r := recover(); r != nil {
			fmt.Printf("WebSocket: panic recovered: %v\n", r)
		}

		s.wsMutex.Lock()
		delete(s.wsConnections, c)
		remainingConnections := len(s.wsConnections)
		s.wsMutex.Unlock()
		c.Close()
		fmt.Printf("WebSocket: connection closed, remaining connections: %d\n", remainingConnections)
	}()

	// Set read timeout to prevent goroutine from hanging forever
	c.SetReadDeadline(time.Now().Add(60 * time.Second))

	// Set pong handler to handle ping/pong for keepalive
	c.SetPongHandler(func(appData string) error {
		c.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// Keep connection alive and handle messages
	for {
		// Set read deadline for each message to prevent hanging
		c.SetReadDeadline(time.Now().Add(60 * time.Second))

		_, _, err := c.ReadMessage()
		if err != nil {
			if !strings.Contains(err.Error(), "close 1001") &&
				!strings.Contains(err.Error(), "timeout") &&
				!websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("WebSocket: read error: %v\n", err)
			}
			break
		}

		// Reset read deadline after successful read
		c.SetReadDeadline(time.Now().Add(60 * time.Second))
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
	var connectionsToRemove []*websocket.Conn

	for conn := range s.wsConnections {
		// Set write deadline to prevent hanging
		conn.SetWriteDeadline(time.Now().Add(5 * time.Second))

		if err := conn.WriteJSON(update); err != nil {
			fmt.Printf("WebSocket broadcast error: failed to send to connection: %v\n", err)
			connectionsToRemove = append(connectionsToRemove, conn)
		} else {
			activeConnections++
		}
	}

	// Remove failed connections outside the range loop to avoid map modification during iteration
	for _, conn := range connectionsToRemove {
		delete(s.wsConnections, conn)
		conn.Close()
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
	var connectionsToRemove []*websocket.Conn

	for conn := range s.wsConnections {
		// Set write deadline to prevent hanging
		conn.SetWriteDeadline(time.Now().Add(5 * time.Second))

		if err := conn.WriteJSON(update); err != nil {
			fmt.Printf("WebSocket broadcast error: failed to send stats update: %v\n", err)
			connectionsToRemove = append(connectionsToRemove, conn)
		} else {
			activeConnections++
		}
	}

	// Remove failed connections outside the range loop to avoid map modification during iteration
	for _, conn := range connectionsToRemove {
		delete(s.wsConnections, conn)
		conn.Close()
	}

	return nil
}

func (s *Server) subscribeEvents(ctx context.Context) error {
	broker := s.receiver.TriggerService()
	sub := broker.Subscribe()

	if sub == nil {
		return fmt.Errorf("failed to subscribe to service updates broker")
	}

	// Ensure unsubscription happens even if panic occurs
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("WebSocket subscribeEvents panic recovered: %v\n", r)
		}
		broker.Unsubscribe(sub)
	}()

	// Use a ticker for periodic cleanup and health checks
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case data, ok := <-sub:
			if !ok {
				// Channel closed, exit gracefully
				return nil
			}

			// reset ticker
			ticker.Reset(30 * time.Second)

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

		case <-ticker.C:
			// Periodic cleanup of dead connections (commented out for now)
			// Can be implemented if needed for additional cleanup logic

		case <-ctx.Done():
			return nil
		}
	}
}
