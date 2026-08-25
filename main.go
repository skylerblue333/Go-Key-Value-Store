package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	defaultMaxKeys  = 10000
	maxKeyLength    = 256
	maxValueLength  = 64 * 1024
	maxRequestBytes = 70 * 1024
)

var (
	errCapacity = errors.New("store capacity reached")
	errVersion  = errors.New("version conflict")
)

type Entry struct {
	Value     string `json:"value"`
	Version   uint64 `json:"version"`
	UpdatedAt string `json:"updated_at"`
}

type Database struct {
	mu      sync.RWMutex
	data    map[string]Entry
	maxKeys int
}

func NewDatabase(maxKeys int) *Database {
	if maxKeys <= 0 {
		maxKeys = defaultMaxKeys
	}
	return &Database{data: make(map[string]Entry), maxKeys: maxKeys}
}

func (db *Database) Set(key, value string, expectedVersion *uint64) (Entry, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	current, exists := db.data[key]
	if !exists && len(db.data) >= db.maxKeys {
		return Entry{}, errCapacity
	}
	if expectedVersion != nil {
		if !exists || current.Version != *expectedVersion {
			return Entry{}, errVersion
		}
	}
	version := uint64(1)
	if exists {
		version = current.Version + 1
	}
	entry := Entry{Value: value, Version: version, UpdatedAt: time.Now().UTC().Format(time.RFC3339Nano)}
	db.data[key] = entry
	return entry, nil
}

func (db *Database) Get(key string) (Entry, bool) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	entry, ok := db.data[key]
	return entry, ok
}

func (db *Database) Delete(key string, expectedVersion *uint64) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	entry, ok := db.data[key]
	if !ok {
		return os.ErrNotExist
	}
	if expectedVersion != nil && entry.Version != *expectedVersion {
		return errVersion
	}
	delete(db.data, key)
	return nil
}

func (db *Database) Len() int {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return len(db.data)
}

type server struct {
	db       *Database
	requests atomic.Uint64
	writes   atomic.Uint64
	reads    atomic.Uint64
}

type setRequest struct {
	Key             string  `json:"key"`
	Value           string  `json:"value"`
	ExpectedVersion *uint64 `json:"expected_version,omitempty"`
}

func normalizeKey(raw string) (string, error) {
	key := strings.TrimSpace(raw)
	if key == "" || len(key) > maxKeyLength {
		return "", fmt.Errorf("key must contain 1-%d characters", maxKeyLength)
	}
	return key, nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func (s *server) handleSet(w http.ResponseWriter, r *http.Request) {
	s.requests.Add(1)
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var req setRequest
	if err := decoder.Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}
	key, err := normalizeKey(req.Key)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if len(req.Value) > maxValueLength {
		writeJSON(w, http.StatusRequestEntityTooLarge, map[string]string{"error": "value exceeds 64 KiB"})
		return
	}
	entry, err := s.db.Set(key, req.Value, req.ExpectedVersion)
	if errors.Is(err, errCapacity) {
		writeJSON(w, http.StatusInsufficientStorage, map[string]string{"error": err.Error()})
		return
	}
	if errors.Is(err, errVersion) {
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}
	s.writes.Add(1)
	writeJSON(w, http.StatusCreated, map[string]any{"key": key, "entry": entry})
}

func (s *server) handleGet(w http.ResponseWriter, r *http.Request) {
	s.requests.Add(1)
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	key, err := normalizeKey(r.URL.Query().Get("key"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	entry, ok := s.db.Get(key)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	s.reads.Add(1)
	writeJSON(w, http.StatusOK, map[string]any{"key": key, "entry": entry})
}

func (s *server) handleDelete(w http.ResponseWriter, r *http.Request) {
	s.requests.Add(1)
	if r.Method != http.MethodDelete {
		w.Header().Set("Allow", http.MethodDelete)
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	key, err := normalizeKey(r.URL.Query().Get("key"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	var expected *uint64
	if raw := r.URL.Query().Get("version"); raw != "" {
		version, parseErr := strconv.ParseUint(raw, 10, 64)
		if parseErr != nil || version == 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "version must be a positive integer"})
			return
		}
		expected = &version
	}
	err = s.db.Delete(key, expected)
	if errors.Is(err, os.ErrNotExist) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	if errors.Is(err, errVersion) {
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}
	s.writes.Add(1)
	w.WriteHeader(http.StatusNoContent)
}

func (s *server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/kv", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			s.handleSet(w, r)
		case http.MethodGet:
			s.handleGet(w, r)
		case http.MethodDelete:
			s.handleDelete(w, r)
		default:
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		}
	})
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "sky-kv"})
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"status": "ready", "keys": s.db.Len()})
	})
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"requests_total": s.requests.Load(),
			"reads_total":    s.reads.Load(),
			"writes_total":   s.writes.Load(),
			"keys":           s.db.Len(),
			"max_keys":       s.db.maxKeys,
		})
	})
	return mux
}

func main() {
	maxKeys := defaultMaxKeys
	if raw := os.Getenv("MAX_KEYS"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 || parsed > 1_000_000 {
			log.Fatal("MAX_KEYS must be an integer between 1 and 1000000")
		}
		maxKeys = parsed
	}
	s := &server{db: NewDatabase(maxKeys)}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	httpServer := &http.Server{
		Addr:              ":" + port,
		Handler:           s.routes(),
		ReadHeaderTimeout: 3 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
	log.Printf("sky-kv listening on %s", httpServer.Addr)
	log.Fatal(httpServer.ListenAndServe())
}
