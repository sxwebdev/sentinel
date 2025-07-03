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
	// Load configuration
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize storage
	stor, err := storage.NewStorage(storage.StorageTypeSQLite, cfg.Database.Path)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer stor.Close()

	// Initialize notifier
	var notif notifier.Notifier
	if cfg.Notifications.Enabled {
		notif, err = notifier.NewNotifier(cfg.Notifications.URLs)
		if err != nil {
			log.Fatalf("Failed to initialize notifier: %v", err)
		}
		log.Printf("Initialized notifications with %d providers", len(cfg.Notifications.URLs))
	}

	// Initialize monitor service
	monitorService := service.NewMonitorService(stor, notif)

	// Initialize scheduler
	sched := scheduler.NewScheduler(monitorService)

	// Start monitoring services
	for _, svcCfg := range cfg.Services {
		if err := sched.AddService(svcCfg); err != nil {
			log.Printf("Failed to add service %s: %v", svcCfg.Name, err)
			continue
		}
	}

	// Start scheduler
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go sched.Start(ctx)

	// Initialize web server
	webServer, err := web.NewServer(monitorService, cfg)
	if err != nil {
		log.Fatalf("Failed to initialize web server: %v", err)
	}

	// Start Fiber server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	go func() {
		log.Printf("Starting Fiber server on %s", addr)
		if err := webServer.App().Listen(addr); err != nil {
			log.Fatalf("Failed to start Fiber server: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")

	// Graceful shutdown

	// Cancel context first to stop scheduler
	log.Println("Stopping scheduler...")
	cancel()
	log.Println("Scheduler stopped")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer shutdownCancel()

	if err := webServer.App().ShutdownWithContext(shutdownCtx); err != nil {
		log.Printf("Failed to shutdown Fiber server: %v", err)
	}

	log.Println("Shutdown complete")
}
