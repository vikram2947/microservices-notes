package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type ServerConfig struct {
	Port          int
	UsersBaseURL  string
	OrdersBaseURL string
	PayBaseURL    string
}

func loadConfig() (ServerConfig, error) {
	portStr := os.Getenv("APP_PORT")
	if portStr == "" {
		portStr = "8200"
	}
	port, err := strconv.Atoi(portStr)
	if err != nil || port <= 0 || port > 65535 {
		return ServerConfig{}, errors.New("invalid APP_PORT")
	}

	usersURL := os.Getenv("USERS_BASE_URL")
	if usersURL == "" {
		usersURL = "http://localhost:8081"
	}
	ordersURL := os.Getenv("ORDERS_BASE_URL")
	if ordersURL == "" {
		ordersURL = "http://localhost:8082"
	}
	payURL := os.Getenv("PAYMENTS_BASE_URL")
	if payURL == "" {
		payURL = "http://localhost:8083"
	}

	return ServerConfig{
		Port:          port,
		UsersBaseURL:  usersURL,
		OrdersBaseURL: ordersURL,
		PayBaseURL:    payURL,
	}, nil
}

func main() {
	logger := log.New(os.Stdout, "[module5-gateway] ", log.LstdFlags|log.Lmicroseconds)

	cfg, err := loadConfig()
	if err != nil {
		logger.Fatalf("config error: %v", err)
	}

	client := &http.Client{
		Timeout: 2 * time.Second,
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

	usersProxy := newReverseProxy(cfg.UsersBaseURL, client)
	ordersProxy := newReverseProxy(cfg.OrdersBaseURL, client)
	payProxy := newReverseProxy(cfg.PayBaseURL, client)

	bucket := newTokenBucket(20, 500*time.Millisecond)

	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Simple routing
	mux.Handle("/users", usersProxy)
	mux.Handle("/users/", usersProxy)

	mux.Handle("/orders", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ordersProxy.ServeHTTP(w, r)
	})))
	mux.Handle("/orders/", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ordersProxy.ServeHTTP(w, r)
	})))

	mux.Handle("/payments", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payProxy.ServeHTTP(w, r)
	})))
	mux.Handle("/payments/", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payProxy.ServeHTTP(w, r)
	})))

	// BFF-style aggregated endpoint
	mux.Handle("/me/overview", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("X-User-ID")
		ctx := r.Context()

		type user struct {
			ID    string `json:"id"`
			Email string `json:"email"`
			Name  string `json:"name"`
		}
		type order struct {
			ID       string  `json:"id"`
			UserID   string  `json:"user_id"`
			Amount   float64 `json:"amount"`
			Currency string  `json:"currency"`
		}

		// Call users service
		var u user
		if err := fetchJSON(ctx, client, cfg.UsersBaseURL+"/users/"+userID, &u); err != nil {
			http.Error(w, "failed to fetch user: "+err.Error(), http.StatusBadGateway)
			return
		}

		// For simplicity, we don't have "list orders" endpoint in module2.
		// We'll simulate "recent orders" by just explaining the idea.
		overview := map[string]any{
			"user": u,
			"orders_info": map[string]string{
				"note": "in a real system we'd call orders service for this user and return the list here",
			},
		}

		writeJSON(w, http.StatusOK, overview)
	})))

	handler := loggingMiddleware(logger, rateLimitMiddleware(bucket, mux))

	server := &http.Server{
		Addr:         ":" + strconv.Itoa(cfg.Port),
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	logger.Printf("starting gateway on %s (users=%s orders=%s payments=%s)",
		server.Addr, cfg.UsersBaseURL, cfg.OrdersBaseURL, cfg.PayBaseURL)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Fatalf("server error: %v", err)
	}
}

func newReverseProxy(target string, client *http.Client) *httputil.ReverseProxy {
	url, err := url.Parse(target)
	if err != nil {
		panic(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(url)
	proxy.Transport = client.Transport

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		// Preserve original path and query
		// NewSingleHostReverseProxy already maps scheme/host; we keep path as is.
	}
	return proxy
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := strings.TrimSpace(r.Header.Get("X-User-ID"))
		if userID == "" {
			http.Error(w, "missing X-User-ID header", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
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

func fetchJSON(ctx context.Context, client *http.Client, url string, dest any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
		return errors.New(resp.Status + ": " + string(body))
	}
	return json.NewDecoder(resp.Body).Decode(dest)
}

