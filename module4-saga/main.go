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
	"sync"
	"time"
)

// Domain models

type OrderStatus string

const (
	StatusPending   OrderStatus = "PENDING"
	StatusConfirmed OrderStatus = "CONFIRMED"
	StatusCancelled OrderStatus = "CANCELLED"
)

type Order struct {
	ID       string      `json:"id"`
	UserID   string      `json:"user_id"`
	Amount   float64     `json:"amount"`
	Currency string      `json:"currency"`
	Status   OrderStatus `json:"status"`
}

// Events

type EventType string

const (
	EventOrderCreated     EventType = "OrderCreated"
	EventPaymentCompleted EventType = "PaymentCompleted"
)

type Event struct {
	ID        string    `json:"id"`
	Type      EventType `json:"type"`
	OrderID   string    `json:"order_id"`
	UserID    string    `json:"user_id,omitempty"`
	Amount    float64   `json:"amount,omitempty"`
	Currency  string    `json:"currency,omitempty"`
	Success   bool      `json:"success,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// Outbox entry

type OutboxEntry struct {
	ID        string
	Event     Event
	Published bool
}

// Stores

type OrderStore struct {
	mu     sync.Mutex
	orders map[string]Order
}

func NewOrderStore() *OrderStore {
	return &OrderStore{orders: make(map[string]Order)}
}

func (s *OrderStore) Save(o Order) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.orders[o.ID] = o
}

func (s *OrderStore) Get(id string) (Order, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	o, ok := s.orders[id]
	return o, ok
}

func (s *OrderStore) UpdateStatus(id string, status OrderStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if o, ok := s.orders[id]; ok {
		o.Status = status
		s.orders[id] = o
	}
}

type OutboxStore struct {
	mu   sync.Mutex
	data []OutboxEntry
}

func NewOutboxStore() *OutboxStore {
	return &OutboxStore{data: make([]OutboxEntry, 0)}
}

func (o *OutboxStore) Add(e Event) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.data = append(o.data, OutboxEntry{
		ID:        generateID("out_"),
		Event:     e,
		Published: false,
	})
}

func (o *OutboxStore) Unpublished() []OutboxEntry {
	o.mu.Lock()
	defer o.mu.Unlock()
	var res []OutboxEntry
	for _, e := range o.data {
		if !e.Published {
			res = append(res, e)
		}
	}
	return res
}

func (o *OutboxStore) MarkPublished(id string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	for i := range o.data {
		if o.data[i].ID == id {
			o.data[i].Published = true
			return
		}
	}
}

// Simple event bus

type EventBus struct {
	subscribers []chan Event
	mu          sync.Mutex
}

func NewEventBus() *EventBus {
	return &EventBus{subscribers: make([]chan Event, 0)}
}

func (b *EventBus) Subscribe() <-chan Event {
	b.mu.Lock()
	defer b.mu.Unlock()
	ch := make(chan Event, 10)
	b.subscribers = append(b.subscribers, ch)
	return ch
}

func (b *EventBus) Publish(e Event) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, ch := range b.subscribers {
		select {
		case ch <- e:
		default:
			// drop if subscriber is too slow for this demo
		}
	}
}

// Read model (CQRS projection)

type OrderSummary struct {
	ID       string      `json:"id"`
	Status   OrderStatus `json:"status"`
	Amount   float64     `json:"amount"`
	Currency string      `json:"currency"`
	Paid     bool        `json:"paid"`
}

type SummaryStore struct {
	mu   sync.Mutex
	data map[string]OrderSummary
}

func NewSummaryStore() *SummaryStore {
	return &SummaryStore{data: make(map[string]OrderSummary)}
}

func (s *SummaryStore) ApplyOrderCreated(o Order) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[o.ID] = OrderSummary{
		ID:       o.ID,
		Status:   o.Status,
		Amount:   o.Amount,
		Currency: o.Currency,
		Paid:     false,
	}
}

func (s *SummaryStore) ApplyOrderStatusUpdated(id string, status OrderStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if sum, ok := s.data[id]; ok {
		sum.Status = status
		s.data[id] = sum
	}
}

func (s *SummaryStore) ApplyPaymentCompleted(id string, success bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if sum, ok := s.data[id]; ok {
		if success {
			sum.Paid = true
			sum.Status = StatusConfirmed
		} else {
			sum.Status = StatusCancelled
		}
		s.data[id] = sum
	}
}

func (s *SummaryStore) Get(id string) (OrderSummary, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sum, ok := s.data[id]
	return sum, ok
}

// HTTP / application wiring

type ServerConfig struct {
	Port int
}

func loadConfig() (ServerConfig, error) {
	portStr := os.Getenv("APP_PORT")
	if portStr == "" {
		portStr = "8100"
	}
	port, err := strconv.Atoi(portStr)
	if err != nil || port <= 0 || port > 65535 {
		return ServerConfig{}, errors.New("invalid APP_PORT")
	}
	return ServerConfig{Port: port}, nil
}

type createOrderRequest struct {
	UserID   string  `json:"user_id"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

func main() {
	logger := log.New(os.Stdout, "[module4-saga] ", log.LstdFlags|log.Lmicroseconds)

	cfg, err := loadConfig()
	if err != nil {
		logger.Fatalf("config error: %v", err)
	}

	orderStore := NewOrderStore()
	outbox := NewOutboxStore()
	eventBus := NewEventBus()
	summaryStore := NewSummaryStore()

	// Outbox publisher
	go func() {
		for {
			time.Sleep(200 * time.Millisecond)
			entries := outbox.Unpublished()
			for _, e := range entries {
				eventBus.Publish(e.Event)
				outbox.MarkPublished(e.ID)
				logger.Printf("published event %s type=%s order=%s", e.ID, e.Event.Type, e.Event.OrderID)
			}
		}
	}()

	// Payment service (simulated) subscribes to events
	paymentEvents := eventBus.Subscribe()
	go paymentService(paymentEvents, eventBus, logger)

	// Saga orchestrator subscribes to events
	sagaEvents := eventBus.Subscribe()
	go sagaOrchestrator(sagaEvents, orderStore, summaryStore, logger)

	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Command: create order
	mux.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req createOrderRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(req.UserID) == "" || req.Amount <= 0 || strings.TrimSpace(req.Currency) == "" {
			http.Error(w, "user_id, amount (>0), currency required", http.StatusBadRequest)
			return
		}
		id := generateID("ord_")
		order := Order{
			ID:       id,
			UserID:   req.UserID,
			Amount:   req.Amount,
			Currency: req.Currency,
			Status:   StatusPending,
		}

		// Local transaction: write order + outbox event
		orderStore.Save(order)
		summaryStore.ApplyOrderCreated(order)

		ev := Event{
			ID:        generateID("evt_"),
			Type:      EventOrderCreated,
			OrderID:   order.ID,
			UserID:    order.UserID,
			Amount:    order.Amount,
			Currency:  order.Currency,
			CreatedAt: time.Now().UTC(),
		}
		outbox.Add(ev)

		writeJSON(w, http.StatusAccepted, order)
	})

	// Query: order (write model)
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
		o, ok := orderStore.Get(id)
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		writeJSON(w, http.StatusOK, o)
	})

	// Query: summary (read model)
	mux.HandleFunc("/summaries/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		id := strings.TrimPrefix(r.URL.Path, "/summaries/")
		if id == "" {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}
		s, ok := summaryStore.Get(id)
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		writeJSON(w, http.StatusOK, s)
	})

	server := &http.Server{
		Addr:         ":" + strconv.Itoa(cfg.Port),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	logger.Printf("starting saga demo service on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Fatalf("server error: %v", err)
	}
}

// paymentService simulates a payment microservice consuming OrderCreated events
func paymentService(events <-chan Event, bus *EventBus, logger *log.Logger) {
	for ev := range events {
		if ev.Type != EventOrderCreated {
			continue
		}
		// simulate processing time and random outcome
		time.Sleep(300 * time.Millisecond)
		success := rand.IntN(10) < 8 // 80% success

		logger.Printf("payment processed for order=%s success=%v", ev.OrderID, success)

		outcome := Event{
			ID:        generateID("evt_"),
			Type:      EventPaymentCompleted,
			OrderID:   ev.OrderID,
			Success:   success,
			CreatedAt: time.Now().UTC(),
		}
		bus.Publish(outcome)
	}
}

// sagaOrchestrator coordinates order status based on PaymentCompleted events
func sagaOrchestrator(events <-chan Event, orders *OrderStore, summaries *SummaryStore, logger *log.Logger) {
	for ev := range events {
		switch ev.Type {
		case EventPaymentCompleted:
			if ev.Success {
				orders.UpdateStatus(ev.OrderID, StatusConfirmed)
				summaries.ApplyPaymentCompleted(ev.OrderID, true)
				logger.Printf("order %s confirmed", ev.OrderID)
			} else {
				orders.UpdateStatus(ev.OrderID, StatusCancelled)
				summaries.ApplyPaymentCompleted(ev.OrderID, false)
				logger.Printf("order %s cancelled", ev.OrderID)
			}
		}
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func generateID(prefix string) string {
	return prefix + strconv.FormatInt(time.Now().UnixNano(), 36)
}

