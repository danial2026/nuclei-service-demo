package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"nuclei-service-demo/internal/config"
	"nuclei-service-demo/internal/model"
	"nuclei-service-demo/internal/repository"
	"nuclei-service-demo/internal/repository/postgres"
	"nuclei-service-demo/internal/service"
)

// Server represents the HTTP server
type Server struct {
	cfg    *config.Config
	logger *zap.Logger
	router *mux.Router
	http   *http.Server
	db     *sql.DB
}

// New creates a new server instance
func New(cfg *config.Config) (*Server, error) {
	// Create logger
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	// Create router
	router := mux.NewRouter()

	// Add middleware
	router.Use(loggingMiddleware(logger))
	router.Use(corsMiddleware())

	// Create server
	srv := &Server{
		cfg:    cfg,
		logger: logger,
		router: router,
		http: &http.Server{
			Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
			Handler:      router,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}

	// Initialize database connection
	db, err := postgres.NewConnection(cfg.DB)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection: %w", err)
	}
	srv.db = db

	// Initialize repositories
	templateRepo := postgres.NewTemplateRepository(db, cfg, logger)
	scanRepo := postgres.NewScanRepository(db, cfg, logger)

	// Initialize services
	nucleiService := service.NewNucleiService(cfg, logger)
	templateService := service.NewTemplateService(templateRepo, cfg, logger)
	scanService := service.NewScanService(scanRepo, templateRepo, nucleiService, cfg, logger)

	// Register routes
	srv.registerRoutes(templateService, scanService, nucleiService)

	return srv, nil
}

// Start starts the server
func (s *Server) Start() error {
	s.logger.Info("Starting server", zap.Int("port", s.cfg.Server.Port))
	return s.http.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down server")
	if err := s.db.Close(); err != nil {
		s.logger.Error("Failed to close database connection", zap.Error(err))
	}
	return s.http.Shutdown(ctx)
}

// registerRoutes registers all HTTP routes
func (s *Server) registerRoutes(
	templateService service.TemplateService,
	scanService service.ScanService,
	nucleiService service.NucleiServiceInterface,
) {
	// Template routes
	s.router.HandleFunc("/api/v1/templates", s.handleListTemplates(templateService)).Methods(http.MethodGet)
	s.router.HandleFunc("/api/v1/templates/{id}", s.handleGetTemplate(templateService)).Methods(http.MethodGet)
	s.router.HandleFunc("/api/v1/templates/refresh", s.handleRefreshTemplates(templateService)).Methods(http.MethodPost)

	// Scan routes
	s.router.HandleFunc("/api/v1/scans", s.handleListScans(scanService)).Methods(http.MethodGet)
	s.router.HandleFunc("/api/v1/scans", s.handleStartScan(scanService, nucleiService)).Methods(http.MethodPost)
	s.router.HandleFunc("/api/v1/scans/{id}", s.handleGetScan(scanService)).Methods(http.MethodGet)
	s.router.HandleFunc("/api/v1/scans/{id}", s.handleDeleteScan(scanService)).Methods(http.MethodDelete)
	s.router.HandleFunc("/api/v1/scans/{id}/results", s.handleGetScanResults(scanService)).Methods(http.MethodGet)
}

// handleListTemplates handles GET /api/v1/templates
func (s *Server) handleListTemplates(service service.TemplateService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get query parameters
		tags := r.URL.Query().Get("tags")
		author := r.URL.Query().Get("author")
		severity := r.URL.Query().Get("severity")
		templateType := r.URL.Query().Get("type")

		// Convert to pointers
		var tagsPtr, authorPtr, severityPtr, typePtr *string
		if tags != "" {
			tagsPtr = &tags
		}
		if author != "" {
			authorPtr = &author
		}
		if severity != "" {
			severityPtr = &severity
		}
		if templateType != "" {
			typePtr = &templateType
		}

		// Get templates
		templates, err := service.List(r.Context(), tagsPtr, authorPtr, severityPtr, typePtr)
		if err != nil {
			s.logger.Error("Failed to list templates", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Write response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(templates); err != nil {
			s.logger.Error("Failed to encode response", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}

// handleGetTemplate handles GET /api/v1/templates/{id}
func (s *Server) handleGetTemplate(service service.TemplateService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get template ID
		vars := mux.Vars(r)
		id := vars["id"]

		// Get template
		template, err := service.Get(r.Context(), id)
		if err != nil {
			if err == repository.ErrNotFound {
				http.Error(w, "Template not found", http.StatusNotFound)
				return
			}
			s.logger.Error("Failed to get template", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Write response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(template); err != nil {
			s.logger.Error("Failed to encode response", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}

// handleRefreshTemplates handles POST /api/v1/templates/refresh
func (s *Server) handleRefreshTemplates(service service.TemplateService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Refresh templates
		if err := service.Refresh(r.Context()); err != nil {
			s.logger.Error("Failed to refresh templates", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Write response
		w.WriteHeader(http.StatusOK)
	}
}

// handleListScans handles GET /api/v1/scans
func (s *Server) handleListScans(service service.ScanService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get query parameters
		status := r.URL.Query().Get("status")
		target := r.URL.Query().Get("target")
		templateID := r.URL.Query().Get("template_id")

		// Convert to pointers
		var statusPtr, targetPtr, templateIDPtr *string
		if status != "" {
			statusPtr = &status
		}
		if target != "" {
			targetPtr = &target
		}
		if templateID != "" {
			templateIDPtr = &templateID
		}

		// Get scans
		scans, err := service.ListScans(r.Context(), statusPtr, targetPtr, templateIDPtr)
		if err != nil {
			s.logger.Error("Failed to list scans", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Write response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(scans); err != nil {
			s.logger.Error("Failed to encode response", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}

// handleStartScan handles POST /api/v1/scans
func (s *Server) handleStartScan(service service.ScanService, nucleiService service.NucleiServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse request body
		var req struct {
			Target      string   `json:"target"`
			TemplateIDs []string `json:"template_ids"`
			Tags        []string `json:"tags"`
			Options     *struct {
				Concurrency     int  `json:"concurrency"`
				RateLimit       int  `json:"rate_limit"`
				Timeout         int  `json:"timeout"`
				Retries         int  `json:"retries"`
				Headless        bool `json:"headless"`
				FollowRedirects bool `json:"follow_redirects"`
			} `json:"options"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Create scan input
		input := model.StartScanInput{
			Target:      req.Target,
			TemplateIDs: req.TemplateIDs,
			Tags:        req.Tags,
		}

		if req.Options != nil {
			input.Options = &model.ScanOptions{
				Concurrency:     req.Options.Concurrency,
				RateLimit:       req.Options.RateLimit,
				Timeout:         req.Options.Timeout,
				Retries:         req.Options.Retries,
				Headless:        req.Options.Headless,
			}
		}

		// Start scan
		scan, err := service.StartScan(r.Context(), input)
		if err != nil {
			s.logger.Error("Failed to start scan worker", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Write response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(scan); err != nil {
			s.logger.Error("Failed to encode response", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}

// handleGetScan handles GET /api/v1/scans/{id}
func (s *Server) handleGetScan(service service.ScanService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get scan ID
		vars := mux.Vars(r)
		id := vars["id"]

		// Get scan
		scan, err := service.GetScan(r.Context(), id)
		if err != nil {
			if err == repository.ErrNotFound {
				http.Error(w, "Scan not found", http.StatusNotFound)
				return
			}
			s.logger.Error("Failed to get scan", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Write response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(scan); err != nil {
			s.logger.Error("Failed to encode response", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}

// handleDeleteScan handles DELETE /api/v1/scans/{id}
func (s *Server) handleDeleteScan(service service.ScanService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get scan ID
		vars := mux.Vars(r)
		id := vars["id"]

		// Delete scan
		deleted, err := service.DeleteScan(r.Context(), id)
		if err != nil {
			if err == repository.ErrNotFound {
				http.Error(w, "Scan not found", http.StatusNotFound)
				return
			}
			s.logger.Error("Failed to delete scan", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if !deleted {
			http.Error(w, "Failed to delete scan", http.StatusInternalServerError)
			return
		}

		// Write response
		w.WriteHeader(http.StatusOK)
	}
}

// handleGetScanResults handles GET /api/v1/scans/{id}/results
func (s *Server) handleGetScanResults(service service.ScanService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get scan ID
		vars := mux.Vars(r)
		id := vars["id"]

		// Get scan
		scan, err := service.GetScan(r.Context(), id)
		if err != nil {
			if err == repository.ErrNotFound {
				http.Error(w, "Scan not found", http.StatusNotFound)
				return
			}
			s.logger.Error("Failed to get scan", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Write response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(scan.Results); err != nil {
			s.logger.Error("Failed to encode response", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}

// loggingMiddleware logs HTTP requests
func loggingMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create response writer wrapper
			rw := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Call next handler
			next.ServeHTTP(rw, r)

			// Log request
			logger.Info("HTTP request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("remote_addr", r.RemoteAddr),
				zap.Int("status", rw.statusCode),
				zap.Duration("duration", time.Since(start)),
			)
		})
	}
}

// corsMiddleware adds CORS headers
func corsMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
