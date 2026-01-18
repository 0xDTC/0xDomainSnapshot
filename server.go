package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	port := flag.Int("port", 8080, "Port to serve on")
	dir := flag.String("dir", ".", "Directory to serve")
	flag.Parse()

	// Get absolute path
	absDir, err := filepath.Abs(*dir)
	if err != nil {
		log.Fatalf("Error resolving directory: %v", err)
	}

	// Check if directory exists
	if _, err := os.Stat(absDir); os.IsNotExist(err) {
		log.Fatalf("Directory does not exist: %s", absDir)
	}

	// Create file server
	fs := http.FileServer(http.Dir(absDir))

	// Wrap with logging middleware
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL.Path)

		// Set cache headers for static assets
		if filepath.Ext(r.URL.Path) == ".css" || filepath.Ext(r.URL.Path) == ".js" {
			w.Header().Set("Cache-Control", "public, max-age=3600")
		}

		// Set JSON content type for .json files
		if filepath.Ext(r.URL.Path) == ".json" {
			w.Header().Set("Content-Type", "application/json")
		}

		fs.ServeHTTP(w, r)
	})

	addr := fmt.Sprintf(":%d", *port)
	fmt.Printf("0xDomainSnapshot Server\n")
	fmt.Printf("=======================\n")
	fmt.Printf("Serving:  %s\n", absDir)
	fmt.Printf("URL:      http://localhost%s\n", addr)
	fmt.Printf("Press Ctrl+C to stop\n\n")

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
