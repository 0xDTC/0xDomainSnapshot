package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// GoDaddyClient handles GoDaddy API interactions
type GoDaddyClient struct {
	APIKey    string
	APISecret string
	BaseURL   string
	client    *http.Client
}

// NewGoDaddyClient creates a new GoDaddy API client
func NewGoDaddyClient(apiKey, apiSecret string) *GoDaddyClient {
	return &GoDaddyClient{
		APIKey:    apiKey,
		APISecret: apiSecret,
		BaseURL:   "https://api.godaddy.com/v1",
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// GoDaddyDomain represents a domain from GoDaddy API
type GoDaddyDomain struct {
	Domain    string    `json:"domain"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// GoDaddyDNSRecord represents a DNS record from GoDaddy API
type GoDaddyDNSRecord struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Data string `json:"data"`
	TTL  int    `json:"ttl"`
}

// GetDomains fetches all domains from GoDaddy with pagination
func (c *GoDaddyClient) GetDomains() ([]GoDaddyDomain, error) {
	var allDomains []GoDaddyDomain
	limit := 1000
	marker := ""
	
	for {
		domains, nextMarker, err := c.getDomainsPage(limit, marker)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch domains page: %w", err)
		}
		
		allDomains = append(allDomains, domains...)
		
		if nextMarker == "" || len(domains) < limit {
			break
		}
		
		marker = nextMarker
	}
	
	return allDomains, nil
}

// getDomainsPage fetches a single page of domains
func (c *GoDaddyClient) getDomainsPage(limit int, marker string) ([]GoDaddyDomain, string, error) {
	params := url.Values{}
	params.Set("limit", strconv.Itoa(limit))
	if marker != "" {
		params.Set("marker", marker)
	}
	
	url := fmt.Sprintf("%s/domains?%s", c.BaseURL, params.Encode())
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, "", err
	}
	
	req.Header.Set("Authorization", fmt.Sprintf("sso-key %s:%s", c.APIKey, c.APISecret))
	req.Header.Set("Accept", "application/json")
	
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	
	var domains []GoDaddyDomain
	if err := json.NewDecoder(resp.Body).Decode(&domains); err != nil {
		return nil, "", err
	}
	
	// Determine next marker (last domain of current page)
	var nextMarker string
	if len(domains) == limit {
		nextMarker = domains[len(domains)-1].Domain
	}
	
	return domains, nextMarker, nil
}

// GetDNSRecords fetches all DNS records for a domain
func (c *GoDaddyClient) GetDNSRecords(domain string) ([]GoDaddyDNSRecord, error) {
	var allRecords []GoDaddyDNSRecord
	limit := 100
	offset := 0
	
	for {
		records, err := c.getDNSRecordsPage(domain, limit, offset)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch DNS records for %s: %w", domain, err)
		}
		
		allRecords = append(allRecords, records...)
		
		if len(records) < limit {
			break
		}
		
		offset += limit
	}
	
	return allRecords, nil
}

// getDNSRecordsPage fetches a single page of DNS records
func (c *GoDaddyClient) getDNSRecordsPage(domain string, limit, offset int) ([]GoDaddyDNSRecord, error) {
	params := url.Values{}
	params.Set("limit", strconv.Itoa(limit))
	params.Set("offset", strconv.Itoa(offset))
	
	url := fmt.Sprintf("%s/domains/%s/records?%s", c.BaseURL, url.QueryEscape(domain), params.Encode())
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Authorization", fmt.Sprintf("sso-key %s:%s", c.APIKey, c.APISecret))
	req.Header.Set("Accept", "application/json")
	
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("domain %s not found", domain)
	}
	
	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("rate limited - too many requests")
	}
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	
	var records []GoDaddyDNSRecord
	if err := json.NewDecoder(resp.Body).Decode(&records); err != nil {
		return nil, err
	}
	
	return records, nil
}

// TestConnection tests the GoDaddy API connection
func (c *GoDaddyClient) TestConnection() error {
	url := fmt.Sprintf("%s/domains?limit=1", c.BaseURL)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	
	req.Header.Set("Authorization", fmt.Sprintf("sso-key %s:%s", c.APIKey, c.APISecret))
	req.Header.Set("Accept", "application/json")
	
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("authentication failed - check API credentials")
	}
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API test failed with status %d: %s", resp.StatusCode, string(body))
	}
	
	return nil
}

// RateLimitedRequest implements exponential backoff for rate limiting
func (c *GoDaddyClient) RateLimitedRequest(req *http.Request) (*http.Response, error) {
	maxRetries := 5
	baseDelay := time.Second
	
	for i := 0; i < maxRetries; i++ {
		resp, err := c.client.Do(req)
		if err != nil {
			return nil, err
		}
		
		if resp.StatusCode != http.StatusTooManyRequests {
			return resp, nil
		}
		
		resp.Body.Close()
		
		// Exponential backoff
		delay := baseDelay * time.Duration(1<<uint(i))
		time.Sleep(delay)
	}
	
	return nil, fmt.Errorf("rate limit exceeded after %d retries", maxRetries)
}