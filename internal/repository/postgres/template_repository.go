package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	"nuclei-service-demo/internal/config"
	"nuclei-service-demo/internal/model"
	"nuclei-service-demo/internal/repository"
)

// TemplateRepository implements repository.TemplateRepository
type TemplateRepository struct {
	db     *sql.DB
	cfg    *config.Config
	logger *zap.Logger
}

// NewTemplateRepository creates a new template repository
func NewTemplateRepository(db *sql.DB, cfg *config.Config, logger *zap.Logger) *TemplateRepository {
	return &TemplateRepository{
		db:     db,
		cfg:    cfg,
		logger: logger,
	}
}

// List returns a list of templates
func (r *TemplateRepository) List(ctx context.Context, tags, author, severity, templateType *string) ([]*model.Template, error) {
	r.logger.Info("Listing templates from database",
		zap.String("tags", safePtr(tags)),
		zap.String("author", safePtr(author)),
		zap.String("severity", safePtr(severity)),
		zap.String("type", safePtr(templateType)))

	// Build query
	query := `
		SELECT t.id, t.path, t.author, t.severity
		FROM templates t
		WHERE 1=1
	`
	args := []interface{}{}

	// Remove tags filter since the column doesn't exist
	// if tags != nil {
	// 	query += ` AND t.tags @> $1`
	// 	args = append(args, *tags)
	// }
	if author != nil {
		query += ` AND t.author = $1`
		args = append(args, *author)
	}
	if severity != nil {
		query += ` AND t.severity = $2`
		args = append(args, *severity)
	}
	// Remove type filter since the column doesn't exist
	// if templateType != nil {
	// 	query += ` AND t.type = $4`
	// 	args = append(args, *templateType)
	// }

	r.logger.Info("Executing template list query",
		zap.String("query", query),
		zap.Int("args_count", len(args)))

	// Execute query
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to execute template list query", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	// Scan results
	var templates []*model.Template
	for rows.Next() {
		var template model.Template
		if err := rows.Scan(
			&template.ID,
			&template.Path,
			&template.Author,
			&template.Severity,
		); err != nil {
			r.logger.Error("Failed to scan template row", zap.Error(err))
			return nil, err
		}
		// Set default values for missing columns
		template.Type = "unknown"
		template.Tags = []string{}
		templates = append(templates, &template)
	}

	r.logger.Info("Retrieved templates from database", zap.Int("count", len(templates)))
	return templates, nil
}

// Get returns a template by ID
func (r *TemplateRepository) Get(ctx context.Context, id string) (*model.Template, error) {
	r.logger.Info("Getting template from database", zap.String("id", id))

	// Build query
	query := `
		SELECT t.id, t.path, t.author, t.severity
		FROM templates t
		WHERE t.id = $1
	`

	r.logger.Info("Executing template get query", zap.String("query", query))

	// Execute query
	var template model.Template
	if err := r.db.QueryRowContext(ctx, query, id).Scan(
		&template.ID,
		&template.Path,
		&template.Author,
		&template.Severity,
	); err != nil {
		if err == sql.ErrNoRows {
			r.logger.Warn("Template not found", zap.String("id", id))
			return nil, repository.ErrNotFound
		}
		r.logger.Error("Failed to get template", zap.Error(err), zap.String("id", id))
		return nil, err
	}

	// Set default values for missing columns
	template.Type = "unknown"
	template.Tags = []string{}

	r.logger.Info("Retrieved template from database", zap.String("id", id))
	return &template, nil
}

// Create creates a new template
func (r *TemplateRepository) Create(ctx context.Context, template *model.Template) error {
	r.logger.Info("Creating template in database",
		zap.String("id", template.ID),
		zap.String("author", template.Author),
		zap.String("severity", template.Severity))

	// Build query
	query := `
		INSERT INTO templates (id, path, author, severity)
		VALUES ($1, $2, $3, $4)
	`

	r.logger.Info("Executing template create query", zap.String("query", query))

	// Execute query
	_, err := r.db.ExecContext(ctx, query,
		template.ID,
		template.Path,
		template.Author,
		template.Severity,
	)
	if err != nil {
		r.logger.Error("Failed to create template", zap.Error(err), zap.String("id", template.ID))
		return err
	}

	r.logger.Info("Successfully created template", zap.String("id", template.ID))
	return nil
}

// Update updates a template
func (r *TemplateRepository) Update(ctx context.Context, template *model.Template) error {
	r.logger.Info("Updating template in database", zap.String("id", template.ID))

	// Build query
	query := `
		UPDATE templates
		SET path = $1, author = $2, severity = $3
		WHERE id = $4
	`

	r.logger.Info("Executing template update query", zap.String("query", query))

	// Execute query
	_, err := r.db.ExecContext(ctx, query,
		template.Path,
		template.Author,
		template.Severity,
		template.ID,
	)
	if err != nil {
		r.logger.Error("Failed to update template", zap.Error(err), zap.String("id", template.ID))
		return err
	}

	r.logger.Info("Successfully updated template", zap.String("id", template.ID))
	return nil
}

// Delete deletes a template by ID
func (r *TemplateRepository) Delete(ctx context.Context, id string) error {
	r.logger.Info("Deleting template from database", zap.String("id", id))

	// Build query
	query := `
		DELETE FROM templates
		WHERE id = $1
	`

	r.logger.Info("Executing template delete query", zap.String("query", query))

	// Execute query
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("Failed to delete template", zap.Error(err), zap.String("id", id))
		return err
	}

	r.logger.Info("Successfully deleted template", zap.String("id", id))
	return nil
}

// Refresh refreshes the template cache
func (r *TemplateRepository) Refresh(ctx context.Context) error {
	r.logger.Info("Refreshing template cache")

	// Build query
	query := `
		TRUNCATE templates
	`

	r.logger.Info("Executing template refresh query", zap.String("query", query))

	// Execute query
	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		r.logger.Error("Failed to refresh template cache", zap.Error(err))
		return err
	}

	r.logger.Info("Successfully refreshed template cache")
	return nil
}

// scanTemplateDirectory scans a directory for template files
func (r *TemplateRepository) scanTemplateDirectory(dir string) ([]*model.Template, error) {
	var templates []*model.Template

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, ".yaml") {
			template, err := r.parseTemplateFile(path)
			if err != nil {
				r.logger.Warn("Failed to parse template file", zap.Error(err), zap.String("path", path))
				return nil
			}
			templates = append(templates, template)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk template directory: %w", err)
	}

	return templates, nil
}

// parseTemplateFile parses a template file
func (r *TemplateRepository) parseTemplateFile(path string) (*model.Template, error) {
	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file: %w", err)
	}

	// Parse YAML
	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse template YAML: %w", err)
	}

	// Extract fields
	info, ok := raw["info"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("missing info section in template")
	}

	// Get ID from file path
	id := ""
	rel, err := filepath.Rel(r.cfg.Nuclei.TemplatesDir, path)
	if err == nil {
		id = strings.TrimSuffix(rel, filepath.Ext(rel))
	} else {
		id = filepath.Base(path)
	}

	// Create template
	template := &model.Template{
		ID:          id,
		Name:        getString(info, "name"),
		Author:      getString(info, "author"),
		Tags:        getStringSlice(info, "tags"),
		Severity:    getString(info, "severity"),
		Type:        getString(info, "type"),
		Description: getString(info, "description"),
		Path:        path,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	return template, nil
}

// Helper functions for parsing template fields
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

func getStringSlice(m map[string]interface{}, key string) []string {
	if val, ok := m[key].([]interface{}); ok {
		result := make([]string, len(val))
		for i, v := range val {
			if s, ok := v.(string); ok {
				result[i] = s
			}
		}
		return result
	}
	return nil
}

// Helper function to safely dereference string pointers for logging
func safePtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
