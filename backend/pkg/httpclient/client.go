package httpclient

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

	"0xdomainsnapshot/internal/config"
)

// Common errors
var (
	ErrQuotaExceeded = errors.New("API quota exceeded")
	ErrRateLimited   = errors.New("rate limited")
	ErrNotFound      = errors.New("resource not found")
)

// Client is an HTTP client with retry and rate limiting support
type Client struct {
	http *http.Client
	cfg  config.RateLimitConfig
}

// New creates a new HTTP client
func New(cfg config.RateLimitConfig) *Client {
	return &Client{
		http: &http.Client{
			Timeout: 60 * time.Second,
		},
		cfg: cfg,
	}
}

// DoWithRetry performs an HTTP request with retry logic
func (c *Client) DoWithRetry(ctx context.Context, method, url string, headers http.Header, body []byte) ([]byte, error) {
	var lastErr error

	for attempt := 0; attempt <= c.cfg.MaxRetries; attempt++ {
		// Check context before each attempt
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// Exponential backoff (skip on first attempt)
		if attempt > 0 {
			sleepTime := time.Duration(math.Pow(c.cfg.BackoffFactor, float64(attempt))) * time.Second
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(sleepTime):
			}
		}

		// Create request
		var bodyReader io.Reader
		if body != nil {
			bodyReader = bytes.NewReader(body)
		}

		req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		// Set headers
		for k, v := range headers {
			req.Header[k] = v
		}

		// Set default headers
		if req.Header.Get("User-Agent") == "" {
			req.Header.Set("User-Agent", "0xDomainSnapshot/1.0")
		}
		if req.Header.Get("Accept") == "" {
			req.Header.Set("Accept", "application/json")
		}

		// Execute request
		resp, err := c.http.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue
		}

		// Read response body
		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("failed to read response: %w", err)
			continue
		}

		// Check for rate limiting (HTTP 429)
		if resp.StatusCode == http.StatusTooManyRequests ||
			strings.Contains(string(respBody), "TOO_MANY_REQUESTS") {
			// Sleep for configured duration
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(c.cfg.SleepOn429):
			}
			lastErr = ErrRateLimited
			continue
		}

		// Check for GoDaddy quota exceeded
		if strings.Contains(string(respBody), "QUOTA_EXCEEDED") {
			return nil, ErrQuotaExceeded
		}

		// Check for 404 Not Found
		if resp.StatusCode == http.StatusNotFound {
			return nil, ErrNotFound
		}

		// Check for success (2xx)
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return respBody, nil
		}

		// Other errors
		lastErr = fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncateString(string(respBody), 500))

		// Don't retry on client errors (4xx) except 429 (already handled)
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return nil, lastErr
		}

		// Server errors (5xx) - retry
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// Get performs a GET request with retry logic
func (c *Client) Get(ctx context.Context, url string, headers http.Header) ([]byte, error) {
	return c.DoWithRetry(ctx, http.MethodGet, url, headers, nil)
}

// Post performs a POST request with retry logic
func (c *Client) Post(ctx context.Context, url string, headers http.Header, body []byte) ([]byte, error) {
	return c.DoWithRetry(ctx, http.MethodPost, url, headers, body)
}

// IsQuotaExceeded checks if the error is a quota exceeded error
func IsQuotaExceeded(err error) bool {
	return errors.Is(err, ErrQuotaExceeded)
}

// IsRateLimited checks if the error is a rate limit error
func IsRateLimited(err error) bool {
	return errors.Is(err, ErrRateLimited)
}

// IsNotFound checks if the error is a not found error
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// truncateString truncates a string to maxLen characters
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
