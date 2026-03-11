package main

import (
	"context"
	"encoding/json"
	"log"
	"math"
	"math/rand/v2"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

const traceHeader = "X-Trace-ID"

type ServerConfig struct {
	Port int
}

func loadConfig() ServerConfig {
	portStr := os.Getenv("APP_PORT")
	if portStr == "" {
		portStr = "8500"
	}
	p, err := strconv.Atoi(portStr)
	if err != nil || p <= 0 || p > 65535 {
		p = 8500
	}
	return ServerConfig{Port: p}
}

// simple metrics store

type routeMetrics struct {
	mu          sync.Mutex
	requests    int64
	errors      int64
	latencyBuckets map[string]int64
}

type metricsRegistry struct {
	mu      sync.Mutex
	routes  map[string]*routeMetrics
}

func newMetricsRegistry() *metricsRegistry {
	return &metricsRegistry{
		routes: make(map[string]*routeMetrics),
	}
}

func (r *metricsRegistry) forRoute(route string) *routeMetrics {
	r.mu.Lock()
	defer r.mu.Unlock()
	if m, ok := r.routes[route]; ok {
		return m
	}
	m := &routeMetrics{
		latencyBuckets: map[string]int64{
			"<100ms": 0,
			"100-500ms": 0,
			"500ms-1s": 0,
			">=1s": 0,
		},
	}
	r.routes[route] = m
	return m
}

func (m *routeMetrics) record(latency time.Duration, err bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requests++
	if err {
		m.errors++
	}
	ms := float64(latency.Milliseconds())
	switch {
	case ms < 100:
		m.latencyBuckets["<100ms"]++
	case ms < 500:
		m.latencyBuckets["100-500ms"]++
	case ms < 1000:
		m.latencyBuckets["500ms-1s"]++
	default:
		m.latencyBuckets[">=1s"]++
	}
}

func (r *metricsRegistry) render() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := ""
	for route, m := range r.routes {
		m.mu.Lock()
		req := m.requests
		errs := m.errors
		buckets := make(map[string]int64, len(m.latencyBuckets))
		for k, v := range m.latencyBuckets {
			buckets[k] = v
		}
		m.mu.Unlock()
		out += "# route " + route + "\n"
		out += "requests_total " + strconv.FormatInt(req, 10) + "\n"
		out += "errors_total " + strconv.FormatInt(errs, 10) + "\n"
		for name, v := range buckets {
			out += "latency_bucket{" + "le=\"" + name + "\"} " + strconv.FormatInt(v, 10) + "\n"
		}
		out += "\n"
	}
	return out
}

// context key for trace id
type ctxKeyTraceID struct{}

func main() {
	logger := log.New(os.Stdout, "[module7-observability] ", log.LstdFlags|log.Lmicroseconds)
	cfg := loadConfig()

	metrics := newMetricsRegistry()

	mux := http.NewServeMux()

	// health
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// metrics endpoint
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		_, _ = w.Write([]byte(metrics.render()))
	})

	// a handler that calls a "downstream" function, propagating trace ID
	mux.Handle("/work", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		traceID := traceIDFromContext(ctx)

		start := time.Now()
		err := doWork(ctx, logger)
		latency := time.Since(start)

		routeKey := "GET /work"
		metrics.forRoute(routeKey).record(latency, err != nil)

		if err != nil {
			logger.Printf("trace_id=%s handler=/work error=%v latency_ms=%.2f", traceID, err, float64(latency.Milliseconds()))
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		logger.Printf("trace_id=%s handler=/work success latency_ms=%.2f", traceID, float64(latency.Milliseconds()))
		writeJSON(w, http.StatusOK, map[string]any{
			"status":   "ok",
			"trace_id": traceID,
			"latency":  latency.String(),
		})
	}))

	handler := tracingMiddleware(logger, metricsMiddleware(metrics, mux))

	server := &http.Server{
		Addr:         ":" + strconv.Itoa(cfg.Port),
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	logger.Printf("starting observability demo on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("server error: %v", err)
	}
}

// doWork simulates some processing and uses the trace ID from context
func doWork(ctx context.Context, logger *log.Logger) error {
	traceID := traceIDFromContext(ctx)
	// simulate variable work
	delay := time.Duration(50+rand.IntN(400)) * time.Millisecond
	time.Sleep(delay)
	// 10% random error
	if rand.Float64() < 0.1 {
		return ctx.Err()
	}
	// pretend we called another service and log with trace ID
	logger.Printf("trace_id=%s downstream_call duration_ms=%.2f", traceID, float64(delay.Milliseconds()))
	return nil
}

// tracingMiddleware extracts or generates a trace ID and logs per-request info
func tracingMiddleware(logger *log.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := r.Header.Get(traceHeader)
		if traceID == "" {
			traceID = generateTraceID()
		}
		// add trace id to context
		ctx := context.WithValue(r.Context(), ctxKeyTraceID{}, traceID)
		r = r.WithContext(ctx)

		start := time.Now()
		logger.Printf("trace_id=%s started method=%s path=%s", traceID, r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
		latency := time.Since(start)
		logger.Printf("trace_id=%s completed method=%s path=%s latency_ms=%.2f", traceID, r.Method, r.URL.Path, float64(latency.Milliseconds()))
	})
}

// metricsMiddleware records basic RED metrics per route (using path+method as key)
func metricsMiddleware(reg *metricsRegistry, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		routeKey := r.Method + " " + r.URL.Path
		errFlag := false

		ww := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(ww, r)

		if ww.status >= 500 {
			errFlag = true
		}
		latency := time.Since(start)
		reg.forRoute(routeKey).record(latency, errFlag)
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func traceIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(ctxKeyTraceID{}).(string); ok {
		return v
	}
	return ""
}

func generateTraceID() string {
	// simple random hex-like string
	return strconv.FormatInt(time.Now().UnixNano(), 16) + "-" + strconv.Itoa(int(math.Abs(float64(rand.Int()))))
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

