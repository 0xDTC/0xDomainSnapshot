package handlers

import (
	"dns-inventory/internal/database"
	"dns-inventory/internal/services"
	"encoding/json"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// EnhancedMigrationHandler handles advanced migration HTTP requests
type EnhancedMigrationHandler struct {
	migrationService *services.EnhancedMigrationService
	userService      *services.UserService
	templates        *template.Template
}

// NewEnhancedMigrationHandler creates a new enhanced migration handler
func NewEnhancedMigrationHandler(migrationService *services.EnhancedMigrationService, userService *services.UserService) *EnhancedMigrationHandler {
	return &EnhancedMigrationHandler{
		migrationService: migrationService,
		userService:      userService,
		templates:        loadTemplates(),
	}
}

// HandleMigrationPage renders the enhanced migration page
func (h *EnhancedMigrationHandler) HandleMigrationPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get system status
	status := map[string]interface{}{
		"total_domains":     0,
		"total_dns_records": 0,
		"total_users":       0,
		"active_jobs":       0,
	}

	// Get migration jobs for status
	if jobs, err := h.migrationService.GetMigrationJobs(); err == nil {
		activeCount := 0
		for _, job := range jobs {
			if job.Status == database.MigrationStatusRunning || job.Status == database.MigrationStatusPending {
				activeCount++
			}
		}
		status["active_jobs"] = activeCount
	}

	// Get users for assignment dropdown
	users, _ := h.userService.GetUsers()

	// Get migration templates
	templates, _ := h.migrationService.GetMigrationTemplates()

	data := map[string]interface{}{
		"Title":             "Data Migration",
		"Page":              "migration",
		"Status":            status,
		"Users":             users,
		"Templates":         templates,
		"MigrationJobs":     []interface{}{}, // Empty for now
		"Settings":          map[string]interface{}{},
	}

	if err := h.templates.ExecuteTemplate(w, "layout.html", data); err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// HandleMigrationAPI handles migration API requests
func (h *EnhancedMigrationHandler) HandleMigrationAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		h.handleGetMigrationStatus(w, r)
	case http.MethodPost:
		h.handleCreateMigration(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGetMigrationStatus returns migration status and job list
func (h *EnhancedMigrationHandler) handleGetMigrationStatus(w http.ResponseWriter, r *http.Request) {
	jobs, err := h.migrationService.GetMigrationJobs()
	if err != nil {
		http.Error(w, "Failed to get migration jobs", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"jobs":              jobs,
		"total_jobs":        len(jobs),
		"total_domains":     0, // TODO: Get actual counts
		"total_dns_records": 0,
		"total_users":       0,
	}

	json.NewEncoder(w).Encode(response)
}

// handleCreateMigration creates a new migration job
func (h *EnhancedMigrationHandler) handleCreateMigration(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form
	if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB max
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get uploaded file
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "No file uploaded", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Read file content
	content, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	// Parse migration settings from form
	settings := database.MigrationSettings{
		BatchSize:           parseInt(r.FormValue("batch_size"), 100),
		DuplicateStrategy:   database.DuplicateStrategy(r.FormValue("duplicate_strategy")),
		ValidateDomains:     r.FormValue("validate_domains") == "true",
		TemplateID:          r.FormValue("template_id"),
		SkipInvalidRecords:  r.FormValue("skip_invalid_records") == "true",
		CreateBackup:        r.FormValue("create_backup") == "true",
	}

	// Parse assigned users
	if userIDs := r.FormValue("assign_to_users"); userIDs != "" {
		settings.AssignToUsers = strings.Split(userIDs, ",")
	}

	// Parse custom mappings if provided
	if mappings := r.FormValue("custom_mappings"); mappings != "" {
		customMappings := make(map[string]string)
		if err := json.Unmarshal([]byte(mappings), &customMappings); err == nil {
			settings.CustomMappings = customMappings
		}
	}

	// Create migration job
	job, err := h.migrationService.CreateMigrationJob(
		header.Filename,
		string(content),
		settings,
		"web-user", // TODO: Get actual user ID from session
	)

	if err != nil {
		http.Error(w, "Failed to create migration job: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"job":     job,
		"message": "Migration job created successfully",
	})
}

// HandleFileAnalysis analyzes uploaded file
func (h *EnhancedMigrationHandler) HandleFileAnalysis(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Parse multipart form
	if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB max
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get uploaded file
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "No file uploaded", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Read file content
	content, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	// Analyze file
	analysis, err := h.migrationService.AnalyzeFile(header.Filename, string(content))
	if err != nil {
		http.Error(w, "Failed to analyze file: "+err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(analysis)
}

// HandleMigrationPreview generates migration preview
func (h *EnhancedMigrationHandler) HandleMigrationPreview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var requestData struct {
		Content  string                     `json:"content"`
		Settings database.MigrationSettings `json:"settings"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	preview, err := h.migrationService.PreviewMigration(requestData.Content, requestData.Settings)
	if err != nil {
		http.Error(w, "Failed to generate preview: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(preview)
}

// HandleMigrationProgress returns real-time migration progress
func (h *EnhancedMigrationHandler) HandleMigrationProgress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	jobID := r.URL.Query().Get("job_id")
	if jobID == "" {
		http.Error(w, "Missing job_id parameter", http.StatusBadRequest)
		return
	}

	progress, err := h.migrationService.GetMigrationProgress(jobID)
	if err != nil {
		http.Error(w, "Failed to get progress: "+err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(progress)
}

// HandleMigrationControl handles job control operations (pause, resume, cancel)
func (h *EnhancedMigrationHandler) HandleMigrationControl(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var requestData struct {
		JobID  string `json:"job_id"`
		Action string `json:"action"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var err error
	var message string

	switch requestData.Action {
	case "pause":
		err = h.migrationService.PauseMigrationJob(requestData.JobID)
		message = "Migration paused successfully"
	case "resume":
		err = h.migrationService.ResumeMigrationJob(requestData.JobID)
		message = "Migration resumed successfully"
	case "cancel":
		err = h.migrationService.CancelMigrationJob(requestData.JobID)
		message = "Migration cancelled successfully"
	default:
		http.Error(w, "Invalid action", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, "Failed to "+requestData.Action+" migration: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": message,
	})
}

// HandleMigrationTemplates handles template operations
func (h *EnhancedMigrationHandler) HandleMigrationTemplates(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		templates, err := h.migrationService.GetMigrationTemplates()
		if err != nil {
			http.Error(w, "Failed to get templates", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(templates)

	case http.MethodPost:
		var template database.MigrationTemplate
		if err := json.NewDecoder(r.Body).Decode(&template); err != nil {
			http.Error(w, "Invalid template data", http.StatusBadRequest)
			return
		}

		if err := h.migrationService.CreateCustomTemplate(&template, "web-user"); err != nil {
			http.Error(w, "Failed to create template: "+err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Template created successfully",
		})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// HandleCreateDefaultUsers creates default users for assignment
func (h *EnhancedMigrationHandler) HandleCreateDefaultUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Create default users if they don't exist
	defaultUsers := []struct {
		Name  string
		Email string
		Group string
	}{
		{"Admin User", "admin@company.com", "administrators"},
		{"DNS Manager", "dns@company.com", "dns-managers"},
		{"Security Team", "security@company.com", "security"},
		{"Operations", "ops@company.com", "operations"},
	}

	created := 0
	for _, userData := range defaultUsers {
		user := &database.User{
			Name:      userData.Name,
			Email:     userData.Email,
			Group:     userData.Group,
			CreatedAt: time.Now(),
		}

		// Check if user exists
		existingUsers, _ := h.userService.GetUsers()
		exists := false
		for _, existing := range existingUsers {
			if existing.Email == user.Email {
				exists = true
				break
			}
		}

		if !exists {
			if _, err := h.userService.CreateUser(user.Name, user.Group, user.Email); err == nil {
				created++
			}
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Created " + strconv.Itoa(created) + " default users",
		"created": created,
	})
}

// HandleMigrationJobs returns list of migration jobs with pagination
func (h *EnhancedMigrationHandler) HandleMigrationJobs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	jobs, err := h.migrationService.GetMigrationJobs()
	if err != nil {
		http.Error(w, "Failed to get migration jobs", http.StatusInternalServerError)
		return
	}

	// Filter and paginate jobs
	page := parseInt(r.URL.Query().Get("page"), 1)
	pageSize := parseInt(r.URL.Query().Get("page_size"), 10)
	
	start := (page - 1) * pageSize
	end := start + pageSize
	
	if start > len(jobs) {
		start = len(jobs)
	}
	if end > len(jobs) {
		end = len(jobs)
	}

	paginatedJobs := jobs[start:end]

	response := map[string]interface{}{
		"jobs":         paginatedJobs,
		"total":        len(jobs),
		"page":         page,
		"page_size":    pageSize,
		"total_pages":  (len(jobs) + pageSize - 1) / pageSize,
	}

	json.NewEncoder(w).Encode(response)
}

// HandleMigrationJob handles individual job operations
func (h *EnhancedMigrationHandler) HandleMigrationJob(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract job ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid job ID", http.StatusBadRequest)
		return
	}
	jobID := pathParts[len(pathParts)-1]

	switch r.Method {
	case http.MethodGet:
		job, err := h.migrationService.GetMigrationJob(jobID)
		if err != nil {
			http.Error(w, "Job not found", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(job)

	case http.MethodDelete:
		// TODO: Implement job deletion
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Job deletion not yet implemented",
		})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Helper function to parse integers with default values
