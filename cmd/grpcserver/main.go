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
	addr := fmt.Sprintf("localhost:%d", *port)
	log.Printf("Starting gRPC server on %s\n", addr)

	lis, err := net.Listen("tcp", addr)
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

	// Start server in a goroutine
	serverErr := make(chan error, 1)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			serverErr <- fmt.Errorf("failed to serve: %w", err)
		}
	}()

	// Wait for interrupt signal or server error
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigChan:
		// Graceful shutdown
	case err := <-serverErr:
		return err
	}

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
