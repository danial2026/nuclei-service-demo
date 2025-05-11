package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"nuclei-service-demo/internal/config"
	"nuclei-service-demo/internal/model"
	"nuclei-service-demo/internal/repository"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// templateService implements the TemplateService interface
type templateService struct {
	repo   repository.TemplateRepository
	cfg    *config.Config
	logger *zap.Logger
}

// NewTemplateService creates a new template service
func NewTemplateService(repo repository.TemplateRepository, cfg *config.Config, logger *zap.Logger) TemplateService {
	return &templateService{
		repo:   repo,
		cfg:    cfg,
		logger: logger,
	}
}

// List returns a list of templates
func (s *templateService) List(ctx context.Context, tags, author, severity, templateType *string) ([]model.Template, error) {
	s.logger.Info("Listing templates",
		zap.String("tags", safePtr(tags)),
		zap.String("author", safePtr(author)),
		zap.String("severity", safePtr(severity)),
		zap.String("type", safePtr(templateType)))

	templates, err := s.repo.List(ctx, tags, author, severity, templateType)
	if err != nil {
		s.logger.Error("Failed to list templates from repository", zap.Error(err))
		return nil, err
	}

	s.logger.Info("Retrieved templates from repository", zap.Int("count", len(templates)))

	// Convert to model.Template
	result := make([]model.Template, len(templates))
	for i, template := range templates {
		result[i] = *template
	}
	return result, nil
}

// Get returns a template by ID
func (s *templateService) Get(ctx context.Context, id string) (*model.Template, error) {
	s.logger.Info("Getting template by ID", zap.String("id", id))

	template, err := s.repo.Get(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get template from repository", zap.Error(err), zap.String("id", id))
		return nil, err
	}

	s.logger.Info("Retrieved template from repository", zap.String("id", id))
	return template, nil
}

// Refresh refreshes the template cache
func (s *templateService) Refresh(ctx context.Context) error {
	s.logger.Info("Starting template refresh")

	// First, clear the existing templates
	if err := s.repo.Refresh(ctx); err != nil {
		s.logger.Error("Failed to clear templates", zap.Error(err))
		return fmt.Errorf("failed to clear templates: %w", err)
	}
	s.logger.Info("Cleared existing templates")

	// Get and validate template directory
	templatesDir := s.cfg.Nuclei.TemplatesDir

	// Check if directory exists
	if stat, err := os.Stat(templatesDir); err != nil {
		s.logger.Error("Templates directory not found", zap.String("dir", templatesDir), zap.Error(err))
		return fmt.Errorf("templates directory not found: %w", err)
	} else if !stat.IsDir() {
		s.logger.Error("Templates path is not a directory", zap.String("dir", templatesDir))
		return fmt.Errorf("templates path is not a directory: %s", templatesDir)
	}

	s.logger.Info("Starting to scan templates directory", zap.String("dir", templatesDir))

	templateCount := 0
	errorCount := 0

	// Walk through template directory
	err := filepath.Walk(templatesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			s.logger.Error("Error accessing path", zap.String("path", path), zap.Error(err))
			errorCount++
			return nil // Continue despite errors
		}

		// Skip directories
		if info.IsDir() {
			s.logger.Info("Skipping directory", zap.String("path", path))
			return nil
		}

		// Skip non-yaml files
		if filepath.Ext(path) != ".yaml" {
			s.logger.Info("Skipping non-yaml file", zap.String("path", path))
			return nil
		}

		s.logger.Info("Parsing template file", zap.String("path", path))

		// Parse template file
		template, err := s.parseTemplateFile(path)
		if err != nil {
			s.logger.Warn("Failed to parse template file", zap.Error(err), zap.String("path", path))
			errorCount++
			return nil // Skip this file but continue with others
		}

		// Save template
		if err := s.repo.Create(ctx, template); err != nil {
			s.logger.Error("Failed to save template", zap.Error(err), zap.String("path", path))
			errorCount++
			return nil // Skip this file but continue with others
		}

		templateCount++
		if templateCount%100 == 0 {
			s.logger.Info("Processing templates", zap.Int("processed", templateCount))
		}

		return nil
	})

	if err != nil {
		s.logger.Error("Failed to walk template directory", zap.Error(err), zap.String("dir", templatesDir))
		return fmt.Errorf("failed to walk template directory: %w", err)
	}

	s.logger.Info("Template refresh completed",
		zap.Int("totalProcessed", templateCount),
		zap.Int("errors", errorCount))
	return nil
}

// parseTemplateFile parses a template file and extracts its metadata
func (s *templateService) parseTemplateFile(path string) (*model.Template, error) {
	// Read template file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file: %w", err)
	}

	// Parse YAML
	var templateData struct {
		ID   string `yaml:"id"`
		Info struct {
			Name        string   `yaml:"name"`
			Description string   `yaml:"description"`
			Severity    string   `yaml:"severity"`
			Author      string   `yaml:"author"`
			Tags        []string `yaml:"tags"`
		} `yaml:"info"`
	}

	if err := yaml.Unmarshal(data, &templateData); err != nil {
		return nil, fmt.Errorf("failed to parse template YAML: %w", err)
	}

	// Extract ID from file path if not specified
	id := templateData.ID
	if id == "" {
		// Extract from file path (e.g., "templates/http/cves/2021/CVE-2021-12345.yaml" -> "http/cves/2021/CVE-2021-12345")
		rel, err := filepath.Rel(s.cfg.Nuclei.TemplatesDir, path)
		if err == nil {
			id = strings.TrimSuffix(rel, filepath.Ext(rel))
		} else {
			// Fallback to base filename without extension
			id = strings.TrimSuffix(filepath.Base(path), filepath.Ext(filepath.Base(path)))
		}
	}

	return &model.Template{
		ID:          id,
		Name:        templateData.Info.Name,
		Description: templateData.Info.Description,
		Severity:    templateData.Info.Severity,
		Author:      templateData.Info.Author,
		Tags:        templateData.Info.Tags,
		Path:        path,
	}, nil
}
