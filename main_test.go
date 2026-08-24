package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestDatabaseVersionAndCapacity(t *testing.T) {
	db := NewDatabase(1)
	first, err := db.Set("alpha", "one", nil)
	if err != nil || first.Version != 1 {
		t.Fatalf("first set = %#v, %v", first, err)
	}
	wrong := uint64(99)
	if _, err := db.Set("alpha", "two", &wrong); err != errVersion {
		t.Fatalf("expected version conflict, got %v", err)
	}
	expected := uint64(1)
	second, err := db.Set("alpha", "two", &expected)
	if err != nil || second.Version != 2 || second.Value != "two" {
		t.Fatalf("second set = %#v, %v", second, err)
	}
	if _, err := db.Set("beta", "blocked", nil); err != errCapacity {
		t.Fatalf("expected capacity error, got %v", err)
	}
}

func TestDatabaseConcurrentWrites(t *testing.T) {
	db := NewDatabase(100)
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := db.Set("shared", "value", nil); err != nil {
				t.Errorf("set failed: %v", err)
			}
		}()
	}
	wg.Wait()
	entry, ok := db.Get("shared")
	if !ok || entry.Version != 50 {
		t.Fatalf("entry = %#v, ok=%v", entry, ok)
	}
}

func requestJSON(t *testing.T, handler http.Handler, method, target string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var data []byte
	if body != nil {
		var err error
		data, err = json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
	}
	req := httptest.NewRequest(method, target, bytes.NewReader(data))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	return res
}

func TestHTTPContract(t *testing.T) {
	s := &server{db: NewDatabase(2)}
	h := s.routes()

	created := requestJSON(t, h, http.MethodPost, "/v1/kv", map[string]any{"key": " alpha ", "value": "one"})
	if created.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", created.Code, created.Body.String())
	}

	read := requestJSON(t, h, http.MethodGet, "/v1/kv?key=alpha", nil)
	if read.Code != http.StatusOK || !bytes.Contains(read.Body.Bytes(), []byte(`"version":1`)) {
		t.Fatalf("read status=%d body=%s", read.Code, read.Body.String())
	}

	conflict := requestJSON(t, h, http.MethodPost, "/v1/kv", map[string]any{"key": "alpha", "value": "two", "expected_version": 99})
	if conflict.Code != http.StatusConflict {
		t.Fatalf("conflict status=%d body=%s", conflict.Code, conflict.Body.String())
	}

	deleted := requestJSON(t, h, http.MethodDelete, "/v1/kv?key=alpha&version=1", nil)
	if deleted.Code != http.StatusNoContent {
		t.Fatalf("delete status=%d body=%s", deleted.Code, deleted.Body.String())
	}

	health := requestJSON(t, h, http.MethodGet, "/healthz", nil)
	if health.Code != http.StatusOK {
		t.Fatalf("health status=%d", health.Code)
	}
}
