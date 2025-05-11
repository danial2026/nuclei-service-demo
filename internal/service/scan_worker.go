package service

import (
	"context"
	"time"

	"nuclei-service-demo/internal/model"
	"nuclei-service-demo/internal/repository"

	"go.uber.org/zap"
)

// ScanWorker handles background processing of pending scans
type ScanWorker struct {
	scanRepo      repository.ScanRepository
	nucleiSvc     NucleiServiceInterface
	logger        *zap.Logger
	checkInterval time.Duration
}

// NewScanWorker creates a new scan worker
func NewScanWorker(scanRepo repository.ScanRepository, nucleiSvc NucleiServiceInterface, logger *zap.Logger) *ScanWorker {
	return &ScanWorker{
		scanRepo:      scanRepo,
		nucleiSvc:     nucleiSvc,
		logger:        logger,
		checkInterval: 20 * time.Second,
	}
}

// Start begins the scan worker
func (w *ScanWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.checkInterval)
	defer ticker.Stop()

	w.logger.Info("Starting scan worker",
		zap.Duration("interval", w.checkInterval),
	)

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Stopping scan worker")
			return
		case <-ticker.C:
			if err := w.processPendingScans(ctx); err != nil {
				w.logger.Error("Error processing pending scans",
					zap.Error(err),
				)
			}
		}
	}
}

// processPendingScans processes all pending scans
func (w *ScanWorker) processPendingScans(ctx context.Context) error {
	// Get pending scans
	status := model.ScanStatusPending
	scans, err := w.scanRepo.List(ctx, &status, nil, nil)
	if err != nil {
		return err
	}

	for _, scan := range scans {
		// Update scan status to running
		scan.Status = "running"
		if err := w.scanRepo.Update(ctx, scan); err != nil {
			w.logger.Error("Failed to update scan status",
				zap.Error(err),
				zap.String("scan_id", scan.ID),
			)
			continue
		}

		// Start scan
		results, err := w.nucleiSvc.StartScan(ctx, scan)
		if err != nil {
			w.logger.Error("Failed to start scan",
				zap.Error(err),
				zap.String("scan_id", scan.ID),
			)
			scan.Status = "failed"
			scan.Error = err.Error()
			if err := w.scanRepo.Update(ctx, scan); err != nil {
				w.logger.Error("Failed to update scan status",
					zap.Error(err),
					zap.String("scan_id", scan.ID),
				)
			}
			continue
		}

		w.logger.Info("Scan completed",
			zap.String("scan_id", scan.ID),
			zap.Int("result_count", len(results)),
		)

		w.logger.Info("Scan results",
			zap.String("scan_id", scan.ID),
			zap.Any("results", results),
		)

		// Store results
		for _, result := range results {
			if err := w.scanRepo.AddResult(ctx, result); err != nil {
				w.logger.Error("Failed to store scan result",
					zap.Error(err),
					zap.String("scan_id", scan.ID),
					zap.String("result_id", result.ID),
				)
				continue
			}
		}

		// Update scan status to completed
		scan.Status = "completed"
		if err := w.scanRepo.Update(ctx, scan); err != nil {
			w.logger.Error("Failed to update scan status",
				zap.Error(err),
				zap.String("scan_id", scan.ID),
			)
		}
	}

	return nil
}
