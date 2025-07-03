package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

var port = flag.Int("port", 50051, "The server port")

func main() {
	flag.Parse()

	if err := run(); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}

func run() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	// Create gRPC server
	grpcServer := grpc.NewServer()

	// Create and register health server
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)

	// Register gRPC reflection service for debugging
	reflection.Register(grpcServer)

	// Set initial serving status
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus("health", grpc_health_v1.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus("test-service", grpc_health_v1.HealthCheckResponse_SERVING)

	log.Printf("gRPC server listening on port %d", *port)
	log.Printf("Health check endpoint: grpc://localhost:%d/grpc.health.v1.Health/Check", *port)
	log.Printf("Available services: '', 'health', 'test-service'")

	// Start server in a goroutine
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("Failed to serve: %v", err)
		}
	}()

	// Simulate service status changes for testing
	// go simulateServiceStatusChanges(healthServer)

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down gRPC server...")

	// Graceful shutdown
	grpcServer.GracefulStop()

	return nil
}

// simulateServiceStatusChanges simulates service status changes for testing
// func simulateServiceStatusChanges(healthServer *health.Server) {
// 	ticker := time.NewTicker(30 * time.Second)
// 	defer ticker.Stop()

// 	status := grpc_health_v1.HealthCheckResponse_SERVING

// 	for range ticker.C {
// 		// Toggle service status for testing
// 		if status == grpc_health_v1.HealthCheckResponse_SERVING {
// 			status = grpc_health_v1.HealthCheckResponse_NOT_SERVING
// 			log.Println("Setting test-service status to NOT_SERVING")
// 		} else {
// 			status = grpc_health_v1.HealthCheckResponse_SERVING
// 			log.Println("Setting test-service status to SERVING")
// 		}

// 		healthServer.SetServingStatus("test-service", status)
// 	}
// }
