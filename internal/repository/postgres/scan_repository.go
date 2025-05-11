package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"
	"go.uber.org/zap"

	"nuclei-service-demo/internal/config"
	"nuclei-service-demo/internal/model"
	"nuclei-service-demo/internal/repository"
)

// ScanRepository implements repository.ScanRepository
type ScanRepository struct {
	db     *sql.DB
	cfg    *config.Config
	logger *zap.Logger
}

// NewScanRepository creates a new scan repository
func NewScanRepository(db *sql.DB, cfg *config.Config, logger *zap.Logger) *ScanRepository {
	return &ScanRepository{
		db:     db,
		cfg:    cfg,
		logger: logger,
	}
}

// List returns a list of scans
func (r *ScanRepository) List(ctx context.Context, status, target, templateID *string) ([]*model.Scan, error) {
	r.logger.Info("Listing scans from database",
		zap.String("status", safePtr(status)),
		zap.String("target", safePtr(target)),
		zap.String("templateID", safePtr(templateID)))

	// Build query
	query := `
		SELECT s.id, s.target, s.status, s.created_at, s.updated_at
		FROM scans s
		WHERE 1=1
	`
	args := []interface{}{}

	if status != nil {
		query += ` AND s.status = $1`
		args = append(args, *status)
	}
	if target != nil {
		query += ` AND s.target = $2`
		args = append(args, *target)
	}
	// Skip templateID check since the column doesn't exist
	// if templateID != nil {
	// 	query += ` AND $3 = ANY(s.template_ids)`
	// 	args = append(args, *templateID)
	// }

	r.logger.Info("Executing scan list query",
		zap.String("query", query),
		zap.Int("args_count", len(args)))

	// Execute query
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to execute scan list query", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	// Scan results
	var scans []*model.Scan
	for rows.Next() {
		var scan model.Scan
		var createdAt, updatedAt time.Time
		var statusStr string
		if err := rows.Scan(
			&scan.ID,
			&scan.Target,
			&statusStr,
			&createdAt,
			&updatedAt,
		); err != nil {
			r.logger.Error("Failed to scan row", zap.Error(err))
			return nil, err
		}

		scan.Status = model.ParseScanStatus(statusStr)
		scan.CreatedAt = createdAt
		scan.UpdatedAt = updatedAt

		// Set default values
		scan.TemplateIDs = []string{}
		scan.Tags = []string{}
		scan.Options = &model.ScanOptions{
			Concurrency: 10,
			RateLimit:   100,
			Timeout:     30,
			Retries:     3,
			Headless:    false,
		}

		scans = append(scans, &scan)
	}

	r.logger.Info("Retrieved scans from database", zap.Int("count", len(scans)))
	return scans, nil
}

// Get returns a scan by ID
func (r *ScanRepository) Get(ctx context.Context, id string) (*model.Scan, error) {
	r.logger.Info("Getting scan from database", zap.String("id", id))

	// Build query
	query := `
		SELECT s.id, s.target, s.status, s.created_at, s.updated_at
		FROM scans s
		WHERE s.id = $1
	`

	r.logger.Info("Executing scan get query", zap.String("query", query))

	// Execute query
	var scan model.Scan
	var createdAt, updatedAt time.Time
	var statusStr string
	if err := r.db.QueryRowContext(ctx, query, id).Scan(
		&scan.ID,
		&scan.Target,
		&statusStr,
		&createdAt,
		&updatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			r.logger.Warn("Scan not found", zap.String("id", id))
			return nil, repository.ErrNotFound
		}
		r.logger.Error("Failed to get scan", zap.Error(err), zap.String("id", id))
		return nil, err
	}

	scan.Status = model.ParseScanStatus(statusStr)
	scan.CreatedAt = createdAt
	scan.UpdatedAt = updatedAt

	// Set default values
	scan.TemplateIDs = []string{}
	scan.Tags = []string{}
	scan.Options = &model.ScanOptions{
		Concurrency: 10,
		RateLimit:   100,
		Timeout:     30,
		Retries:     3,
		Headless:    false,
	}

	r.logger.Info("Retrieved scan from database", zap.String("id", id))
	return &scan, nil
}

// Create creates a new scan
func (r *ScanRepository) Create(ctx context.Context, scan *model.Scan) error {
	r.logger.Info("Creating scan in database",
		zap.String("id", scan.ID),
		zap.String("target", scan.Target),
		zap.String("status", string(scan.Status)))

	// Build query
	query := `
		INSERT INTO scans (id, target, status, created_at, updated_at, template_ids, tags)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	r.logger.Info("Executing scan create query", zap.String("query", query))

	// Execute query
	now := time.Now()
	var id string
	err := r.db.QueryRowContext(ctx, query,
		scan.ID,
		scan.Target,
		scan.Status,
		now,
		now,
		pq.Array(scan.TemplateIDs),
		pq.Array(scan.Tags),
	).Scan(&id)
	if err != nil {
		r.logger.Error("Failed to create scan", zap.Error(err), zap.String("id", scan.ID))
		return err
	}

	// Update scan ID with the returned value
	scan.ID = id

	r.logger.Info("Successfully created scan", zap.String("id", scan.ID))
	return nil
}

// Update updates a scan
func (r *ScanRepository) Update(ctx context.Context, scan *model.Scan) error {
	r.logger.Info("Updating scan in database",
		zap.String("id", scan.ID),
		zap.String("status", string(scan.Status)))

	// Build query
	query := `
		UPDATE scans
		SET target = $1, status = $2, updated_at = $3
		WHERE id = $4
	`

	r.logger.Info("Executing scan update query", zap.String("query", query))

	// Execute query
	now := time.Now()
	_, err := r.db.ExecContext(ctx, query,
		scan.Target,
		scan.Status,
		now,
		scan.ID,
	)
	if err != nil {
		r.logger.Error("Failed to update scan", zap.Error(err), zap.String("id", scan.ID))
		return err
	}

	r.logger.Info("Successfully updated scan", zap.String("id", scan.ID))
	return nil
}

// Delete deletes a scan by ID
func (r *ScanRepository) Delete(ctx context.Context, id string) error {
	r.logger.Info("Deleting scan from database", zap.String("id", id))

	// Build query
	query := `
		DELETE FROM scans
		WHERE id = $1
	`

	r.logger.Info("Executing scan delete query", zap.String("query", query))

	// Execute query
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("Failed to delete scan", zap.Error(err), zap.String("id", id))
		return err
	}

	r.logger.Info("Successfully deleted scan", zap.String("id", id))
	return nil
}

// AddResult adds a scan result
func (r *ScanRepository) AddResult(ctx context.Context, result *model.ScanResult) error {
	r.logger.Info("Adding scan result to database",
		zap.String("scan_id", result.ScanID),
		zap.String("template_id", result.TemplateID),
		zap.String("severity", result.Severity))

	// Build query
	query := `
		INSERT INTO scan_results (scan_id, template_id, template_name, severity, matched, host, matched_at, matcher_name, extracted_results, request, response)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	r.logger.Info("Executing scan result create query", zap.String("query", query))

	// Execute query
	_, err := r.db.ExecContext(ctx, query,
		result.ScanID,
		result.TemplateID,
		result.TemplateName,
		result.Severity,
		result.Matched,
		result.Host,
		result.MatchedAt,
		result.MatcherName,
		pq.Array(result.ExtractedResults),
		result.Request,
		result.Response,
		// result.Metadata,
	)
	if err != nil {
		r.logger.Error("Failed to add scan result",
			zap.Error(err),
			zap.String("scan_id", result.ScanID),
			zap.String("template_id", result.TemplateID))
		return err
	}

	r.logger.Info("Successfully added scan result",
		zap.String("scan_id", result.ScanID),
		zap.String("template_id", result.TemplateID))
	return nil
}

// GetResults returns scan results for a scan
func (r *ScanRepository) GetResults(ctx context.Context, scanID string) ([]*model.ScanResult, error) {
	r.logger.Info("Getting scan results from database", zap.String("scan_id", scanID))

	// Build query
	query := `
		SELECT r.scan_id, r.template_id, r.template_name, r.severity, r.matched, r.host, r.matched_at, r.matcher_name, r.extracted_results, r.request, r.response, r.metadata
		FROM scan_results r
		WHERE r.scan_id = $1
	`

	r.logger.Info("Executing scan results get query", zap.String("query", query))

	// Execute query
	rows, err := r.db.QueryContext(ctx, query, scanID)
	if err != nil {
		r.logger.Error("Failed to get scan results", zap.Error(err), zap.String("scan_id", scanID))
		return nil, err
	}
	defer rows.Close()

	// Scan results
	var results []*model.ScanResult
	for rows.Next() {
		var result model.ScanResult
		if err := rows.Scan(
			&result.ScanID,
			&result.TemplateID,
			&result.TemplateName,
			&result.Severity,
			&result.Matched,
			&result.Host,
			&result.MatchedAt,
			&result.MatcherName,
			&result.ExtractedResults,
			&result.Request,
			&result.Response,
			&result.Metadata,
		); err != nil {
			r.logger.Error("Failed to scan result row", zap.Error(err))
			return nil, err
		}
		results = append(results, &result)
	}

	r.logger.Info("Retrieved scan results from database",
		zap.String("scan_id", scanID),
		zap.Int("count", len(results)))
	return results, nil
}

// Helper function to safely dereference string pointers for logging
// func safePtr(s *string) string {
// 	if s == nil {
// 		return ""
// 	}
// 	return *s
// }
