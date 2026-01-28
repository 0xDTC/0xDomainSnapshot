package merger

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"0xdomainsnapshot/internal/collector"
	"0xdomainsnapshot/internal/database"
)

// MergeStats holds statistics about a merge operation
type MergeStats struct {
	Added   int
	Updated int
	Removed int
}

// Merger handles merging new records with existing database records
type Merger struct {
	db *database.DB
}

// New creates a new Merger
func New(db *database.DB) *Merger {
	return &Merger{db: db}
}

// MergeDomains merges new domains with existing records
// - Preserves discovery_date for existing records
// - Marks missing records as "removed"
func (m *Merger) MergeDomains(ctx context.Context, source string, domains []collector.Domain) (*MergeStats, error) {
	stats := &MergeStats{}

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	today := time.Now().Format("2006-01-02")

	// Track which domains we've seen in this sync
	seen := make(map[string]bool)

	for _, d := range domains {
		seen[d.Domain] = true

		// Serialize raw data to JSON
		var rawJSON []byte
		if d.RawData != nil {
			rawJSON, _ = json.Marshal(d.RawData)
		}

		// Try to find existing record
		var existingID string
		err := tx.QueryRowContext(ctx, `
			SELECT id FROM domains
			WHERE domain = $1 AND registrar = $2
		`, d.Domain, source).Scan(&existingID)

		if err == sql.ErrNoRows {
			// New domain - insert
			_, err = tx.ExecContext(ctx, `
				INSERT INTO domains (domain, registrar, status, expiry_date, discovery_date, last_seen, raw_data)
				VALUES ($1, $2, 'active', $3, $4, $4, $5)
			`, d.Domain, source, d.ExpiryDate, today, rawJSON)
			if err != nil {
				return nil, fmt.Errorf("insert domain %s: %w", d.Domain, err)
			}
			stats.Added++
		} else if err == nil {
			// Existing domain - update (preserve discovery_date)
			_, err = tx.ExecContext(ctx, `
				UPDATE domains
				SET status = 'active', expiry_date = $1, last_seen = $2, raw_data = $3, updated_at = NOW()
				WHERE id = $4
			`, d.ExpiryDate, today, rawJSON, existingID)
			if err != nil {
				return nil, fmt.Errorf("update domain %s: %w", d.Domain, err)
			}
			stats.Updated++
		} else {
			return nil, fmt.Errorf("query domain %s: %w", d.Domain, err)
		}
	}

	// Mark missing domains from this source as removed
	result, err := tx.ExecContext(ctx, `
		UPDATE domains
		SET status = 'removed', updated_at = NOW()
		WHERE registrar = $1 AND status = 'active' AND last_seen < $2
	`, source, today)
	if err != nil {
		return nil, fmt.Errorf("mark removed domains: %w", err)
	}

	removed, _ := result.RowsAffected()
	stats.Removed = int(removed)

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return stats, nil
}

// MergeDNSRecords merges new DNS records with existing records
// - Uses signature (domain, subdomain, type, data, source) for matching
// - Preserves discovery_date for existing records
// - Marks missing records as "removed"
func (m *Merger) MergeDNSRecords(ctx context.Context, source string, records []collector.DNSRecord) (*MergeStats, error) {
	stats := &MergeStats{}

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	today := time.Now().Format("2006-01-02")

	// Track which records we've seen (for marking removed)
	seenDomains := make(map[string]bool)

	for _, r := range records {
		seenDomains[r.Domain] = true

		// Serialize raw data to JSON
		var rawJSON []byte
		if r.RawData != nil {
			rawJSON, _ = json.Marshal(r.RawData)
		}

		// Try to find existing record by signature
		var existingID string
		err := tx.QueryRowContext(ctx, `
			SELECT id FROM dns_records
			WHERE domain = $1 AND subdomain = $2 AND record_type = $3 AND data = $4 AND source = $5
		`, r.Domain, r.Subdomain, r.RecordType, r.Data, source).Scan(&existingID)

		if err == sql.ErrNoRows {
			// New record - insert
			_, err = tx.ExecContext(ctx, `
				INSERT INTO dns_records
				(domain, subdomain, record_type, data, ttl, priority, source, status, discovery_date, last_seen, raw_data)
				VALUES ($1, $2, $3, $4, $5, $6, $7, 'active', $8, $8, $9)
			`, r.Domain, r.Subdomain, r.RecordType, r.Data, r.TTL, r.Priority, source, today, rawJSON)
			if err != nil {
				return nil, fmt.Errorf("insert record %s.%s: %w", r.Subdomain, r.Domain, err)
			}
			stats.Added++
		} else if err == nil {
			// Existing record - update (preserve discovery_date)
			_, err = tx.ExecContext(ctx, `
				UPDATE dns_records
				SET status = 'active', ttl = $1, priority = $2, last_seen = $3, raw_data = $4, updated_at = NOW()
				WHERE id = $5
			`, r.TTL, r.Priority, today, rawJSON, existingID)
			if err != nil {
				return nil, fmt.Errorf("update record %s.%s: %w", r.Subdomain, r.Domain, err)
			}
			stats.Updated++
		} else {
			return nil, fmt.Errorf("query record %s.%s: %w", r.Subdomain, r.Domain, err)
		}
	}

	// Mark missing records from this source as removed
	// Only for domains we actually checked
	for domain := range seenDomains {
		result, err := tx.ExecContext(ctx, `
			UPDATE dns_records
			SET status = 'removed', updated_at = NOW()
			WHERE source = $1 AND domain = $2 AND status = 'active' AND last_seen < $3
		`, source, domain, today)
		if err != nil {
			return nil, fmt.Errorf("mark removed records for %s: %w", domain, err)
		}

		removed, _ := result.RowsAffected()
		stats.Removed += int(removed)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return stats, nil
}
