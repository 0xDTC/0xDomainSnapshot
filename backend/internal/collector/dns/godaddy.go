package dns

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"0xdomainsnapshot/internal/collector"
	"0xdomainsnapshot/internal/config"
	"0xdomainsnapshot/pkg/httpclient"
)

// GoDaddyCollector collects DNS records from GoDaddy
type GoDaddyCollector struct {
	cfg    config.GoDaddyConfig
	rate   config.RateLimitConfig
	client *httpclient.Client
}

// NewGoDaddyCollector creates a new GoDaddy collector
func NewGoDaddyCollector(cfg config.GoDaddyConfig, rate config.RateLimitConfig) *GoDaddyCollector {
	return &GoDaddyCollector{
		cfg:    cfg,
		rate:   rate,
		client: httpclient.New(rate),
	}
}

// Name returns the collector name
func (g *GoDaddyCollector) Name() string {
	return "godaddy_dns"
}

// Type returns the collector type
func (g *GoDaddyCollector) Type() collector.CollectorType {
	return collector.CollectorTypeDNSRecords
}

// Source returns the source name
func (g *GoDaddyCollector) Source() string {
	return "GoDaddy"
}

// Validate checks if the collector is properly configured
func (g *GoDaddyCollector) Validate() error {
	if g.cfg.APIKey == "" {
		return fmt.Errorf("GODADDY_API_KEY is required")
	}
	if g.cfg.APISecret == "" {
		return fmt.Errorf("GODADDY_API_SECRET is required")
	}
	return nil
}

// Collect performs the DNS record collection
func (g *GoDaddyCollector) Collect(ctx context.Context) (*collector.CollectorResult, error) {
	result := &collector.CollectorResult{
		StartTime: time.Now(),
	}

	// Step 1: Fetch all domains using marker-based pagination
	log.Printf("[GoDaddy] Fetching domains...")
	domains, err := g.fetchAllDomains(ctx)
	if err != nil {
		result.Error = err
		result.EndTime = time.Now()
		return result, err
	}
	log.Printf("[GoDaddy] Found %d domains", len(domains))

	// Convert to collector.Domain
	now := time.Now()
	for _, d := range domains {
		result.Domains = append(result.Domains, collector.Domain{
			Domain:        d.domain,
			Registrar:     "GoDaddy",
			Status:        "active",
			ExpiryDate:    d.expires,
			DiscoveryDate: now,
			LastSeen:      now,
			RawData:       d.raw,
		})
	}

	// Step 2: Fetch DNS records for each domain
	log.Printf("[GoDaddy] Fetching DNS records for %d domains...", len(domains))
	quotaExceeded := false

	for i, domain := range domains {
		if ctx.Err() != nil {
			result.Error = ctx.Err()
			break
		}

		if quotaExceeded {
			break
		}

		records, err := g.fetchDNSRecords(ctx, domain.domain)
		if err != nil {
			if httpclient.IsQuotaExceeded(err) {
				log.Printf("[GoDaddy] Quota exceeded after %d domains", i+1)
				quotaExceeded = true
				break
			}
			if httpclient.IsNotFound(err) {
				log.Printf("[GoDaddy] Domain %s not found, skipping", domain.domain)
				continue
			}
			log.Printf("[GoDaddy] Error fetching records for %s: %v", domain.domain, err)
			continue
		}

		result.DNSRecords = append(result.DNSRecords, records...)

		if (i+1)%50 == 0 {
			log.Printf("[GoDaddy] Processed %d/%d domains, %d records so far",
				i+1, len(domains), len(result.DNSRecords))
		}
	}

	result.EndTime = time.Now()
	log.Printf("[GoDaddy] Collection complete: %d domains, %d DNS records in %v",
		len(result.Domains), len(result.DNSRecords), result.Duration())

	return result, nil
}

// authHeader returns the authorization header for GoDaddy API
func (g *GoDaddyCollector) authHeader() http.Header {
	return http.Header{
		"Authorization": []string{fmt.Sprintf("sso-key %s:%s", g.cfg.APIKey, g.cfg.APISecret)},
		"Accept":        []string{"application/json"},
	}
}

// godaddyDomain holds domain info from GoDaddy API
type godaddyDomain struct {
	domain  string
	expires *time.Time
	raw     map[string]interface{}
}

// fetchAllDomains fetches all domains using marker-based pagination
func (g *GoDaddyCollector) fetchAllDomains(ctx context.Context) ([]godaddyDomain, error) {
	var allDomains []godaddyDomain
	seen := make(map[string]bool)
	var marker string

	for {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// Build URL with pagination
		params := url.Values{
			"limit": {fmt.Sprintf("%d", g.cfg.DomainsLimit)},
		}
		if marker != "" {
			params.Set("marker", marker)
		}

		reqURL := fmt.Sprintf("%s/v1/domains?%s", g.cfg.BaseURL, params.Encode())

		body, err := g.client.Get(ctx, reqURL, g.authHeader())
		if err != nil {
			return nil, fmt.Errorf("fetch domains: %w", err)
		}

		var domains []map[string]interface{}
		if err := json.Unmarshal(body, &domains); err != nil {
			return nil, fmt.Errorf("parse domains response: %w", err)
		}

		if len(domains) == 0 {
			break
		}

		for _, d := range domains {
			domainName, ok := d["domain"].(string)
			if !ok || domainName == "" {
				continue
			}

			// Skip if already seen (shouldn't happen but defensive)
			if seen[domainName] {
				continue
			}
			seen[domainName] = true

			// Skip test domains
			if IsTestDomain(domainName) {
				continue
			}

			gd := godaddyDomain{
				domain: domainName,
				raw:    d,
			}

			// Parse expiry date if present
			if expires, ok := d["expires"].(string); ok && expires != "" {
				if t, err := time.Parse(time.RFC3339, expires); err == nil {
					gd.expires = &t
				}
			}

			allDomains = append(allDomains, gd)
		}

		// Check if we got fewer than limit (last page)
		if len(domains) < g.cfg.DomainsLimit {
			break
		}

		// Set marker for next page (last domain name)
		marker = allDomains[len(allDomains)-1].domain
	}

	return allDomains, nil
}

// fetchDNSRecords fetches DNS records for a domain using offset-based pagination
func (g *GoDaddyCollector) fetchDNSRecords(ctx context.Context, domain string) ([]collector.DNSRecord, error) {
	var allRecords []collector.DNSRecord
	offset := 0
	now := time.Now()

	for {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// Build URL with pagination
		params := url.Values{
			"limit":  {fmt.Sprintf("%d", g.cfg.RecordsLimit)},
			"offset": {fmt.Sprintf("%d", offset)},
		}

		reqURL := fmt.Sprintf("%s/v1/domains/%s/records?%s",
			g.cfg.BaseURL, url.PathEscape(domain), params.Encode())

		body, err := g.client.Get(ctx, reqURL, g.authHeader())
		if err != nil {
			return nil, err
		}

		var records []map[string]interface{}
		if err := json.Unmarshal(body, &records); err != nil {
			return nil, fmt.Errorf("parse records response for %s: %w", domain, err)
		}

		if len(records) == 0 {
			break
		}

		for _, r := range records {
			name, _ := r["name"].(string)
			recType, _ := r["type"].(string)
			data, _ := r["data"].(string)
			ttl, _ := r["ttl"].(float64)
			priority, _ := r["priority"].(float64)

			// Normalize subdomain (@ becomes empty string)
			subdomain := NormalizeSubdomain(name)

			allRecords = append(allRecords, collector.DNSRecord{
				Domain:        domain,
				Subdomain:     subdomain,
				RecordType:    NormalizeRecordType(recType),
				Data:          data,
				TTL:           int(ttl),
				Priority:      int(priority),
				Source:        "GoDaddy",
				Status:        "active",
				DiscoveryDate: now,
				LastSeen:      now,
				RawData:       r,
			})
		}

		// Check if we got fewer than limit (last page)
		if len(records) < g.cfg.RecordsLimit {
			break
		}

		offset += g.cfg.RecordsLimit
	}

	return allRecords, nil
}
