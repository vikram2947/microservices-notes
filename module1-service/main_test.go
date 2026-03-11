package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestLoadConfig_DefaultPort(t *testing.T) {
	t.Setenv("APP_PORT", "")
	cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.Port != 8080 {
		t.Fatalf("expected default port 8080, got %d", cfg.Port)
	}
}

func TestLoadConfig_InvalidPort(t *testing.T) {
	t.Setenv("APP_PORT", "not-a-number")
	_, err := loadConfig()
	if err == nil {
		t.Fatalf("expected error for invalid port")
	}
}

func TestHealthzHandler_Contract(t *testing.T) {
	logger := newTestLogger()
	cfg := ServerConfig{Port: 8080}
	app := newAppServer(cfg, logger)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()

	app.httpServer.Handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json response: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("expected status=ok, got %v", body["status"])
	}
}

func newTestLogger() *log.Logger {
	return log.New(os.Stdout, "[test] ", 0)
}

