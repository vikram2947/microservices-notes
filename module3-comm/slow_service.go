package main

import (
	"encoding/json"
	"math/rand/v2"
	"net/http"
	"time"
)

type slowService struct{}

func newSlowService() *slowService {
	return &slowService{}
}

func (s *slowService) handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("/work", func(w http.ResponseWriter, r *http.Request) {
		// Simulate variable latency and failure
		delay := time.Duration(100+rand.IntN(900)) * time.Millisecond
		time.Sleep(delay)

		switch x := rand.IntN(10); {
		case x < 2:
			// 20% failure
			http.Error(w, "temporary error", http.StatusInternalServerError)
			return
		default:
			writeJSON(w, http.StatusOK, map[string]any{
				"status": "ok",
				"delay":  delay.String(),
			})
		}
	})

	return mux
}

