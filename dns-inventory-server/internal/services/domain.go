package services

import (
	"dns-inventory/internal/api"
	"dns-inventory/internal/database"
	"fmt"
	"log"
	"strings"
	"time"
)

// DomainService handles domain-related operations
type DomainService struct {
	db             *database.FileDB
	godaddyClient  *api.GoDaddyClient
	cloudflareClient *api.CloudflareClient
}

// NewDomainService creates a new domain service
func NewDomainService(db *database.FileDB, godaddyClient *api.GoDaddyClient, cloudflareClient *api.CloudflareClient) *DomainService {
	return &DomainService{
		db:               db,
		godaddyClient:    godaddyClient,
		cloudflareClient: cloudflareClient,
	}
}

// CollectDomains fetches domains from both providers and updates the database
func (s *DomainService) CollectDomains() error {
	log.Println("Starting domain collection...")
	
	// Collect from GoDaddy
	godaddyDomains := make(map[string]string)
	if err := s.collectGoDaddyDomains(godaddyDomains); err != nil {
		log.Printf("Failed to collect GoDaddy domains: %v", err)
		// Continue with Cloudflare even if GoDaddy fails
	}
	
	// Collect from Cloudflare
	cloudflareDomains := make(map[string]string)
	if err := s.collectCloudflareDomains(cloudflareDomains); err != nil {
		log.Printf("Failed to collect Cloudflare domains: %v", err)
		// Continue processing with whatever we have
	}
	
	// Merge and update database
	return s.updateDomainDatabase(godaddyDomains, cloudflareDomains)
}

// collectGoDaddyDomains fetches domains from GoDaddy
func (s *DomainService) collectGoDaddyDomains(domains map[string]string) error {
	log.Println("Fetching domains from GoDaddy...")
	
	gdDomains, err := s.godaddyClient.GetDomains()
	if err != nil {
		return fmt.Errorf("failed to fetch GoDaddy domains: %w", err)
	}
	
	for _, domain := range gdDomains {
		if !s.isTestDomain(domain.Domain) {
			domains[domain.Domain] = domain.Status
		}
	}
	
	log.Printf("Collected %d domains from GoDaddy", len(domains))
	return nil
}

// collectCloudflareDomains fetches domains from Cloudflare
func (s *DomainService) collectCloudflareDomains(domains map[string]string) error {
	log.Println("Fetching zones from Cloudflare...")
	
	zones, err := s.cloudflareClient.GetZones()
	if err != nil {
		return fmt.Errorf("failed to fetch Cloudflare zones: %w", err)
	}
	
	for _, zone := range zones {
		if !s.isTestDomain(zone.Name) {
			// Map Cloudflare status to our status
			status := "active"
			if zone.Status != "active" {
				status = "removed"
			}
			domains[zone.Name] = status
		}
	}
	
	log.Printf("Collected %d zones from Cloudflare", len(domains))
	return nil
}

// updateDomainDatabase merges collected domains with existing data
func (s *DomainService) updateDomainDatabase(godaddyDomains, cloudflareDomains map[string]string) error {
	log.Println("Updating domain database...")
	
	// Get existing domains to preserve discovery dates
	existingDomains, _, err := s.db.GetDomains(10000, 0, "") // Get all existing
	if err != nil {
		return fmt.Errorf("failed to fetch existing domains: %w", err)
	}
	
	existingMap := make(map[string]*database.Domain)
	for _, d := range existingDomains {
		existingMap[d.Domain.Domain] = &d.Domain
	}
	
	// Collect all unique domains
	allDomains := make(map[string]bool)
	for domain := range godaddyDomains {
		allDomains[domain] = true
	}
	for domain := range cloudflareDomains {
		allDomains[domain] = true
	}
	
	// Process each domain
	now := time.Now()
	updated := 0
	created := 0
	
	for domainName := range allDomains {
		godaddyStatus := godaddyDomains[domainName]
		cloudflareStatus := cloudflareDomains[domainName]
		
		if existing, exists := existingMap[domainName]; exists {
			// Update existing domain
			needsUpdate := false
			
			if existing.GodaddyStatus != godaddyStatus {
				existing.GodaddyStatus = godaddyStatus
				needsUpdate = true
			}
			
			if existing.CloudflareStatus != cloudflareStatus {
				existing.CloudflareStatus = cloudflareStatus
				needsUpdate = true
			}
			
			if needsUpdate {
				existing.LastSeen = now
				if err := s.db.UpdateDomain(existing); err != nil {
					log.Printf("Failed to update domain %s: %v", domainName, err)
					continue
				}
				updated++
			}
		} else {
			// Create new domain
			domain := &database.Domain{
				Domain:           domainName,
				GodaddyStatus:    godaddyStatus,
				CloudflareStatus: cloudflareStatus,
				DiscoveryDate:    now,
				LastSeen:         now,
			}
			
			if err := s.db.CreateDomain(domain); err != nil {
				log.Printf("Failed to create domain %s: %v", domainName, err)
				continue
			}
			created++
		}
	}
	
	log.Printf("Domain update complete: %d created, %d updated", created, updated)
	
	// Create snapshot
	if err := s.db.CreateSnapshot(); err != nil {
		log.Printf("Failed to create snapshot: %v", err)
	}
	
	return nil
}

// GetDomains retrieves domains with pagination and search
func (s *DomainService) GetDomains(limit, offset int, search string) ([]database.DomainWithUsers, int, error) {
	return s.db.GetDomains(limit, offset, search)
}

// AssignUserToDomain assigns a user to a domain
func (s *DomainService) AssignUserToDomain(domainID, userID int) error {
	return s.db.AssignDomainToUser(domainID, userID)
}

// GetDomainStats returns domain statistics
func (s *DomainService) GetDomainStats() (*database.Stats, error) {
	return s.db.GetStats()
}

// isTestDomain checks if a domain is a test/example domain
func (s *DomainService) isTestDomain(domain string) bool {
	testDomains := []string{
		"example.com", "example.org", "example.net",
		"test.com", "test.org", "test.net",
		"domain.com", "domain.org", "domain.net",
		"localhost", "invalid", "example", "test",
	}
	
	domainLower := strings.ToLower(domain)
	
	// Check exact matches
	for _, testDomain := range testDomains {
		if domainLower == testDomain {
			return true
		}
	}
	
	// Check for common test patterns
	if strings.HasPrefix(domainLower, "test-") || 
	   strings.HasPrefix(domainLower, "example-") {
		return true
	}
	
	return false
}

// ExportDomainsCSV exports domains to CSV format
func (s *DomainService) ExportDomainsCSV() ([]byte, error) {
	domains, _, err := s.db.GetDomains(10000, 0, "") // Get all domains
	if err != nil {
		return nil, err
	}
	
	var csv strings.Builder
	csv.WriteString("Domain,GoDaddy Status,Cloudflare Status,Discovery Date,Last Seen,Assigned Users\n")
	
	for _, domain := range domains {
		var userNames []string
		for _, user := range domain.AssignedUsers {
			userNames = append(userNames, user.Name)
		}
		
		godaddyStatus := domain.GodaddyStatus
		if godaddyStatus == "" {
			godaddyStatus = "-"
		}
		
		cloudflareStatus := domain.CloudflareStatus
		if cloudflareStatus == "" {
			cloudflareStatus = "-"
		}
		
		csv.WriteString(fmt.Sprintf("%s,%s,%s,%s,%s,\"%s\"\n",
			domain.Domain.Domain,
			godaddyStatus,
			cloudflareStatus,
			domain.Domain.DiscoveryDate.Format("2006-01-02"),
			domain.Domain.LastSeen.Format("2006-01-02"),
			strings.Join(userNames, "; "),
		))
	}
	
	return []byte(csv.String()), nil
}