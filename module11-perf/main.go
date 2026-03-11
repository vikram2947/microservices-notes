package main

import (
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

type ServerConfig struct {
	Port int
}

func loadConfig() ServerConfig {
	portStr := os.Getenv("APP_PORT")
	if portStr == "" {
		portStr = "8700"
	}
	p, err := strconv.Atoi(portStr)
	if err != nil || p <= 0 || p > 65535 {
		p = 8700
	}
	return ServerConfig{Port: p}
}

// Simulated slow data source

func fetchFromSlowSource(key string) (string, time.Duration) {
	// simulate variable latency 50-200ms
	delay := time.Duration(50+rand.IntN(150)) * time.Millisecond
	time.Sleep(delay)
	return "value-for-" + key, delay
}

// Simple in-memory cache with TTL

type cacheEntry struct {
	value     string
	expiresAt time.Time
}

type cache struct {
	mu    sync.Mutex
	data  map[string]cacheEntry
	ttl   time.Duration
	hits  int64
	miss  int64
}

func newCache(ttl time.Duration) *cache {
	return &cache{
		data: make(map[string]cacheEntry),
		ttl:  ttl,
	}
}

func (c *cache) get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.data[key]
	if !ok || time.Now().After(entry.expiresAt) {
		if ok {
			delete(c.data, key)
		}
		c.miss++
		return "", false
	}
	c.hits++
	return entry.value, true
}

func (c *cache) set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = cacheEntry{
		value:     value,
		expiresAt: time.Now().Add(c.ttl),
	}
}

func (c *cache) stats() (hits, miss int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.hits, c.miss
}

// Simple latency tracker for load generator

type latencyStats struct {
	mu      sync.Mutex
	values  []float64
}

func newLatencyStats() *latencyStats {
	return &latencyStats{values: make([]float64, 0, 1024)}
}

func (s *latencyStats) add(ms float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values = append(s.values, ms)
}

func (s *latencyStats) summary() (avg, p95 float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.values) == 0 {
		return 0, 0
	}
	sum := 0.0
	for _, v := range s.values {
		sum += v
	}
	avg = sum / float64(len(s.values))

	// copy and sort for percentile
	cp := make([]float64, len(s.values))
	copy(cp, s.values)
	slowSort(cp)

	idx := int(math.Ceil(0.95*float64(len(cp)))) - 1
	if idx < 0 {
		idx = 0
	}
	p95 = cp[idx]
	return avg, p95
}

func slowSort(a []float64) {
	// simple insertion sort for small arrays
	for i := 1; i < len(a); i++ {
		j := i
		for j > 0 && a[j-1] > a[j] {
			a[j-1], a[j] = a[j], a[j-1]
			j--
		}
	}
}

func main() {
	logger := log.New(os.Stdout, "[module11-perf] ", log.LstdFlags|log.Lmicroseconds)
	cfg := loadConfig()

	c := newCache(2 * time.Second)

	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Without cache
	mux.HandleFunc("/data/raw", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if key == "" {
			key = "default"
		}
		value, delay := fetchFromSlowSource(key)
		writeJSON(w, http.StatusOK, map[string]any{
			"key":    key,
			"value":  value,
			"source": "slow",
			"delay":  delay.String(),
		})
	})

	// With cache
	mux.HandleFunc("/data/cached", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if key == "" {
			key = "default"
		}
		start := time.Now()
		if v, ok := c.get(key); ok {
			writeJSON(w, http.StatusOK, map[string]any{
				"key":    key,
				"value":  v,
				"source": "cache",
				"delay":  time.Since(start).String(),
			})
			return
		}
		value, _ := fetchFromSlowSource(key)
		c.set(key, value)
		writeJSON(w, http.StatusOK, map[string]any{
			"key":    key,
			"value":  value,
			"source": "slow",
			"delay":  time.Since(start).String(),
		})
	})

	// Simple built-in load generator
	mux.HandleFunc("/bench", func(w http.ResponseWriter, r *http.Request) {
		target := r.URL.Query().Get("target")
		if target != "raw" && target != "cached" {
			target = "raw"
		}
		nStr := r.URL.Query().Get("n")
		n, err := strconv.Atoi(nStr)
		if err != nil || n <= 0 {
			n = 100
		}
		client := &http.Client{Timeout: 2 * time.Second}
		stats := newLatencyStats()

		for i := 0; i < n; i++ {
			start := time.Now()
			resp, err := client.Get("http://localhost:" + strconv.Itoa(cfg.Port) + "/data/" + target + "?key=bench")
			if err != nil {
				continue
			}
			_ = resp.Body.Close()
			elapsed := time.Since(start)
			stats.add(float64(elapsed.Milliseconds()))
		}

		avg, p95 := stats.summary()
		hits, miss := c.stats()
		writeJSON(w, http.StatusOK, map[string]any{
			"target":       target,
			"requests":     n,
			"avg_ms":       avg,
			"p95_ms":       p95,
			"cache_hits":   hits,
			"cache_misses": miss,
		})
	})

	server := &http.Server{
		Addr:         ":" + strconv.Itoa(cfg.Port),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	logger.Printf("starting perf demo on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("server error: %v", err)
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

