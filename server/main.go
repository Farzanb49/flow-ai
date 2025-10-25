package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Deployment struct {
	ID          string    `json:"id"`
	Project     string    `json:"project"`
	Namespace   string    `json:"namespace"`
	Image       string    `json:"image"`
	Status      string    `json:"status"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
}

type Store struct {
	mu sync.RWMutex
	deployments map[string]Deployment
}

func main() {
	store := &Store{deployments: map[string]Deployment{}}
	r := chi.NewRouter()
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	r.Post("/deployments", func(w http.ResponseWriter, r *http.Request) {
		var d Deployment
		if err := json.NewDecoder(r.Body).Decode(&d); err != nil { http.Error(w, err.Error(), http.StatusBadRequest); return }
		if d.ID == "" { d.ID = uuid.NewString() }
		store.mu.Lock(); store.deployments[d.ID] = d; store.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(d)
	})
	r.Get("/deployments", func(w http.ResponseWriter, r *http.Request) {
		store.mu.RLock(); list := make([]Deployment, 0, len(store.deployments));
		for _, d := range store.deployments { list = append(list, d) }
		store.mu.RUnlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(list)
	})
	log.Println("API server on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
