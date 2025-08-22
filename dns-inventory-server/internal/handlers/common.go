package handlers

import (
	"encoding/json"
	"html/template"
	"os"
	"path/filepath"
	"strconv"
)

// parseInt parses a string to int with default value
func parseInt(s string, defaultValue int) int {
	if i, err := strconv.Atoi(s); err == nil {
		return i
	}
	return defaultValue
}

// loadTemplates loads and parses all HTML templates
func loadTemplates() *template.Template {
	// Define template functions
	funcMap := template.FuncMap{
		"formatDate": func(t interface{}) string {
			if timeVal, ok := t.(string); ok {
				return timeVal[:10] // Return just the date part
			}
			return ""
		},
		"formatSource": func(godaddy, cloudflare string) string {
			if godaddy != "" && cloudflare != "" {
				return "GoDaddy, Cloudflare"
			}
			if godaddy != "" {
				return "GoDaddy"
			}
			if cloudflare != "" {
				return "Cloudflare"
			}
			return ""
		},
		"formatStatus": func(godaddy, cloudflare string) string {
			if godaddy != "" && cloudflare != "" {
				return godaddy + ", " + cloudflare
			}
			if godaddy != "" {
				return godaddy
			}
			if cloudflare != "" {
				return cloudflare
			}
			return ""
		},
		"formatSubdomain": func(subdomain string) string {
			if subdomain == "" {
				return "@"
			}
			return subdomain
		},
		"json": func(v interface{}) template.JS {
			// Convert data to JSON for JavaScript
			b, _ := json.Marshal(v)
			return template.JS(b)
		},
		"join": func(users interface{}, sep string) string {
			// Helper to join user names
			return "" // Will be implemented in templates using range
		},
	}

	// Load templates with functions - use working directory for better compatibility
	templatesPath := filepath.Join("web", "templates", "*.html")
	
	// Try to get absolute path
	if workDir, err := os.Getwd(); err == nil {
		templatesPath = filepath.Join(workDir, "web", "templates", "*.html")
	}
	
	templates := template.Must(template.New("").Funcs(funcMap).ParseGlob(templatesPath))
	return templates
}