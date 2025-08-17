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
func Run(t *testing.T, router chi.Router, testCases []TestCase) {
	t.Helper()
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Helper()
			log.Printf("### Running test: %s %s", tc.Method, tc.URL)
			req := newRequest(t, &tc)

			// Set custom headers if provided
			for k, v := range tc.Headers {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			checkResponse(t, rec, &tc)
		})
	}
}

// newRequest creates a new HTTP request for a test case.
func newRequest(t *testing.T, test *TestCase) *http.Request {
	t.Helper()
	req := httptest.NewRequest(test.Method, test.URL, strings.NewReader(test.Body))
	req.Header.Set(ContentType, ApplicationJSON)
	req.Header.Set(ContentLength, strconv.Itoa(len(test.Body)))
	return req
}

// checkResponse validates the HTTP response for a test case.
func checkResponse(t *testing.T, rec *httptest.ResponseRecorder, test *TestCase) {
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
