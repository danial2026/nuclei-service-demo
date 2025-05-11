package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	nucleiLib "github.com/projectdiscovery/nuclei/v3/lib"
	"github.com/projectdiscovery/nuclei/v3/pkg/output"

	"go.uber.org/zap"

	"nuclei-service-demo/internal/config"
	"nuclei-service-demo/internal/model"
)

// NucleiServiceInterface defines the interface for nuclei operations
type NucleiServiceInterface interface {
	StartScan(ctx context.Context, scan *model.Scan) ([]*model.ScanResult, error)
	CancelScan(ctx context.Context, scanID string) error
}

// nucleiService implements the NucleiServiceInterface
type nucleiService struct {
	cfg     *config.Config
	logger  *zap.Logger
	mu      sync.Mutex
	cancels map[string]context.CancelFunc
}

// NewNucleiService creates a new nuclei service
func NewNucleiService(cfg *config.Config, logger *zap.Logger) NucleiServiceInterface {
	return &nucleiService{
		cfg:     cfg,
		logger:  logger,
		cancels: make(map[string]context.CancelFunc),
	}
}

// StartScan starts a new nuclei scan using the nuclei library
func (s *nucleiService) StartScan(ctx context.Context, scan *model.Scan) ([]*model.ScanResult, error) {
	// Create cancellable context and store cancel function
	scanCtx, cancel := context.WithCancel(ctx)
	s.mu.Lock()
	s.cancels[scan.ID] = cancel
	s.mu.Unlock()

	s.logger.Info("Starting nuclei scan",
		zap.String("scan_id", scan.ID),
		zap.String("target", scan.Target),
		zap.Strings("template_ids", scan.TemplateIDs),
	)

	// Build SDK options
	opts := []nucleiLib.NucleiSDKOptions{
		// filter by severity
		nucleiLib.WithTemplateFilters(nucleiLib.TemplateFilters{
			IDs:      scan.TemplateIDs,
			Severity: "critical,high,medium,low,info",
		}),
		// load templates from directory
		nucleiLib.WithTemplatesOrWorkflows(nucleiLib.TemplateSources{
			Templates: []string{s.cfg.Nuclei.TemplatesDir},
		}),
		// disable update checks
		nucleiLib.DisableUpdateCheck(),
	}

	// add concurrency
	if scan.Options != nil {
		// if scan.Options.Concurrency > 0 {
		// 	opts = append(opts, nucleiLib.WithConcurrency(nucleiLib.Concurrency(scan.Options.Concurrency)))
		// }
		// rate limit
		if scan.Options.RateLimit > 0 {
			opts = append(opts, nucleiLib.WithGlobalRateLimitCtx(scanCtx, scan.Options.RateLimit, time.Second))
		}
		// headless
		if scan.Options.Headless {
			hopts := nucleiLib.HeadlessOpts{}
			opts = append(opts, nucleiLib.EnableHeadlessWithOpts(&hopts))
		}
	}

	// initialize engine
	engine, err := nucleiLib.NewNucleiEngineCtx(scanCtx, opts...)

	if err != nil {
		s.logger.Error("Failed to initialize nuclei engine", zap.Error(err))
		return nil, fmt.Errorf("initializing nuclei engine: %w", err)
	}
	defer engine.Close()

	engine.LoadAllTemplates()

	// load targets
	engine.LoadTargets([]string{scan.Target}, false)

	// collect results
	var results []*model.ScanResult
	callback := func(event *output.ResultEvent) {
		if event == nil {
			s.logger.Warn("Received nil event in callback")
			return
		}
		s.logger.Info("Received nuclei event", zap.Any("event", event))
		// map event to ScanResult
		result := &model.ScanResult{
			ID:         event.MatcherName + ":" + fmt.Sprint(time.Now().UnixNano()),
			ScanID:     scan.ID,
			TemplateID: event.TemplateID,
			// Severity:   event.Info.Severity,
			Host:      event.Host,
			MatchedAt: time.Now(),
			// add other fields as needed
		}
		results = append(results, result)
		s.logger.Info("Processed scan result",
			zap.String("scan_id", scan.ID),
			zap.String("result_id", result.ID),
			zap.String("template_id", result.TemplateID),
			zap.String("severity", result.Severity),
		)
	}

	// execute scan
	s.logger.Info("Executing nuclei scan", zap.String("scan_id", scan.ID))
	err = engine.ExecuteCallbackWithCtx(scanCtx, callback)
	if err != nil {
		// remove cancel
		s.mu.Lock()
		delete(s.cancels, scan.ID)
		s.mu.Unlock()
		s.logger.Error("Nuclei execution failed", zap.Error(err))
		return nil, fmt.Errorf("nuclei execution: %w", err)
	}

	// cleanup cancel
	s.mu.Lock()
	delete(s.cancels, scan.ID)
	s.mu.Unlock()

	s.logger.Info("Completed nuclei scan",
		zap.String("scan_id", scan.ID),
		zap.Int("result_count", len(results)),
	)
	return results, nil
}

// CancelScan cancels a running scan
func (s *nucleiService) CancelScan(ctx context.Context, scanID string) error {
	s.mu.Lock()
	cancel, exists := s.cancels[scanID]
	if exists {
		cancel()
		delete(s.cancels, scanID)
		s.mu.Unlock()
		s.logger.Info("Cancelled nuclei scan", zap.String("scan_id", scanID))
		return nil
	}
	s.mu.Unlock()
	return fmt.Errorf("no running scan found with ID %s", scanID)
}
