package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Okja-Engineering/go-service-kit/pkg/problem"
	"github.com/go-chi/chi/v5"
)

func TestNewBase(t *testing.T) {
	base := NewBase("TestService", "1.0.0", "test-build", true)

	if base.ServiceName != "TestService" {
		t.Errorf("Expected ServiceName 'TestService', got '%s'", base.ServiceName)
	}

	if base.Version != "1.0.0" {
		t.Errorf("Expected Version '1.0.0', got '%s'", base.Version)
	}

	if base.BuildInfo != "test-build" {
		t.Errorf("Expected BuildInfo 'test-build', got '%s'", base.BuildInfo)
	}

	if !base.Healthy {
		t.Error("Expected Healthy to be true")
	}
}

func TestReturnJSON(t *testing.T) {
	base := NewBase("TestService", "1.0.0", "test-build", true)

	testData := map[string]string{
		"message": "Hello, World!",
		"status":  "success",
	}

	w := httptest.NewRecorder()

	base.ReturnJSON(w, testData)

	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", w.Header().Get("Content-Type"))
	}

	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["message"] != "Hello, World!" {
		t.Errorf("Expected message 'Hello, World!', got '%s'", response["message"])
	}

	if response["status"] != "success" {
		t.Errorf("Expected status 'success', got '%s'", response["status"])
	}
}

func TestReturnJSONWithError(t *testing.T) {
	base := NewBase("TestService", "1.0.0", "test-build", true)

	// Create a channel that can't be marshaled to JSON
	unmarshallable := make(chan int)

	w := httptest.NewRecorder()

	base.ReturnJSON(w, unmarshallable)

	// Should return a problem response
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	var problem problem.Problem
	if err := json.Unmarshal(w.Body.Bytes(), &problem); err != nil {
		t.Fatalf("Failed to unmarshal problem response: %v", err)
	}

	if problem.Type != "json-encoding" {
		t.Errorf("Expected problem type 'json-encoding', got '%s'", problem.Type)
	}
}

func TestReturnText(t *testing.T) {
	base := NewBase("TestService", "1.0.0", "test-build", true)

	w := httptest.NewRecorder()

	base.ReturnText(w, "Hello, World!")

	if w.Header().Get("Content-Type") != "text/plain" {
		t.Errorf("Expected Content-Type 'text/plain', got '%s'", w.Header().Get("Content-Type"))
	}

	if w.Body.String() != "Hello, World!" {
		t.Errorf("Expected body 'Hello, World!', got '%s'", w.Body.String())
	}
}

func TestReturnErrorJSON(t *testing.T) {
	base := NewBase("TestService", "1.0.0", "test-build", true)

	w := httptest.NewRecorder()

	testError := problem.New("test-error", "Test Error", 400, "This is a test error", "test-instance")
	base.ReturnErrorJSON(w, testError)

	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["error"] != testError.Error() {
		t.Errorf("Expected error '%s', got '%s'", testError.Error(), response["error"])
	}
}

func TestReturnOKJSON(t *testing.T) {
	base := NewBase("TestService", "1.0.0", "test-build", true)

	w := httptest.NewRecorder()

	base.ReturnOKJSON(w)

	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["result"] != "ok" {
		t.Errorf("Expected result 'ok', got '%s'", response["result"])
	}
}

func TestStartServer(t *testing.T) {
	base := NewBase("TestService", "1.0.0", "test-build", true)

	router := chi.NewRouter()
	router.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("test")); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	})

	// Start server in a goroutine so we can test it
	go func() {
		// Use a short timeout for testing
		base.StartServer(0, router, 100*time.Millisecond)
	}()

	// Give the server a moment to start
	time.Sleep(10 * time.Millisecond)
}
