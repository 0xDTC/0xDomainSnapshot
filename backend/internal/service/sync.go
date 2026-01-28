package service

import (
	"context"
	"fmt"
	"log"

	"0xdomainsnapshot/internal/collector"
	"0xdomainsnapshot/internal/database"
	"0xdomainsnapshot/internal/merger"
)

// SyncStats holds statistics about a sync operation
type SyncStats struct {
	Found   int
	Added   int
	Updated int
	Removed int
}

// SyncService orchestrates data synchronization
type SyncService struct {
	db     *database.DB
	merger *merger.Merger
}

// NewSyncService creates a new SyncService
func NewSyncService(db *database.DB) *SyncService {
	return &SyncService{
		db:     db,
		merger: merger.New(db),
	}
}

// RunCollector runs a collector and merges the results
func (s *SyncService) RunCollector(ctx context.Context, c collector.Collector) (*SyncStats, error) {
	log.Printf("[Sync] Starting collector: %s", c.Name())

	// Run the collector
	result, err := c.Collect(ctx)
	if err != nil {
		return nil, fmt.Errorf("collector %s failed: %w", c.Name(), err)
	}

	stats := &SyncStats{
		Found: len(result.Domains) + len(result.DNSRecords),
	}

	// Merge domains if any were collected
	if len(result.Domains) > 0 {
		log.Printf("[Sync] Merging %d domains from %s", len(result.Domains), c.Source())
		domainStats, err := s.merger.MergeDomains(ctx, c.Source(), result.Domains)
		if err != nil {
			return stats, fmt.Errorf("merge domains: %w", err)
		}
		stats.Added += domainStats.Added
		stats.Updated += domainStats.Updated
		stats.Removed += domainStats.Removed
		log.Printf("[Sync] Domains: added=%d updated=%d removed=%d",
			domainStats.Added, domainStats.Updated, domainStats.Removed)
	}

	// Merge DNS records if any were collected
	if len(result.DNSRecords) > 0 {
		log.Printf("[Sync] Merging %d DNS records from %s", len(result.DNSRecords), c.Source())
		recordStats, err := s.merger.MergeDNSRecords(ctx, c.Source(), result.DNSRecords)
		if err != nil {
			return stats, fmt.Errorf("merge DNS records: %w", err)
		}
		stats.Added += recordStats.Added
		stats.Updated += recordStats.Updated
		stats.Removed += recordStats.Removed
		log.Printf("[Sync] DNS Records: added=%d updated=%d removed=%d",
			recordStats.Added, recordStats.Updated, recordStats.Removed)
	}

	log.Printf("[Sync] Collector %s complete: found=%d added=%d updated=%d removed=%d",
		c.Name(), stats.Found, stats.Added, stats.Updated, stats.Removed)

	return stats, nil
}

// GetDomains retrieves domains from the database
func (s *SyncService) GetDomains(ctx context.Context, status, source string) ([]map[string]interface{}, error) {
	query := `
		SELECT domain, registrar, status, expiry_date, discovery_date, last_seen
		FROM domains
		WHERE 1=1
	`
	args := []interface{}{}
	argNum := 1

	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argNum)
		args = append(args, status)
		argNum++
	}
	if source != "" {
		query += fmt.Sprintf(" AND registrar = $%d", argNum)
		args = append(args, source)
		argNum++
	}

	query += " ORDER BY domain"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var domain, registrar, status string
		var expiryDate, discoveryDate, lastSeen interface{}

		if err := rows.Scan(&domain, &registrar, &status, &expiryDate, &discoveryDate, &lastSeen); err != nil {
			return nil, err
		}

		result := map[string]interface{}{
			"domain":    domain,
			"registrar": registrar,
			"status":    status,
		}

		if expiryDate != nil {
			result["expiry_date"] = formatDate(expiryDate)
		}
		if discoveryDate != nil {
			result["discovery_date"] = formatDate(discoveryDate)
		}
		if lastSeen != nil {
			result["last_seen"] = formatDate(lastSeen)
		}

		results = append(results, result)
	}

	return results, rows.Err()
}

// GetDNSRecords retrieves DNS records from the database
func (s *SyncService) GetDNSRecords(ctx context.Context, status, source, domain string) ([]map[string]interface{}, error) {
	query := `
		SELECT domain, subdomain, record_type, data, source, status, discovery_date, last_seen
		FROM dns_records
		WHERE 1=1
	`
	args := []interface{}{}
	argNum := 1

	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argNum)
		args = append(args, status)
		argNum++
	}
	if source != "" {
		query += fmt.Sprintf(" AND source = $%d", argNum)
		args = append(args, source)
		argNum++
	}
	if domain != "" {
		query += fmt.Sprintf(" AND domain = $%d", argNum)
		args = append(args, domain)
		argNum++
	}

	query += " ORDER BY domain, subdomain"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var domainVal, subdomain, recType, data, source, status string
		var discoveryDate, lastSeen interface{}

		if err := rows.Scan(&domainVal, &subdomain, &recType, &data, &source, &status, &discoveryDate, &lastSeen); err != nil {
			return nil, err
		}

		result := map[string]interface{}{
			"domain":    domainVal,
			"subdomain": subdomain,
			"type":      recType,
			"data":      data,
			"source":    source,
			"status":    status,
		}

		if discoveryDate != nil {
			result["discovery_date"] = formatDate(discoveryDate)
		}
		if lastSeen != nil {
			result["last_seen"] = formatDate(lastSeen)
		}

		results = append(results, result)
	}

	return results, rows.Err()
}

// formatDate formats a date value as YYYY-MM-DD
func formatDate(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case []byte:
		return string(t)
	default:
		return fmt.Sprintf("%v", v)
	}
}
