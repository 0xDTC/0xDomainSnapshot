package handlers

import (
	"dns-inventory/internal/services"
	"fmt"
	"net/http"
	"time"
)

// ExportHandler handles export-related HTTP requests
type ExportHandler struct {
	domainService *services.DomainService
	dnsService    *services.DNSService
}

// NewExportHandler creates a new export handler
func NewExportHandler(domainService *services.DomainService, dnsService *services.DNSService) *ExportHandler {
	return &ExportHandler{
		domainService: domainService,
		dnsService:    dnsService,
	}
}

// HandleDomainsCSV exports domains to CSV
func (h *ExportHandler) HandleDomainsCSV(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Generate CSV data
	csvData, err := h.domainService.ExportDomainsCSV()
	if err != nil {
		http.Error(w, "Failed to export domains", http.StatusInternalServerError)
		return
	}

	// Set headers for file download
	filename := fmt.Sprintf("domains_%s.csv", time.Now().Format("2006-01-02"))
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(csvData)))

	// Write CSV data
	w.Write(csvData)
}

// HandleDNSCSV exports DNS records to CSV
func (h *ExportHandler) HandleDNSCSV(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Generate CSV data
	csvData, err := h.dnsService.ExportDNSRecordsCSV()
	if err != nil {
		http.Error(w, "Failed to export DNS records", http.StatusInternalServerError)
		return
	}

	// Set headers for file download
	filename := fmt.Sprintf("dns_records_%s.csv", time.Now().Format("2006-01-02"))
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(csvData)))

	// Write CSV data
	w.Write(csvData)
}