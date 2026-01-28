package dns

import (
	"strings"
)

// testDomains is a list of test/example domains to filter out
var testDomains = map[string]bool{
	"example.com":  true,
	"example.org":  true,
	"example.net":  true,
	"test.com":     true,
	"test.org":     true,
	"test.net":     true,
	"domain.com":   true,
	"domain.org":   true,
	"domain.net":   true,
	"localhost":    true,
	"invalid":      true,
	"example":      true,
	"test":         true,
	"local":        true,
	"internal":     true,
	"localdomain":  true,
}

// testPrefixes are domain prefixes that indicate test domains
var testPrefixes = []string{
	"test-",
	"test.",
	"example-",
	"example.",
	"demo-",
	"demo.",
	"staging-",
	"dev-",
}

// IsTestDomain checks if a domain is a test/example domain that should be filtered out
func IsTestDomain(domain string) bool {
	d := strings.ToLower(strings.TrimSpace(domain))

	// Check exact match
	if testDomains[d] {
		return true
	}

	// Check prefixes
	for _, prefix := range testPrefixes {
		if strings.HasPrefix(d, prefix) {
			return true
		}
	}

	return false
}

// NormalizeSubdomain normalizes a subdomain value
// - Converts "@" to empty string (root domain)
// - Trims whitespace
// - Converts to lowercase
func NormalizeSubdomain(subdomain string) string {
	s := strings.TrimSpace(subdomain)
	if s == "@" {
		return ""
	}
	return strings.ToLower(s)
}

// ExtractSubdomain extracts the subdomain from a full hostname
// given the parent domain.
// Example: ExtractSubdomain("www.example.com", "example.com") returns "www"
func ExtractSubdomain(hostname, domain string) string {
	hostname = strings.ToLower(strings.TrimSpace(hostname))
	domain = strings.ToLower(strings.TrimSpace(domain))

	// Remove trailing dots
	hostname = strings.TrimSuffix(hostname, ".")
	domain = strings.TrimSuffix(domain, ".")

	// If hostname equals domain, it's the root
	if hostname == domain {
		return ""
	}

	// Check if hostname ends with .domain
	suffix := "." + domain
	if strings.HasSuffix(hostname, suffix) {
		return strings.TrimSuffix(hostname, suffix)
	}

	// Fallback: return hostname as-is
	return hostname
}

// ValidRecordTypes is a list of valid DNS record types
var ValidRecordTypes = map[string]bool{
	"A":     true,
	"AAAA":  true,
	"CNAME": true,
	"MX":    true,
	"TXT":   true,
	"NS":    true,
	"SOA":   true,
	"SRV":   true,
	"CAA":   true,
	"PTR":   true,
	"NAPTR": true,
	"DNSKEY": true,
	"DS":    true,
	"TLSA":  true,
	"SSHFP": true,
	"SPF":   true,  // Deprecated but still used
}

// IsValidRecordType checks if a record type is valid
func IsValidRecordType(recordType string) bool {
	return ValidRecordTypes[strings.ToUpper(recordType)]
}

// NormalizeRecordType normalizes a record type to uppercase
func NormalizeRecordType(recordType string) string {
	return strings.ToUpper(strings.TrimSpace(recordType))
}
