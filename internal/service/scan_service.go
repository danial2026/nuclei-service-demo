package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"nuclei-service-demo/internal/config"
	"nuclei-service-demo/internal/model"
	"nuclei-service-demo/internal/repository"
)

// scanService implements the ScanService interface
type scanService struct {
	scanRepo     repository.ScanRepository
	templateRepo repository.TemplateRepository
	nucleiSvc    NucleiServiceInterface
	cfg          *config.Config
	logger       *zap.Logger
}

// NewScanService creates a new scan service
func NewScanService(
	scanRepo repository.ScanRepository,
	templateRepo repository.TemplateRepository,
	nucleiSvc NucleiServiceInterface,
	cfg *config.Config,
	logger *zap.Logger,
) ScanService {
	return &scanService{
		scanRepo:     scanRepo,
		templateRepo: templateRepo,
		nucleiSvc:    nucleiSvc,
		cfg:          cfg,
		logger:       logger,
	}
}

// ListScans lists scans
func (s *scanService) ListScans(ctx context.Context, status, target, templateID *string) ([]model.Scan, error) {
	s.logger.Info("Listing scans",
		zap.String("status", safePtr(status)),
		zap.String("target", safePtr(target)),
		zap.String("templateID", safePtr(templateID)))

	scans, err := s.scanRepo.List(ctx, status, target, templateID)
	if err != nil {
		s.logger.Error("Failed to list scans from repository", zap.Error(err))
		return nil, err
	}

	s.logger.Info("Retrieved scans from repository", zap.Int("count", len(scans)))

	// Convert []*model.Scan to []model.Scan
	result := make([]model.Scan, len(scans))
	for i, scan := range scans {
		result[i] = *scan
	}

	return result, nil
}

// GetScan gets a scan by ID
func (s *scanService) GetScan(ctx context.Context, id string) (*model.Scan, error) {
	s.logger.Info("Getting scan", zap.String("id", id))

	scan, err := s.scanRepo.Get(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get scan from repository", zap.Error(err), zap.String("id", id))
		return nil, err
	}

	s.logger.Info("Retrieved scan from repository", zap.String("id", id))
	return scan, nil
}

// StartScan starts a new scan
func (s *scanService) StartScan(ctx context.Context, input model.StartScanInput) (*model.Scan, error) {
	s.logger.Info("Starting scan",
		zap.String("target", input.Target),
		zap.Strings("templateIDs", input.TemplateIDs),
		zap.Strings("tags", input.Tags))

	// Create scan
	scan := &model.Scan{
		ID:          uuid.New().String(),
		Target:      input.Target,
		TemplateIDs: input.TemplateIDs,
		Tags:        input.Tags,
		Options:     input.Options,
		Status:      model.ScanStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save scan
	if err := s.scanRepo.Create(ctx, scan); err != nil {
		s.logger.Error("Failed to create scan in repository", zap.Error(err))
		return nil, err
	}

	s.logger.Info("Created scan in repository", zap.String("id", scan.ID))
	return scan, nil
}

// DeleteScan deletes a scan
func (s *scanService) DeleteScan(ctx context.Context, id string) (bool, error) {
	s.logger.Info("Deleting scan", zap.String("id", id))

	// Get scan
	scan, err := s.scanRepo.Get(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get scan from repository", zap.Error(err), zap.String("id", id))
		return false, err
	}

	// Cancel scan if running
	if scan.Status == model.ScanStatusRunning {
		if err := s.nucleiSvc.CancelScan(ctx, id); err != nil {
			s.logger.Error("Failed to cancel scan", zap.Error(err))
		}
	}

	// Delete scan
	if err := s.scanRepo.Delete(ctx, id); err != nil {
		s.logger.Error("Failed to delete scan from repository", zap.Error(err), zap.String("id", id))
		return false, err
	}

	s.logger.Info("Deleted scan from repository", zap.String("id", id))
	return true, nil
}

// GetScanResults returns scan results for a scan
func (s *scanService) GetScanResults(ctx context.Context, scanID string) ([]*model.ScanResult, error) {
	s.logger.Info("Getting scan results", zap.String("scan_id", scanID))

	results, err := s.scanRepo.GetResults(ctx, scanID)
	if err != nil {
		s.logger.Error("Failed to get scan results from repository", zap.Error(err), zap.String("scan_id", scanID))
		return nil, err
	}

	s.logger.Info("Retrieved scan results from repository", zap.String("scan_id", scanID), zap.Int("count", len(results)))
	return results, nil
}
