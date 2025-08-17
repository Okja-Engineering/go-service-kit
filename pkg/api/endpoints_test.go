package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestAddOKEndpoint(t *testing.T) {
	base := NewBase("TestService", "1.0.0", "test-build", true)
	router := chi.NewRouter()

	base.AddOKEndpoint(router, "ok")

	req := httptest.NewRequest("GET", "/ok", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "OK" {
		t.Errorf("Expected body 'OK', got '%s'", w.Body.String())
	}
}

func TestAddHealthEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		healthy  bool
		expected int
		body     string
	}{
		{
			name:     "healthy service",
			healthy:  true,
			expected: http.StatusOK,
			body:     "OK: Service is healthy",
		},
		{
			name:     "unhealthy service",
			healthy:  false,
			expected: http.StatusServiceUnavailable,
			body:     "Error: Service is not healthy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := NewBase("TestService", "1.0.0", "test-build", tt.healthy)
			router := chi.NewRouter()

			base.AddHealthEndpoint(router, "health")

			req := httptest.NewRequest("GET", "/health", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expected {
				t.Errorf("Expected status %d, got %d", tt.expected, w.Code)
			}

			if w.Body.String() != tt.body {
				t.Errorf("Expected body '%s', got '%s'", tt.body, w.Body.String())
			}
		})
	}
}

func TestAddStatusEndpoint(t *testing.T) {
	base := NewBase("TestService", "1.0.0", "test-build", true)
	router := chi.NewRouter()

	base.AddStatusEndpoint(router, "status")

	req := httptest.NewRequest("GET", "/status", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var status Status
	if err := json.Unmarshal(w.Body.Bytes(), &status); err != nil {
		t.Fatalf("Failed to unmarshal status response: %v", err)
	}

	// Test basic fields
	testBasicFields(t, status)

	// Test system fields
	testSystemFields(t, status)
}

func testBasicFields(t *testing.T, status Status) {
	if status.Service != "TestService" {
		t.Errorf("Expected service 'TestService', got '%s'", status.Service)
	}

	if status.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", status.Version)
	}

	if status.BuildInfo != "test-build" {
		t.Errorf("Expected buildInfo 'test-build', got '%s'", status.BuildInfo)
	}

	if !status.Healthy {
		t.Error("Expected healthy to be true")
	}
}

func testSystemFields(t *testing.T, status Status) {
	if status.Hostname == "" {
		t.Error("Expected hostname to be set")
	}

	if status.OS == "" {
		t.Error("Expected OS to be set")
	}

	if status.Architecture == "" {
		t.Error("Expected architecture to be set")
	}

	if status.CPUCount <= 0 {
		t.Error("Expected CPU count to be greater than 0")
	}

	if status.GoVersion == "" {
		t.Error("Expected Go version to be set")
	}

	if status.ClientAddr == "" {
		t.Error("Expected client address to be set")
	}

	if status.ServerHost == "" {
		t.Error("Expected server host to be set")
	}

	if status.Uptime == "" {
		t.Error("Expected uptime to be set")
	}
}

func TestAddMetricsEndpoint(t *testing.T) {
	base := NewBase("TestService", "1.0.0", "test-build", true)
	router := chi.NewRouter()

	base.AddMetricsEndpoint(router, "metrics")

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Metrics endpoint should return Prometheus metrics
	body := w.Body.String()
	if body == "" {
		t.Error("Expected metrics response to have content")
	}

	// Should contain some basic Prometheus metrics
	if len(body) < 100 {
		t.Error("Expected metrics response to be substantial")
	}
}
