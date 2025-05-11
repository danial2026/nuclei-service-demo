package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"nuclei-service-demo/internal/config"
	"os"
	"os/exec"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// DemoServer represents a server with intentionally vulnerable endpoints for testing
type DemoServer struct {
	logger *zap.Logger
	router *mux.Router
	http   *http.Server
}

// NewDemoServer creates a new demo server instance
func NewDemoServer(cfg *config.Config) (*DemoServer, error) {
	// Create logger
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	// Create router
	router := mux.NewRouter()

	// Create server
	srv := &DemoServer{
		logger: logger,
		router: router,
		http: &http.Server{
			Addr:         fmt.Sprintf(":%d", cfg.Server.DemoPort),
			Handler:      router,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}

	// Register routes
	srv.registerRoutes()

	return srv, nil
}

// Start starts the demo server
func (s *DemoServer) Start() error {
	s.logger.Info("Starting demo server", zap.String("addr", s.http.Addr))
	return s.http.ListenAndServe()
}

// Shutdown gracefully shuts down the demo server
func (s *DemoServer) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down demo server")
	return s.http.Shutdown(ctx)
}

// registerRoutes registers all vulnerable endpoints
func (s *DemoServer) registerRoutes() {
	// 1. Open Redirect (generic)
	s.router.HandleFunc("/vuln/openredirect", s.handleOpenRedirect()).Methods(http.MethodGet)

	// 2. Oracle Fatwire LFI
	s.router.HandleFunc("/vuln/lfi-fatwire", s.handleFatwireLFI()).Methods(http.MethodGet)

	// 3. HiBoss RCE
	s.router.HandleFunc("/vuln/hiboss-rce", s.handleHiBossRCE()).Methods(http.MethodGet)

	// 4. ThinkPHP Arbitrary File Write
	s.router.HandleFunc("/vuln/thinkphp-write", s.handleThinkPHPWrite()).Methods(http.MethodGet)

	// 5. Zyxel Unauthenticated LFI
	s.router.HandleFunc("/vuln/zyxel-lfi", s.handleZyxelLFI()).Methods(http.MethodGet)

	// 6. Nuxt.js XSS
	s.router.HandleFunc("/vuln/nuxt-xss", s.handleNuxtXSS()).Methods(http.MethodGet)

	// 7. Sick-Beard XSS
	s.router.HandleFunc("/vuln/sickbeard-xss", s.handleSickBeardXSS()).Methods(http.MethodGet)

	// 8. Fastjson Deserialization RCE
	s.router.HandleFunc("/vuln/fastjson-rce", s.handleFastjsonRCE()).Methods(http.MethodPost)

	// 9. BeyondTrust XSS
	s.router.HandleFunc("/vuln/beyondtrust-xss", s.handleBeyondTrustXSS()).Methods(http.MethodGet)

	// 10. WordPress Brandfolder Open Redirect
	s.router.HandleFunc("/vuln/brandfolder-redirect", s.handleBrandfolderRedirect()).Methods(http.MethodGet)
}

func (s *DemoServer) handleOpenRedirect() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dest := r.URL.Query().Get("redirect")
		http.Redirect(w, r, dest, http.StatusFound)
	}
}

func (s *DemoServer) handleFatwireLFI() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn := r.URL.Query().Get("fn")
		data, err := os.ReadFile(fn)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.Write(data)
	}
}

func (s *DemoServer) handleHiBossRCE() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := r.URL.Query().Get("ip")
		out, err := exec.Command("sh", "-c", "ping -c 1 "+ip).CombinedOutput()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(out)
	}
}

func (s *DemoServer) handleThinkPHPWrite() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		content := r.URL.Query().Get("content")
		filename := "pwned.txt"
		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintln(w, "Wrote to", filename)
	}
}

func (s *DemoServer) handleZyxelLFI() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		file := r.URL.Query().Get("path")
		data, err := os.ReadFile(file)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.Write(data)
	}
}

func (s *DemoServer) handleNuxtXSS() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stack := r.URL.Query().Get("stack")
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, "<html><body>Error stack: %s</body></html>", stack)
	}
}

func (s *DemoServer) handleSickBeardXSS() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pattern := r.URL.Query().Get("pattern")
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, "<div>Pattern: %s</div>", pattern)
	}
}

func (s *DemoServer) handleFastjsonRCE() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.Write(body)
	}
}

func (s *DemoServer) handleBeyondTrustXSS() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		input := r.URL.Query().Get("input")
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, "<h1>Challenge: %s</h1>", input)
	}
}

func (s *DemoServer) handleBrandfolderRedirect() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.Query().Get("url")
		http.Redirect(w, r, url, http.StatusFound)
	}
}
