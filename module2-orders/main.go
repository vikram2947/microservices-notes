package main

import (
	"encoding/json"
	"errors"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Order struct {
	ID        string  `json:"id"`
	UserID    string  `json:"user_id"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	CreatedAt string  `json:"created_at"`
}

type createOrderRequest struct {
	UserID   string  `json:"user_id"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

type ServerConfig struct {
	Port int
}

type OrderStore struct {
	data map[string]Order
}

func NewOrderStore() *OrderStore {
	return &OrderStore{data: make(map[string]Order)}
}

func (s *OrderStore) Create(o Order) Order {
	s.data[o.ID] = o
	return o
}

func (s *OrderStore) Get(id string) (Order, bool) {
	o, ok := s.data[id]
	return o, ok
}

func loadConfig() (ServerConfig, error) {
	portStr := os.Getenv("APP_PORT")
	if portStr == "" {
		portStr = "8082"
	}
	port, err := strconv.Atoi(portStr)
	if err != nil || port <= 0 || port > 65535 {
		return ServerConfig{}, errors.New("invalid APP_PORT")
	}
	return ServerConfig{Port: port}, nil
}

func main() {
	logger := log.New(os.Stdout, "[module2-orders] ", log.LstdFlags|log.Lmicroseconds)

	cfg, err := loadConfig()
	if err != nil {
		logger.Fatalf("config error: %v", err)
	}

	store := NewOrderStore()

	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			var req createOrderRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "invalid json", http.StatusBadRequest)
				return
			}
			if strings.TrimSpace(req.UserID) == "" || req.Amount <= 0 || strings.TrimSpace(req.Currency) == "" {
				http.Error(w, "user_id, amount (>0) and currency required", http.StatusBadRequest)
				return
			}
			id := generateID("ord_")
			order := Order{
				ID:        id,
				UserID:    req.UserID,
				Amount:    req.Amount,
				Currency:  req.Currency,
				CreatedAt: time.Now().UTC().Format(time.RFC3339),
			}
			store.Create(order)
			writeJSON(w, http.StatusCreated, order)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/orders/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		id := strings.TrimPrefix(r.URL.Path, "/orders/")
		if id == "" {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}
		o, ok := store.Get(id)
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		writeJSON(w, http.StatusOK, o)
	})

	server := &http.Server{
		Addr:         ":" + strconv.Itoa(cfg.Port),
		Handler:      loggingMiddleware(logger, mux),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	logger.Printf("starting orders service on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Fatalf("server error: %v", err)
	}
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

func generateID(prefix string) string {
	return prefix + strconv.FormatInt(time.Now().UnixNano(), 36) + strconv.FormatUint(rand.Uint64(), 36)
}

