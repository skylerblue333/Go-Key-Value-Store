package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSetAndGet(t *testing.T) {
	db = NewDatabase() // reset global db for tests
	
	// Test Set
	setReqBody := []byte(`{"key":"test","value":"data"}`)
	req, _ := http.NewRequest("POST", "/set", bytes.NewBuffer(setReqBody))
	rr := httptest.NewRecorder()
	handleSet(rr, req)
	if rr.Code != http.StatusCreated {
		t.Errorf("Expected 201 Created, got %d", rr.Code)
	}

	// Test Get
	req, _ = http.NewRequest("GET", "/get?key=test", nil)
	rr = httptest.NewRecorder()
	handleGet(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", rr.Code)
	}

	var resp struct {
		Value string `json:"value"`
	}
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Value != "data" {
		t.Errorf("Expected 'data', got '%s'", resp.Value)
	}
}
