package logging

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/go-chi/chi/middleware"
)

// MockLogger for testing
type MockLogger struct {
	output *bytes.Buffer
}

func (m *MockLogger) Printf(format string, v ...interface{}) {
	m.output.WriteString(format)
}

func (m *MockLogger) Println(v ...interface{}) {
	m.output.WriteString("log")
}

// MockURLFilter for testing
type MockURLFilter struct {
	shouldFilter bool
}

func (m *MockURLFilter) ShouldFilter(url string) bool {
	return m.shouldFilter
}

func TestNewRequestLogger(t *testing.T) {
	// Test default logger
	logger := NewRequestLogger()
	if logger.config.Logger == nil {
		t.Error("Expected logger to be set")
	}
	if logger.config.Formatter == nil {
		t.Error("Expected formatter to be set")
	}
	if logger.config.URLFilter != nil {
		t.Error("Expected URLFilter to be nil by default")
	}
}

func TestNewRequestLoggerWithOptions(t *testing.T) {
	mockLogger := &MockLogger{output: &bytes.Buffer{}}
	mockFilter := &MockURLFilter{shouldFilter: true}

	logger := NewRequestLogger(
		WithLogger(mockLogger),
		WithURLFilter(mockFilter),
		WithNoColor(true),
	)

	if logger.config.Logger != mockLogger {
		t.Error("Expected custom logger to be set")
	}
	if logger.config.URLFilter != mockFilter {
		t.Error("Expected custom URL filter to be set")
	}
	if !logger.config.NoColor {
		t.Error("Expected NoColor to be true")
	}
}

func TestRequestLoggerMiddleware(t *testing.T) {
	// Test without filtering
	logger := NewRequestLogger()
	middleware := logger.Middleware()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("test"))
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	middleware(handler).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestRequestLoggerWithURLFilter(t *testing.T) {
	// Test with URL filtering
	mockFilter := &MockURLFilter{shouldFilter: true}
	logger := NewRequestLogger(WithURLFilter(mockFilter))
	middleware := logger.Middleware()

	handlerCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("test"))
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	middleware(handler).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	if !handlerCalled {
		t.Error("Expected handler to be called")
	}
}

func TestRegexURLFilter(t *testing.T) {
	pattern := regexp.MustCompile(`/health`)
	filter := &RegexURLFilter{pattern: pattern}

	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{
			name:     "should filter health endpoint",
			url:      "/health",
			expected: true,
		},
		{
			name:     "should filter health with query",
			url:      "/health?ready=1",
			expected: true,
		},
		{
			name:     "should not filter other endpoints",
			url:      "/api/users",
			expected: false,
		},
		{
			name:     "should not filter empty URL",
			url:      "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filter.ShouldFilter(tt.url)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestNewLoggingConfig(t *testing.T) {
	// Test with no options (should use defaults)
	config := NewLoggingConfig()
	if config.Logger == nil {
		t.Error("Expected logger to be set")
	}
	if config.Formatter == nil {
		t.Error("Expected formatter to be set")
	}
	if config.URLFilter != nil {
		t.Error("Expected URLFilter to be nil by default")
	}
	if config.NoColor {
		t.Error("Expected NoColor to be false by default")
	}

	// Test with custom options
	mockLogger := &MockLogger{output: &bytes.Buffer{}}
	mockFilter := &MockURLFilter{shouldFilter: true}

	config = NewLoggingConfig(
		WithLogger(mockLogger),
		WithURLFilter(mockFilter),
		WithNoColor(true),
	)

	if config.Logger != mockLogger {
		t.Error("Expected custom logger to be set")
	}
	if config.URLFilter != mockFilter {
		t.Error("Expected custom URL filter to be set")
	}
	if !config.NoColor {
		t.Error("Expected NoColor to be true")
	}
}

func TestWithRegexFilter(t *testing.T) {
	pattern := regexp.MustCompile(`/metrics`)
	config := NewLoggingConfig(WithRegexFilter(pattern))

	if config.URLFilter == nil {
		t.Error("Expected URLFilter to be set")
	}

	regexFilter, ok := config.URLFilter.(*RegexURLFilter)
	if !ok {
		t.Error("Expected URLFilter to be RegexURLFilter")
	}

	if regexFilter.pattern != pattern {
		t.Error("Expected pattern to match")
	}
}

// Legacy function tests (existing tests)
func TestNewFilteredRequestLogger(t *testing.T) {
	pattern := regexp.MustCompile(`/health`)
	middleware := NewFilteredRequestLogger(pattern)

	handlerCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("test"))
	})

	// Test filtered URL
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	middleware(handler).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	if !handlerCalled {
		t.Error("Expected handler to be called")
	}
}

func TestFilteredRequestLoggerWithFilteredURL(t *testing.T) {
	pattern := regexp.MustCompile(`/health`)
	formatter := middleware.DefaultLogFormatter{
		Logger:  log.New(&bytes.Buffer{}, "", 0),
		NoColor: true,
	}

	middleware := FilteredRequestLogger(&formatter, pattern)

	handlerCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("test"))
	})

	// Test filtered URL
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	middleware(handler).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	if !handlerCalled {
		t.Error("Expected handler to be called")
	}
}

func TestFilteredRequestLoggerWithNonFilteredURL(t *testing.T) {
	pattern := regexp.MustCompile(`/health`)
	output := &bytes.Buffer{}
	formatter := middleware.DefaultLogFormatter{
		Logger:  log.New(output, "", 0),
		NoColor: true,
	}

	middleware := FilteredRequestLogger(&formatter, pattern)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("test"))
	})

	// Test non-filtered URL
	req := httptest.NewRequest("GET", "/api/test", nil)
	w := httptest.NewRecorder()

	middleware(handler).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Should have logged the request
	if output.Len() == 0 {
		t.Error("Expected request to be logged")
	}
}

func TestFilteredRequestLoggerWithMultipleFilters(t *testing.T) {
	pattern := regexp.MustCompile(`/(health|metrics)`)
	formatter := middleware.DefaultLogFormatter{
		Logger:  log.New(&bytes.Buffer{}, "", 0),
		NoColor: true,
	}

	middleware := FilteredRequestLogger(&formatter, pattern)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("test"))
	})

	tests := []struct {
		name string
		url  string
	}{
		{"health endpoint", "/health"},
		{"metrics endpoint", "/metrics"},
		{"api endpoint", "/api/test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.url, nil)
			w := httptest.NewRecorder()

			middleware(handler).ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", w.Code)
			}
		})
	}
}

func TestFilteredRequestLoggerWithNilFilter(t *testing.T) {
	output := &bytes.Buffer{}
	formatter := middleware.DefaultLogFormatter{
		Logger:  log.New(output, "", 0),
		NoColor: true,
	}

	middleware := FilteredRequestLogger(&formatter, nil)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("test"))
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	middleware(handler).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Should have logged the request
	if output.Len() == 0 {
		t.Error("Expected request to be logged")
	}
}

func TestFilteredRequestLoggerWithComplexRegex(t *testing.T) {
	pattern := regexp.MustCompile(`^/(health|metrics)(/.*)?$`)
	formatter := middleware.DefaultLogFormatter{
		Logger:  log.New(&bytes.Buffer{}, "", 0),
		NoColor: true,
	}

	middleware := FilteredRequestLogger(&formatter, pattern)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("test"))
	})

	tests := []struct {
		name string
		url  string
	}{
		{"health root", "/health"},
		{"health detailed", "/health/detailed"},
		{"metrics", "/metrics"},
		{"api endpoint", "/api/users"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.url, nil)
			w := httptest.NewRecorder()

			middleware(handler).ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", w.Code)
			}
		})
	}
}
