package repository

import (
	"context"
	"errors"
	"nuclei-service-demo/internal/model"
)

// Common errors
var (
	ErrNotFound = errors.New("not found")
)

// TemplateRepository defines the interface for template operations
type TemplateRepository interface {
	// List returns a list of templates
	List(ctx context.Context, tags, author, severity, templateType *string) ([]*model.Template, error)
	// Get returns a template by ID
	Get(ctx context.Context, id string) (*model.Template, error)
	// Create creates a new template
	Create(ctx context.Context, template *model.Template) error
	// Update updates a template
	Update(ctx context.Context, template *model.Template) error
	// Delete deletes a template by ID
	Delete(ctx context.Context, id string) error
	// Refresh refreshes the template cache
	Refresh(ctx context.Context) error
}

// ScanRepository defines the interface for scan operations
type ScanRepository interface {
	// List returns a list of scans
	List(ctx context.Context, status, target, templateID *string) ([]*model.Scan, error)
	// Get returns a scan by ID
	Get(ctx context.Context, id string) (*model.Scan, error)
	// Create creates a new scan
	Create(ctx context.Context, scan *model.Scan) error
	// Update updates a scan
	Update(ctx context.Context, scan *model.Scan) error
	// Delete deletes a scan by ID
	Delete(ctx context.Context, id string) error
	// AddResult adds a scan result
	AddResult(ctx context.Context, result *model.ScanResult) error
	// GetResults returns scan results for a scan
	GetResults(ctx context.Context, scanID string) ([]*model.ScanResult, error)
}
