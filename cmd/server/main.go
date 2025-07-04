package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sxwebdev/sentinel/internal/config"
	"github.com/sxwebdev/sentinel/internal/notifier"
	"github.com/sxwebdev/sentinel/internal/scheduler"
	"github.com/sxwebdev/sentinel/internal/service"
	"github.com/sxwebdev/sentinel/internal/storage"
	"github.com/sxwebdev/sentinel/internal/web"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := run(ctx); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	// Load configuration
	cfg, err := config.Load("config.yaml")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize storage
	stor, err := storage.NewStorage(storage.StorageTypeSQLite, cfg.Database.Path)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	defer stor.Close()

	// Initialize notifier
	var notif notifier.Notifier
	if cfg.Notifications.Enabled {
		notif, err = notifier.NewNotifier(cfg.Notifications.URLs)
		if err != nil {
			return fmt.Errorf("failed to initialize notifier: %w", err)
		}
	}

	// Initialize monitor service
	monitorService := service.NewMonitorService(stor, notif)

	// Initialize scheduler
	sched := scheduler.NewScheduler(monitorService)

	// Set scheduler in monitor service
	monitorService.SetScheduler(sched)

	// Load services from database
	if err := monitorService.LoadServicesFromStorage(ctx); err != nil {
		return fmt.Errorf("failed to load services from storage: %w", err)
	}

	// Start scheduler with initial check
	shutdownCtx, shutdownCancel := context.WithCancel(ctx)
	defer shutdownCancel()

	schedulerErr := make(chan error, 1)
	go func() {
		if err := sched.StartWithInitialCheck(shutdownCtx); err != nil {
			schedulerErr <- fmt.Errorf("scheduler error: %w", err)
		}
	}()

	// Initialize web server
	webServer, err := web.NewServer(monitorService, cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize web server: %w", err)
	}

	// Set web server in monitor service for WebSocket broadcasts
	monitorService.SetWebServer(webServer)

	// Start Fiber server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	serverErr := make(chan error, 1)
	go func() {
		if err := webServer.App().Listen(addr); err != nil {
			serverErr <- fmt.Errorf("failed to start Fiber server: %w", err)
		}
	}()

	// Wait for interrupt signal or server error
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigChan:
		// Graceful shutdown
		shutdownCancel()
	case err := <-serverErr:
		shutdownCancel()
		return err
	case err := <-schedulerErr:
		return err
	}

	// Graceful shutdown
	shutdownCtx, shutdownCancel = context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer shutdownCancel()

	if err := webServer.App().ShutdownWithContext(shutdownCtx); err != nil {
		return fmt.Errorf("failed to shutdown Fiber server: %w", err)
	}

	return nil
}
