package handlers

import (
	"dns-inventory/internal/database"
	"dns-inventory/internal/services"
	"encoding/json"
	"html/template"
	"net/http"
	"time"
)

// DomainHandler handles domain-related HTTP requests
type DomainHandler struct {
	domainService *services.DomainService
	userService   *services.UserService
	templates     *template.Template
}

// NewDomainHandler creates a new domain handler
func NewDomainHandler(domainService *services.DomainService, userService *services.UserService) *DomainHandler {
	return &DomainHandler{
		domainService: domainService,
		userService:   userService,
		templates:     loadTemplates(),
	}
}

// DomainPageData represents data for the domain page
type DomainPageData struct {
	Title     string                          `json:"title"`
	Stats     *database.Stats                 `json:"stats"`
	Users     []database.User                 `json:"users"`
	Domains   []database.DomainWithUsers      `json:"domains"`
	Page      int                             `json:"page"`
	Limit     int                             `json:"limit"`
	Total     int                             `json:"total"`
	Search    string                          `json:"search"`
	HasMore   bool                            `json:"has_more"`
}

// HandleDomainsPage renders the domains page
func (h *DomainHandler) HandleDomainsPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	page := parseInt(r.URL.Query().Get("page"), 1)
	limit := parseInt(r.URL.Query().Get("limit"), 50)
	search := r.URL.Query().Get("search")

	// Calculate offset
	offset := (page - 1) * limit

	// Get data
	stats, err := h.domainService.GetDomainStats()
	if err != nil {
		http.Error(w, "Failed to get domain stats", http.StatusInternalServerError)
		return
	}

	users, err := h.userService.GetActiveUsers()
	if err != nil {
		http.Error(w, "Failed to get users", http.StatusInternalServerError)
		return
	}

	domains, total, err := h.domainService.GetDomains(limit, offset, search)
	if err != nil {
		http.Error(w, "Failed to get domains", http.StatusInternalServerError)
		return
	}

	data := DomainPageData{
		Title:   "Domain Management",
		Stats:   stats,
		Users:   users,
		Domains: domains,
		Page:    page,
		Limit:   limit,
		Total:   total,
		Search:  search,
		HasMore: offset+limit < total,
	}

	w.Header().Set("Content-Type", "text/html")
	if err := h.templates.ExecuteTemplate(w, "layout.html", data); err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// HandleDomainsAPI handles domain API requests
func (h *DomainHandler) HandleDomainsAPI(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleGetDomains(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGetDomains returns domain data as JSON
func (h *DomainHandler) handleGetDomains(w http.ResponseWriter, r *http.Request) {
	page := parseInt(r.URL.Query().Get("page"), 1)
	limit := parseInt(r.URL.Query().Get("limit"), 50)
	search := r.URL.Query().Get("search")

	offset := (page - 1) * limit

	domains, total, err := h.domainService.GetDomains(limit, offset, search)
	if err != nil {
		http.Error(w, "Failed to get domains", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"domains":  domains,
		"total":    total,
		"page":     page,
		"limit":    limit,
		"has_more": offset+limit < total,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleAssignDomain assigns a user to a domain
func (h *DomainHandler) HandleAssignDomain(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		DomainID int `json:"domain_id"`
		UserID   int `json:"user_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.domainService.AssignUserToDomain(req.DomainID, req.UserID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// HandleCollectDomains triggers domain collection
func (h *DomainHandler) HandleCollectDomains(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Run collection in background
	go func() {
		if err := h.domainService.CollectDomains(); err != nil {
			// Log error but don't block response
			println("Domain collection failed:", err.Error())
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "started",
		"message": "Domain collection started in background",
		"time":    time.Now().Format(time.RFC3339),
	})
}