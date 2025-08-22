package database

import "time"

// Domain represents a domain with dual provider status
type Domain struct {
	ID               int       `json:"id"`
	Domain           string    `json:"domain"`
	GodaddyStatus    string    `json:"godaddy_status"`    // Active, Cancelled, etc.
	CloudflareStatus string    `json:"cloudflare_status"` // Active, Removed, or empty if not on CF
	DiscoveryDate    time.Time `json:"discovery_date"`
	LastSeen         time.Time `json:"last_seen"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// DNSRecord represents a DNS record from either provider
type DNSRecord struct {
	ID            int       `json:"id"`
	Domain        string    `json:"domain"`
	Subdomain     string    `json:"subdomain"`
	Type          string    `json:"type"`
	Data          string    `json:"data"`
	Source        string    `json:"source"`        // GoDaddy, Cloudflare
	Status        string    `json:"status"`        // active, removed
	DiscoveryDate time.Time `json:"discovery_date"`
	LastSeen      time.Time `json:"last_seen"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// User represents an employee or group
type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Group     string    `json:"group"`
	Email     string    `json:"email"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DomainAssignment links domains to users
type DomainAssignment struct {
	ID       int       `json:"id"`
	DomainID int       `json:"domain_id"`
	UserID   int       `json:"user_id"`
	Domain   string    `json:"domain"`
	UserName string    `json:"user_name"`
	UserGroup string   `json:"user_group"`
	AssignedAt time.Time `json:"assigned_at"`
}

// DNSAssignment links DNS records to users
type DNSAssignment struct {
	ID          int       `json:"id"`
	DNSRecordID int       `json:"dns_record_id"`
	UserID      int       `json:"user_id"`
	Domain      string    `json:"domain"`
	Subdomain   string    `json:"subdomain"`
	UserName    string    `json:"user_name"`
	UserGroup   string    `json:"user_group"`
	AssignedAt  time.Time `json:"assigned_at"`
}

// Snapshot tracks historical data for reporting
type Snapshot struct {
	ID               int       `json:"id"`
	Date             time.Time `json:"date"`
	TotalDomains     int       `json:"total_domains"`
	GodaddyDomains   int       `json:"godaddy_domains"`
	CloudflareDomains int       `json:"cloudflare_domains"`
	TotalDNSRecords  int       `json:"total_dns_records"`
	GodaddyDNSRecords int       `json:"godaddy_dns_records"`
	CloudflareDNSRecords int    `json:"cloudflare_dns_records"`
	CreatedAt        time.Time `json:"created_at"`
}

// DomainWithUsers represents domain data with assigned users
type DomainWithUsers struct {
	Domain
	AssignedUsers []User `json:"assigned_users"`
}

// DNSRecordWithUsers represents DNS record data with assigned users
type DNSRecordWithUsers struct {
	DNSRecord
	AssignedUsers []User `json:"assigned_users"`
}

// Stats represents dashboard statistics
type Stats struct {
	Date              time.Time `json:"date"`
	TotalDomains      int       `json:"total_domains"`
	GodaddyDomains    int       `json:"godaddy_domains"`
	CloudflareDomains int       `json:"cloudflare_domains"`
	TotalDNSRecords   int       `json:"total_dns_records"`
	GodaddyDNSRecords int       `json:"godaddy_dns_records"`
	CloudflareDNSRecords int    `json:"cloudflare_dns_records"`
	Status            string    `json:"status"` // ✅ for success
}

// MigrationJob represents a migration job with progress tracking
type MigrationJob struct {
	ID              string              `json:"id"`
	Status          MigrationStatus     `json:"status"`
	Type            string              `json:"type"`
	Filename        string              `json:"filename"`
	TotalRecords    int                 `json:"total_records"`
	ProcessedRecords int                `json:"processed_records"`
	SuccessRecords  int                 `json:"success_records"`
	FailedRecords   int                 `json:"failed_records"`
	SkippedRecords  int                 `json:"skipped_records"`
	Progress        float64             `json:"progress"`
	StartTime       time.Time           `json:"start_time"`
	EndTime         *time.Time          `json:"end_time,omitempty"`
	CreatedBy       string              `json:"created_by"`
	Settings        MigrationSettings   `json:"settings"`
	Errors          []string            `json:"errors"`
	Warnings        []string            `json:"warnings"`
	CurrentBatch    int                 `json:"current_batch"`
	TotalBatches    int                 `json:"total_batches"`
	LastCheckpoint  *time.Time          `json:"last_checkpoint,omitempty"`
	ResumeData      map[string]interface{} `json:"resume_data,omitempty"`
}

// MigrationStatus represents the status of a migration job
type MigrationStatus string

const (
	MigrationStatusPending   MigrationStatus = "pending"
	MigrationStatusRunning   MigrationStatus = "running"
	MigrationStatusPaused    MigrationStatus = "paused"
	MigrationStatusCompleted MigrationStatus = "completed"
	MigrationStatusFailed    MigrationStatus = "failed"
	MigrationStatusCancelled MigrationStatus = "cancelled"
)

// MigrationSettings represents migration configuration
type MigrationSettings struct {
	BatchSize           int                   `json:"batch_size"`
	DuplicateStrategy   DuplicateStrategy     `json:"duplicate_strategy"`
	AssignToUsers       []string              `json:"assign_to_users"`
	ValidateDomains     bool                  `json:"validate_domains"`
	TemplateID          string                `json:"template_id,omitempty"`
	CustomMappings      map[string]string     `json:"custom_mappings,omitempty"`
	SkipInvalidRecords  bool                  `json:"skip_invalid_records"`
	CreateBackup        bool                  `json:"create_backup"`
}

// DuplicateStrategy defines how to handle duplicate records
type DuplicateStrategy string

const (
	DuplicateStrategySkip     DuplicateStrategy = "skip"
	DuplicateStrategyReplace  DuplicateStrategy = "replace"
	DuplicateStrategyMerge    DuplicateStrategy = "merge"
	DuplicateStrategyAppend   DuplicateStrategy = "append"
)

// MigrationTemplate represents predefined migration rules
type MigrationTemplate struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Type        string            `json:"type"`
	FieldMappings map[string]string `json:"field_mappings"`
	DefaultSettings MigrationSettings `json:"default_settings"`
	ValidationRules []ValidationRule   `json:"validation_rules"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	CreatedBy   string            `json:"created_by"`
	IsDefault   bool              `json:"is_default"`
}

// ValidationRule defines field validation rules
type ValidationRule struct {
	Field       string `json:"field"`
	Required    bool   `json:"required"`
	Pattern     string `json:"pattern,omitempty"`
	MinLength   int    `json:"min_length,omitempty"`
	MaxLength   int    `json:"max_length,omitempty"`
	AllowedValues []string `json:"allowed_values,omitempty"`
}

// MigrationProgress represents real-time migration progress
type MigrationProgress struct {
	JobID           string    `json:"job_id"`
	CurrentRecord   int       `json:"current_record"`
	TotalRecords    int       `json:"total_records"`
	CurrentBatch    int       `json:"current_batch"`
	TotalBatches    int       `json:"total_batches"`
	Progress        float64   `json:"progress"`
	Status          MigrationStatus `json:"status"`
	Message         string    `json:"message"`
	LastUpdate      time.Time `json:"last_update"`
	EstimatedTimeRemaining string `json:"estimated_time_remaining"`
}