package testhelper

import (
	"io"
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestTestCaseValidate(t *testing.T) {
	tests := []struct {
		name    string
		method  string
		isValid bool
	}{
		{"GET", http.MethodGet, true},
		{"POST", http.MethodPost, true},
		{"PUT", http.MethodPut, true},
		{"DELETE", http.MethodDelete, true},
		{"PATCH", http.MethodPatch, true},
		{"HEAD", http.MethodHead, true},
		{"OPTIONS", http.MethodOptions, true},
		{"INVALID", "INVALID", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := TestCase{Method: tt.method}
			err := tc.Validate()

			if tt.isValid && err != nil {
				t.Errorf("Expected valid method '%s', got error: %v", tt.method, err)
			}

			if !tt.isValid && err == nil {
				t.Errorf("Expected invalid method '%s' to return error", tt.method)
			}
		})
	}
}

func TestRun(t *testing.T) {
	// Create a test router
	router := chi.NewRouter()
	router.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("test response")); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	})

	router.Post("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		if _, err := w.Write([]byte("created")); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	})

	router.Get("/notfound", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		if _, err := w.Write([]byte("not found")); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	})

	testCases := []TestCase{
		{
			Name:           "GET success",
			URL:            "/test",
			Method:         http.MethodGet,
			CheckStatus:    http.StatusOK,
			CheckBody:      "test response",
			CheckBodyCount: 1,
		},
		{
			Name:           "POST success",
			URL:            "/test",
			Method:         http.MethodPost,
			CheckStatus:    http.StatusCreated,
			CheckBody:      "created",
			CheckBodyCount: 1,
		},
		{
			Name:           "404 response",
			URL:            "/notfound",
			Method:         http.MethodGet,
			CheckStatus:    http.StatusNotFound,
			CheckBody:      "not found",
			CheckBodyCount: 1,
		},
	}

	Run(t, router, testCases)
}

func TestRunWithHeaders(t *testing.T) {
	// Create a test router that checks headers
	router := chi.NewRouter()
	router.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		userAgent := r.Header.Get("User-Agent")
		contentType := r.Header.Get("Content-Type")

		if userAgent == "test-agent" && contentType == "application/json" {
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte("headers ok")); err != nil {
				t.Errorf("Failed to write response: %v", err)
			}
		} else {
			w.WriteHeader(http.StatusBadRequest)
			if _, err := w.Write([]byte("headers missing")); err != nil {
				t.Errorf("Failed to write response: %v", err)
			}
		}
	})

	testCases := []TestCase{
		{
			Name:   "with headers",
			URL:    "/test",
			Method: http.MethodGet,
			Headers: map[string]string{
				"User-Agent":   "test-agent",
				"Content-Type": "application/json",
			},
			CheckStatus:    http.StatusOK,
			CheckBody:      "headers ok",
			CheckBodyCount: 1,
		},
		{
			Name:           "without headers",
			URL:            "/test",
			Method:         http.MethodGet,
			CheckStatus:    http.StatusBadRequest,
			CheckBody:      "headers missing",
			CheckBodyCount: 1,
		},
	}

	Run(t, router, testCases)
}

func TestRunWithBody(t *testing.T) {
	// Create a test router that echoes the request body
	router := chi.NewRouter()
	router.Post("/echo", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(body); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	})

	testCases := []TestCase{
		{
			Name:           "echo body",
			URL:            "/echo",
			Method:         http.MethodPost,
			Body:           `{"test": "data"}`,
			CheckStatus:    http.StatusOK,
			CheckBody:      `{"test": "data"}`,
			CheckBodyCount: 1,
		},
		{
			Name:           "empty body",
			URL:            "/echo",
			Method:         http.MethodPost,
			Body:           "",
			CheckStatus:    http.StatusOK,
			CheckBody:      "",
			CheckBodyCount: 0,
		},
	}

	Run(t, router, testCases)
}

func TestRunWithRegexBody(t *testing.T) {
	// Create a test router that returns JSON
	router := chi.NewRouter()
	router.Get("/json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"name": "test", "value": 123, "items": ["a", "b", "c"]}`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	})

	testCases := []TestCase{
		{
			Name:           "regex match",
			URL:            "/json",
			Method:         http.MethodGet,
			CheckStatus:    http.StatusOK,
			CheckBody:      `"name": "test"`,
			CheckBodyCount: 1,
		},
		{
			Name:           "regex no match",
			URL:            "/json",
			Method:         http.MethodGet,
			CheckStatus:    http.StatusOK,
			CheckBody:      `"nonexistent"`,
			CheckBodyCount: 0,
		},
	}

	Run(t, router, testCases)
}
