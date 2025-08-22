package services

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"dns-inventory/internal/config"
	"dns-inventory/internal/database"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// NotificationService handles email notifications via AWS SES
type NotificationService struct {
	enabled      bool
	region       string
	accessKey    string
	secretKey    string
	fromEmail    string
	toEmail      string
	serviceName  string
	client       *http.Client
}

// NotificationEvent represents different types of notifications
type NotificationEvent string

const (
	EventMigrationCompleted NotificationEvent = "migration_completed"
	EventMigrationFailed    NotificationEvent = "migration_failed"
	EventMigrationStarted   NotificationEvent = "migration_started"
	EventDataCollection     NotificationEvent = "data_collection"
	EventAPIError           NotificationEvent = "api_error"
	EventSystemAlert        NotificationEvent = "system_alert"
)

// EmailNotification represents an email notification
type EmailNotification struct {
	Event     NotificationEvent `json:"event"`
	Subject   string            `json:"subject"`
	Message   string            `json:"message"`
	Timestamp time.Time         `json:"timestamp"`
	JobID     string            `json:"job_id,omitempty"`
	UserID    string            `json:"user_id,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// NewNotificationService creates a new notification service based on configuration
func NewNotificationService(cfg *config.Config) *NotificationService {
	// Check if SES configuration is present
	enabled := cfg.HasSESConfig()

	service := &NotificationService{
		enabled:     enabled,
		region:      cfg.AWSRegion,
		accessKey:   cfg.AWSAccessKey,
		secretKey:   cfg.AWSSecretKey,
		fromEmail:   cfg.NotificationFromEmail,
		toEmail:     cfg.NotificationToEmail,
		serviceName: "ses",
		client:      &http.Client{Timeout: 30 * time.Second},
	}

	if enabled {
		log.Printf("✅ Email notifications enabled - sending to %s", cfg.NotificationToEmail)
	} else {
		log.Println("📧 Email notifications disabled - missing AWS SES configuration in .env file")
		log.Println("   Add AWS_SES_REGION, AWS_SES_ACCESS_KEY, AWS_SES_SECRET_KEY,")
		log.Println("   NOTIFICATION_FROM_EMAIL, and NOTIFICATION_TO_EMAIL to .env to enable")
	}

	return service
}

// IsEnabled returns whether notifications are enabled
func (n *NotificationService) IsEnabled() bool {
	return n.enabled
}

// SendMigrationCompleted sends notification when migration completes
func (n *NotificationService) SendMigrationCompleted(job *database.MigrationJob) error {
	if !n.enabled {
		return nil
	}

	notification := &EmailNotification{
		Event:     EventMigrationCompleted,
		Subject:   fmt.Sprintf("✅ Migration Completed: %s", job.Filename),
		Timestamp: time.Now(),
		JobID:     job.ID,
		UserID:    job.CreatedBy,
		Details: map[string]interface{}{
			"filename":         job.Filename,
			"total_records":    job.TotalRecords,
			"success_records":  job.SuccessRecords,
			"failed_records":   job.FailedRecords,
			"skipped_records":  job.SkippedRecords,
			"start_time":       job.StartTime.Format("2006-01-02 15:04:05"),
			"duration":         time.Since(job.StartTime).String(),
		},
	}

	notification.Message = n.formatMigrationCompletedMessage(job)
	return n.sendEmail(notification)
}

// SendMigrationFailed sends notification when migration fails
func (n *NotificationService) SendMigrationFailed(job *database.MigrationJob, errorMsg string) error {
	if !n.enabled {
		return nil
	}

	notification := &EmailNotification{
		Event:     EventMigrationFailed,
		Subject:   fmt.Sprintf("❌ Migration Failed: %s", job.Filename),
		Timestamp: time.Now(),
		JobID:     job.ID,
		UserID:    job.CreatedBy,
		Details: map[string]interface{}{
			"filename":        job.Filename,
			"total_records":   job.TotalRecords,
			"processed":       job.ProcessedRecords,
			"error":           errorMsg,
			"start_time":      job.StartTime.Format("2006-01-02 15:04:05"),
			"failure_time":    time.Now().Format("2006-01-02 15:04:05"),
		},
	}

	notification.Message = n.formatMigrationFailedMessage(job, errorMsg)
	return n.sendEmail(notification)
}

// SendMigrationStarted sends notification when migration starts
func (n *NotificationService) SendMigrationStarted(job *database.MigrationJob) error {
	if !n.enabled {
		return nil
	}

	notification := &EmailNotification{
		Event:     EventMigrationStarted,
		Subject:   fmt.Sprintf("🚀 Migration Started: %s", job.Filename),
		Timestamp: time.Now(),
		JobID:     job.ID,
		UserID:    job.CreatedBy,
		Details: map[string]interface{}{
			"filename":      job.Filename,
			"total_records": job.TotalRecords,
			"batch_size":    job.Settings.BatchSize,
			"start_time":    job.StartTime.Format("2006-01-02 15:04:05"),
		},
	}

	notification.Message = n.formatMigrationStartedMessage(job)
	return n.sendEmail(notification)
}

// SendDataCollectionSummary sends daily data collection summary
func (n *NotificationService) SendDataCollectionSummary(domains, dnsRecords int, errors []string) error {
	if !n.enabled {
		return nil
	}

	notification := &EmailNotification{
		Event:     EventDataCollection,
		Subject:   fmt.Sprintf("📊 Daily Data Collection Summary - %s", time.Now().Format("2006-01-02")),
		Timestamp: time.Now(),
		Details: map[string]interface{}{
			"domains_collected":     domains,
			"dns_records_collected": dnsRecords,
			"errors":                errors,
			"collection_date":       time.Now().Format("2006-01-02"),
		},
	}

	notification.Message = n.formatDataCollectionMessage(domains, dnsRecords, errors)
	return n.sendEmail(notification)
}

// SendAPIError sends notification for API connection errors
func (n *NotificationService) SendAPIError(provider, errorMsg string) error {
	if !n.enabled {
		return nil
	}

	notification := &EmailNotification{
		Event:     EventAPIError,
		Subject:   fmt.Sprintf("⚠️ API Error: %s Connection Failed", provider),
		Timestamp: time.Now(),
		Details: map[string]interface{}{
			"provider": provider,
			"error":    errorMsg,
			"time":     time.Now().Format("2006-01-02 15:04:05"),
		},
	}

	notification.Message = n.formatAPIErrorMessage(provider, errorMsg)
	return n.sendEmail(notification)
}

// SendSystemAlert sends general system alerts
func (n *NotificationService) SendSystemAlert(title, message string, details map[string]interface{}) error {
	if !n.enabled {
		return nil
	}

	notification := &EmailNotification{
		Event:     EventSystemAlert,
		Subject:   fmt.Sprintf("🔔 System Alert: %s", title),
		Message:   message,
		Timestamp: time.Now(),
		Details:   details,
	}

	return n.sendEmail(notification)
}

// sendEmail sends email via AWS SES API
func (n *NotificationService) sendEmail(notification *EmailNotification) error {
	if !n.enabled {
		return nil
	}

	// Create SES SendEmail request
	requestBody := n.createSESRequest(notification)

	// Create HTTP request
	req, err := n.createHTTPRequest(requestBody)
	if err != nil {
		log.Printf("Failed to create HTTP request: %v", err)
		return err
	}

	// Send request
	resp, err := n.client.Do(req)
	if err != nil {
		log.Printf("Failed to send email notification: %v", err)
		return err
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("SES API error (%d): %s", resp.StatusCode, string(body))
		return fmt.Errorf("SES API error: %d", resp.StatusCode)
	}

	log.Printf("📧 Email notification sent: %s", notification.Subject)
	return nil
}

// createSESRequest creates AWS SES SendEmail request parameters
func (n *NotificationService) createSESRequest(notification *EmailNotification) url.Values {
	params := url.Values{}
	params.Set("Action", "SendEmail")
	params.Set("Version", "2010-12-01")
	
	// Source and destination
	params.Set("Source", n.fromEmail)
	params.Set("Destination.ToAddresses.member.1", n.toEmail)
	
	// Subject
	params.Set("Message.Subject.Data", notification.Subject)
	params.Set("Message.Subject.Charset", "UTF-8")
	
	// Body (both text and HTML)
	params.Set("Message.Body.Text.Data", notification.Message)
	params.Set("Message.Body.Text.Charset", "UTF-8")
	params.Set("Message.Body.Html.Data", n.formatHTMLMessage(notification))
	params.Set("Message.Body.Html.Charset", "UTF-8")

	return params
}

// createHTTPRequest creates signed HTTP request for SES
func (n *NotificationService) createHTTPRequest(params url.Values) (*http.Request, error) {
	endpoint := fmt.Sprintf("https://email.%s.amazonaws.com/", n.region)
	
	req, err := http.NewRequest("POST", endpoint, strings.NewReader(params.Encode()))
	if err != nil {
		return nil, err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Amzn-Authorization", n.createAuthHeader(params))

	return req, nil
}

// createAuthHeader creates AWS Signature Version 2 authorization header
func (n *NotificationService) createAuthHeader(params url.Values) string {
	// Create canonical string
	canonicalString := "POST\n"
	canonicalString += fmt.Sprintf("email.%s.amazonaws.com\n", n.region)
	canonicalString += "/\n"
	canonicalString += params.Encode()

	// Create signature
	mac := hmac.New(sha256.New, []byte(n.secretKey))
	mac.Write([]byte(canonicalString))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	// Create auth header
	return fmt.Sprintf("AWS3-HTTPS AWSAccessKeyId=%s,Algorithm=HmacSHA256,Signature=%s",
		n.accessKey, signature)
}

// Message formatting functions

func (n *NotificationService) formatMigrationCompletedMessage(job *database.MigrationJob) string {
	var buffer bytes.Buffer
	
	buffer.WriteString("🎉 Migration completed successfully!\n\n")
	buffer.WriteString(fmt.Sprintf("File: %s\n", job.Filename))
	buffer.WriteString(fmt.Sprintf("Started by: %s\n", job.CreatedBy))
	buffer.WriteString(fmt.Sprintf("Total Records: %d\n", job.TotalRecords))
	buffer.WriteString(fmt.Sprintf("✅ Successfully imported: %d\n", job.SuccessRecords))
	buffer.WriteString(fmt.Sprintf("❌ Failed: %d\n", job.FailedRecords))
	buffer.WriteString(fmt.Sprintf("⏭️ Skipped: %d\n", job.SkippedRecords))
	buffer.WriteString(fmt.Sprintf("⏱️ Duration: %v\n", time.Since(job.StartTime)))
	
	if len(job.Warnings) > 0 {
		buffer.WriteString(fmt.Sprintf("\n⚠️ Warnings (%d):\n", len(job.Warnings)))
		for i, warning := range job.Warnings {
			if i < 5 { // Limit to first 5 warnings
				buffer.WriteString(fmt.Sprintf("- %s\n", warning))
			}
		}
		if len(job.Warnings) > 5 {
			buffer.WriteString(fmt.Sprintf("... and %d more warnings\n", len(job.Warnings)-5))
		}
	}
	
	buffer.WriteString(fmt.Sprintf("\n📊 View details: http://localhost:8080/migration\n"))
	buffer.WriteString("\n--\nDNS Inventory\n")
	
	return buffer.String()
}

func (n *NotificationService) formatMigrationFailedMessage(job *database.MigrationJob, errorMsg string) string {
	var buffer bytes.Buffer
	
	buffer.WriteString("❌ Migration failed!\n\n")
	buffer.WriteString(fmt.Sprintf("File: %s\n", job.Filename))
	buffer.WriteString(fmt.Sprintf("Started by: %s\n", job.CreatedBy))
	buffer.WriteString(fmt.Sprintf("Records Processed: %d / %d\n", job.ProcessedRecords, job.TotalRecords))
	buffer.WriteString(fmt.Sprintf("Error: %s\n", errorMsg))
	
	if len(job.Errors) > 0 {
		buffer.WriteString(fmt.Sprintf("\nRecent Errors (%d):\n", len(job.Errors)))
		for i, err := range job.Errors {
			if i < 3 { // Limit to first 3 errors
				buffer.WriteString(fmt.Sprintf("- %s\n", err))
			}
		}
		if len(job.Errors) > 3 {
			buffer.WriteString(fmt.Sprintf("... and %d more errors\n", len(job.Errors)-3))
		}
	}
	
	buffer.WriteString(fmt.Sprintf("\n🔄 Resume migration: http://localhost:8080/migration\n"))
	buffer.WriteString("\n--\nDNS Inventory\n")
	
	return buffer.String()
}

func (n *NotificationService) formatMigrationStartedMessage(job *database.MigrationJob) string {
	var buffer bytes.Buffer
	
	buffer.WriteString("🚀 Migration started!\n\n")
	buffer.WriteString(fmt.Sprintf("File: %s\n", job.Filename))
	buffer.WriteString(fmt.Sprintf("Started by: %s\n", job.CreatedBy))
	buffer.WriteString(fmt.Sprintf("Total Records: %d\n", job.TotalRecords))
	buffer.WriteString(fmt.Sprintf("Batch Size: %d\n", job.Settings.BatchSize))
	buffer.WriteString(fmt.Sprintf("Duplicate Strategy: %s\n", job.Settings.DuplicateStrategy))
	
	if len(job.Settings.AssignToUsers) > 0 {
		buffer.WriteString(fmt.Sprintf("Assigning to users: %s\n", strings.Join(job.Settings.AssignToUsers, ", ")))
	}
	
	buffer.WriteString(fmt.Sprintf("\n📈 Monitor progress: http://localhost:8080/migration\n"))
	buffer.WriteString("\n--\nDNS Inventory\n")
	
	return buffer.String()
}

func (n *NotificationService) formatDataCollectionMessage(domains, dnsRecords int, errors []string) string {
	var buffer bytes.Buffer
	
	buffer.WriteString("📊 Daily Data Collection Summary\n\n")
	buffer.WriteString(fmt.Sprintf("Date: %s\n", time.Now().Format("2006-01-02")))
	buffer.WriteString(fmt.Sprintf("🌐 Domains collected: %d\n", domains))
	buffer.WriteString(fmt.Sprintf("📡 DNS records collected: %d\n", dnsRecords))
	
	if len(errors) > 0 {
		buffer.WriteString(fmt.Sprintf("\n⚠️ Errors encountered (%d):\n", len(errors)))
		for i, err := range errors {
			if i < 5 { // Limit to first 5 errors
				buffer.WriteString(fmt.Sprintf("- %s\n", err))
			}
		}
		if len(errors) > 5 {
			buffer.WriteString(fmt.Sprintf("... and %d more errors\n", len(errors)-5))
		}
	}
	
	buffer.WriteString(fmt.Sprintf("\n📊 View dashboard: http://localhost:8080\n"))
	buffer.WriteString("\n--\nDNS Inventory\n")
	
	return buffer.String()
}

func (n *NotificationService) formatAPIErrorMessage(provider, errorMsg string) string {
	var buffer bytes.Buffer
	
	buffer.WriteString(fmt.Sprintf("⚠️ %s API Connection Error\n\n", provider))
	buffer.WriteString(fmt.Sprintf("Time: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	buffer.WriteString(fmt.Sprintf("Error: %s\n", errorMsg))
	buffer.WriteString("\nPlease check your API credentials and network connectivity.\n")
	buffer.WriteString(fmt.Sprintf("\n📊 View dashboard: http://localhost:8080\n"))
	buffer.WriteString("\n--\nDNS Inventory\n")
	
	return buffer.String()
}

func (n *NotificationService) formatHTMLMessage(notification *EmailNotification) string {
	// Simple HTML template
	html := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #007bff; color: white; padding: 20px; border-radius: 5px; }
        .content { background: #f8f9fa; padding: 20px; border-radius: 5px; margin: 20px 0; }
        .footer { color: #666; font-size: 12px; text-align: center; margin-top: 20px; }
        .success { color: #28a745; }
        .error { color: #dc3545; }
        .warning { color: #ffc107; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h2>%s</h2>
        </div>
        <div class="content">
            <pre>%s</pre>
        </div>
        <div class="footer">
            DNS Inventory - Automated Notification
        </div>
    </div>
</body>
</html>`

	return fmt.Sprintf(html, notification.Subject, notification.Message)
}

