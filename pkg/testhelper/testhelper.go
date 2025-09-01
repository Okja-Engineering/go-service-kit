// Package testhelper provides helpers for table-driven HTTP endpoint testing in Go.
package testhelper

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

const (
	ContentType     = "Content-Type"
	ContentLength   = "Content-Length"
	ApplicationJSON = "application/json"
)

// Logger defines the interface for logging operations
type Logger interface {
	Printf(format string, v ...interface{})
}

// DefaultLogger implements Logger using the standard log package
type DefaultLogger struct{}

// Printf logs using the standard log package
func (l *DefaultLogger) Printf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

// TestRunner defines the interface for running test cases
type TestRunner interface {
	Run(t *testing.T, router chi.Router, testCases []TestCase)
}

// ResponseValidator defines the interface for validating HTTP responses
type ResponseValidator interface {
	Validate(t *testing.T, rec *httptest.ResponseRecorder, test *TestCase)
}

// DefaultResponseValidator implements ResponseValidator with standard validation
type DefaultResponseValidator struct{}

// Validate validates the HTTP response for a test case
func (v *DefaultResponseValidator) Validate(t *testing.T, rec *httptest.ResponseRecorder, test *TestCase) {
	t.Helper()
	resp := rec.Result()
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Failed to read body: %v", err)
		return
	}

	if resp.StatusCode != test.CheckStatus {
		t.Errorf("Got status %d wanted %d\nBody: %s", resp.StatusCode, test.CheckStatus, string(body))
	}

	if test.CheckBody != "" {
		bodyCheckRegex, err := regexp.Compile(test.CheckBody)
		if err != nil {
			t.Errorf("Invalid body check regex: %v", err)
			return
		}

		matches := bodyCheckRegex.FindAllStringIndex(string(body), -1)

		if len(matches) != test.CheckBodyCount {
			t.Errorf("'%s' not found %d times in body\nBODY: %s", test.CheckBody, test.CheckBodyCount, body)
		}
	}
}

// TestHelperOption is a functional option for test helper configuration
type TestHelperOption func(*TestHelperConfig)

// TestHelperConfig holds configuration for test helper behavior
type TestHelperConfig struct {
	Logger            Logger
	ResponseValidator ResponseValidator
	LogTestExecution  bool
	DefaultHeaders    map[string]string
}

// DefaultTestHelperConfig provides sensible defaults
func DefaultTestHelperConfig() *TestHelperConfig {
	return &TestHelperConfig{
		Logger:            &DefaultLogger{},
		ResponseValidator: &DefaultResponseValidator{},
		LogTestExecution:  true,
		DefaultHeaders: map[string]string{
			ContentType: ApplicationJSON,
		},
	}
}

// WithLogger sets a custom logger
func WithLogger(logger Logger) TestHelperOption {
	return func(config *TestHelperConfig) {
		config.Logger = logger
	}
}

// WithResponseValidator sets a custom response validator
func WithResponseValidator(validator ResponseValidator) TestHelperOption {
	return func(config *TestHelperConfig) {
		config.ResponseValidator = validator
	}
}

// WithLogTestExecution enables/disables test execution logging
func WithLogTestExecution(logTest bool) TestHelperOption {
	return func(config *TestHelperConfig) {
		config.LogTestExecution = logTest
	}
}

// WithDefaultHeaders sets default headers for all requests
func WithDefaultHeaders(headers map[string]string) TestHelperOption {
	return func(config *TestHelperConfig) {
		config.DefaultHeaders = headers
	}
}

// NewTestHelperConfig creates a new test helper config with options
func NewTestHelperConfig(options ...TestHelperOption) *TestHelperConfig {
	config := DefaultTestHelperConfig()
	for _, option := range options {
		option(config)
	}
	return config
}

// TestHelper handles HTTP test execution with configuration
type TestHelper struct {
	config *TestHelperConfig
}

// NewTestHelper creates a new test helper with options
func NewTestHelper(options ...TestHelperOption) *TestHelper {
	config := NewTestHelperConfig(options...)
	return &TestHelper{config: config}
}

// TestCase represents a single HTTP test case for use with Run.
type TestCase struct {
	// Name is a description of the test case.
	Name string
	// URL is the endpoint under test (can include query params).
	URL string
	// Method is the HTTP method to use (GET, POST, etc).
	Method string
	// Body is the optional request body for POST, PUT, etc.
	Body string
	// Headers is an optional map of headers to set on the request.
	Headers map[string]string
	// CheckBody is a regex to match against the response body.
	CheckBody string
	// CheckBodyCount is the number of expected matches for CheckBody.
	CheckBodyCount int
	// CheckStatus is the expected HTTP status code.
	CheckStatus int
}

// Validate checks if the HTTP method of the test case is valid.
func (tc *TestCase) Validate() error {
	switch tc.Method {
	case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete,
		http.MethodPatch, http.MethodHead, http.MethodOptions:
		return nil
	default:
		return fmt.Errorf("invalid method: %s", tc.Method)
	}
}

// Run executes the provided test cases against the given chi.Router.
// Each test case is run as a subtest. All checks are reported as errors, not fatals.
func (th *TestHelper) Run(t *testing.T, router chi.Router, testCases []TestCase) {
	t.Helper()
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Helper()
			if th.config.LogTestExecution {
				th.config.Logger.Printf("### Running test: %s %s", tc.Method, tc.URL)
			}
			req := th.newRequest(t, &tc)

			// Set default headers first
			for k, v := range th.config.DefaultHeaders {
				req.Header.Set(k, v)
			}

			// Set custom headers if provided (override defaults)
			for k, v := range tc.Headers {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			th.config.ResponseValidator.Validate(t, rec, &tc)
		})
	}
}

// newRequest creates a new HTTP request for a test case.
func (th *TestHelper) newRequest(t *testing.T, test *TestCase) *http.Request {
	t.Helper()
	req := httptest.NewRequest(test.Method, test.URL, strings.NewReader(test.Body))
	req.Header.Set(ContentLength, strconv.Itoa(len(test.Body)))
	return req
}

// Legacy functions for backward compatibility
func Run(t *testing.T, router chi.Router, testCases []TestCase) {
	helper := NewTestHelper()
	helper.Run(t, router, testCases)
}
