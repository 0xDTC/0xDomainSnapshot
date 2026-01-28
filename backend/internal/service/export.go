package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"0xdomainsnapshot/internal/config"
)

// ExportService handles exporting data to JSON files
type ExportService struct {
	syncSvc   *SyncService
	outputDir string
}

// NewExportService creates a new ExportService
func NewExportService(syncSvc *SyncService, cfg config.ExportConfig) *ExportService {
	return &ExportService{
		syncSvc:   syncSvc,
		outputDir: cfg.OutputDir,
	}
}

// ExportAll exports all data to JSON files for the frontend
func (e *ExportService) ExportAll(ctx context.Context) error {
	log.Printf("[Export] Starting export to %s", e.outputDir)

	// Ensure output directory exists
	if err := os.MkdirAll(e.outputDir, 0755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	// Export domains.json
	log.Printf("[Export] Exporting domains.json")
	domains, err := e.syncSvc.GetDomains(ctx, "", "")
	if err != nil {
		return fmt.Errorf("get domains: %w", err)
	}
	if err := e.writeJSON("domains.json", domains); err != nil {
		return fmt.Errorf("write domains.json: %w", err)
	}
	log.Printf("[Export] Exported %d domains", len(domains))

	// Export subdomains.json (all DNS records)
	log.Printf("[Export] Exporting subdomains.json")
	records, err := e.syncSvc.GetDNSRecords(ctx, "", "", "")
	if err != nil {
		return fmt.Errorf("get DNS records: %w", err)
	}
	if err := e.writeJSON("subdomains.json", records); err != nil {
		return fmt.Errorf("write subdomains.json: %w", err)
	}
	log.Printf("[Export] Exported %d DNS records", len(records))

	// Export removed.json
	log.Printf("[Export] Exporting removed.json")
	removed, err := e.getRemovedAssets(ctx)
	if err != nil {
		return fmt.Errorf("get removed assets: %w", err)
	}
	if err := e.writeJSON("removed.json", removed); err != nil {
		return fmt.Errorf("write removed.json: %w", err)
	}
	log.Printf("[Export] Exported %d removed assets", len(removed))

	// Update metadata.json
	log.Printf("[Export] Updating metadata.json")
	if err := e.updateMetadata(ctx, len(domains), len(records)); err != nil {
		return fmt.Errorf("update metadata: %w", err)
	}

	log.Printf("[Export] Export complete")
	return nil
}

// writeJSON writes data to a JSON file with pretty formatting
func (e *ExportService) writeJSON(filename string, data interface{}) error {
	path := filepath.Join(e.outputDir, filename)

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "    ")

	// Handle nil data
	if data == nil {
		data = []interface{}{}
	}

	return encoder.Encode(data)
}

// getRemovedAssets gets all removed assets for the removed.json file
func (e *ExportService) getRemovedAssets(ctx context.Context) ([]map[string]interface{}, error) {
	var removed []map[string]interface{}

	// Get removed domains
	domains, err := e.syncSvc.GetDomains(ctx, "removed", "")
	if err != nil {
		return nil, err
	}

	for _, d := range domains {
		removed = append(removed, map[string]interface{}{
			"asset_type":     "domain",
			"name":           d["domain"],
			"provider":       d["registrar"],
			"details":        "Domain removed from registrar",
			"discovery_date": d["discovery_date"],
			"removed_date":   d["last_seen"],
			"status":         "removed",
		})
	}

	// Get removed DNS records
	records, err := e.syncSvc.GetDNSRecords(ctx, "removed", "", "")
	if err != nil {
		return nil, err
	}

	for _, r := range records {
		name := r["domain"].(string)
		if subdomain, ok := r["subdomain"].(string); ok && subdomain != "" {
			name = subdomain + "." + name
		}

		removed = append(removed, map[string]interface{}{
			"asset_type":     "subdomain",
			"name":           name,
			"provider":       r["source"],
			"details":        fmt.Sprintf("%s record - %s", r["type"], r["data"]),
			"discovery_date": r["discovery_date"],
			"removed_date":   r["last_seen"],
			"status":         "removed",
		})
	}

	return removed, nil
}

// updateMetadata updates the metadata.json file
func (e *ExportService) updateMetadata(ctx context.Context, domainCount, recordCount int) error {
	metadataPath := filepath.Join(e.outputDir, "metadata.json")

	// Try to read existing metadata
	var metadata map[string]interface{}
	if data, err := os.ReadFile(metadataPath); err == nil {
		json.Unmarshal(data, &metadata)
	}

	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	// Ensure services map exists
	services, ok := metadata["services"].(map[string]interface{})
	if !ok {
		services = make(map[string]interface{})
	}

	now := time.Now().UTC().Format(time.RFC3339)

	// Update DNS services
	services["dns"] = map[string]interface{}{
		"name":         "DNS",
		"provider":     "GoDaddy/Cloudflare",
		"schedule":     "daily",
		"last_updated": now,
		"services": map[string]interface{}{
			"domains": map[string]interface{}{
				"last_updated": now,
				"count":        domainCount,
			},
			"subdomains": map[string]interface{}{
				"last_updated": now,
				"count":        recordCount,
			},
		},
	}

	metadata["services"] = services
	metadata["last_updated"] = now

	return e.writeJSON("metadata.json", metadata)
}

// ExportDomains exports only domains to domains.json
func (e *ExportService) ExportDomains(ctx context.Context) error {
	domains, err := e.syncSvc.GetDomains(ctx, "", "")
	if err != nil {
		return err
	}
	return e.writeJSON("domains.json", domains)
}

// ExportDNSRecords exports only DNS records to subdomains.json
func (e *ExportService) ExportDNSRecords(ctx context.Context) error {
	records, err := e.syncSvc.GetDNSRecords(ctx, "", "", "")
	if err != nil {
		return err
	}
	return e.writeJSON("subdomains.json", records)
}
