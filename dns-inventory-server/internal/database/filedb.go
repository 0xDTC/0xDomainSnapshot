package database

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// FileDB implements a simple file-based database using JSON
type FileDB struct {
	dataDir string
	mu      sync.RWMutex
}

// NewFileDB creates a new file-based database
func NewFileDB(dataDir string) (*FileDB, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}
	
	db := &FileDB{
		dataDir: dataDir,
	}
	
	// Initialize empty files if they don't exist
	files := []string{
		"domains.json",
		"dns_records.json", 
		"users.json",
		"domain_assignments.json",
		"dns_assignments.json",
		"snapshots.json",
		"migration_jobs.json",
		"migration_templates.json",
	}
	
	for _, file := range files {
		path := filepath.Join(dataDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			if err := os.WriteFile(path, []byte("[]"), 0644); err != nil {
				return nil, fmt.Errorf("failed to create %s: %w", file, err)
			}
		}
	}
	
	return db, nil
}

// Domain operations
func (db *FileDB) GetDomains(limit, offset int, search string) ([]DomainWithUsers, int, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	
	var domains []Domain
	if err := db.loadJSON("domains.json", &domains); err != nil {
		return nil, 0, err
	}
	
	// Filter by search if provided
	var filtered []Domain
	for _, domain := range domains {
		if search == "" || contains(domain.Domain, search) ||
			contains(domain.GodaddyStatus, search) ||
			contains(domain.CloudflareStatus, search) {
			filtered = append(filtered, domain)
		}
	}
	
	total := len(filtered)
	
	// Apply pagination
	start := offset
	end := offset + limit
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}
	
	// Get paginated results with users
	var result []DomainWithUsers
	for i := start; i < end; i++ {
		domain := filtered[i]
		users, _ := db.getDomainUsers(domain.ID)
		result = append(result, DomainWithUsers{
			Domain:        domain,
			AssignedUsers: users,
		})
	}
	
	return result, total, nil
}

func (db *FileDB) CreateDomain(domain *Domain) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	
	var domains []Domain
	if err := db.loadJSON("domains.json", &domains); err != nil {
		return err
	}
	
	// Find max ID
	maxID := 0
	for _, d := range domains {
		if d.ID > maxID {
			maxID = d.ID
		}
	}
	
	domain.ID = maxID + 1
	domain.CreatedAt = time.Now()
	domain.UpdatedAt = time.Now()
	
	domains = append(domains, *domain)
	return db.saveJSON("domains.json", domains)
}

func (db *FileDB) UpdateDomain(domain *Domain) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	
	var domains []Domain
	if err := db.loadJSON("domains.json", &domains); err != nil {
		return err
	}
	
	for i, d := range domains {
		if d.ID == domain.ID {
			domain.UpdatedAt = time.Now()
			domains[i] = *domain
			return db.saveJSON("domains.json", domains)
		}
	}
	
	return fmt.Errorf("domain not found")
}

// DNS Record operations
func (db *FileDB) GetDNSRecords(limit, offset int, search string) ([]DNSRecordWithUsers, int, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	
	var records []DNSRecord
	if err := db.loadJSON("dns_records.json", &records); err != nil {
		return nil, 0, err
	}
	
	// Filter by search if provided
	var filtered []DNSRecord
	for _, record := range records {
		if search == "" || contains(record.Domain, search) ||
			contains(record.Subdomain, search) ||
			contains(record.Type, search) ||
			contains(record.Data, search) ||
			contains(record.Source, search) ||
			contains(record.Status, search) {
			filtered = append(filtered, record)
		}
	}
	
	total := len(filtered)
	
	// Apply pagination
	start := offset
	end := offset + limit
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}
	
	// Get paginated results with users
	var result []DNSRecordWithUsers
	for i := start; i < end; i++ {
		record := filtered[i]
		users, _ := db.getDNSRecordUsers(record.ID)
		result = append(result, DNSRecordWithUsers{
			DNSRecord:     record,
			AssignedUsers: users,
		})
	}
	
	return result, total, nil
}

func (db *FileDB) CreateDNSRecord(record *DNSRecord) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	
	var records []DNSRecord
	if err := db.loadJSON("dns_records.json", &records); err != nil {
		return err
	}
	
	// Find max ID
	maxID := 0
	for _, r := range records {
		if r.ID > maxID {
			maxID = r.ID
		}
	}
	
	record.ID = maxID + 1
	record.CreatedAt = time.Now()
	record.UpdatedAt = time.Now()
	
	records = append(records, *record)
	return db.saveJSON("dns_records.json", records)
}

// User operations
func (db *FileDB) GetUsers() ([]User, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	
	var users []User
	err := db.loadJSON("users.json", &users)
	return users, err
}

func (db *FileDB) CreateUser(user *User) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	
	var users []User
	if err := db.loadJSON("users.json", &users); err != nil {
		return err
	}
	
	// Find max ID
	maxID := 0
	for _, u := range users {
		if u.ID > maxID {
			maxID = u.ID
		}
	}
	
	user.ID = maxID + 1
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	
	users = append(users, *user)
	return db.saveJSON("users.json", users)
}

func (db *FileDB) UpdateUser(user *User) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	
	var users []User
	if err := db.loadJSON("users.json", &users); err != nil {
		return err
	}
	
	for i, u := range users {
		if u.ID == user.ID {
			user.UpdatedAt = time.Now()
			users[i] = *user
			return db.saveJSON("users.json", users)
		}
	}
	
	return fmt.Errorf("user not found")
}

func (db *FileDB) DeleteUser(id int) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	
	var users []User
	if err := db.loadJSON("users.json", &users); err != nil {
		return err
	}
	
	var filtered []User
	for _, user := range users {
		if user.ID != id {
			filtered = append(filtered, user)
		}
	}
	
	return db.saveJSON("users.json", filtered)
}

// Assignment operations
func (db *FileDB) UpdateDNSRecord(record *DNSRecord) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	
	var records []DNSRecord
	if err := db.loadJSON("dns_records.json", &records); err != nil {
		return err
	}
	
	for i, r := range records {
		if r.ID == record.ID {
			record.UpdatedAt = time.Now()
			records[i] = *record
			return db.saveJSON("dns_records.json", records)
		}
	}
	
	return fmt.Errorf("DNS record not found")
}

func (db *FileDB) AssignDNSToUser(recordID, userID int) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	
	var assignments []DNSAssignment
	if err := db.loadJSON("dns_assignments.json", &assignments); err != nil {
		return err
	}
	
	// Check if assignment already exists
	for _, a := range assignments {
		if a.DNSRecordID == recordID && a.UserID == userID {
			return nil // Already assigned
		}
	}
	
	// Find max ID
	maxID := 0
	for _, a := range assignments {
		if a.ID > maxID {
			maxID = a.ID
		}
	}
	
	assignment := DNSAssignment{
		ID:          maxID + 1,
		DNSRecordID: recordID,
		UserID:      userID,
		AssignedAt:  time.Now(),
	}
	
	assignments = append(assignments, assignment)
	return db.saveJSON("dns_assignments.json", assignments)
}

func (db *FileDB) AssignDomainToUser(domainID, userID int) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	
	var assignments []DomainAssignment
	if err := db.loadJSON("domain_assignments.json", &assignments); err != nil {
		return err
	}
	
	// Check if assignment already exists
	for _, a := range assignments {
		if a.DomainID == domainID && a.UserID == userID {
			return nil // Already assigned
		}
	}
	
	// Find max ID
	maxID := 0
	for _, a := range assignments {
		if a.ID > maxID {
			maxID = a.ID
		}
	}
	
	assignment := DomainAssignment{
		ID:         maxID + 1,
		DomainID:   domainID,
		UserID:     userID,
		AssignedAt: time.Now(),
	}
	
	assignments = append(assignments, assignment)
	return db.saveJSON("domain_assignments.json", assignments)
}

// Stats operations
func (db *FileDB) GetStats() (*Stats, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	
	var snapshots []Snapshot
	if err := db.loadJSON("snapshots.json", &snapshots); err != nil {
		return nil, err
	}
	
	if len(snapshots) == 0 {
		return db.calculateCurrentStats()
	}
	
	// Get latest snapshot
	sort.Slice(snapshots, func(i, j int) bool {
		return snapshots[i].Date.After(snapshots[j].Date)
	})
	
	latest := snapshots[0]
	return &Stats{
		Date:                 latest.Date,
		TotalDomains:         latest.TotalDomains,
		GodaddyDomains:       latest.GodaddyDomains,
		CloudflareDomains:    latest.CloudflareDomains,
		TotalDNSRecords:      latest.TotalDNSRecords,
		GodaddyDNSRecords:    latest.GodaddyDNSRecords,
		CloudflareDNSRecords: latest.CloudflareDNSRecords,
		Status:               "✅",
	}, nil
}

func (db *FileDB) CreateSnapshot() error {
	db.mu.Lock()
	defer db.mu.Unlock()
	
	stats, err := db.calculateCurrentStats()
	if err != nil {
		return err
	}
	
	var snapshots []Snapshot
	if err := db.loadJSON("snapshots.json", &snapshots); err != nil {
		return err
	}
	
	// Find max ID
	maxID := 0
	for _, s := range snapshots {
		if s.ID > maxID {
			maxID = s.ID
		}
	}
	
	snapshot := Snapshot{
		ID:                   maxID + 1,
		Date:                 time.Now(),
		TotalDomains:         stats.TotalDomains,
		GodaddyDomains:       stats.GodaddyDomains,
		CloudflareDomains:    stats.CloudflareDomains,
		TotalDNSRecords:      stats.TotalDNSRecords,
		GodaddyDNSRecords:    stats.GodaddyDNSRecords,
		CloudflareDNSRecords: stats.CloudflareDNSRecords,
		CreatedAt:            time.Now(),
	}
	
	snapshots = append(snapshots, snapshot)
	return db.saveJSON("snapshots.json", snapshots)
}

// Helper methods
func (db *FileDB) loadJSON(filename string, v interface{}) error {
	path := filepath.Join(db.dataDir, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

func (db *FileDB) saveJSON(filename string, v interface{}) error {
	path := filepath.Join(db.dataDir, filename)
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (db *FileDB) getDomainUsers(domainID int) ([]User, error) {
	var assignments []DomainAssignment
	if err := db.loadJSON("domain_assignments.json", &assignments); err != nil {
		return nil, err
	}
	
	var users []User
	if err := db.loadJSON("users.json", &users); err != nil {
		return nil, err
	}
	
	var result []User
	for _, assignment := range assignments {
		if assignment.DomainID == domainID {
			for _, user := range users {
				if user.ID == assignment.UserID {
					result = append(result, user)
					break
				}
			}
		}
	}
	
	return result, nil
}

func (db *FileDB) getDNSRecordUsers(recordID int) ([]User, error) {
	var assignments []DNSAssignment
	if err := db.loadJSON("dns_assignments.json", &assignments); err != nil {
		return nil, err
	}
	
	var users []User
	if err := db.loadJSON("users.json", &users); err != nil {
		return nil, err
	}
	
	var result []User
	for _, assignment := range assignments {
		if assignment.DNSRecordID == recordID {
			for _, user := range users {
				if user.ID == assignment.UserID {
					result = append(result, user)
					break
				}
			}
		}
	}
	
	return result, nil
}

func (db *FileDB) calculateCurrentStats() (*Stats, error) {
	var domains []Domain
	if err := db.loadJSON("domains.json", &domains); err != nil {
		return nil, err
	}
	
	var records []DNSRecord
	if err := db.loadJSON("dns_records.json", &records); err != nil {
		return nil, err
	}
	
	stats := &Stats{
		Date:   time.Now(),
		Status: "✅",
	}
	
	// Count domains
	for _, domain := range domains {
		stats.TotalDomains++
		if domain.GodaddyStatus != "" {
			stats.GodaddyDomains++
		}
		if domain.CloudflareStatus != "" {
			stats.CloudflareDomains++
		}
	}
	
	// Count DNS records
	for _, record := range records {
		if record.Status == "active" {
			stats.TotalDNSRecords++
			if record.Source == "GoDaddy" {
				stats.GodaddyDNSRecords++
			} else if record.Source == "Cloudflare" {
				stats.CloudflareDNSRecords++
			}
		}
	}
	
	return stats, nil
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(substr) > 0 && len(s) > 0 && 
			fmt.Sprintf("%s", s)[0:min(len(s), len(substr))] == substr))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Migration Jobs methods

func (db *FileDB) CreateMigrationJob(job *MigrationJob) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	var jobs []MigrationJob
	if err := db.loadJSON("migration_jobs.json", &jobs); err != nil {
		return err
	}

	jobs = append(jobs, *job)
	return db.saveJSON("migration_jobs.json", jobs)
}

func (db *FileDB) GetMigrationJob(id string) (*MigrationJob, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var jobs []MigrationJob
	if err := db.loadJSON("migration_jobs.json", &jobs); err != nil {
		return nil, err
	}

	for _, job := range jobs {
		if job.ID == id {
			return &job, nil
		}
	}

	return nil, fmt.Errorf("migration job not found")
}

func (db *FileDB) UpdateMigrationJob(job *MigrationJob) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	var jobs []MigrationJob
	if err := db.loadJSON("migration_jobs.json", &jobs); err != nil {
		return err
	}

	for i, j := range jobs {
		if j.ID == job.ID {
			jobs[i] = *job
			return db.saveJSON("migration_jobs.json", jobs)
		}
	}

	return fmt.Errorf("migration job not found")
}

func (db *FileDB) GetMigrationJobs() ([]MigrationJob, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var jobs []MigrationJob
	if err := db.loadJSON("migration_jobs.json", &jobs); err != nil {
		return nil, err
	}

	// Sort by start time (newest first)
	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].StartTime.After(jobs[j].StartTime)
	})

	return jobs, nil
}

func (db *FileDB) DeleteMigrationJob(id string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	var jobs []MigrationJob
	if err := db.loadJSON("migration_jobs.json", &jobs); err != nil {
		return err
	}

	filtered := make([]MigrationJob, 0)
	for _, job := range jobs {
		if job.ID != id {
			filtered = append(filtered, job)
		}
	}

	return db.saveJSON("migration_jobs.json", filtered)
}

// Migration Templates methods

func (db *FileDB) CreateMigrationTemplate(template *MigrationTemplate) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	var templates []MigrationTemplate
	if err := db.loadJSON("migration_templates.json", &templates); err != nil {
		return err
	}

	templates = append(templates, *template)
	return db.saveJSON("migration_templates.json", templates)
}

func (db *FileDB) GetMigrationTemplate(id string) (*MigrationTemplate, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var templates []MigrationTemplate
	if err := db.loadJSON("migration_templates.json", &templates); err != nil {
		return nil, err
	}

	for _, template := range templates {
		if template.ID == id {
			return &template, nil
		}
	}

	return nil, fmt.Errorf("migration template not found")
}

func (db *FileDB) GetMigrationTemplates() ([]MigrationTemplate, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var templates []MigrationTemplate
	if err := db.loadJSON("migration_templates.json", &templates); err != nil {
		return nil, err
	}

	// Sort by name
	sort.Slice(templates, func(i, j int) bool {
		return templates[i].Name < templates[j].Name
	})

	return templates, nil
}

func (db *FileDB) UpdateMigrationTemplate(template *MigrationTemplate) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	var templates []MigrationTemplate
	if err := db.loadJSON("migration_templates.json", &templates); err != nil {
		return err
	}

	for i, t := range templates {
		if t.ID == template.ID {
			templates[i] = *template
			templates[i].UpdatedAt = time.Now()
			return db.saveJSON("migration_templates.json", templates)
		}
	}

	return fmt.Errorf("migration template not found")
}

func (db *FileDB) DeleteMigrationTemplate(id string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	var templates []MigrationTemplate
	if err := db.loadJSON("migration_templates.json", &templates); err != nil {
		return err
	}

	filtered := make([]MigrationTemplate, 0)
	for _, template := range templates {
		if template.ID != id {
			filtered = append(filtered, template)
		}
	}

	return db.saveJSON("migration_templates.json", filtered)
}

func (db *FileDB) GetDefaultMigrationTemplates() ([]MigrationTemplate, error) {
	templates, err := db.GetMigrationTemplates()
	if err != nil {
		return nil, err
	}

	defaults := make([]MigrationTemplate, 0)
	for _, template := range templates {
		if template.IsDefault {
			defaults = append(defaults, template)
		}
	}

	return defaults, nil
}

// Additional methods needed by migration service

func (db *FileDB) GetDomainByName(name string) (*Domain, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	
	var domains []Domain
	if err := db.loadJSON("domains.json", &domains); err != nil {
		return nil, err
	}
	
	for _, domain := range domains {
		if domain.Domain == name {
			return &domain, nil
		}
	}
	
	return nil, fmt.Errorf("domain not found")
}


func (db *FileDB) GetDNSRecordsByDomain(domain string) ([]DNSRecord, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	
	var records []DNSRecord
	if err := db.loadJSON("dns_records.json", &records); err != nil {
		return nil, err
	}
	
	var domainRecords []DNSRecord
	for _, record := range records {
		if record.Domain == domain {
			domainRecords = append(domainRecords, record)
		}
	}
	
	return domainRecords, nil
}

func (db *FileDB) CreateDomainAssignment(assignment *DomainAssignment) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	
	var assignments []DomainAssignment
	if err := db.loadJSON("domain_assignments.json", &assignments); err != nil {
		return err
	}
	
	// Check if assignment already exists
	for _, a := range assignments {
		if a.Domain == assignment.Domain && a.UserID == assignment.UserID {
			return nil // Already assigned
		}
	}
	
	// Find max ID
	maxID := 0
	for _, a := range assignments {
		if a.ID > maxID {
			maxID = a.ID
		}
	}
	
	assignment.ID = maxID + 1
	assignments = append(assignments, *assignment)
	return db.saveJSON("domain_assignments.json", assignments)
}

func (db *FileDB) CreateDNSAssignment(assignment *DNSAssignment) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	
	var assignments []DNSAssignment
	if err := db.loadJSON("dns_assignments.json", &assignments); err != nil {
		return err
	}
	
	// Check if assignment already exists
	for _, a := range assignments {
		if a.DNSRecordID == assignment.DNSRecordID && a.UserID == assignment.UserID {
			return nil // Already assigned
		}
	}
	
	// Find max ID
	maxID := 0
	for _, a := range assignments {
		if a.ID > maxID {
			maxID = a.ID
		}
	}
	
	assignment.ID = maxID + 1
	assignments = append(assignments, *assignment)
	return db.saveJSON("dns_assignments.json", assignments)
}