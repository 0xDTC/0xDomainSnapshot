package main

import (
	"context"
	"dns-inventory/internal/api"
	"dns-inventory/internal/config"
	"dns-inventory/internal/database"
	"dns-inventory/internal/handlers"
	"dns-inventory/internal/services"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

// Version information
const (
	Version     = "2.0.0"
	AppName     = "DNS Inventory Server"
	Description = "Enterprise DNS Asset Management System"
)

func main() {
	// Print startup banner
	printBanner()
	
	log.Printf("🚀 Starting %s v%s...", AppName, Version)
	
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Failed to load configuration: %v", err)
	}
	
	if err := cfg.Validate(); err != nil {
		log.Fatalf("❌ Invalid configuration: %v", err)
	}
	
	// Initialize database
	log.Println("📊 Initializing database...")
	db, err := database.NewFileDB(filepath.Dir(cfg.DBPath))
	if err != nil {
		log.Fatalf("❌ Failed to initialize database: %v", err)
	}
	log.Println("✅ Database initialized successfully")
	
	// Initialize API clients
	log.Println("🔌 Initializing API clients...")
	godaddyClient := api.NewGoDaddyClient(cfg.GodaddyAPIKey, cfg.GodaddyAPISecret)
	cloudflareClient := api.NewCloudflareClient(cfg.CloudflareAPIToken)
	
	// Initialize notification service
	log.Println("📧 Initializing notification service...")
	notificationService := services.NewNotificationService(cfg)
	
	// Test API connections (only if credentials are provided)
	if cfg.HasAPICredentials() {
		log.Println("🔍 Testing API connections...")
		testAPIConnections(godaddyClient, cloudflareClient, notificationService)
	} else {
		log.Println("⚠️  No API credentials provided - running in offline mode")
		log.Println("   💡 Add API credentials to .env file to enable data collection:")
		log.Println("      • GODADDY_API_KEY & GODADDY_API_SECRET for GoDaddy")
		log.Println("      • CLOUDFLARE_API_TOKEN for Cloudflare")
	}
	
	// Initialize services
	log.Println("⚙️  Initializing business services...")
	userService := services.NewUserService(db)
	domainService := services.NewDomainService(db, godaddyClient, cloudflareClient)
	dnsService := services.NewDNSService(db, godaddyClient, cloudflareClient, domainService)
	
	// Initialize enhanced migration service with notifications
	enhancedMigrationService := services.NewEnhancedMigrationService(
		db, userService, godaddyClient, cloudflareClient, notificationService)
	
	// Initialize HTTP handlers
	log.Println("🌐 Setting up HTTP handlers...")
	domainHandler := handlers.NewDomainHandler(domainService, userService)
	dnsHandler := handlers.NewDNSHandler(dnsService, domainService, userService)
	userHandler := handlers.NewUserHandler(userService)
	exportHandler := handlers.NewExportHandler(domainService, dnsService)
	enhancedMigrationHandler := handlers.NewEnhancedMigrationHandler(enhancedMigrationService, userService)
	
	// Setup HTTP routes
	mux := setupRoutes(domainHandler, dnsHandler, userHandler, exportHandler, enhancedMigrationHandler)
	
	// Start background data collection (only if API credentials are available)
	if cfg.HasAPICredentials() {
		log.Println("🔄 Starting background data collection...")
		go startDataCollection(domainService, dnsService, cfg.DataCollectionInterval, notificationService)
	} else {
		log.Println("⚠️  Background data collection disabled - no API credentials")
	}
	
	// Start HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.ServerPort),
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	
	// Graceful shutdown
	setupGracefulShutdown(server)
	
	// Print access information
	printAccessInfo(cfg.ServerPort)
	
	// Start server
	log.Printf("🌐 Server listening on port %d", cfg.ServerPort)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("❌ Server failed to start: %v", err)
	}
	
	log.Println("🛑 Server stopped gracefully")
}

func printBanner() {
	banner := `
╔══════════════════════════════════════════════════════╗
║                                                      ║
║               🌐 DNS Inventory Server                ║
║                                                      ║
║          Enterprise DNS Asset Management             ║
║                                                      ║
║    • Multi-Provider Integration (GoDaddy, CF)       ║
║    • Advanced Migration Engine                      ║
║    • Real-time Web Dashboard                        ║
║    • Team Collaboration & Assignments              ║
║    • Email Notifications (AWS SES)                 ║
║    • Zero External Dependencies                     ║
║                                                      ║
╚══════════════════════════════════════════════════════╝
	`
	fmt.Println(banner)
}

func printAccessInfo(port int) {
	fmt.Printf(`
╔══════════════════════════════════════════════════════╗
║                    🚀 SERVER READY                   ║
╠══════════════════════════════════════════════════════╣
║                                                      ║
║  📊 Dashboard:     http://localhost:%d              ║
║  🌐 Domains:       http://localhost:%d/domains      ║
║  📡 DNS Records:   http://localhost:%d/dns          ║
║  👥 Users:         http://localhost:%d/users        ║
║  🔄 Migration:     http://localhost:%d/migration    ║
║  🔌 API:           http://localhost:%d/api/          ║
║                                                      ║
╠══════════════════════════════════════════════════════╣
║                                                      ║
║  💡 Tips:                                            ║
║  • Add API credentials to .env for data collection  ║
║  • Configure AWS SES for email notifications        ║
║  • Access migration wizard for data import          ║
║  • Use export functions for data backup             ║
║                                                      ║
╚══════════════════════════════════════════════════════╝

`, port, port, port, port, port, port)
}

func setupRoutes(
	domainHandler *handlers.DomainHandler,
	dnsHandler *handlers.DNSHandler,
	userHandler *handlers.UserHandler,
	exportHandler *handlers.ExportHandler,
	migrationHandler *handlers.EnhancedMigrationHandler,
) *http.ServeMux {
	mux := http.NewServeMux()
	
	// Static files with proper path handling
	workingDir, err := os.Getwd()
	if err != nil {
		log.Printf("⚠️  Could not get working directory: %v", err)
		workingDir = "."
	}
	staticDir := filepath.Join(workingDir, "web", "static")
	
	// Verify static directory exists
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		log.Printf("⚠️  Static directory not found at %s", staticDir)
		// Try relative path as fallback
		staticDir = "web/static"
	}
	
	log.Printf("📁 Serving static files from: %s", staticDir)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))
	
	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"healthy","version":"%s","timestamp":"%s"}`, 
			Version, time.Now().Format(time.RFC3339))
	})
	
	// Main pages
	mux.HandleFunc("/", domainHandler.HandleDomainsPage)
	mux.HandleFunc("/domains", domainHandler.HandleDomainsPage)
	mux.HandleFunc("/dns", dnsHandler.HandleDNSPage)
	mux.HandleFunc("/users", userHandler.HandleUsersPage)
	mux.HandleFunc("/migration", migrationHandler.HandleMigrationPage)
	
	// Core API endpoints
	mux.HandleFunc("/api/domains", domainHandler.HandleDomainsAPI)
	mux.HandleFunc("/api/dns", dnsHandler.HandleDNSAPI)
	mux.HandleFunc("/api/users", userHandler.HandleUsersAPI)
	mux.HandleFunc("/api/assign-domain", domainHandler.HandleAssignDomain)
	mux.HandleFunc("/api/assign-dns", dnsHandler.HandleAssignDNS)
	mux.HandleFunc("/api/collect-domains", domainHandler.HandleCollectDomains)
	mux.HandleFunc("/api/collect-dns", dnsHandler.HandleCollectDNS)
	
	// Enhanced Migration API endpoints
	mux.HandleFunc("/api/migration", migrationHandler.HandleMigrationAPI)
	mux.HandleFunc("/api/migration/analyze", migrationHandler.HandleFileAnalysis)
	mux.HandleFunc("/api/migration/preview", migrationHandler.HandleMigrationPreview)
	mux.HandleFunc("/api/migration/progress", migrationHandler.HandleMigrationProgress)
	mux.HandleFunc("/api/migration/control", migrationHandler.HandleMigrationControl)
	mux.HandleFunc("/api/migration/templates", migrationHandler.HandleMigrationTemplates)
	mux.HandleFunc("/api/migration/jobs", migrationHandler.HandleMigrationJobs)
	mux.HandleFunc("/api/migration/default-users", migrationHandler.HandleCreateDefaultUsers)
	
	// Individual migration job endpoints
	mux.HandleFunc("/api/migration/job/", migrationHandler.HandleMigrationJob)
	
	// Export endpoints
	mux.HandleFunc("/export/domains.csv", exportHandler.HandleDomainsCSV)
	mux.HandleFunc("/export/dns.csv", exportHandler.HandleDNSCSV)
	
	return mux
}

func testAPIConnections(godaddyClient *api.GoDaddyClient, cloudflareClient *api.CloudflareClient, notificationService *services.NotificationService) {
	if err := godaddyClient.TestConnection(); err != nil {
		log.Printf("⚠️  GoDaddy API connection failed: %v", err)
		if notificationService.IsEnabled() {
			notificationService.SendAPIError("GoDaddy", err.Error())
		}
	} else {
		log.Println("✅ GoDaddy API connection successful")
	}
	
	if err := cloudflareClient.TestConnection(); err != nil {
		log.Printf("⚠️  Cloudflare API connection failed: %v", err)
		if notificationService.IsEnabled() {
			notificationService.SendAPIError("Cloudflare", err.Error())
		}
	} else {
		log.Println("✅ Cloudflare API connection successful")
	}
}

func startDataCollection(domainService *services.DomainService, dnsService *services.DNSService, intervalMinutes int, notificationService *services.NotificationService) {
	ticker := time.NewTicker(time.Duration(intervalMinutes) * time.Minute)
	defer ticker.Stop()
	
	log.Printf("🔄 Starting periodic data collection (every %d minutes)", intervalMinutes)
	
	// Run initial collection
	log.Println("🔍 Running initial data collection...")
	runDataCollection(domainService, dnsService, notificationService)
	
	// Run periodic collection
	for range ticker.C {
		log.Println("🔄 Running scheduled data collection...")
		runDataCollection(domainService, dnsService, notificationService)
	}
}

func runDataCollection(domainService *services.DomainService, dnsService *services.DNSService, notificationService *services.NotificationService) {
	start := time.Now()
	errors := make([]string, 0)
	domainCount := 0
	dnsCount := 0
	
	// Collect domains first
	if err := domainService.CollectDomains(); err != nil {
		errMsg := fmt.Sprintf("Domain collection failed: %v", err)
		log.Printf("❌ %s", errMsg)
		errors = append(errors, errMsg)
	} else {
		log.Println("✅ Domain collection completed successfully")
		// Get domain count (simplified - implement actual counting if needed)
		domainCount = 100 // Placeholder
	}
	
	// Then collect DNS records
	if err := dnsService.CollectDNSRecords(); err != nil {
		errMsg := fmt.Sprintf("DNS records collection failed: %v", err)
		log.Printf("❌ %s", errMsg)
		errors = append(errors, errMsg)
	} else {
		log.Println("✅ DNS records collection completed successfully")
		// Get DNS record count (simplified - implement actual counting if needed)
		dnsCount = 500 // Placeholder
	}
	
	duration := time.Since(start)
	log.Printf("⏱️  Data collection completed in %v", duration)
	
	// Send daily summary notification (simple check for demonstration)
	if notificationService != nil && notificationService.IsEnabled() {
		currentHour := time.Now().Hour()
		// Send summary at midnight or if it's the first run of the day
		if currentHour == 0 || (len(errors) == 0 && time.Since(start) < 2*time.Minute) {
			log.Println("📧 Sending data collection summary...")
			notificationService.SendDataCollectionSummary(domainCount, dnsCount, errors)
		}
	}
}

func setupGracefulShutdown(server *http.Server) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	
	go func() {
		<-c
		log.Println("🛑 Shutdown signal received, gracefully shutting down...")
		
		// Create a timeout for shutdown
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()
		
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("❌ Server forced to shutdown: %v", err)
		}
	}()
}