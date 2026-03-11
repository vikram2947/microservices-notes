package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

type ServerConfig struct {
	Port int
}

type AppServer struct {
	httpServer *http.Server
	logger     *log.Logger
}

type healthResponse struct {
	Status string `json:"status"`
}

func loadConfig() (ServerConfig, error) {
	portStr := os.Getenv("APP_PORT")
	if portStr == "" {
		portStr = "8080"
	}
	port, err := strconv.Atoi(portStr)
	if err != nil || port <= 0 || port > 65535 {
		return ServerConfig{}, errors.New("invalid APP_PORT")
	}
	return ServerConfig{Port: port}, nil
}

func newAppServer(cfg ServerConfig, logger *log.Logger) *AppServer {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
	})

	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, healthResponse{Status: "ready"})
	})

	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		select {
		case <-time.After(100 * time.Millisecond):
			writeJSON(w, http.StatusOK, map[string]string{"message": "hello from module1-service"})
		case <-ctx.Done():
			logger.Printf("request canceled: %v", ctx.Err())
			http.Error(w, "request canceled", http.StatusRequestTimeout)
			return
		}
	})

	srv := &http.Server{
		Addr:              ":" + strconv.Itoa(cfg.Port),
		Handler:           loggingMiddleware(logger, mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	return &AppServer{
		httpServer: srv,
		logger:     logger,
	}
}

func (s *AppServer) Start() error {
	s.logger.Printf("starting server on %s", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

func (s *AppServer) Shutdown(ctx context.Context) error {
	s.logger.Println("shutting down server gracefully")
	return s.httpServer.Shutdown(ctx)
}

func loggingMiddleware(logger *log.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		logger.Printf("started %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
		logger.Printf("completed %s %s in %s", r.Method, r.URL.Path, time.Since(start))
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func main() {
	logger := log.New(os.Stdout, "[module1-service] ", log.LstdFlags|log.Lmicroseconds)

	cfg, err := loadConfig()
	if err != nil {
		logger.Fatalf("failed to load config: %v", err)
	}

	app := newAppServer(cfg, logger)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := app.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.Shutdown(shutdownCtx); err != nil {
		logger.Printf("graceful shutdown failed: %v", err)
	}

	logger.Println("server stopped")
}

