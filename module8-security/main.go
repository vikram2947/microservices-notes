package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type ServerConfig struct {
	AuthPort    int
	GatewayPort int
	BackendPort int
	SecretKey   []byte
}

type tokenClaims struct {
	UserID string   `json:"user_id"`
	Roles  []string `json:"roles"`
	Exp    int64    `json:"exp"`
}

func loadConfig() (ServerConfig, error) {
	secret := os.Getenv("AUTH_SECRET")
	if secret == "" {
		secret = "dev-secret-change-me"
	}
	authPort := getenvInt("AUTH_PORT", 8600)
	gatewayPort := getenvInt("GATEWAY_PORT", 8601)
	backendPort := getenvInt("BACKEND_PORT", 8602)
	if authPort <= 0 || gatewayPort <= 0 || backendPort <= 0 {
		return ServerConfig{}, errors.New("invalid ports")
	}
	return ServerConfig{
		AuthPort:    authPort,
		GatewayPort: gatewayPort,
		BackendPort: backendPort,
		SecretKey:   []byte(secret),
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

// very simple HMAC-SHA256 "JWT-like" token: base64(json).base64(sig)

func signToken(secret []byte, claims tokenClaims) (string, error) {
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	payloadB64 := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, secret)
	mac.Write(payload)
	sig := mac.Sum(nil)
	sigB64 := base64.RawURLEncoding.EncodeToString(sig)
	return payloadB64 + "." + sigB64, nil
}

func parseToken(secret []byte, token string) (tokenClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return tokenClaims{}, errors.New("invalid token format")
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return tokenClaims{}, errors.New("invalid payload encoding")
	}
	sigBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return tokenClaims{}, errors.New("invalid signature encoding")
	}
	mac := hmac.New(sha256.New, secret)
	mac.Write(payloadBytes)
	expected := mac.Sum(nil)
	if !hmac.Equal(sigBytes, expected) {
		return tokenClaims{}, errors.New("invalid signature")
	}
	var claims tokenClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return tokenClaims{}, errors.New("invalid payload")
	}
	if time.Now().Unix() > claims.Exp {
		return tokenClaims{}, errors.New("token expired")
	}
	return claims, nil
}

func main() {
	logger := log.New(os.Stdout, "[module8-security] ", log.LstdFlags|log.Lmicroseconds)

	cfg, err := loadConfig()
	if err != nil {
		logger.Fatalf("config error: %v", err)
	}

	// Auth service
	authMux := http.NewServeMux()
	authMux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	authMux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var body struct {
			UserID string   `json:"user_id"`
			Roles  []string `json:"roles"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(body.UserID) == "" {
			http.Error(w, "user_id required", http.StatusBadRequest)
			return
		}
		if len(body.Roles) == 0 {
			body.Roles = []string{"user"}
		}
		claims := tokenClaims{
			UserID: body.UserID,
			Roles:  body.Roles,
			Exp:    time.Now().Add(30 * time.Minute).Unix(),
		}
		tok, err := signToken(cfg.SecretKey, claims)
		if err != nil {
			http.Error(w, "failed to issue token", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"token": tok})
	})

	authSrv := &http.Server{
		Addr:    ":" + strconv.Itoa(cfg.AuthPort),
		Handler: authMux,
	}

	// Backend service
	backendMux := http.NewServeMux()
	backendMux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	backendMux.HandleFunc("/profile", func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("X-User-ID")
		if userID == "" {
			http.Error(w, "missing user", http.StatusUnauthorized)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"user_id": userID,
			"message": "this is a user profile",
		})
	})
	backendMux.HandleFunc("/admin", func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("X-User-ID")
		roles := strings.Split(r.Header.Get("X-User-Roles"), ",")
		if userID == "" {
			http.Error(w, "missing user", http.StatusUnauthorized)
			return
		}
		if !hasRole(roles, "admin") {
			http.Error(w, "forbidden: admin role required", http.StatusForbidden)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"user_id": userID,
			"message": "welcome, admin",
		})
	})

	backendSrv := &http.Server{
		Addr:    ":" + strconv.Itoa(cfg.BackendPort),
		Handler: backendMux,
	}

	// Gateway with token validation
	gatewayMux := http.NewServeMux()
	gatewayMux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	gatewayMux.Handle("/profile", authMiddlewareHandler(cfg.SecretKey, "user", "user", "admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		forwardToBackend(w, r, cfg.BackendPort, "/profile")
	})))

	gatewayMux.Handle("/admin", authMiddlewareHandler(cfg.SecretKey, "admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		forwardToBackend(w, r, cfg.BackendPort, "/admin")
	})))

	gatewaySrv := &http.Server{
		Addr:    ":" + strconv.Itoa(cfg.GatewayPort),
		Handler: gatewayMux,
	}

	go func() {
		log.Printf("auth service on :%d", cfg.AuthPort)
		if err := authSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("auth error: %v", err)
		}
	}()
	go func() {
		log.Printf("backend service on :%d", cfg.BackendPort)
		if err := backendSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("backend error: %v", err)
		}
	}()
	log.Printf("gateway on :%d", cfg.GatewayPort)
	if err := gatewaySrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("gateway error: %v", err)
	}
}

func authMiddlewareHandler(secret []byte, requiredRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "missing bearer token", http.StatusUnauthorized)
				return
			}
			token := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := parseToken(secret, token)
			if err != nil {
				http.Error(w, "invalid token: "+err.Error(), http.StatusUnauthorized)
				return
			}
			if len(requiredRoles) > 0 && !anyRole(claims.Roles, requiredRoles) {
				http.Error(w, "forbidden: missing required role", http.StatusForbidden)
				return
			}
			r.Header.Set("X-User-ID", claims.UserID)
			r.Header.Set("X-User-Roles", strings.Join(claims.Roles, ","))
			next.ServeHTTP(w, r)
		})
	}
}

func anyRole(userRoles []string, required []string) bool {
	for _, r := range required {
		if hasRole(userRoles, r) {
			return true
		}
	}
	return false
}

func hasRole(roles []string, role string) bool {
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

func forwardToBackend(w http.ResponseWriter, r *http.Request, backendPort int, path string) {
	req, err := http.NewRequestWithContext(r.Context(), r.Method, "http://localhost:"+strconv.Itoa(backendPort)+path, nil)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	req.Header = r.Header.Clone()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "backend error", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	for k, v := range resp.Header {
		for _, vv := range v {
			w.Header().Add(k, vv)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

