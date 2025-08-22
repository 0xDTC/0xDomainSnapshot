package handlers

import (
	"dns-inventory/internal/database"
	"dns-inventory/internal/services"
	"encoding/json"
	"html/template"
	"net/http"
	"time"
)

// DNSHandler handles DNS record-related HTTP requests
type DNSHandler struct {
	dnsService    *services.DNSService
	domainService *services.DomainService
	userService   *services.UserService
	templates     *template.Template
}

// NewDNSHandler creates a new DNS handler
func NewDNSHandler(dnsService *services.DNSService, domainService *services.DomainService, userService *services.UserService) *DNSHandler {
	return &DNSHandler{
		dnsService:    dnsService,
		domainService: domainService,
		userService:   userService,
		templates:     loadTemplates(),
	}
}

// DNSPageData represents data for the DNS page
type DNSPageData struct {
	Title     string                         `json:"title"`
	Stats     *database.Stats                `json:"stats"`
	Users     []database.User                `json:"users"`
	Records   []database.DNSRecordWithUsers  `json:"records"`
	Page      int                            `json:"page"`
	Limit     int                            `json:"limit"`
	Total     int                            `json:"total"`
	Search    string                         `json:"search"`
	HasMore   bool                           `json:"has_more"`
}

// HandleDNSPage renders the DNS records page
func (h *DNSHandler) HandleDNSPage(w http.ResponseWriter, r *http.Request) {
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

	// Get data - reuse domain stats for DNS page
	stats, err := h.domainService.GetDomainStats()
	if err != nil {
		http.Error(w, "Failed to get DNS stats", http.StatusInternalServerError)
		return
	}

	users, err := h.userService.GetActiveUsers()
	if err != nil {
		http.Error(w, "Failed to get users", http.StatusInternalServerError)
		return
	}

	records, total, err := h.dnsService.GetDNSRecords(limit, offset, search)
	if err != nil {
		http.Error(w, "Failed to get DNS records", http.StatusInternalServerError)
		return
	}

	data := DNSPageData{
		Title:   "DNS Records Management",
		Stats:   stats,
		Users:   users,
		Records: records,
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

// HandleDNSAPI handles DNS record API requests
func (h *DNSHandler) HandleDNSAPI(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleGetDNSRecords(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGetDNSRecords returns DNS record data as JSON
func (h *DNSHandler) handleGetDNSRecords(w http.ResponseWriter, r *http.Request) {
	page := parseInt(r.URL.Query().Get("page"), 1)
	limit := parseInt(r.URL.Query().Get("limit"), 50)
	search := r.URL.Query().Get("search")

	offset := (page - 1) * limit

	records, total, err := h.dnsService.GetDNSRecords(limit, offset, search)
	if err != nil {
		http.Error(w, "Failed to get DNS records", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"records":  records,
		"total":    total,
		"page":     page,
		"limit":    limit,
		"has_more": offset+limit < total,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleAssignDNS assigns a user to a DNS record
func (h *DNSHandler) HandleAssignDNS(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		RecordID int `json:"record_id"`
		UserID   int `json:"user_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.dnsService.AssignUserToDNSRecord(req.RecordID, req.UserID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// HandleCollectDNS triggers DNS records collection
func (h *DNSHandler) HandleCollectDNS(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Run collection in background
	go func() {
		if err := h.dnsService.CollectDNSRecords(); err != nil {
			// Log error but don't block response
			println("DNS collection failed:", err.Error())
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "started",
		"message": "DNS collection started in background",
		"time":    time.Now().Format(time.RFC3339),
	})
}