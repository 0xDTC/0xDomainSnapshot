package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Response helpers

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

// Health check

// handleHealth handles GET /api/v1/health
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
	})
}

// Sync endpoints

// handleSyncStatus handles GET /api/v1/sync/status
func (s *Server) handleSyncStatus(w http.ResponseWriter, r *http.Request) {
	statuses, err := s.scheduler.GetAllStatus(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if statuses == nil {
		statuses = []interface{}{}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"collectors": statuses,
	})
}

// handleCollectorStatus handles GET /api/v1/sync/status/{collector}
func (s *Server) handleCollectorStatus(w http.ResponseWriter, r *http.Request) {
	collectorName := chi.URLParam(r, "collector")
	if collectorName == "" {
		respondError(w, http.StatusBadRequest, "collector name required")
		return
	}

	// Check if running
	running, err := s.scheduler.IsCollectorRunning(r.Context(), collectorName)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Get last status
	status, err := s.scheduler.GetCollectorStatus(r.Context(), collectorName)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Get next run time
	nextRun := s.scheduler.GetNextRun(collectorName)

	response := map[string]interface{}{
		"collector":  collectorName,
		"is_running": running,
		"next_run":   nextRun,
	}

	if status != nil {
		response["last_run"] = status
	}

	respondJSON(w, http.StatusOK, response)
}

// handleTriggerSync handles POST /api/v1/sync/trigger/{collector}
func (s *Server) handleTriggerSync(w http.ResponseWriter, r *http.Request) {
	collectorName := chi.URLParam(r, "collector")
	if collectorName == "" {
		respondError(w, http.StatusBadRequest, "collector name required")
		return
	}

	// Check if already running
	running, err := s.scheduler.IsCollectorRunning(r.Context(), collectorName)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if running {
		respondJSON(w, http.StatusConflict, map[string]interface{}{
			"status":    "already_running",
			"collector": collectorName,
			"message":   "Sync is already in progress",
		})
		return
	}

	// Trigger sync
	err = s.scheduler.TriggerSync(r.Context(), collectorName)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusAccepted, map[string]interface{}{
		"status":    "started",
		"collector": collectorName,
		"message":   "Sync started in background",
	})
}

// handleTriggerSyncAll handles POST /api/v1/sync/trigger-all
func (s *Server) handleTriggerSyncAll(w http.ResponseWriter, r *http.Request) {
	err := s.scheduler.TriggerSyncAll(r.Context())
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusAccepted, map[string]interface{}{
		"status":  "started",
		"message": "All syncs started in background",
	})
}

// Data endpoints

// handleGetDomains handles GET /api/v1/domains
func (s *Server) handleGetDomains(w http.ResponseWriter, r *http.Request) {
	// Query parameters
	status := r.URL.Query().Get("status")
	source := r.URL.Query().Get("source")

	domains, err := s.syncSvc.GetDomains(r.Context(), status, source)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if domains == nil {
		domains = []map[string]interface{}{}
	}

	respondJSON(w, http.StatusOK, domains)
}

// handleGetDNSRecords handles GET /api/v1/dns-records
func (s *Server) handleGetDNSRecords(w http.ResponseWriter, r *http.Request) {
	// Query parameters
	status := r.URL.Query().Get("status")
	source := r.URL.Query().Get("source")
	domain := r.URL.Query().Get("domain")

	records, err := s.syncSvc.GetDNSRecords(r.Context(), status, source, domain)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if records == nil {
		records = []map[string]interface{}{}
	}

	respondJSON(w, http.StatusOK, records)
}

// Export endpoint

// handleExport handles POST /api/v1/export
func (s *Server) handleExport(w http.ResponseWriter, r *http.Request) {
	err := s.exportSvc.ExportAll(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "JSON files exported successfully",
	})
}

// Scheduler endpoints

// handleSchedulerJobs handles GET /api/v1/scheduler/jobs
func (s *Server) handleSchedulerJobs(w http.ResponseWriter, r *http.Request) {
	jobs := s.scheduler.GetScheduledJobs()

	if jobs == nil {
		jobs = []interface{}{}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"jobs": jobs,
	})
}
