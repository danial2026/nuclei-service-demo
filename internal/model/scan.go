package model

import (
	"time"

	"github.com/google/uuid"
)

// ScanStatus represents the status of a scan
type ScanStatus = string

const (
	// ScanStatusPending indicates a scan is waiting to start
	ScanStatusPending = "pending"
	// ScanStatusRunning indicates a scan is currently running
	ScanStatusRunning = "running"
	// ScanStatusCompleted indicates a scan has completed successfully
	ScanStatusCompleted = "completed"
	// ScanStatusFailed indicates a scan has failed
	ScanStatusFailed = "failed"
	// ScanStatusCancelled indicates a scan has been cancelled
	ScanStatusCancelled = "cancelled"
)

// Scan represents a nuclei scan
type Scan struct {
	ID          string       `json:"id" db:"id"`
	Target      string       `json:"target" db:"target"`
	Status      string       `json:"status" db:"status"`
	TemplateIDs []string     `json:"template_ids" db:"template_ids"`
	Tags        []string     `json:"tags" db:"tags"`
	Options     *ScanOptions `json:"options" db:"options"`
	Error       string       `json:"error,omitempty" db:"error"`
	CreatedAt   time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at" db:"updated_at"`
	StartedAt   *time.Time   `json:"started_at,omitempty" db:"started_at"`
	CompletedAt *time.Time   `json:"completed_at,omitempty" db:"completed_at"`
	Results     []ScanResult `json:"results,omitempty" db:"-"`
}

// ScanOptions represents the options for a scan
type ScanOptions struct {
	Concurrency int  `json:"concurrency"`
	RateLimit   int  `json:"rate_limit"`
	Timeout     int  `json:"timeout"`
	Retries     int  `json:"retries"`
	Headless    bool `json:"headless"`
}

// ScanResult represents a result from a nuclei scan
type ScanResult struct {
	ID               string                 `json:"id"`
	ScanID           string                 `json:"scan_id"`
	TemplateID       string                 `json:"template_id"`
	TemplateName     string                 `json:"template_name"`
	Severity         string                 `json:"severity"`
	Matched          bool                   `json:"matched"`
	Host             string                 `json:"host"`
	MatchedAt        time.Time              `json:"matched_at"`
	MatcherName      string                 `json:"matcher_name"`
	ExtractedResults []string               `json:"extracted_results"`
	Request          string                 `json:"request,omitempty"`
	Response         string                 `json:"response,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// StartScanInput represents the input for starting a scan
type StartScanInput struct {
	Target      string       `json:"target"`
	TemplateIDs []string     `json:"template_ids"`
	Tags        []string     `json:"tags"`
	Options     *ScanOptions `json:"options"`
}

// ParseScanStatus parses a string into a ScanStatus
func ParseScanStatus(s string) ScanStatus {
	switch s {
	case "pending":
		return ScanStatusPending
	case "running":
		return ScanStatusRunning
	case "completed":
		return ScanStatusCompleted
	case "failed":
		return ScanStatusFailed
	case "cancelled":
		return ScanStatusCancelled
	default:
		return ScanStatusPending
	}
}

// ScanFilter represents scan filtering options
type ScanFilter struct {
	Status []string `json:"status,omitempty"`
	Target string   `json:"target,omitempty"`
	From   string   `json:"from,omitempty"`
	To     string   `json:"to,omitempty"`
}

// NewUUID generates a new UUID string
func NewUUID() string {
	return uuid.New().String()
}
