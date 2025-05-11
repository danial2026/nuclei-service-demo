package service

import (
	"context"
	"time"

	"nuclei-service-demo/internal/model"
)

// TemplateService defines the interface for template operations
type TemplateService interface {
	// List returns a list of templates
	List(ctx context.Context, tags, author, severity, templateType *string) ([]model.Template, error)
	// Get returns a template by ID
	Get(ctx context.Context, id string) (*model.Template, error)
	// Refresh refreshes the template cache
	Refresh(ctx context.Context) error
}

// ScanService defines the interface for scan operations
type ScanService interface {
	// List returns a list of scans
	ListScans(ctx context.Context, status, target, templateID *string) ([]model.Scan, error)
	// Get returns a scan by ID
	GetScan(ctx context.Context, id string) (*model.Scan, error)
	// Start starts a new scan
	StartScan(ctx context.Context, input model.StartScanInput) (*model.Scan, error)
	// Delete deletes a scan by ID
	DeleteScan(ctx context.Context, id string) (bool, error)
}

// NucleiService handles running nuclei scans
type NucleiService interface {
	// CancelScan cancels a running scan
	CancelScan(ctx context.Context, scanID string) error
	// StartScan starts a new nuclei scan
	StartScan(ctx context.Context, scan *model.Scan) error
}

// Template represents a Nuclei template
type Template struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Author      string    `json:"author"`
	Tags        []string  `json:"tags"`
	Severity    string    `json:"severity"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Scan represents a Nuclei scan
type Scan struct {
	ID          string       `json:"id"`
	Target      string       `json:"target"`
	Status      string       `json:"status"`
	StartedAt   time.Time    `json:"started_at"`
	CompletedAt *time.Time   `json:"completed_at,omitempty"`
	Results     []ScanResult `json:"results"`
}

// ScanResult represents a result from a Nuclei scan
type ScanResult struct {
	TemplateID       string                 `json:"template_id"`
	TemplateName     string                 `json:"template_name"`
	Severity         string                 `json:"severity"`
	MatchedAt        time.Time              `json:"matched_at"`
	MatcherName      string                 `json:"matcher_name"`
	ExtractedResults []string               `json:"extracted_results"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// ScanOptions represents the options for a scan
type ScanOptions struct {
	Concurrency     int  `json:"concurrency"`
	RateLimit       int  `json:"rate_limit"`
	Timeout         int  `json:"timeout"`
	Retries         int  `json:"retries"`
	Headless        bool `json:"headless"`
	FollowRedirects bool `json:"follow_redirects"`
}
