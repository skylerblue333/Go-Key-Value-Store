package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

// KVStore is a thread-safe in-memory key-value store
type KVStore struct {
	mu   sync.RWMutex
	data map[string]string
}

func NewKVStore() *KVStore {
	return &KVStore{data: make(map[string]string)}
}

func (kv *KVStore) Set(key, value string) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	kv.data[key] = value
}

func (kv *KVStore) Get(key string) (string, bool) {
	kv.mu.RLock()
	defer kv.mu.RUnlock()
	val, ok := kv.data[key]
	return val, ok
}

func (kv *KVStore) Delete(key string) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	delete(kv.data, key)
}

var store = NewKVStore()

func handleSet(w http.ResponseWriter, r *http.Request) {
	var body map[string]string
	json.NewDecoder(r.Body).Decode(&body)
	store.Set(body["key"], body["value"])
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleGet(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	val, ok := store.Get(key)
	if !ok {
		http.Error(w, "key not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"key": key, "value": val})
}

func main() {
	http.HandleFunc("/set", handleSet)
	http.HandleFunc("/get", handleGet)
	log.Println("KV Store listening on :6379")
	log.Fatal(http.ListenAndServe(":6379", nil))
}
