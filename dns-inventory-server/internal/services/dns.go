package services

import (
	"dns-inventory/internal/api"
	"dns-inventory/internal/database"
	"fmt"
	"log"
	"strings"
	"time"
)

// DNSService handles DNS record operations
type DNSService struct {
	db               *database.FileDB
	godaddyClient    *api.GoDaddyClient
	cloudflareClient *api.CloudflareClient
	domainService    *DomainService
}

// NewDNSService creates a new DNS service
func NewDNSService(db *database.FileDB, godaddyClient *api.GoDaddyClient, cloudflareClient *api.CloudflareClient, domainService *DomainService) *DNSService {
	return &DNSService{
		db:               db,
		godaddyClient:    godaddyClient,
		cloudflareClient: cloudflareClient,
		domainService:    domainService,
	}
}

// CollectDNSRecords fetches DNS records from both providers and updates the database
func (s *DNSService) CollectDNSRecords() error {
	log.Println("Starting DNS records collection...")
	
	// Get all domains to fetch DNS records for
	domains, _, err := s.db.GetDomains(10000, 0, "") // Get all domains
	if err != nil {
		return fmt.Errorf("failed to fetch domains: %w", err)
	}
	
	var collectedRecords []*database.DNSRecord
	
	// Collect DNS records from both providers
	for _, domainWithUsers := range domains {
		domain := domainWithUsers.Domain.Domain
		
		// Collect from GoDaddy if domain has GoDaddy status
		if domainWithUsers.Domain.GodaddyStatus != "" {
			if records, err := s.collectGoDaddyDNSRecords(domain); err != nil {
				log.Printf("Failed to collect GoDaddy DNS records for %s: %v", domain, err)
			} else {
				collectedRecords = append(collectedRecords, records...)
			}
		}
		
		// Collect from Cloudflare if domain has Cloudflare status
		if domainWithUsers.Domain.CloudflareStatus != "" {
			if records, err := s.collectCloudflareDNSRecords(domain); err != nil {
				log.Printf("Failed to collect Cloudflare DNS records for %s: %v", domain, err)
			} else {
				collectedRecords = append(collectedRecords, records...)
			}
		}
	}
	
	// Update database with collected records
	return s.updateDNSDatabase(collectedRecords)
}

// collectGoDaddyDNSRecords fetches DNS records from GoDaddy for a domain
func (s *DNSService) collectGoDaddyDNSRecords(domain string) ([]*database.DNSRecord, error) {
	log.Printf("Fetching DNS records from GoDaddy for %s", domain)
	
	gdRecords, err := s.godaddyClient.GetDNSRecords(domain)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch GoDaddy DNS records: %w", err)
	}
	
	var records []*database.DNSRecord
	now := time.Now()
	
	for _, record := range gdRecords {
		subdomain := ""
		if record.Name != "@" {
			subdomain = record.Name
		}
		
		dnsRecord := &database.DNSRecord{
			Domain:        domain,
			Subdomain:     subdomain,
			Type:          record.Type,
			Data:          record.Data,
			Source:        "GoDaddy",
			Status:        "active",
			DiscoveryDate: now,
			LastSeen:      now,
		}
		
		records = append(records, dnsRecord)
	}
	
	log.Printf("Collected %d DNS records from GoDaddy for %s", len(records), domain)
	return records, nil
}

// collectCloudflareDNSRecords fetches DNS records from Cloudflare for a domain
func (s *DNSService) collectCloudflareDNSRecords(domain string) ([]*database.DNSRecord, error) {
	log.Printf("Fetching DNS records from Cloudflare for %s", domain)
	
	// First get the zone ID for this domain
	zones, err := s.cloudflareClient.GetZones()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Cloudflare zones: %w", err)
	}
	
	var zoneID string
	for _, zone := range zones {
		if zone.Name == domain {
			zoneID = zone.ID
			break
		}
	}
	
	if zoneID == "" {
		return nil, fmt.Errorf("zone not found for domain %s", domain)
	}
	
	cfRecords, err := s.cloudflareClient.GetDNSRecords(zoneID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Cloudflare DNS records: %w", err)
	}
	
	var records []*database.DNSRecord
	now := time.Now()
	
	for _, record := range cfRecords {
		subdomain := ""
		if record.Name != domain {
			// Extract subdomain by removing domain suffix
			if strings.HasSuffix(record.Name, "."+domain) {
				subdomain = strings.TrimSuffix(record.Name, "."+domain)
			} else if record.Name != domain {
				subdomain = record.Name
			}
		}
		
		dnsRecord := &database.DNSRecord{
			Domain:        domain,
			Subdomain:     subdomain,
			Type:          record.Type,
			Data:          record.Content,
			Source:        "Cloudflare",
			Status:        "active",
			DiscoveryDate: now,
			LastSeen:      now,
		}
		
		records = append(records, dnsRecord)
	}
	
	log.Printf("Collected %d DNS records from Cloudflare for %s", len(records), domain)
	return records, nil
}

// updateDNSDatabase merges collected DNS records with existing data
func (s *DNSService) updateDNSDatabase(collectedRecords []*database.DNSRecord) error {
	log.Println("Updating DNS records database...")
	
	// Get existing DNS records to preserve discovery dates
	existingRecords, _, err := s.db.GetDNSRecords(100000, 0, "") // Get all existing
	if err != nil {
		return fmt.Errorf("failed to fetch existing DNS records: %w", err)
	}
	
	// Create map for quick lookup
	existingMap := make(map[string]*database.DNSRecord)
	for _, r := range existingRecords {
		key := s.createRecordKey(r.DNSRecord)
		existingMap[key] = &r.DNSRecord
	}
	
	// Track what we've seen in this collection
	seenKeys := make(map[string]bool)
	now := time.Now()
	updated := 0
	created := 0
	
	// Process collected records
	for _, record := range collectedRecords {
		key := s.createRecordKey(*record)
		seenKeys[key] = true
		
		if existing, exists := existingMap[key]; exists {
			// Update existing record
			existing.LastSeen = now
			existing.Status = "active" // Mark as active since we found it
			if err := s.db.UpdateDNSRecord(existing); err != nil {
				log.Printf("Failed to update DNS record %s: %v", key, err)
				continue
			}
			updated++
		} else {
			// Create new record
			if err := s.db.CreateDNSRecord(record); err != nil {
				log.Printf("Failed to create DNS record %s: %v", key, err)
				continue
			}
			created++
		}
	}
	
	// Mark records not seen as removed (but only if their source was checked)
	removed := 0
	for _, existingRecord := range existingRecords {
		key := s.createRecordKey(existingRecord.DNSRecord)
		if !seenKeys[key] && existingRecord.DNSRecord.Status == "active" {
			// Check if we collected data for this domain and source
			if s.shouldMarkAsRemoved(existingRecord.DNSRecord) {
				existingRecord.DNSRecord.Status = "removed"
				if err := s.db.UpdateDNSRecord(&existingRecord.DNSRecord); err != nil {
					log.Printf("Failed to mark DNS record as removed %s: %v", key, err)
					continue
				}
				removed++
			}
		}
	}
	
	log.Printf("DNS records update complete: %d created, %d updated, %d marked as removed", created, updated, removed)
	
	// Create snapshot
	if err := s.db.CreateSnapshot(); err != nil {
		log.Printf("Failed to create snapshot: %v", err)
	}
	
	return nil
}

// createRecordKey creates a unique key for a DNS record
func (s *DNSService) createRecordKey(record database.DNSRecord) string {
	return fmt.Sprintf("%s|%s|%s|%s|%s", 
		record.Domain, 
		record.Subdomain, 
		record.Type, 
		record.Data, 
		record.Source)
}

// shouldMarkAsRemoved determines if a record should be marked as removed
func (s *DNSService) shouldMarkAsRemoved(record database.DNSRecord) bool {
	// Get the domain to check if we collected data for it
	domains, _, err := s.db.GetDomains(10000, 0, record.Domain)
	if err != nil {
		return false
	}
	
	for _, domain := range domains {
		if domain.Domain.Domain == record.Domain {
			// Check if we would have collected from this source
			if record.Source == "GoDaddy" && domain.Domain.GodaddyStatus != "" {
				return true
			}
			if record.Source == "Cloudflare" && domain.Domain.CloudflareStatus != "" {
				return true
			}
		}
	}
	
	return false
}

// GetDNSRecords retrieves DNS records with pagination and search
func (s *DNSService) GetDNSRecords(limit, offset int, search string) ([]database.DNSRecordWithUsers, int, error) {
	return s.db.GetDNSRecords(limit, offset, search)
}

// AssignUserToDNSRecord assigns a user to a DNS record
func (s *DNSService) AssignUserToDNSRecord(recordID, userID int) error {
	return s.db.AssignDNSToUser(recordID, userID)
}

// ExportDNSRecordsCSV exports DNS records to CSV format
func (s *DNSService) ExportDNSRecordsCSV() ([]byte, error) {
	records, _, err := s.db.GetDNSRecords(100000, 0, "") // Get all records
	if err != nil {
		return nil, err
	}
	
	var csv strings.Builder
	csv.WriteString("Domain,Subdomain,Type,Data,Source,Status,Discovery Date,Last Seen,Assigned Users\n")
	
	for _, record := range records {
		var userNames []string
		for _, user := range record.AssignedUsers {
			userNames = append(userNames, user.Name)
		}
		
		subdomain := record.DNSRecord.Subdomain
		if subdomain == "" {
			subdomain = "@"
		}
		
		csv.WriteString(fmt.Sprintf("%s,%s,%s,\"%s\",%s,%s,%s,%s,\"%s\"\n",
			record.DNSRecord.Domain,
			subdomain,
			record.DNSRecord.Type,
			record.DNSRecord.Data,
			record.DNSRecord.Source,
			record.DNSRecord.Status,
			record.DNSRecord.DiscoveryDate.Format("2006-01-02"),
			record.DNSRecord.LastSeen.Format("2006-01-02"),
			strings.Join(userNames, "; "),
		))
	}
	
	return []byte(csv.String()), nil
}