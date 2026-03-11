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

type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type createUserRequest struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type ServerConfig struct {
	Port int
}

type UserStore struct {
	data map[string]User
}

func NewUserStore() *UserStore {
	return &UserStore{data: make(map[string]User)}
}

func (s *UserStore) Create(u User) User {
	s.data[u.ID] = u
	return u
}

func (s *UserStore) Get(id string) (User, bool) {
	u, ok := s.data[id]
	return u, ok
}

func loadConfig() (ServerConfig, error) {
	portStr := os.Getenv("APP_PORT")
	if portStr == "" {
		portStr = "8081"
	}
	port, err := strconv.Atoi(portStr)
	if err != nil || port <= 0 || port > 65535 {
		return ServerConfig{}, errors.New("invalid APP_PORT")
	}
	return ServerConfig{Port: port}, nil
}

func main() {
	logger := log.New(os.Stdout, "[module2-users] ", log.LstdFlags|log.Lmicroseconds)

	cfg, err := loadConfig()
	if err != nil {
		logger.Fatalf("config error: %v", err)
	}

	store := NewUserStore()

	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			var req createUserRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "invalid json", http.StatusBadRequest)
				return
			}
			if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Name) == "" {
				http.Error(w, "email and name required", http.StatusBadRequest)
				return
			}
			id := generateID("usr_")
			user := User{
				ID:    id,
				Email: req.Email,
				Name:  req.Name,
			}
			store.Create(user)
			writeJSON(w, http.StatusCreated, user)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/users/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		id := strings.TrimPrefix(r.URL.Path, "/users/")
		if id == "" {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}
		u, ok := store.Get(id)
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		writeJSON(w, http.StatusOK, u)
	})

	server := &http.Server{
		Addr:         ":" + strconv.Itoa(cfg.Port),
		Handler:      loggingMiddleware(logger, mux),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	logger.Printf("starting users service on %s", server.Addr)
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

