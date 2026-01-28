package api

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"0xdomainsnapshot/internal/config"
	"0xdomainsnapshot/internal/scheduler"
	"0xdomainsnapshot/internal/service"
)

// Server is the HTTP API server
type Server struct {
	router    chi.Router
	cfg       config.ServerConfig
	scheduler *scheduler.Scheduler
	syncSvc   *service.SyncService
	exportSvc *service.ExportService
}

// NewServer creates a new API server
func NewServer(
	cfg config.ServerConfig,
	sched *scheduler.Scheduler,
	syncSvc *service.SyncService,
	exportSvc *service.ExportService,
) *Server {
	s := &Server{
		router:    chi.NewRouter(),
		cfg:       cfg,
		scheduler: sched,
		syncSvc:   syncSvc,
		exportSvc: exportSvc,
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

// setupMiddleware configures middleware
func (s *Server) setupMiddleware() {
	// Request ID
	s.router.Use(middleware.RequestID)

	// Logging
	s.router.Use(middleware.Logger)

	// Panic recovery
	s.router.Use(middleware.Recoverer)

	// CORS
	s.router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"Link", "X-Request-ID"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Set content type for JSON responses
	s.router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only set for API routes
			if strings.HasPrefix(r.URL.Path, "/api/") {
				w.Header().Set("Content-Type", "application/json")
			}
			next.ServeHTTP(w, r)
		})
	})
}

// setupRoutes configures API routes
func (s *Server) setupRoutes() {
	// API routes
	s.router.Route("/api/v1", func(r chi.Router) {
		// Health check
		r.Get("/health", s.handleHealth)

		// Sync endpoints
		r.Route("/sync", func(r chi.Router) {
			r.Get("/status", s.handleSyncStatus)
			r.Get("/status/{collector}", s.handleCollectorStatus)
			r.Post("/trigger/{collector}", s.handleTriggerSync)
			r.Post("/trigger-all", s.handleTriggerSyncAll)
		})

		// Data endpoints
		r.Get("/domains", s.handleGetDomains)
		r.Get("/dns-records", s.handleGetDNSRecords)

		// Export endpoint
		r.Post("/export", s.handleExport)

		// Scheduler info
		r.Get("/scheduler/jobs", s.handleSchedulerJobs)
	})

	// Serve data/*.json files with JSON content type
	s.router.Get("/data/*", s.handleDataFiles)

	// Serve static files (HTML, CSS, JS, etc.)
	s.setupStaticFiles()
}

// setupStaticFiles configures static file serving
func (s *Server) setupStaticFiles() {
	staticDir := s.cfg.StaticDir
	if staticDir == "" {
		staticDir = "."
	}

	// Resolve to absolute path
	absPath, err := filepath.Abs(staticDir)
	if err != nil {
		log.Printf("[Server] Warning: could not resolve static dir: %v", err)
		absPath = staticDir
	}

	log.Printf("[Server] Serving static files from: %s", absPath)

	// Create file server
	fs := http.FileServer(http.Dir(absPath))

	// Serve files with custom handler for caching
	s.router.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Set cache headers based on file type
		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".css", ".js":
			w.Header().Set("Cache-Control", "public, max-age=3600")
		case ".json":
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Cache-Control", "no-cache")
		case ".html", ".htm":
			w.Header().Set("Cache-Control", "no-cache")
		case ".png", ".jpg", ".jpeg", ".gif", ".svg", ".ico":
			w.Header().Set("Cache-Control", "public, max-age=86400")
		}

		fs.ServeHTTP(w, r)
	})
}

// handleDataFiles serves JSON files from the data directory
func (s *Server) handleDataFiles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")

	// Get the file path
	path := chi.URLParam(r, "*")
	if path == "" {
		path = r.URL.Path[len("/data/"):]
	}

	// Serve from static dir
	filePath := filepath.Join(s.cfg.StaticDir, "data", path)
	http.ServeFile(w, r, filePath)
}

// ServeHTTP implements http.Handler
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// ListenAndServe starts the HTTP server
func (s *Server) ListenAndServe() error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
	log.Printf("[Server] Starting on http://%s", addr)
	return http.ListenAndServe(addr, s)
}

// Addr returns the server address
func (s *Server) Addr() string {
	return fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
}
