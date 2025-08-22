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

// CloudflareClient handles Cloudflare API interactions
type CloudflareClient struct {
	APIToken string
	BaseURL  string
	client   *http.Client
}

// NewCloudflareClient creates a new Cloudflare API client
func NewCloudflareClient(apiToken string) *CloudflareClient {
	return &CloudflareClient{
		APIToken: apiToken,
		BaseURL:  "https://api.cloudflare.com/client/v4",
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// CloudflareResponse represents the standard Cloudflare API response
type CloudflareResponse struct {
	Success    bool        `json:"success"`
	Errors     []string    `json:"errors"`
	Messages   []string    `json:"messages"`
	Result     interface{} `json:"result"`
	ResultInfo *ResultInfo `json:"result_info,omitempty"`
}

// ResultInfo contains pagination information
type ResultInfo struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	TotalPages int `json:"total_pages"`
	Count      int `json:"count"`
	TotalCount int `json:"total_count"`
}

// CloudflareZone represents a zone from Cloudflare API
type CloudflareZone struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

// CloudflareDNSRecord represents a DNS record from Cloudflare API
type CloudflareDNSRecord struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
	TTL     int    `json:"ttl"`
	ZoneID  string `json:"zone_id"`
}

// GetZones fetches all zones from Cloudflare with pagination
func (c *CloudflareClient) GetZones() ([]CloudflareZone, error) {
	var allZones []CloudflareZone
	page := 1
	perPage := 50
	
	for {
		zones, hasMore, err := c.getZonesPage(page, perPage)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch zones page %d: %w", page, err)
		}
		
		allZones = append(allZones, zones...)
		
		if !hasMore {
			break
		}
		
		page++
	}
	
	return allZones, nil
}

// getZonesPage fetches a single page of zones
func (c *CloudflareClient) getZonesPage(page, perPage int) ([]CloudflareZone, bool, error) {
	params := url.Values{}
	params.Set("page", strconv.Itoa(page))
	params.Set("per_page", strconv.Itoa(perPage))
	
	url := fmt.Sprintf("%s/zones?%s", c.BaseURL, params.Encode())
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, false, err
	}
	
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIToken))
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, false, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, false, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	
	var response CloudflareResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, false, err
	}
	
	if !response.Success {
		return nil, false, fmt.Errorf("API request failed: %v", response.Errors)
	}
	
	// Parse zones from result
	resultBytes, err := json.Marshal(response.Result)
	if err != nil {
		return nil, false, err
	}
	
	var zones []CloudflareZone
	if err := json.Unmarshal(resultBytes, &zones); err != nil {
		return nil, false, err
	}
	
	// Check if there are more pages
	hasMore := response.ResultInfo != nil && 
		response.ResultInfo.Page < response.ResultInfo.TotalPages
	
	return zones, hasMore, nil
}

// GetDNSRecords fetches all DNS records for a zone
func (c *CloudflareClient) GetDNSRecords(zoneID string) ([]CloudflareDNSRecord, error) {
	var allRecords []CloudflareDNSRecord
	page := 1
	perPage := 1000
	
	for {
		records, hasMore, err := c.getDNSRecordsPage(zoneID, page, perPage)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch DNS records for zone %s: %w", zoneID, err)
		}
		
		allRecords = append(allRecords, records...)
		
		if !hasMore {
			break
		}
		
		page++
	}
	
	return allRecords, nil
}

// getDNSRecordsPage fetches a single page of DNS records
func (c *CloudflareClient) getDNSRecordsPage(zoneID string, page, perPage int) ([]CloudflareDNSRecord, bool, error) {
	params := url.Values{}
	params.Set("page", strconv.Itoa(page))
	params.Set("per_page", strconv.Itoa(perPage))
	
	url := fmt.Sprintf("%s/zones/%s/dns_records?%s", c.BaseURL, zoneID, params.Encode())
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, false, err
	}
	
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIToken))
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, false, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, false, fmt.Errorf("rate limited - too many requests")
	}
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, false, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	
	var response CloudflareResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, false, err
	}
	
	if !response.Success {
		return nil, false, fmt.Errorf("API request failed: %v", response.Errors)
	}
	
	// Parse DNS records from result
	resultBytes, err := json.Marshal(response.Result)
	if err != nil {
		return nil, false, err
	}
	
	var records []CloudflareDNSRecord
	if err := json.Unmarshal(resultBytes, &records); err != nil {
		return nil, false, err
	}
	
	// Check if there are more pages
	hasMore := response.ResultInfo != nil && 
		response.ResultInfo.Page < response.ResultInfo.TotalPages
	
	return records, hasMore, nil
}

// TestConnection tests the Cloudflare API connection
func (c *CloudflareClient) TestConnection() error {
	url := fmt.Sprintf("%s/user/tokens/verify", c.BaseURL)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIToken))
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("authentication failed - check API token")
	}
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API test failed with status %d: %s", resp.StatusCode, string(body))
	}
	
	var response CloudflareResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return err
	}
	
	if !response.Success {
		return fmt.Errorf("token verification failed: %v", response.Errors)
	}
	
	return nil
}

// RateLimitedRequest implements exponential backoff for rate limiting
func (c *CloudflareClient) RateLimitedRequest(req *http.Request) (*http.Response, error) {
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