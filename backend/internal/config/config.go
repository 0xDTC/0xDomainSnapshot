package config

import (
	"fmt"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

// Config holds all application configuration
type Config struct {
	Server     ServerConfig
	Database   DatabaseConfig
	GoDaddy    GoDaddyConfig
	Cloudflare CloudflareConfig
	RateLimit  RateLimitConfig
	Scheduler  SchedulerConfig
	Export     ExportConfig
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port      int    `envconfig:"SERVER_PORT" default:"8080"`
	Host      string `envconfig:"SERVER_HOST" default:"0.0.0.0"`
	StaticDir string `envconfig:"STATIC_DIR" default:".."`
}

// DatabaseConfig holds PostgreSQL configuration
type DatabaseConfig struct {
	URL            string `envconfig:"DATABASE_URL" required:"true"`
	MaxConnections int    `envconfig:"DATABASE_MAX_CONNECTIONS" default:"25"`
	MaxIdle        int    `envconfig:"DATABASE_MAX_IDLE" default:"5"`
}

// GoDaddyConfig holds GoDaddy API configuration
type GoDaddyConfig struct {
	APIKey       string `envconfig:"GODADDY_API_KEY"`
	APISecret    string `envconfig:"GODADDY_API_SECRET"`
	BaseURL      string `envconfig:"GODADDY_BASE_URL" default:"https://api.godaddy.com"`
	DomainsLimit int    `envconfig:"GODADDY_DOMAINS_LIMIT" default:"1000"`
	RecordsLimit int    `envconfig:"GODADDY_RECORDS_LIMIT" default:"100"`
}

// IsConfigured returns true if GoDaddy credentials are provided
func (g GoDaddyConfig) IsConfigured() bool {
	return g.APIKey != "" && g.APISecret != ""
}

// CloudflareConfig holds Cloudflare API configuration
type CloudflareConfig struct {
	APIToken       string `envconfig:"CLOUDFLARE_API_TOKEN"`
	BaseURL        string `envconfig:"CLOUDFLARE_BASE_URL" default:"https://api.cloudflare.com/client/v4"`
	ZonesPerPage   int    `envconfig:"CLOUDFLARE_ZONES_PER_PAGE" default:"50"`
	RecordsPerPage int    `envconfig:"CLOUDFLARE_RECORDS_PER_PAGE" default:"1000"`
}

// IsConfigured returns true if Cloudflare credentials are provided
func (c CloudflareConfig) IsConfigured() bool {
	return c.APIToken != ""
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	SleepOn429    time.Duration `envconfig:"RATE_LIMIT_SLEEP_ON_429" default:"30s"`
	MaxRetries    int           `envconfig:"RATE_LIMIT_MAX_RETRIES" default:"5"`
	BackoffFactor float64       `envconfig:"RATE_LIMIT_BACKOFF_FACTOR" default:"1.5"`
}

// SchedulerConfig holds scheduler configuration
type SchedulerConfig struct {
	Enabled     bool   `envconfig:"SCHEDULER_ENABLED" default:"true"`
	DNSCron     string `envconfig:"SCHEDULER_DNS_CRON" default:"0 6 * * *"`
	DomainsCron string `envconfig:"SCHEDULER_DOMAINS_CRON" default:"0 0 * * 0"`
}

// ExportConfig holds JSON export configuration
type ExportConfig struct {
	OutputDir string `envconfig:"JSON_OUTPUT_DIR" default:"../data"`
}

// Load loads configuration from environment variables and .env file
func Load() (*Config, error) {
	// Load .env file if it exists (optional - environment variables take precedence)
	_ = godotenv.Load()

	var cfg Config

	// Process server config
	if err := envconfig.Process("", &cfg.Server); err != nil {
		return nil, fmt.Errorf("failed to process server config: %w", err)
	}

	// Process database config
	if err := envconfig.Process("", &cfg.Database); err != nil {
		return nil, fmt.Errorf("failed to process database config: %w", err)
	}

	// Process GoDaddy config (optional)
	if err := envconfig.Process("", &cfg.GoDaddy); err != nil {
		return nil, fmt.Errorf("failed to process GoDaddy config: %w", err)
	}

	// Process Cloudflare config (optional)
	if err := envconfig.Process("", &cfg.Cloudflare); err != nil {
		return nil, fmt.Errorf("failed to process Cloudflare config: %w", err)
	}

	// Process rate limit config
	if err := envconfig.Process("", &cfg.RateLimit); err != nil {
		return nil, fmt.Errorf("failed to process rate limit config: %w", err)
	}

	// Process scheduler config
	if err := envconfig.Process("", &cfg.Scheduler); err != nil {
		return nil, fmt.Errorf("failed to process scheduler config: %w", err)
	}

	// Process export config
	if err := envconfig.Process("", &cfg.Export); err != nil {
		return nil, fmt.Errorf("failed to process export config: %w", err)
	}

	return &cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Database.URL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}

	if !c.GoDaddy.IsConfigured() && !c.Cloudflare.IsConfigured() {
		return fmt.Errorf("at least one provider (GoDaddy or Cloudflare) must be configured")
	}

	return nil
}
