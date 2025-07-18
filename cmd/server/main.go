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
	"github.com/sxwebdev/sentinel/internal/receiver"
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

	// set default timezone
	time.Local, err = time.LoadLocation(cfg.Timezone)
	if err != nil {
		return fmt.Errorf("failed to set timezone: %w", err)
	}

	// Initialize storage
	stor, err := storage.NewStorage(storage.StorageTypeSQLite, cfg.Database.Path)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Initialize notifier
	var notif notifier.Notifier
	if cfg.Notifications.Enabled {
		notif, err = notifier.NewNotifier(cfg.Notifications.URLs)
		if err != nil {
			return fmt.Errorf("failed to initialize notifier: %w", err)
		}
	}

	rc := receiver.New()
	defer func() {
		_ = rc.Stop(context.Background())
	}()

	if err := rc.Start(ctx); err != nil {
		return err
	}

	// Create monitor service
	monitorService := service.NewMonitorService(stor, cfg, notif, rc)

	// Initialize scheduler
	sched := scheduler.New(monitorService, rc)
	defer func() {
		sched.Stop(context.Background())
	}()

	// Start scheduler with initial check
	shutdownCtx, shutdownCancel := context.WithCancel(ctx)
	defer shutdownCancel()

	schedulerErr := make(chan error, 1)
	go func() {
		if err := sched.Start(shutdownCtx); err != nil {
			schedulerErr <- fmt.Errorf("scheduler error: %w", err)
		}
	}()

	// Create web server
	webServer, err := web.NewServer(cfg, monitorService, stor, rc)
	if err != nil {
		return fmt.Errorf("failed to initialize web server: %w", err)
	}
	go func() {
		schedulerErr <- webServer.Start(ctx)
	}()
	defer func() {
		_ = webServer.Stop(context.Background())
	}()

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

	// Graceful shutdown - сначала останавливаем WebSocket, потом базу данных
	shutdownCtx, shutdownCancel = context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer shutdownCancel()

	if err := webServer.App().ShutdownWithContext(shutdownCtx); err != nil {
		return fmt.Errorf("failed to shutdown Fiber server: %w", err)
	}

	// Закрываем базу данных в последнюю очередь
	defer stor.Close()

	return nil
}
