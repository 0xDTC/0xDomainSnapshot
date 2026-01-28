package collector

import (
	"context"
	"time"
)

// CollectorType identifies the type of resources collected
type CollectorType string

const (
	// CollectorTypeDomains collects registered domain information
	CollectorTypeDomains CollectorType = "domains"
	// CollectorTypeDNSRecords collects DNS records (subdomains)
	CollectorTypeDNSRecords CollectorType = "dns_records"
)

// Domain represents a registered domain
type Domain struct {
	Domain        string                 `json:"domain"`
	Registrar     string                 `json:"registrar"`
	Status        string                 `json:"status"`
	ExpiryDate    *time.Time             `json:"expiry_date,omitempty"`
	DiscoveryDate time.Time              `json:"discovery_date"`
	LastSeen      time.Time              `json:"last_seen"`
	RawData       map[string]interface{} `json:"raw_data,omitempty"`
}

// DNSRecord represents a DNS record
type DNSRecord struct {
	Domain        string                 `json:"domain"`
	Subdomain     string                 `json:"subdomain"`
	RecordType    string                 `json:"type"`
	Data          string                 `json:"data"`
	TTL           int                    `json:"ttl,omitempty"`
	Priority      int                    `json:"priority,omitempty"`
	Source        string                 `json:"source"`
	Status        string                 `json:"status"`
	DiscoveryDate time.Time              `json:"discovery_date"`
	LastSeen      time.Time              `json:"last_seen"`
	RawData       map[string]interface{} `json:"raw_data,omitempty"`
}

// CollectorResult holds the results of a collection run
type CollectorResult struct {
	Domains    []Domain
	DNSRecords []DNSRecord
	StartTime  time.Time
	EndTime    time.Time
	Error      error
}

// Stats returns statistics about the collection result
func (r *CollectorResult) Stats() (domains, records int) {
	return len(r.Domains), len(r.DNSRecords)
}

// Duration returns the duration of the collection run
func (r *CollectorResult) Duration() time.Duration {
	return r.EndTime.Sub(r.StartTime)
}

// Collector is the interface that all collectors must implement
type Collector interface {
	// Name returns the unique identifier for this collector
	// Example: "godaddy_dns", "cloudflare_dns"
	Name() string

	// Type returns what kind of resources this collector handles
	Type() CollectorType

	// Source returns the provider name (GoDaddy, Cloudflare, etc.)
	Source() string

	// Collect performs the actual data collection
	Collect(ctx context.Context) (*CollectorResult, error)

	// Validate checks if the collector is properly configured
	Validate() error
}

// CollectorStatus represents the current state of a collector
type CollectorStatus struct {
	Name       string        `json:"name"`
	Type       CollectorType `json:"type"`
	Source     string        `json:"source"`
	IsRunning  bool          `json:"is_running"`
	LastRun    *time.Time    `json:"last_run,omitempty"`
	LastError  string        `json:"last_error,omitempty"`
	NextRun    *time.Time    `json:"next_run,omitempty"`
}
