package logging

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestNewFilteredRequestLogger(t *testing.T) {
	// Test with a regex that filters out health endpoints
	filterRegex := regexp.MustCompile(`/health`)
	logger := NewFilteredRequestLogger(filterRegex)

	if logger == nil {
		t.Error("Expected logger to not be nil")
	}
}

func TestFilteredRequestLoggerWithFilteredURL(t *testing.T) {
	// Create a regex that filters out health endpoints
	filterRegex := regexp.MustCompile(`/health`)
	logger := NewFilteredRequestLogger(filterRegex)

	// Create a test router
	router := chi.NewRouter()

	// Add the filtered logger middleware
	router.Use(logger)

	// Add a test endpoint
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("healthy")); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	})

	// Test request to filtered URL
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should still work normally, just not log
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "healthy" {
		t.Errorf("Expected body 'healthy', got '%s'", w.Body.String())
	}
}

func TestFilteredRequestLoggerWithNonFilteredURL(t *testing.T) {
	// Create a regex that filters out health endpoints
	filterRegex := regexp.MustCompile(`/health`)
	logger := NewFilteredRequestLogger(filterRegex)

	// Create a test router
	router := chi.NewRouter()

	// Add the filtered logger middleware
	router.Use(logger)

	// Add a test endpoint
	router.Get("/api/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("test")); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	})

	// Test request to non-filtered URL
	req := httptest.NewRequest("GET", "/api/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should work normally and log
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "test" {
		t.Errorf("Expected body 'test', got '%s'", w.Body.String())
	}
}

func TestFilteredRequestLoggerWithMultipleFilters(t *testing.T) {
	// Create a regex that filters out health and metrics endpoints
	filterRegex := regexp.MustCompile(`/(health|metrics)`)
	logger := NewFilteredRequestLogger(filterRegex)

	// Create a test router
	router := chi.NewRouter()

	// Add the filtered logger middleware
	router.Use(logger)

	// Add test endpoints
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("healthy")); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	})

	router.Get("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("metrics")); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	})

	router.Get("/api/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("test")); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	})

	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{"health endpoint", "/health", "healthy"},
		{"metrics endpoint", "/metrics", "metrics"},
		{"api endpoint", "/api/test", "test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.url, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", w.Code)
			}

			if w.Body.String() != tt.expected {
				t.Errorf("Expected body '%s', got '%s'", tt.expected, w.Body.String())
			}
		})
	}
}

func TestFilteredRequestLoggerWithNilFilter(t *testing.T) {
	// Test with nil filter (should not panic)
	logger := NewFilteredRequestLogger(nil)

	// Create a test router
	router := chi.NewRouter()

	// Add the filtered logger middleware
	router.Use(logger)

	// Add a test endpoint
	router.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("test")); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	})

	// Test request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should work normally
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "test" {
		t.Errorf("Expected body 'test', got '%s'", w.Body.String())
	}
}

func TestFilteredRequestLoggerWithComplexRegex(t *testing.T) {
	// Create a complex regex that filters out multiple patterns
	filterRegex := regexp.MustCompile(`^/(health|metrics|ready|live)(/.*)?$`)
	logger := NewFilteredRequestLogger(filterRegex)

	// Create a test router
	router := chi.NewRouter()

	// Add the filtered logger middleware
	router.Use(logger)

	// Add test endpoints
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("health")); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	})

	router.Get("/health/detailed", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("health-detailed")); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	})

	router.Get("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("metrics")); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	})

	router.Get("/api/users", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("users")); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	})

	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{"health root", "/health", "health"},
		{"health detailed", "/health/detailed", "health-detailed"},
		{"metrics", "/metrics", "metrics"},
		{"api endpoint", "/api/users", "users"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.url, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", w.Code)
			}

			if w.Body.String() != tt.expected {
				t.Errorf("Expected body '%s', got '%s'", tt.expected, w.Body.String())
			}
		})
	}
}
