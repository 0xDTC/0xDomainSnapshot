package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"0xdomainsnapshot/internal/api"
	"0xdomainsnapshot/internal/collector"
	"0xdomainsnapshot/internal/collector/dns"
	"0xdomainsnapshot/internal/config"
	"0xdomainsnapshot/internal/database"
	"0xdomainsnapshot/internal/scheduler"
	"0xdomainsnapshot/internal/service"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting 0xDomainSnapshot Backend...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	log.Printf("Configuration loaded successfully")
	log.Printf("  Server: %s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("  Static directory: %s", cfg.Server.StaticDir)
	log.Printf("  Scheduler enabled: %v", cfg.Scheduler.Enabled)

	// Connect to database
	log.Println("Connecting to database...")
	db, err := database.New(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("Database connected successfully")

	// Run migrations
	log.Println("Running database migrations...")
	ctx := context.Background()
	if err := db.RunMigrations(ctx); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Migrations completed")

	// Create services
	syncSvc := service.NewSyncService(db)
	exportSvc := service.NewExportService(syncSvc, cfg.Export)
	syncLock := scheduler.NewSyncLock(db)

	// Create collector registry
	registry := collector.NewRegistry()

	// Register DNS collectors
	if cfg.GoDaddy.IsConfigured() {
		gdCollector := dns.NewGoDaddyCollector(cfg.GoDaddy, cfg.RateLimit)
		if err := registry.Register(gdCollector); err != nil {
			log.Printf("Warning: Failed to register GoDaddy collector: %v", err)
		} else {
			log.Println("GoDaddy DNS collector registered")
		}
	} else {
		log.Println("GoDaddy collector skipped (not configured)")
	}

	if cfg.Cloudflare.IsConfigured() {
		cfCollector := dns.NewCloudflareCollector(cfg.Cloudflare, cfg.RateLimit)
		if err := registry.Register(cfCollector); err != nil {
			log.Printf("Warning: Failed to register Cloudflare collector: %v", err)
		} else {
			log.Println("Cloudflare DNS collector registered")
		}
	} else {
		log.Println("Cloudflare collector skipped (not configured)")
	}

	log.Printf("Registered %d collectors: %v", registry.Count(), registry.Names())

	// Create scheduler
	sched := scheduler.New(registry, syncSvc, exportSvc, syncLock, cfg.Scheduler)

	// Create API server
	server := api.NewServer(cfg.Server, sched, syncSvc, exportSvc)

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start scheduler in background
	go func() {
		if err := sched.Start(ctx); err != nil {
			log.Printf("Scheduler error: %v", err)
		}
	}()

	// Start server in background
	serverErr := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil {
			serverErr <- err
		}
	}()

	// Print startup complete
	log.Println("")
	log.Println("=================================================")
	log.Println("  0xDomainSnapshot Backend Started Successfully")
	log.Println("=================================================")
	log.Printf("  Dashboard: http://%s", server.Addr())
	log.Printf("  API:       http://%s/api/v1/health", server.Addr())
	log.Println("")
	log.Println("Available API Endpoints:")
	log.Println("  GET  /api/v1/health              - Health check")
	log.Println("  GET  /api/v1/sync/status         - All collector statuses")
	log.Println("  GET  /api/v1/sync/status/{name}  - Single collector status")
	log.Println("  POST /api/v1/sync/trigger/{name} - Trigger manual sync")
	log.Println("  POST /api/v1/sync/trigger-all    - Trigger all syncs")
	log.Println("  GET  /api/v1/domains             - Get domains")
	log.Println("  GET  /api/v1/dns-records         - Get DNS records")
	log.Println("  POST /api/v1/export              - Export JSON files")
	log.Println("  GET  /api/v1/scheduler/jobs      - Scheduled jobs")
	log.Println("")
	log.Println("Press Ctrl+C to stop")
	log.Println("")

	// Wait for shutdown signal or server error
	select {
	case sig := <-sigChan:
		log.Printf("Received signal: %v", sig)
	case err := <-serverErr:
		log.Printf("Server error: %v", err)
	}

	// Graceful shutdown
	log.Println("Initiating graceful shutdown...")
	cancel()

	// Give services time to stop
	time.Sleep(2 * time.Second)

	log.Println("Shutdown complete")
}
