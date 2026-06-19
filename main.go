package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

type Store struct {
	mu   sync.RWMutex
	data map[string]string
}

func NewStore() *Store {
	return &Store{
		data: make(map[string]string),
	}
}

func (s *Store) Set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

func (s *Store) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.data[key]
	return val, ok
}

func (s *Store) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
}

type Request struct {
	Key   string `json:"key"`
	Value string `json:"value,omitempty"`
}

func (s *Store) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		key := r.URL.Query().Get("key")
		if val, ok := s.Get(key); ok {
			json.NewEncoder(w).Encode(map[string]string{"key": key, "value": val})
			return
		}
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	}

	if r.Method == http.MethodPost {
		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		s.Set(req.Key, req.Value)
		w.WriteHeader(http.StatusCreated)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func main() {
	store := NewStore()
	http.Handle("/kv", store)
	log.Println("KV Store listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
