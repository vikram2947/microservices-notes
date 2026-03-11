package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
}

type ServiceStatus string

const (
	StatusHealthy   ServiceStatus = "HEALTHY"
	StatusUnhealthy ServiceStatus = "UNHEALTHY"
)

type ServiceInstance struct {
	Name   string        `json:"name"`
	URL    string        `json:"url"`
	Status ServiceStatus `json:"status"`
}

type Registry struct {
	mu        sync.Mutex
	instances map[string][]ServiceInstance
}

func NewRegistry() *Registry {
	return &Registry{instances: make(map[string][]ServiceInstance)}
}

func (r *Registry) Register(svc ServiceInstance) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.instances[svc.Name] = append(r.instances[svc.Name], svc)
}

func (r *Registry) SetStatus(name, url string, status ServiceStatus) {
	r.mu.Lock()
	defer r.mu.Unlock()
	list := r.instances[name]
	for i := range list {
		if list[i].URL == url {
			list[i].Status = status
		}
	}
	r.instances[name] = list
}

func (r *Registry) HealthyInstance(name string) (ServiceInstance, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, inst := range r.instances[name] {
		if inst.Status == StatusHealthy {
			return inst, true
		}
	}
	return ServiceInstance{}, false
}

type Config struct {
	DiscoveryPort int `json:"discovery_port"`
	GatewayPort   int `json:"gateway_port"`
}

func loadConfig() (Config, error) {
	f, err := os.Open("config.json")
	if err != nil {
		return Config{}, err
	}
	defer f.Close()
	var cfg Config
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return Config{}, err
	}
	if v := os.Getenv("DISCOVERY_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			cfg.DiscoveryPort = p
		}
	}
	if v := os.Getenv("GATEWAY_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			cfg.GatewayPort = p
		}
	}
	if cfg.DiscoveryPort == 0 || cfg.GatewayPort == 0 {
		return Config{}, errors.New("invalid ports in config")
	}
	return cfg, nil
}

func main() {
	logger := log.New(os.Stdout, "[module6-discovery] ", log.LstdFlags|log.Lmicroseconds)

	cfg, err := loadConfig()
	if err != nil {
		logger.Fatalf("config error: %v", err)
	}

	registry := NewRegistry()

	// seed registry with static instances (conceptual: in real life they'd register themselves)
	registry.Register(ServiceInstance{Name: "users", URL: "http://localhost:8081", Status: StatusHealthy})
	registry.Register(ServiceInstance{Name: "orders", URL: "http://localhost:8082", Status: StatusHealthy})
	registry.Register(ServiceInstance{Name: "payments", URL: "http://localhost:8083", Status: StatusHealthy})

	// discovery HTTP API
	discoveryMux := http.NewServeMux()
	discoveryMux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	discoveryMux.HandleFunc("/discover/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		name := r.URL.Path[len("/discover/"):]
		if name == "" {
			http.Error(w, "missing service name", http.StatusBadRequest)
			return
		}
		inst, ok := registry.HealthyInstance(name)
		if !ok {
			http.Error(w, "no healthy instance", http.StatusNotFound)
			return
		}
		writeJSON(w, http.StatusOK, inst)
	})

	discoveryServer := &http.Server{
		Addr:         ":" + strconv.Itoa(cfg.DiscoveryPort),
		Handler:      discoveryMux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// gateway using discovery
	gatewayMux := http.NewServeMux()
	client := &http.Client{Timeout: 2 * time.Second}

	gatewayMux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// example: route /users/** via discovery
	gatewayMux.HandleFunc("/users/", func(w http.ResponseWriter, r *http.Request) {
		inst, ok := registry.HealthyInstance("users")
		if !ok {
			http.Error(w, "users service unavailable", http.StatusServiceUnavailable)
			return
		}
		target := inst.URL + r.URL.Path
		req, err := http.NewRequestWithContext(r.Context(), r.Method, target, r.Body)
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		req.Header = r.Header.Clone()
		resp, err := client.Do(req)
		if err != nil {
			registry.SetStatus("users", inst.URL, StatusUnhealthy)
			http.Error(w, "downstream error", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()
		w.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(w, resp.Body)
	})

	gatewayServer := &http.Server{
		Addr:         ":" + strconv.Itoa(cfg.GatewayPort),
		Handler:      gatewayMux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Printf("starting discovery on :%d", cfg.DiscoveryPort)
		if err := discoveryServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf("discovery error: %v", err)
		}
	}()

	go func() {
		logger.Printf("starting discovery-aware gateway on :%d", cfg.GatewayPort)
		if err := gatewayServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf("gateway error: %v", err)
		}
	}()

	<-ctx.Done()
	stop()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_ = discoveryServer.Shutdown(shutdownCtx)
	_ = gatewayServer.Shutdown(shutdownCtx)

	logger.Println("module6-discovery stopped")
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

