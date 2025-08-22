package services

import (
	"crypto/rand"
	"dns-inventory/internal/api"
	"dns-inventory/internal/database"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// EnhancedMigrationService provides advanced migration functionality
type EnhancedMigrationService struct {
	db              *database.FileDB
	userService     *UserService
	godaddyClient   *api.GoDaddyClient
	cloudflareClient *api.CloudflareClient
	notificationService *NotificationService
	progressMap     map[string]*database.MigrationProgress
	jobQueue        chan *database.MigrationJob
	workers         int
	mu              sync.RWMutex
}

// NewEnhancedMigrationService creates a new enhanced migration service
func NewEnhancedMigrationService(db *database.FileDB, userService *UserService, godaddyClient *api.GoDaddyClient, cloudflareClient *api.CloudflareClient, notificationService *NotificationService) *EnhancedMigrationService {
	service := &EnhancedMigrationService{
		db:               db,
		userService:      userService,
		godaddyClient:    godaddyClient,
		cloudflareClient: cloudflareClient,
		notificationService: notificationService,
		progressMap:      make(map[string]*database.MigrationProgress),
		jobQueue:         make(chan *database.MigrationJob, 10),
		workers:          3, // Number of concurrent migration workers
	}

	// Start worker goroutines
	for i := 0; i < service.workers; i++ {
		go service.migrationWorker()
	}

	// Initialize default templates
	go service.initializeDefaultTemplates()

	return service
}

// CreateMigrationJob creates and queues a new migration job
func (s *EnhancedMigrationService) CreateMigrationJob(filename, content string, settings database.MigrationSettings, createdBy string) (*database.MigrationJob, error) {
	// Generate unique job ID
	jobID, err := s.generateJobID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate job ID: %w", err)
	}

	// Analyze content to determine type and count
	analysis, err := s.analyzeContent(content, filename)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze content: %w", err)
	}

	// Calculate batch size and total batches
	batchSize := settings.BatchSize
	if batchSize <= 0 {
		batchSize = 100 // Default batch size
	}
	
	totalBatches := (analysis.TotalRecords + batchSize - 1) / batchSize

	// Create migration job
	job := &database.MigrationJob{
		ID:              jobID,
		Status:          database.MigrationStatusPending,
		Type:            analysis.Type,
		Filename:        filename,
		TotalRecords:    analysis.TotalRecords,
		ProcessedRecords: 0,
		SuccessRecords:   0,
		FailedRecords:    0,
		SkippedRecords:   0,
		Progress:         0.0,
		StartTime:        time.Now(),
		CreatedBy:        createdBy,
		Settings:         settings,
		Errors:           make([]string, 0),
		Warnings:         make([]string, 0),
		CurrentBatch:     0,
		TotalBatches:     totalBatches,
		ResumeData: map[string]interface{}{
			"content":        content,
			"last_processed": 0,
		},
	}

	// Save job to database
	if err := s.db.CreateMigrationJob(job); err != nil {
		return nil, fmt.Errorf("failed to create migration job: %w", err)
	}

	// Add to processing queue
	select {
	case s.jobQueue <- job:
		// Job queued successfully
		// Send started notification
		if s.notificationService != nil {
			s.notificationService.SendMigrationStarted(job)
		}
	default:
		// Queue is full, update status to indicate waiting
		job.Status = database.MigrationStatusPending
		s.db.UpdateMigrationJob(job)
	}

	return job, nil
}

// GetMigrationProgress returns real-time progress for a job
func (s *EnhancedMigrationService) GetMigrationProgress(jobID string) (*database.MigrationProgress, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if progress, exists := s.progressMap[jobID]; exists {
		return progress, nil
	}

	// If not in memory, get from database
	job, err := s.db.GetMigrationJob(jobID)
	if err != nil {
		return nil, err
	}

	return &database.MigrationProgress{
		JobID:         job.ID,
		CurrentRecord: job.ProcessedRecords,
		TotalRecords:  job.TotalRecords,
		CurrentBatch:  job.CurrentBatch,
		TotalBatches:  job.TotalBatches,
		Progress:      job.Progress,
		Status:        job.Status,
		LastUpdate:    time.Now(),
	}, nil
}

// ResumeMigrationJob resumes a paused or failed migration
func (s *EnhancedMigrationService) ResumeMigrationJob(jobID string) error {
	job, err := s.db.GetMigrationJob(jobID)
	if err != nil {
		return err
	}

	if job.Status != database.MigrationStatusPaused && job.Status != database.MigrationStatusFailed {
		return fmt.Errorf("job is not in a resumable state")
	}

	job.Status = database.MigrationStatusPending
	if err := s.db.UpdateMigrationJob(job); err != nil {
		return err
	}

	// Add back to queue
	select {
	case s.jobQueue <- job:
		return nil
	default:
		return fmt.Errorf("migration queue is full")
	}
}

// PauseMigrationJob pauses a running migration
func (s *EnhancedMigrationService) PauseMigrationJob(jobID string) error {
	job, err := s.db.GetMigrationJob(jobID)
	if err != nil {
		return err
	}

	if job.Status != database.MigrationStatusRunning {
		return fmt.Errorf("job is not running")
	}

	job.Status = database.MigrationStatusPaused
	return s.db.UpdateMigrationJob(job)
}

// CancelMigrationJob cancels a migration
func (s *EnhancedMigrationService) CancelMigrationJob(jobID string) error {
	job, err := s.db.GetMigrationJob(jobID)
	if err != nil {
		return err
	}

	job.Status = database.MigrationStatusCancelled
	endTime := time.Now()
	job.EndTime = &endTime

	return s.db.UpdateMigrationJob(job)
}

// migrationWorker processes migration jobs from the queue
func (s *EnhancedMigrationService) migrationWorker() {
	for job := range s.jobQueue {
		if err := s.processMigrationJob(job); err != nil {
			job.Status = database.MigrationStatusFailed
			job.Errors = append(job.Errors, fmt.Sprintf("Migration failed: %v", err))
			endTime := time.Now()
			job.EndTime = &endTime
			s.db.UpdateMigrationJob(job)
			
			// Send failure notification
			if s.notificationService != nil {
				s.notificationService.SendMigrationFailed(job, err.Error())
			}
		}
	}
}

// processMigrationJob processes a single migration job with batch processing
func (s *EnhancedMigrationService) processMigrationJob(job *database.MigrationJob) error {
	// Update status to running
	job.Status = database.MigrationStatusRunning
	if err := s.db.UpdateMigrationJob(job); err != nil {
		return err
	}

	// Get content from resume data
	content, ok := job.ResumeData["content"].(string)
	if !ok {
		return fmt.Errorf("missing content in resume data")
	}

	lastProcessed, _ := job.ResumeData["last_processed"].(float64)
	startIndex := int(lastProcessed)

	// Parse content based on type
	var records []interface{}
	if err := json.Unmarshal([]byte(content), &records); err != nil {
		return fmt.Errorf("failed to parse content: %w", err)
	}

	// Apply template if specified
	if job.Settings.TemplateID != "" {
		template, err := s.db.GetMigrationTemplate(job.Settings.TemplateID)
		if err != nil {
			job.Warnings = append(job.Warnings, fmt.Sprintf("Template %s not found, using default mapping", job.Settings.TemplateID))
		} else {
			records = s.applyTemplate(records, template)
		}
	}

	// Process in batches
	batchSize := job.Settings.BatchSize
	if batchSize <= 0 {
		batchSize = 100
	}

	totalRecords := len(records)
	for i := startIndex; i < totalRecords; i += batchSize {
		// Check if job was cancelled or paused
		currentJob, err := s.db.GetMigrationJob(job.ID)
		if err != nil {
			return err
		}

		if currentJob.Status == database.MigrationStatusCancelled {
			return fmt.Errorf("migration was cancelled")
		}

		if currentJob.Status == database.MigrationStatusPaused {
			return fmt.Errorf("migration was paused")
		}

		// Process batch
		end := i + batchSize
		if end > totalRecords {
			end = totalRecords
		}

		batch := records[i:end]
		batchResults := s.processBatch(batch, job)

		// Update job statistics
		job.ProcessedRecords += len(batch)
		job.SuccessRecords += batchResults.Success
		job.FailedRecords += batchResults.Failed
		job.SkippedRecords += batchResults.Skipped
		job.CurrentBatch = (i / batchSize) + 1
		job.Progress = float64(job.ProcessedRecords) / float64(job.TotalRecords) * 100
		job.Errors = append(job.Errors, batchResults.Errors...)
		job.Warnings = append(job.Warnings, batchResults.Warnings...)

		// Update resume data
		job.ResumeData["last_processed"] = float64(i + batchSize)
		checkpoint := time.Now()
		job.LastCheckpoint = &checkpoint

		// Update progress in memory and database
		s.updateProgress(job)
		if err := s.db.UpdateMigrationJob(job); err != nil {
			return fmt.Errorf("failed to update job progress: %w", err)
		}

		// Small delay to prevent overwhelming the system
		time.Sleep(10 * time.Millisecond)
	}

	// Mark job as completed
	job.Status = database.MigrationStatusCompleted
	job.Progress = 100.0
	endTime := time.Now()
	job.EndTime = &endTime

	// Send completion notification
	if s.notificationService != nil {
		s.notificationService.SendMigrationCompleted(job)
	}

	return s.db.UpdateMigrationJob(job)
}

// BatchResults holds results from processing a batch
type BatchResults struct {
	Success  int
	Failed   int
	Skipped  int
	Errors   []string
	Warnings []string
}

// processBatch processes a batch of records
func (s *EnhancedMigrationService) processBatch(batch []interface{}, job *database.MigrationJob) BatchResults {
	results := BatchResults{
		Errors:   make([]string, 0),
		Warnings: make([]string, 0),
	}

	for _, record := range batch {
		if err := s.processRecord(record, job); err != nil {
			if job.Settings.SkipInvalidRecords {
				results.Skipped++
				results.Warnings = append(results.Warnings, fmt.Sprintf("Skipped invalid record: %v", err))
			} else {
				results.Failed++
				results.Errors = append(results.Errors, fmt.Sprintf("Failed to process record: %v", err))
			}
		} else {
			results.Success++
		}
	}

	return results
}

// processRecord processes a single record based on job type and settings
func (s *EnhancedMigrationService) processRecord(record interface{}, job *database.MigrationJob) error {
	recordMap, ok := record.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid record format")
	}

	switch job.Type {
	case "domains":
		return s.processDomainRecord(recordMap, job)
	case "dns":
		return s.processDNSRecord(recordMap, job)
	case "org_assets":
		return s.processOrgAssetRecord(recordMap, job)
	default:
		return fmt.Errorf("unknown migration type: %s", job.Type)
	}
}

// processDomainRecord processes a domain record
func (s *EnhancedMigrationService) processDomainRecord(record map[string]interface{}, job *database.MigrationJob) error {
	domain := &database.Domain{}

	// Map fields
	if name, ok := record["domain"].(string); ok {
		domain.Domain = name
	} else if name, ok := record["name"].(string); ok {
		domain.Domain = name
	} else {
		return fmt.Errorf("missing domain name")
	}

	// Validate domain if required
	if job.Settings.ValidateDomains {
		if err := s.validateDomain(domain.Domain); err != nil {
			return fmt.Errorf("domain validation failed: %w", err)
		}
	}

	// Set other fields with defaults
	domain.GodaddyStatus = getStringField(record, "godaddy_status", "unknown")
	domain.CloudflareStatus = getStringField(record, "cloudflare_status", "unknown")

	// Parse discovery date
	if discoveryDate := getStringField(record, "discovery_date", ""); discoveryDate != "" {
		if parsed, err := parseFlexibleDate(discoveryDate); err == nil {
			domain.DiscoveryDate = parsed
		}
	}

	if domain.DiscoveryDate.IsZero() {
		domain.DiscoveryDate = time.Now()
	}

	domain.UpdatedAt = time.Now()
	domain.CreatedAt = time.Now()

	// Handle duplicates based on strategy
	existing, err := s.db.GetDomainByName(domain.Domain)
	if err == nil {
		// Domain exists, apply duplicate strategy
		switch job.Settings.DuplicateStrategy {
		case database.DuplicateStrategySkip:
			return nil // Skip without error
		case database.DuplicateStrategyReplace:
			domain.ID = existing.ID
			if err := s.db.UpdateDomain(domain); err != nil {
				return fmt.Errorf("failed to replace domain: %w", err)
			}
		case database.DuplicateStrategyMerge:
			// Keep original discovery date but update other fields
			if existing.DiscoveryDate.Before(domain.DiscoveryDate) {
				domain.DiscoveryDate = existing.DiscoveryDate
			}
			domain.ID = existing.ID
			if err := s.db.UpdateDomain(domain); err != nil {
				return fmt.Errorf("failed to merge domain: %w", err)
			}
		case database.DuplicateStrategyAppend:
			// Create with modified name
			domain.Domain = fmt.Sprintf("%s-imported-%d", domain.Domain, time.Now().Unix())
			if err := s.db.CreateDomain(domain); err != nil {
				return fmt.Errorf("failed to append domain: %w", err)
			}
		}
	} else {
		// New domain
		if err := s.db.CreateDomain(domain); err != nil {
			return fmt.Errorf("failed to create domain: %w", err)
		}
	}

	// Assign users if specified
	if len(job.Settings.AssignToUsers) > 0 {
		s.assignDomainToUsers(domain.Domain, job.Settings.AssignToUsers)
	}

	return nil
}

// processDNSRecord processes a DNS record
func (s *EnhancedMigrationService) processDNSRecord(record map[string]interface{}, job *database.MigrationJob) error {
	dnsRecord := &database.DNSRecord{}

	// Map required fields
	dnsRecord.Domain = getStringField(record, "domain", "")
	dnsRecord.Subdomain = getStringField(record, "subdomain", "@")
	dnsRecord.Type = getStringField(record, "type", "A")
	dnsRecord.Data = getStringField(record, "data", "")

	if dnsRecord.Domain == "" || dnsRecord.Data == "" {
		return fmt.Errorf("missing required fields (domain or data)")
	}

	// Set other fields
	dnsRecord.Source = getStringField(record, "source", "Migration")
	dnsRecord.Status = getStringField(record, "status", "active")

	// Parse discovery date
	if discoveryDate := getStringField(record, "discovery_date", ""); discoveryDate != "" {
		if parsed, err := parseFlexibleDate(discoveryDate); err == nil {
			dnsRecord.DiscoveryDate = parsed
		}
	}

	if dnsRecord.DiscoveryDate.IsZero() {
		dnsRecord.DiscoveryDate = time.Now()
	}

	dnsRecord.UpdatedAt = time.Now()
	dnsRecord.CreatedAt = time.Now()

	// Handle duplicates
	existingRecords, err := s.db.GetDNSRecordsByDomain(dnsRecord.Domain)
	if err == nil {
		for _, existing := range existingRecords {
			if existing.Subdomain == dnsRecord.Subdomain && 
			   existing.Type == dnsRecord.Type && 
			   existing.Data == dnsRecord.Data {
				// Duplicate found
				switch job.Settings.DuplicateStrategy {
				case database.DuplicateStrategySkip:
					return nil
				case database.DuplicateStrategyReplace:
					dnsRecord.ID = existing.ID
					return s.db.UpdateDNSRecord(dnsRecord)
				case database.DuplicateStrategyMerge:
					if existing.DiscoveryDate.Before(dnsRecord.DiscoveryDate) {
						dnsRecord.DiscoveryDate = existing.DiscoveryDate
					}
					dnsRecord.ID = existing.ID
					return s.db.UpdateDNSRecord(dnsRecord)
				}
				break
			}
		}
	}

	// Create new record
	if err := s.db.CreateDNSRecord(dnsRecord); err != nil {
		return fmt.Errorf("failed to create DNS record: %w", err)
	}

	// Assign users if specified
	if len(job.Settings.AssignToUsers) > 0 {
		s.assignDNSToUsers(dnsRecord.ID, job.Settings.AssignToUsers)
	}

	return nil
}

// processOrgAssetRecord processes an org asset record (combined domain/DNS)
func (s *EnhancedMigrationService) processOrgAssetRecord(record map[string]interface{}, job *database.MigrationJob) error {
	// First process as domain
	if err := s.processDomainRecord(record, job); err != nil {
		// Don't fail completely, just log warning
		job.Warnings = append(job.Warnings, fmt.Sprintf("Failed to process domain part: %v", err))
	}

	// Then process DNS records if present
	if recordType := getStringField(record, "type", ""); recordType != "" && recordType != "domain" {
		return s.processDNSRecord(record, job)
	}

	return nil
}

// Utility functions

func (s *EnhancedMigrationService) generateJobID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

type ContentAnalysis struct {
	Type         string
	TotalRecords int
	SampleRecord map[string]interface{}
	Fields       []string
}

func (s *EnhancedMigrationService) analyzeContent(content, filename string) (*ContentAnalysis, error) {
	var records []map[string]interface{}
	if err := json.Unmarshal([]byte(content), &records); err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("no records found")
	}

	// Determine type based on filename or content
	migType := "unknown"
	if strings.Contains(strings.ToLower(filename), "domain") {
		migType = "domains"
	} else if strings.Contains(strings.ToLower(filename), "dns") {
		migType = "dns"
	} else if strings.Contains(strings.ToLower(filename), "org") {
		migType = "org_assets"
	}

	// Extract field names from first record
	fields := make([]string, 0)
	for key := range records[0] {
		fields = append(fields, key)
	}

	return &ContentAnalysis{
		Type:         migType,
		TotalRecords: len(records),
		SampleRecord: records[0],
		Fields:       fields,
	}, nil
}

func (s *EnhancedMigrationService) validateDomain(domain string) error {
	// Basic domain validation
	if len(domain) == 0 || len(domain) > 253 {
		return fmt.Errorf("invalid domain length")
	}

	// Use regex for basic validation
	domainRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`)
	if !domainRegex.MatchString(domain) {
		return fmt.Errorf("invalid domain format")
	}

	// Optional: Check with API if clients are available
	if s.godaddyClient != nil {
		// Try to validate with GoDaddy API
		// This is optional and can be extended based on API capabilities
	}

	return nil
}

func (s *EnhancedMigrationService) assignDomainToUsers(domainName string, userIDs []string) {
	for _, userIDStr := range userIDs {
		if userID, err := strconv.Atoi(userIDStr); err == nil {
			assignment := &database.DomainAssignment{
				Domain:     domainName,
				UserID:     userID,
				AssignedAt: time.Now(),
			}
			s.db.CreateDomainAssignment(assignment) // Ignore errors for assignments
		}
	}
}

func (s *EnhancedMigrationService) assignDNSToUsers(dnsRecordID int, userIDs []string) {
	for _, userIDStr := range userIDs {
		if userID, err := strconv.Atoi(userIDStr); err == nil {
			assignment := &database.DNSAssignment{
				DNSRecordID: dnsRecordID,
				UserID:      userID,
				AssignedAt:  time.Now(),
			}
			s.db.CreateDNSAssignment(assignment) // Ignore errors for assignments
		}
	}
}

func (s *EnhancedMigrationService) updateProgress(job *database.MigrationJob) {
	s.mu.Lock()
	defer s.mu.Unlock()

	progress := &database.MigrationProgress{
		JobID:         job.ID,
		CurrentRecord: job.ProcessedRecords,
		TotalRecords:  job.TotalRecords,
		CurrentBatch:  job.CurrentBatch,
		TotalBatches:  job.TotalBatches,
		Progress:      job.Progress,
		Status:        job.Status,
		LastUpdate:    time.Now(),
	}

	// Calculate estimated time remaining
	if job.Progress > 0 {
		elapsed := time.Since(job.StartTime)
		totalEstimated := time.Duration(float64(elapsed) * 100.0 / job.Progress)
		remaining := totalEstimated - elapsed
		if remaining > 0 {
			progress.EstimatedTimeRemaining = remaining.Round(time.Second).String()
		}
	}

	s.progressMap[job.ID] = progress
}

func (s *EnhancedMigrationService) applyTemplate(records []interface{}, template *database.MigrationTemplate) []interface{} {
	mappedRecords := make([]interface{}, len(records))
	
	for i, record := range records {
		recordMap, ok := record.(map[string]interface{})
		if !ok {
			mappedRecords[i] = record
			continue
		}

		// Apply field mappings from template
		mappedRecord := make(map[string]interface{})
		for sourceField, targetField := range template.FieldMappings {
			if value, exists := recordMap[sourceField]; exists {
				mappedRecord[targetField] = value
			}
		}

		// Copy unmapped fields
		for key, value := range recordMap {
			if _, mapped := template.FieldMappings[key]; !mapped {
				mappedRecord[key] = value
			}
		}

		mappedRecords[i] = mappedRecord
	}

	return mappedRecords
}

// Helper functions for field extraction
func getStringField(record map[string]interface{}, key, defaultValue string) string {
	if value, ok := record[key]; ok {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return defaultValue
}

func getFloatField(record map[string]interface{}, key string, defaultValue float64) float64 {
	if value, ok := record[key]; ok {
		switch v := value.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		case string:
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				return f
			}
		}
	}
	return defaultValue
}

func parseFlexibleDate(dateStr string) (time.Time, error) {
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02 15:04:05",
		"2006-01-02",
		"01/02/2006",
		"2006/01/02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

// initializeDefaultTemplates creates default migration templates
func (s *EnhancedMigrationService) initializeDefaultTemplates() {
	templates := []database.MigrationTemplate{
		{
			ID:          "godaddy-domains-v1",
			Name:        "GoDaddy Domains (Legacy)",
			Description: "Template for migrating old GoDaddy domain files (domains.json)",
			Type:        "domains",
			FieldMappings: map[string]string{
				"domain":         "domain",
				"status":         "status",
				"discovery_date": "discovery_date",
				"created_at":     "discovery_date",
			},
			DefaultSettings: database.MigrationSettings{
				BatchSize:           50,
				DuplicateStrategy:   database.DuplicateStrategyMerge,
				ValidateDomains:     true,
				SkipInvalidRecords:  true,
				CreateBackup:        true,
			},
			ValidationRules: []database.ValidationRule{
				{Field: "domain", Required: true, Pattern: `^[a-zA-Z0-9][a-zA-Z0-9-]*[a-zA-Z0-9]*\.[a-zA-Z]{2,}$`},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			CreatedBy: "system",
			IsDefault: true,
		},
		{
			ID:          "dns-records-v1",
			Name:        "DNS Records (Legacy)",
			Description: "Template for migrating old DNS record files (dns_records.json)",
			Type:        "dns",
			FieldMappings: map[string]string{
				"domain":         "domain",
				"subdomain":      "subdomain",
				"type":           "type",
				"data":           "data",
				"ttl":            "ttl",
				"priority":       "priority",
				"discovery_date": "discovery_date",
				"created_at":     "discovery_date",
			},
			DefaultSettings: database.MigrationSettings{
				BatchSize:           100,
				DuplicateStrategy:   database.DuplicateStrategySkip,
				ValidateDomains:     false,
				SkipInvalidRecords:  true,
				CreateBackup:        true,
			},
			ValidationRules: []database.ValidationRule{
				{Field: "domain", Required: true},
				{Field: "type", Required: true, AllowedValues: []string{"A", "AAAA", "CNAME", "MX", "TXT", "NS", "SOA", "SRV", "PTR"}},
				{Field: "data", Required: true},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			CreatedBy: "system",
			IsDefault: true,
		},
		{
			ID:          "org-assets-v1",
			Name:        "Organization Assets (Combined)",
			Description: "Template for migrating combined org asset files (org_assets.json)",
			Type:        "org_assets",
			FieldMappings: map[string]string{
				"domain":         "domain",
				"subdomain":      "subdomain",
				"type":           "type",
				"data":           "data",
				"source":         "source",
				"discovery_date": "discovery_date",
				"created_at":     "discovery_date",
			},
			DefaultSettings: database.MigrationSettings{
				BatchSize:           75,
				DuplicateStrategy:   database.DuplicateStrategyMerge,
				ValidateDomains:     true,
				SkipInvalidRecords:  true,
				CreateBackup:        true,
			},
			ValidationRules: []database.ValidationRule{
				{Field: "domain", Required: true},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			CreatedBy: "system",
			IsDefault: true,
		},
		{
			ID:          "custom-flexible-v1",
			Name:        "Flexible Import",
			Description: "Flexible template that adapts to various JSON structures",
			Type:        "flexible",
			FieldMappings: map[string]string{
				// Auto-mapping based on common field names
				"name":           "domain",
				"hostname":       "domain",
				"domain_name":    "domain",
				"record_name":    "subdomain",
				"record_type":    "type",
				"record_data":    "data",
				"record_value":   "data",
				"time_created":   "discovery_date",
				"date_created":   "discovery_date",
				"created":        "discovery_date",
			},
			DefaultSettings: database.MigrationSettings{
				BatchSize:           25,
				DuplicateStrategy:   database.DuplicateStrategySkip,
				ValidateDomains:     false,
				SkipInvalidRecords:  true,
				CreateBackup:        true,
			},
			ValidationRules: []database.ValidationRule{},
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
			CreatedBy:       "system",
			IsDefault:       true,
		},
	}

	// Create templates if they don't exist
	for _, template := range templates {
		existing, err := s.db.GetMigrationTemplate(template.ID)
		if err != nil || existing == nil {
			s.db.CreateMigrationTemplate(&template)
		}
	}
}

// GetMigrationTemplates returns all available templates
func (s *EnhancedMigrationService) GetMigrationTemplates() ([]database.MigrationTemplate, error) {
	return s.db.GetMigrationTemplates()
}

// GetMigrationTemplate returns a specific template
func (s *EnhancedMigrationService) GetMigrationTemplate(id string) (*database.MigrationTemplate, error) {
	return s.db.GetMigrationTemplate(id)
}

// CreateCustomTemplate creates a custom migration template
func (s *EnhancedMigrationService) CreateCustomTemplate(template *database.MigrationTemplate, createdBy string) error {
	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()
	template.CreatedBy = createdBy
	template.IsDefault = false

	return s.db.CreateMigrationTemplate(template)
}

// GetMigrationJobs returns all migration jobs
func (s *EnhancedMigrationService) GetMigrationJobs() ([]database.MigrationJob, error) {
	return s.db.GetMigrationJobs()
}

// GetMigrationJob returns a specific migration job
func (s *EnhancedMigrationService) GetMigrationJob(id string) (*database.MigrationJob, error) {
	return s.db.GetMigrationJob(id)
}

// AnalyzeFile analyzes uploaded file and suggests migration approach
func (s *EnhancedMigrationService) AnalyzeFile(filename, content string) (map[string]interface{}, error) {
	analysis, err := s.analyzeContent(content, filename)
	if err != nil {
		return nil, err
	}

	// Suggest template based on analysis
	templates, _ := s.db.GetDefaultMigrationTemplates()
	suggestedTemplate := ""
	
	for _, template := range templates {
		if template.Type == analysis.Type {
			suggestedTemplate = template.ID
			break
		}
	}

	// If no direct match, use flexible template
	if suggestedTemplate == "" {
		suggestedTemplate = "custom-flexible-v1"
	}

	return map[string]interface{}{
		"filename":         filename,
		"type":            analysis.Type,
		"count":           analysis.TotalRecords,
		"fields":          analysis.Fields,
		"sample_record":   analysis.SampleRecord,
		"suggested_template": suggestedTemplate,
		"filesize":        len(content),
	}, nil
}

// PreviewMigration generates a preview of what would be imported
func (s *EnhancedMigrationService) PreviewMigration(content string, settings database.MigrationSettings) (map[string]interface{}, error) {
	var records []interface{}
	if err := json.Unmarshal([]byte(content), &records); err != nil {
		return nil, err
	}

	// Apply template if specified
	if settings.TemplateID != "" {
		template, err := s.db.GetMigrationTemplate(settings.TemplateID)
		if err == nil {
			records = s.applyTemplate(records, template)
		}
	}

	// Take first 5 records as preview
	sampleCount := 5
	if len(records) < sampleCount {
		sampleCount = len(records)
	}

	samples := make([]map[string]interface{}, sampleCount)
	for i := 0; i < sampleCount; i++ {
		if recordMap, ok := records[i].(map[string]interface{}); ok {
			samples[i] = recordMap
		}
	}

	return map[string]interface{}{
		"type":         determineRecordType(samples[0]),
		"total_count":  len(records),
		"sample_count": sampleCount,
		"samples":      samples,
	}, nil
}

func determineRecordType(record map[string]interface{}) string {
	// Check if it looks like a DNS record
	if _, hasType := record["type"]; hasType {
		if _, hasData := record["data"]; hasData {
			return "DNS Record"
		}
	}
	
	// Check if it looks like a domain
	if _, hasDomain := record["domain"]; hasDomain {
		return "Domain"
	}
	
	return "Mixed/Unknown"
}