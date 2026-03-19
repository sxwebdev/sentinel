package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/sxwebdev/sentinel/internal/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

var (
	enableTCP  = flag.Bool("tcp", false, "Enable TCP server")
	enableGRPC = flag.Bool("grpc", false, "Enable gRPC server")
	enableHTTP = flag.Bool("http", false, "Enable HTTP server")
	tcpPort    = flag.Int("tcp-port", 12345, "TCP server port")
	grpcPort   = flag.Int("grpc-port", 50051, "gRPC server port")
	httpPort   = flag.Int("http-port", 8085, "HTTP server port")
)

func main() {
	flag.Parse()

	// Check if at least one server is enabled
	if !*enableTCP && !*enableGRPC && !*enableHTTP {
		log.Fatal("At least one server must be enabled. Use -tcp, -grpc, or -http flags.")
	}

	if err := run(); err != nil {
		log.Fatalf("Failed to run servers: %v", err)
	}
}

func run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	errChan := make(chan error, 3) // Buffer for all possible servers

	// Start enabled servers
	if *enableTCP {
		wg.Go(func() {
			log.Printf("Starting TCP server on port %d", *tcpPort)
			if err := runTCPServer(ctx, *tcpPort); err != nil {
				errChan <- fmt.Errorf("TCP server error: %w", err)
			}
		})
	}

	if *enableGRPC {
		wg.Go(func() {
			log.Printf("Starting gRPC server on port %d", *grpcPort)
			if err := runGRPCServer(ctx, *grpcPort); err != nil {
				errChan <- fmt.Errorf("gRPC server error: %w", err)
			}
		})
	}

	if *enableHTTP {
		wg.Go(func() {
			log.Printf("Starting HTTP server on port %d", *httpPort)
			if err := runHTTPServer(ctx, *httpPort); err != nil {
				errChan <- fmt.Errorf("HTTP server error: %w", err)
			}
		})
	}

	// Wait for interrupt signal or server error
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigChan:
		log.Println("Received interrupt signal, shutting down...")
		cancel() // Signal all servers to stop
	case err := <-errChan:
		log.Printf("Server error: %v", err)
		cancel()
		return err
	}

	// Wait for all servers to shut down gracefully
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("All servers shut down gracefully")
	case <-time.After(10 * time.Second):
		log.Println("Timeout waiting for servers to shut down")
	}

	return nil
}

// runTCPServer runs the TCP server with original logic from cmd/tcpserver/main.go
func runTCPServer(ctx context.Context, port int) error {
	addr := fmt.Sprintf("localhost:%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to start TCP server: %w", err)
	}
	defer listener.Close()

	// Channel to signal when to stop accepting new connections
	stopChan := make(chan struct{})

	// Goroutine to handle context cancellation
	go func() {
		<-ctx.Done()
		close(stopChan)
		listener.Close() // This will cause Accept() to return an error
	}()

	for {
		select {
		case <-stopChan:
			return nil
		default:
		}

		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-stopChan:
				return nil // Expected error due to context cancellation
			default:
				log.Printf("Failed to accept TCP connection: %v", err)
				continue
			}
		}

		// log.Printf("TCP client connected: %s", conn.RemoteAddr())
		go handleTCPConnection(conn)
	}
}

// handleTCPConnection handles individual TCP connections (original logic from cmd/tcpserver/main.go)
func handleTCPConnection(conn net.Conn) {
	defer func() {
		conn.Close()
		// log.Printf("TCP connection closed: %s\n", conn.RemoteAddr())
	}()

	// Set connection timeout - close if no data received in 5 seconds
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	var accumulated []byte
	reader := bufio.NewReader(conn)

	for {
		buffer := make([]byte, 1024)
		n, err := reader.Read(buffer)
		if err != nil {
			if utils.IsErrTimeout(err) {
				break
			}
			log.Printf("Failed to read from TCP connection: %v", err)
			return
		}

		if n == 0 {
			break
		}

		// Accumulate received data
		accumulated = append(accumulated, buffer[:n]...)

		// If no more data buffered, client likely finished sending
		if reader.Buffered() == 0 {
			break
		}
	}

	// Now process the complete message
	if len(accumulated) == 0 {
		log.Println("No data received from TCP client")
		return
	}

	msg := string(accumulated)
	// log.Printf("TCP complete message received: %s", msg)

	// Simple ping-pong protocol - exact match
	switch msg {
	case "ping":
		// log.Println("TCP: Sending pong")
		_, err := conn.Write([]byte("pong"))
		if err != nil {
			log.Printf("Failed to send TCP response: %v", err)
		}
	case "noresponse":
		log.Println("TCP: No response expected")
	default:
		log.Printf("TCP: Unknown message '%s', sending ok", msg)
		_, err := conn.Write([]byte("ok"))
		if err != nil {
			log.Printf("Failed to send TCP response: %v", err)
		}
	}
}

// runGRPCServer runs the gRPC server with original logic from cmd/grpcserver/main.go
func runGRPCServer(ctx context.Context, port int) error {
	addr := fmt.Sprintf("localhost:%d", port)

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

	// Wait for context cancellation or server error
	select {
	case <-ctx.Done():
		// Graceful shutdown
		log.Println("Shutting down gRPC server...")
		grpcServer.GracefulStop()
		return nil
	case err := <-serverErr:
		return err
	}
}

// runHTTPServer runs a simple HTTP server using standard library
func runHTTPServer(ctx context.Context, port int) error {
	mux := http.NewServeMux()

	// Simple handler that returns {"ok":true}
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	// Start server in a goroutine
	serverErr := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- fmt.Errorf("failed to start HTTP server: %w", err)
		}
	}()

	// Wait for context cancellation or server error
	select {
	case <-ctx.Done():
		log.Println("Shutting down HTTP server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("failed to shutdown HTTP server: %w", err)
		}
		return nil
	case err := <-serverErr:
		return err
	}
}
