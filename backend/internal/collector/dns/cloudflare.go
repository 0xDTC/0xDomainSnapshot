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

// CloudflareCollector collects DNS records from Cloudflare
type CloudflareCollector struct {
	cfg    config.CloudflareConfig
	rate   config.RateLimitConfig
	client *httpclient.Client
}

// NewCloudflareCollector creates a new Cloudflare collector
func NewCloudflareCollector(cfg config.CloudflareConfig, rate config.RateLimitConfig) *CloudflareCollector {
	return &CloudflareCollector{
		cfg:    cfg,
		rate:   rate,
		client: httpclient.New(rate),
	}
}

// Name returns the collector name
func (c *CloudflareCollector) Name() string {
	return "cloudflare_dns"
}

// Type returns the collector type
func (c *CloudflareCollector) Type() collector.CollectorType {
	return collector.CollectorTypeDNSRecords
}

// Source returns the source name
func (c *CloudflareCollector) Source() string {
	return "Cloudflare"
}

// Validate checks if the collector is properly configured
func (c *CloudflareCollector) Validate() error {
	if c.cfg.APIToken == "" {
		return fmt.Errorf("CLOUDFLARE_API_TOKEN is required")
	}
	return nil
}

// Collect performs the DNS record collection
func (c *CloudflareCollector) Collect(ctx context.Context) (*collector.CollectorResult, error) {
	result := &collector.CollectorResult{
		StartTime: time.Now(),
	}

	// Verify token first
	if err := c.verifyToken(ctx); err != nil {
		result.Error = fmt.Errorf("token verification failed: %w", err)
		result.EndTime = time.Now()
		return result, result.Error
	}

	// Step 1: Fetch all zones using page-based pagination
	log.Printf("[Cloudflare] Fetching zones...")
	zones, err := c.fetchAllZones(ctx)
	if err != nil {
		result.Error = err
		result.EndTime = time.Now()
		return result, err
	}
	log.Printf("[Cloudflare] Found %d zones", len(zones))

	// Convert zones to collector.Domain
	now := time.Now()
	for _, z := range zones {
		result.Domains = append(result.Domains, collector.Domain{
			Domain:        z.name,
			Registrar:     "Cloudflare",
			Status:        "active",
			DiscoveryDate: now,
			LastSeen:      now,
			RawData:       z.raw,
		})
	}

	// Step 2: Fetch DNS records for each zone
	log.Printf("[Cloudflare] Fetching DNS records for %d zones...", len(zones))

	for i, zone := range zones {
		if ctx.Err() != nil {
			result.Error = ctx.Err()
			break
		}

		records, err := c.fetchDNSRecords(ctx, zone.id, zone.name)
		if err != nil {
			log.Printf("[Cloudflare] Error fetching records for %s: %v", zone.name, err)
			continue
		}

		result.DNSRecords = append(result.DNSRecords, records...)

		if (i+1)%20 == 0 {
			log.Printf("[Cloudflare] Processed %d/%d zones, %d records so far",
				i+1, len(zones), len(result.DNSRecords))
		}
	}

	result.EndTime = time.Now()
	log.Printf("[Cloudflare] Collection complete: %d zones, %d DNS records in %v",
		len(result.Domains), len(result.DNSRecords), result.Duration())

	return result, nil
}

// authHeader returns the authorization header for Cloudflare API
func (c *CloudflareCollector) authHeader() http.Header {
	return http.Header{
		"Authorization": []string{fmt.Sprintf("Bearer %s", c.cfg.APIToken)},
		"Content-Type":  []string{"application/json"},
	}
}

// cloudflareResponse is the standard Cloudflare API response structure
type cloudflareResponse struct {
	Success    bool                     `json:"success"`
	Result     []map[string]interface{} `json:"result"`
	ResultInfo struct {
		Page       int `json:"page"`
		PerPage    int `json:"per_page"`
		TotalPages int `json:"total_pages"`
		Count      int `json:"count"`
		TotalCount int `json:"total_count"`
	} `json:"result_info"`
	Errors []struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"errors"`
}

// verifyToken verifies the Cloudflare API token
func (c *CloudflareCollector) verifyToken(ctx context.Context) error {
	reqURL := fmt.Sprintf("%s/user/tokens/verify", c.cfg.BaseURL)

	body, err := c.client.Get(ctx, reqURL, c.authHeader())
	if err != nil {
		return err
	}

	var resp struct {
		Success bool `json:"success"`
		Errors  []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("parse token verify response: %w", err)
	}

	if !resp.Success {
		if len(resp.Errors) > 0 {
			return fmt.Errorf("token verification failed: %s", resp.Errors[0].Message)
		}
		return fmt.Errorf("token verification failed")
	}

	return nil
}

// cloudflareZone holds zone info from Cloudflare API
type cloudflareZone struct {
	id   string
	name string
	raw  map[string]interface{}
}

// fetchAllZones fetches all zones using page-based pagination
func (c *CloudflareCollector) fetchAllZones(ctx context.Context) ([]cloudflareZone, error) {
	var allZones []cloudflareZone
	page := 1

	for {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// Build URL with pagination
		params := url.Values{
			"page":     {fmt.Sprintf("%d", page)},
			"per_page": {fmt.Sprintf("%d", c.cfg.ZonesPerPage)},
		}

		reqURL := fmt.Sprintf("%s/zones?%s", c.cfg.BaseURL, params.Encode())

		body, err := c.client.Get(ctx, reqURL, c.authHeader())
		if err != nil {
			return nil, fmt.Errorf("fetch zones page %d: %w", page, err)
		}

		var resp cloudflareResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, fmt.Errorf("parse zones response: %w", err)
		}

		if !resp.Success {
			if len(resp.Errors) > 0 {
				return nil, fmt.Errorf("cloudflare API error: %s", resp.Errors[0].Message)
			}
			return nil, fmt.Errorf("cloudflare API error: unknown")
		}

		for _, z := range resp.Result {
			id, _ := z["id"].(string)
			name, _ := z["name"].(string)

			if id == "" || name == "" {
				continue
			}

			// Skip test domains
			if IsTestDomain(name) {
				continue
			}

			allZones = append(allZones, cloudflareZone{
				id:   id,
				name: name,
				raw:  z,
			})
		}

		// Check if we've reached the last page
		if page >= resp.ResultInfo.TotalPages || len(resp.Result) == 0 {
			break
		}

		page++
	}

	return allZones, nil
}

// fetchDNSRecords fetches DNS records for a zone using page-based pagination
func (c *CloudflareCollector) fetchDNSRecords(ctx context.Context, zoneID, zoneName string) ([]collector.DNSRecord, error) {
	var allRecords []collector.DNSRecord
	page := 1
	now := time.Now()

	for {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// Build URL with pagination
		params := url.Values{
			"page":     {fmt.Sprintf("%d", page)},
			"per_page": {fmt.Sprintf("%d", c.cfg.RecordsPerPage)},
		}

		reqURL := fmt.Sprintf("%s/zones/%s/dns_records?%s",
			c.cfg.BaseURL, zoneID, params.Encode())

		body, err := c.client.Get(ctx, reqURL, c.authHeader())
		if err != nil {
			return nil, err
		}

		var resp cloudflareResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, fmt.Errorf("parse records response for zone %s: %w", zoneName, err)
		}

		if !resp.Success {
			if len(resp.Errors) > 0 {
				return nil, fmt.Errorf("cloudflare API error: %s", resp.Errors[0].Message)
			}
			return nil, fmt.Errorf("cloudflare API error: unknown")
		}

		for _, r := range resp.Result {
			name, _ := r["name"].(string)
			recType, _ := r["type"].(string)
			content, _ := r["content"].(string)
			ttl, _ := r["ttl"].(float64)
			priority, _ := r["priority"].(float64)

			// Extract subdomain from full hostname
			subdomain := ExtractSubdomain(name, zoneName)

			allRecords = append(allRecords, collector.DNSRecord{
				Domain:        zoneName,
				Subdomain:     subdomain,
				RecordType:    NormalizeRecordType(recType),
				Data:          content,
				TTL:           int(ttl),
				Priority:      int(priority),
				Source:        "Cloudflare",
				Status:        "active",
				DiscoveryDate: now,
				LastSeen:      now,
				RawData:       r,
			})
		}

		// Check if we've reached the last page
		if page >= resp.ResultInfo.TotalPages || len(resp.Result) == 0 {
			break
		}

		page++
	}

	return allRecords, nil
}
