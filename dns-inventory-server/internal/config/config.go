package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all application configuration
type Config struct {
	// API Configuration (Optional)
	GodaddyAPIKey      string
	GodaddyAPISecret   string
	CloudflareAPIToken string
	
	// Database Configuration
	DBPath string
	
	// Server Configuration
	ServerPort int
	
	// Application Configuration
	DataCollectionInterval int // minutes
	PageSize              int // for pagination
	
	// AWS SES Notification Configuration (Optional)
	AWSRegion             string
	AWSAccessKey          string
	AWSSecretKey          string
	NotificationFromEmail string
	NotificationToEmail   string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	config := &Config{
		// Default values
		DBPath:                 "./data",
		ServerPort:             8080,
		DataCollectionInterval: 60, // 1 hour
		PageSize:              50,
	}
	
	// Optional API credentials
	config.GodaddyAPIKey = os.Getenv("GODADDY_API_KEY")
	config.GodaddyAPISecret = os.Getenv("GODADDY_API_SECRET")
	config.CloudflareAPIToken = os.Getenv("CLOUDFLARE_API_TOKEN")
	
	// Optional AWS SES configuration
	config.AWSRegion = os.Getenv("AWS_SES_REGION")
	config.AWSAccessKey = os.Getenv("AWS_SES_ACCESS_KEY")
	config.AWSSecretKey = os.Getenv("AWS_SES_SECRET_KEY")
	config.NotificationFromEmail = os.Getenv("NOTIFICATION_FROM_EMAIL")
	config.NotificationToEmail = os.Getenv("NOTIFICATION_TO_EMAIL")
	
	// Optional configurations
	if dbPath := os.Getenv("DB_PATH"); dbPath != "" {
		config.DBPath = dbPath
	}
	
	if portStr := os.Getenv("SERVER_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil && port > 0 {
			config.ServerPort = port
		}
	}
	
	if intervalStr := os.Getenv("DATA_COLLECTION_INTERVAL"); intervalStr != "" {
		if interval, err := strconv.Atoi(intervalStr); err == nil && interval > 0 {
			config.DataCollectionInterval = interval
		}
	}
	
	if pageSizeStr := os.Getenv("PAGE_SIZE"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 {
			config.PageSize = pageSize
		}
	}
	
	return config, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.ServerPort <= 0 || c.ServerPort > 65535 {
		return fmt.Errorf("invalid server port: %d", c.ServerPort)
	}
	
	return nil
}

// HasAPICredentials checks if API credentials are configured
func (c *Config) HasAPICredentials() bool {
	return c.GodaddyAPIKey != "" && c.GodaddyAPISecret != "" && c.CloudflareAPIToken != ""
}

// HasSESConfig checks if AWS SES configuration is present
func (c *Config) HasSESConfig() bool {
	return c.AWSRegion != "" && c.AWSAccessKey != "" && c.AWSSecretKey != "" && 
		   c.NotificationFromEmail != "" && c.NotificationToEmail != ""
}