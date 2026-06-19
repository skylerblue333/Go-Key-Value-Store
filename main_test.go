package main

import "testing"

func TestStore(t *testing.T) {
	store := NewStore()
	
	store.Set("key1", "val1")
	val, ok := store.Get("key1")
	
	if !ok || val != "val1" {
		t.Errorf("Expected val1, got %v", val)
	}
	
	store.Delete("key1")
	_, ok = store.Get("key1")
	if ok {
		t.Error("Expected key to be deleted")
	}
}
