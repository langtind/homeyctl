package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUpdateDevice(t *testing.T) {
	// Create a test server
	var receivedPath string
	var receivedBody map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&receivedBody); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	// Create client with test server
	client := &Client{
		baseURL:    server.URL,
		token:      "test-token",
		httpClient: server.Client(),
	}

	// Test UpdateDevice
	deviceID := "test-device-123"
	updates := map[string]interface{}{
		"name": "New Device Name",
	}

	err := client.UpdateDevice(deviceID, updates)
	if err != nil {
		t.Fatalf("UpdateDevice failed: %v", err)
	}

	// Verify the request
	expectedPath := "/api/manager/devices/device/test-device-123"
	if receivedPath != expectedPath {
		t.Errorf("expected path %s, got %s", expectedPath, receivedPath)
	}

	if receivedBody["name"] != "New Device Name" {
		t.Errorf("expected name 'New Device Name', got %v", receivedBody["name"])
	}
}

func TestUpdateDevice_Error(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "device not found"}`))
	}))
	defer server.Close()

	client := &Client{
		baseURL:    server.URL,
		token:      "test-token",
		httpClient: server.Client(),
	}

	err := client.UpdateDevice("nonexistent", map[string]interface{}{"name": "Test"})
	if err == nil {
		t.Error("expected error for nonexistent device")
	}
}

func TestUpdateZone(t *testing.T) {
	var receivedPath string
	var receivedBody map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&receivedBody); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := &Client{
		baseURL:    server.URL,
		token:      "test-token",
		httpClient: server.Client(),
	}

	zoneID := "test-zone-123"
	updates := map[string]interface{}{
		"name": "New Zone Name",
	}

	err := client.UpdateZone(zoneID, updates)
	if err != nil {
		t.Fatalf("UpdateZone failed: %v", err)
	}

	expectedPath := "/api/manager/zones/zone/test-zone-123"
	if receivedPath != expectedPath {
		t.Errorf("expected path %s, got %s", expectedPath, receivedPath)
	}

	if receivedBody["name"] != "New Zone Name" {
		t.Errorf("expected name 'New Zone Name', got %v", receivedBody["name"])
	}
}

func TestUpdateZone_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "zone not found"}`))
	}))
	defer server.Close()

	client := &Client{
		baseURL:    server.URL,
		token:      "test-token",
		httpClient: server.Client(),
	}

	err := client.UpdateZone("nonexistent", map[string]interface{}{"name": "Test"})
	if err == nil {
		t.Error("expected error for nonexistent zone")
	}
}
