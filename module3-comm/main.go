package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"math"
	"math/rand/v2"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
)

type ServerConfig struct {
	APIPort      int
	SlowPort     int
	QueueWorkers int
}

type job struct {
	ID   string `json:"id"`
	Data string `json:"data"`
}

type jobResult struct {
	JobID  string `json:"job_id"`
	Status string `json:"status"`
}

type jobStore struct {
	mu     sync.Mutex
	status map[string]jobResult
}

func newJobStore() *jobStore {
	return &jobStore{status: make(map[string]jobResult)}
}

func (s *jobStore) set(res jobResult) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status[res.JobID] = res
}

func (s *jobStore) get(id string) (jobResult, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.status[id]
	return r, ok
}

func loadConfig() (ServerConfig, error) {
	apiPort := getenvInt("API_PORT", 8090)
	slowPort := getenvInt("SLOW_PORT", 8091)
	workers := getenvInt("QUEUE_WORKERS", 2)
	if apiPort <= 0 || slowPort <= 0 {
		return ServerConfig{}, errors.New("invalid ports")
	}
	return ServerConfig{
		APIPort:      apiPort,
		SlowPort:     slowPort,
		QueueWorkers: workers,
	}, nil
}

func getenvInt(name string, def int) int {
	v := os.Getenv(name)
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return i
}

func main() {
	logger := log.New(os.Stdout, "[module3-comm] ", log.LstdFlags|log.Lmicroseconds)

	cfg, err := loadConfig()
	if err != nil {
		logger.Fatalf("config error: %v", err)
	}

	// start slow downstream service
	slowSrv := &http.Server{
		Addr:         ":" + strconv.Itoa(cfg.SlowPort),
		Handler:      newSlowService().handler(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	go func() {
		logger.Printf("starting slow-service on %s", slowSrv.Addr)
		if err := slowSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf("slow-service error: %v", err)
		}
	}()

	// client with timeout and connection pooling
	httpClient := &http.Client{
		Timeout: 1 * time.Second,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   500 * time.Millisecond,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	cb := newCircuitBreaker(3, 5*time.Second)
	bucket := newTokenBucket(5, 500*time.Millisecond)

	jobQueue := make(chan job, 100)
	store := newJobStore()

	for i := 0; i < cfg.QueueWorkers; i++ {
		go worker(jobQueue, store, logger)
	}

	apiMux := http.NewServeMux()

	apiMux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	apiMux.Handle("/do-work", rateLimitMiddleware(bucket, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !cb.allow() {
			http.Error(w, "downstream temporarily unavailable (circuit open)", http.StatusServiceUnavailable)
			return
		}

		resp, err := callSlowServiceWithRetry(r.Context(), httpClient, "http://localhost:"+strconv.Itoa(cfg.SlowPort)+"/work", 3, 100*time.Millisecond)
		if err != nil {
			cb.onFailure()
			http.Error(w, "downstream error: "+err.Error(), http.StatusBadGateway)
			return
		}
		cb.onSuccess()
		writeJSON(w, http.StatusOK, resp)
	})))

	apiMux.Handle("/enqueue", rateLimitMiddleware(bucket, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var payload struct {
			Data string `json:"data"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		id := generateID("job_")
		j := job{ID: id, Data: payload.Data}
		select {
		case jobQueue <- j:
			writeJSON(w, http.StatusAccepted, map[string]string{"job_id": id})
		default:
			http.Error(w, "queue full", http.StatusServiceUnavailable)
		}
	})))

	apiMux.HandleFunc("/jobs/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		id := r.URL.Path[len("/jobs/"):]
		if id == "" {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}
		if res, ok := store.get(id); ok {
			writeJSON(w, http.StatusOK, res)
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	})

	apiSrv := &http.Server{
		Addr:         ":" + strconv.Itoa(cfg.APIPort),
		Handler:      apiMux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Printf("starting api service on %s", apiSrv.Addr)
		if err := apiSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf("api service error: %v", err)
		}
	}()

	<-ctx.Done()
	stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_ = apiSrv.Shutdown(shutdownCtx)
	_ = slowSrv.Shutdown(shutdownCtx)

	logger.Println("module3-comm services stopped")
}

func worker(queue <-chan job, store *jobStore, logger *log.Logger) {
	for j := range queue {
		logger.Printf("processing job %s", j.ID)
		// simple idempotent behavior: status is set based on last attempt
		var attempt int
		for {
			attempt++
			if processJobOnce(j) {
				store.set(jobResult{JobID: j.ID, Status: "DONE"})
				logger.Printf("job %s completed on attempt %d", j.ID, attempt)
				break
			}
			if attempt >= 3 {
				store.set(jobResult{JobID: j.ID, Status: "FAILED"})
				logger.Printf("job %s failed after %d attempts", j.ID, attempt)
				break
			}
			backoff := time.Duration(math.Pow(2, float64(attempt))) * 100 * time.Millisecond
			jitter := time.Duration(rand.IntN(100)) * time.Millisecond
			time.Sleep(backoff + jitter)
		}
	}
}

func processJobOnce(j job) bool {
	// 70% chance of success
	return rand.IntN(10) < 7
}

func callSlowServiceWithRetry(ctx context.Context, client *http.Client, url string, maxAttempts int, baseBackoff time.Duration) (map[string]any, error) {
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		resp, err := client.Do(req)
		if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			defer resp.Body.Close()
			var out map[string]any
			if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
				return nil, err
			}
			return out, nil
		}
		if err != nil {
			lastErr = err
		} else {
			lastErr = errors.New(resp.Status)
			resp.Body.Close()
		}
		if attempt == maxAttempts {
			break
		}
		backoff := time.Duration(math.Pow(2, float64(attempt-1))) * baseBackoff
		jitter := time.Duration(rand.IntN(100)) * time.Millisecond
		select {
		case <-time.After(backoff + jitter):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return nil, lastErr
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func generateID(prefix string) string {
	return prefix + strconv.FormatInt(time.Now().UnixNano(), 36)
}

