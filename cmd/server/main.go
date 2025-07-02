package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
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
	stor, err := storage.NewChainDBStorage(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer stor.Close()

	// Initialize notifier
	var notif notifier.Notifier
	if cfg.Telegram.Enabled {
		notif, err = notifier.NewTelegramNotifier(cfg.Telegram.BotToken, cfg.Telegram.ChatID)
		if err != nil {
			log.Fatalf("Failed to initialize telegram notifier: %v", err)
		}
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
	webServer := web.NewServer(monitorService, cfg)

	// Start web server
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler: webServer.Router(),
	}

	go func() {
		log.Printf("Starting web server on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start web server: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Failed to shutdown web server: %v", err)
	}

	cancel() // Stop scheduler
	log.Println("Shutdown complete")
}
