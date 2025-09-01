package testhelper

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

// MockLogger for testing
type MockLogger struct {
	output *bytes.Buffer
}

func (m *MockLogger) Printf(format string, v ...interface{}) {
	m.output.WriteString(format)
}

// MockResponseValidator for testing
type MockResponseValidator struct {
	validateCalled bool
	lastRecorder   *httptest.ResponseRecorder
	lastTestCase   *TestCase
}

func (m *MockResponseValidator) Validate(t *testing.T, rec *httptest.ResponseRecorder, test *TestCase) {
	m.validateCalled = true
	m.lastRecorder = rec
	m.lastTestCase = test
}

func TestNewTestHelper(t *testing.T) {
	// Test default test helper
	helper := NewTestHelper()
	if helper.config.Logger == nil {
		t.Error("Expected logger to be set")
	}
	if helper.config.ResponseValidator == nil {
		t.Error("Expected response validator to be set")
	}
	if !helper.config.LogTestExecution {
		t.Error("Expected LogTestExecution to be true by default")
	}
	if helper.config.DefaultHeaders[ContentType] != ApplicationJSON {
		t.Error("Expected default Content-Type header")
	}
}

func TestNewTestHelperWithOptions(t *testing.T) {
	mockLogger := &MockLogger{output: &bytes.Buffer{}}
	mockValidator := &MockResponseValidator{}

	helper := NewTestHelper(
		WithLogger(mockLogger),
		WithResponseValidator(mockValidator),
		WithLogTestExecution(false),
		WithDefaultHeaders(map[string]string{"X-Test": "value"}),
	)

	if helper.config.Logger != mockLogger {
		t.Error("Expected custom logger to be set")
	}
	// We can't directly compare interfaces, but we can verify the config was set
	if helper.config.ResponseValidator == nil {
		t.Error("Expected response validator to be set")
	}
	if helper.config.LogTestExecution {
		t.Error("Expected LogTestExecution to be false")
	}
	if helper.config.DefaultHeaders["X-Test"] != "value" {
		t.Error("Expected custom default header to be set")
	}
}

func TestTestHelperRun(t *testing.T) {
	// Create a test router
	router := chi.NewRouter()
	router.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("test response"))
	})

	helper := NewTestHelper()

	testCases := []TestCase{
		{
			Name:           "GET success",
			URL:            "/test",
			Method:         http.MethodGet,
			CheckStatus:    http.StatusOK,
			CheckBody:      "test response",
			CheckBodyCount: 1,
		},
	}

	helper.Run(t, router, testCases)
}

func TestTestHelperRunWithCustomLogger(t *testing.T) {
	// Create a test router
	router := chi.NewRouter()
	router.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("test response"))
	})

	mockLogger := &MockLogger{output: &bytes.Buffer{}}
	helper := NewTestHelper(WithLogger(mockLogger))

	testCases := []TestCase{
		{
			Name:           "GET success",
			URL:            "/test",
			Method:         http.MethodGet,
			CheckStatus:    http.StatusOK,
			CheckBody:      "test response",
			CheckBodyCount: 1,
		},
	}

	helper.Run(t, router, testCases)

	// Check that logger was called
	if mockLogger.output.Len() == 0 {
		t.Error("Expected logger to be called")
	}
}

func TestTestHelperRunWithCustomValidator(t *testing.T) {
	// Create a test router
	router := chi.NewRouter()
	router.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("test response"))
	})

	mockValidator := &MockResponseValidator{}
	helper := NewTestHelper(WithResponseValidator(mockValidator))

	testCases := []TestCase{
		{
			Name:           "GET success",
			URL:            "/test",
			Method:         http.MethodGet,
			CheckStatus:    http.StatusOK,
			CheckBody:      "test response",
			CheckBodyCount: 1,
		},
	}

	helper.Run(t, router, testCases)

	// Check that validator was called
	if !mockValidator.validateCalled {
		t.Error("Expected validator to be called")
	}
	if mockValidator.lastRecorder == nil {
		t.Error("Expected validator to receive response recorder")
	}
	if mockValidator.lastTestCase == nil {
		t.Error("Expected validator to receive test case")
	}
}

func TestTestHelperRunWithoutLogging(t *testing.T) {
	// Create a test router
	router := chi.NewRouter()
	router.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("test response"))
	})

	mockLogger := &MockLogger{output: &bytes.Buffer{}}
	helper := NewTestHelper(
		WithLogger(mockLogger),
		WithLogTestExecution(false),
	)

	testCases := []TestCase{
		{
			Name:           "GET success",
			URL:            "/test",
			Method:         http.MethodGet,
			CheckStatus:    http.StatusOK,
			CheckBody:      "test response",
			CheckBodyCount: 1,
		},
	}

	helper.Run(t, router, testCases)

	// Check that logger was not called
	if mockLogger.output.Len() > 0 {
		t.Error("Expected logger not to be called")
	}
}

func TestTestHelperRunWithDefaultHeaders(t *testing.T) {
	// Create a test router that checks headers
	router := chi.NewRouter()
	router.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")
		customHeader := r.Header.Get("X-Custom")

		if contentType == "application/json" && customHeader == "default" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("headers ok"))
		} else {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("headers missing"))
		}
	})

	helper := NewTestHelper(
		WithDefaultHeaders(map[string]string{
			"Content-Type": "application/json",
			"X-Custom":     "default",
		}),
	)

	testCases := []TestCase{
		{
			Name:           "with default headers",
			URL:            "/test",
			Method:         http.MethodGet,
			CheckStatus:    http.StatusOK,
			CheckBody:      "headers ok",
			CheckBodyCount: 1,
		},
	}

	helper.Run(t, router, testCases)
}

func TestNewTestHelperConfig(t *testing.T) {
	// Test with no options (should use defaults)
	config := NewTestHelperConfig()
	if config.Logger == nil {
		t.Error("Expected logger to be set")
	}
	if config.ResponseValidator == nil {
		t.Error("Expected response validator to be set")
	}
	if !config.LogTestExecution {
		t.Error("Expected LogTestExecution to be true by default")
	}

	// Test with custom options
	mockLogger := &MockLogger{output: &bytes.Buffer{}}
	mockValidator := &MockResponseValidator{}

	config = NewTestHelperConfig(
		WithLogger(mockLogger),
		WithResponseValidator(mockValidator),
		WithLogTestExecution(false),
	)

	if config.Logger != mockLogger {
		t.Error("Expected custom logger to be set")
	}
	// We can't directly compare interfaces, but we can verify the config was set
	if config.ResponseValidator == nil {
		t.Error("Expected response validator to be set")
	}
	if config.LogTestExecution {
		t.Error("Expected LogTestExecution to be false")
	}
}

// Legacy function tests (existing tests)
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
