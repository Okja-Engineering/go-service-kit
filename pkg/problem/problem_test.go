package problem

import (
	"bytes"
	"errors"
	"net/http/httptest"
	"testing"
)

// MockLogger for testing
type MockLogger struct {
	output *bytes.Buffer
}

func (m *MockLogger) Printf(format string, v ...interface{}) {
	m.output.WriteString(format)
}

func TestNewProblemManager(t *testing.T) {
	// Test default manager
	manager := NewProblemManager()
	if manager.config.Logger == nil {
		t.Error("Expected logger to be set")
	}
	if manager.config.LogPrefix != "### 💥 API" {
		t.Errorf("Expected log prefix '### 💥 API', got '%s'", manager.config.LogPrefix)
	}
	if !manager.config.LogErrors {
		t.Error("Expected LogErrors to be true by default")
	}
}

func TestNewProblemManagerWithOptions(t *testing.T) {
	mockLogger := &MockLogger{output: &bytes.Buffer{}}

	manager := NewProblemManager(
		WithLogger(mockLogger),
		WithLogPrefix("[ERROR]"),
		WithLogErrors(false),
	)

	if manager.config.Logger != mockLogger {
		t.Error("Expected custom logger to be set")
	}
	if manager.config.LogPrefix != "[ERROR]" {
		t.Errorf("Expected log prefix '[ERROR]', got '%s'", manager.config.LogPrefix)
	}
	if manager.config.LogErrors {
		t.Error("Expected LogErrors to be false")
	}
}

func TestProblemManagerNew(t *testing.T) {
	manager := NewProblemManager()

	problem := manager.New("test-type", "Test Title", 400, "Test detail", "test-instance")

	if problem.Type != "test-type" {
		t.Errorf("Expected type 'test-type', got '%s'", problem.Type)
	}
	if problem.Title != "Test Title" {
		t.Errorf("Expected title 'Test Title', got '%s'", problem.Title)
	}
	if problem.Status != 400 {
		t.Errorf("Expected status 400, got %d", problem.Status)
	}
	if problem.Detail != "Test detail" {
		t.Errorf("Expected detail 'Test detail', got '%s'", problem.Detail)
	}
	if problem.Instance != "test-instance" {
		t.Errorf("Expected instance 'test-instance', got '%s'", problem.Instance)
	}
}

func TestProblemManagerSend(t *testing.T) {
	mockLogger := &MockLogger{output: &bytes.Buffer{}}
	manager := NewProblemManager(WithLogger(mockLogger))

	problem := manager.New("test-type", "Test Title", 400, "Test detail", "test-instance")

	w := httptest.NewRecorder()

	manager.Send(problem, w)

	if w.Code != 400 {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/problem+json" {
		t.Errorf("Expected content type 'application/problem+json', got '%s'", contentType)
	}

	// Check that error was logged
	if mockLogger.output.Len() == 0 {
		t.Error("Expected error to be logged")
	}
}

func TestProblemManagerSendWithoutLogging(t *testing.T) {
	mockLogger := &MockLogger{output: &bytes.Buffer{}}
	manager := NewProblemManager(
		WithLogger(mockLogger),
		WithLogErrors(false),
	)

	problem := manager.New("test-type", "Test Title", 400, "Test detail", "test-instance")

	w := httptest.NewRecorder()

	manager.Send(problem, w)

	if w.Code != 400 {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	// Check that error was not logged
	if mockLogger.output.Len() > 0 {
		t.Error("Expected error not to be logged")
	}
}

func TestProblemManagerWrap(t *testing.T) {
	manager := NewProblemManager()

	testError := errors.New("test error")
	problem := manager.Wrap(500, "server-error", "test-instance", testError)

	if problem.Type != "server-error" {
		t.Errorf("Expected type 'server-error', got '%s'", problem.Type)
	}
	if problem.Status != 500 {
		t.Errorf("Expected status 500, got %d", problem.Status)
	}
	if problem.Detail != "test error" {
		t.Errorf("Expected detail 'test error', got '%s'", problem.Detail)
	}
	if problem.Instance != "test-instance" {
		t.Errorf("Expected instance 'test-instance', got '%s'", problem.Instance)
	}
}

func TestProblemManagerWrapWithNilError(t *testing.T) {
	manager := NewProblemManager()

	problem := manager.Wrap(500, "server-error", "test-instance", nil)

	if problem.Type != "server-error" {
		t.Errorf("Expected type 'server-error', got '%s'", problem.Type)
	}
	if problem.Status != 500 {
		t.Errorf("Expected status 500, got %d", problem.Status)
	}
	if problem.Detail != "Other error occurred" {
		t.Errorf("Expected detail 'Other error occurred', got '%s'", problem.Detail)
	}
	if problem.Instance != "test-instance" {
		t.Errorf("Expected instance 'test-instance', got '%s'", problem.Instance)
	}
}

func TestNewProblemConfig(t *testing.T) {
	// Test with no options (should use defaults)
	config := NewProblemConfig()
	if config.Logger == nil {
		t.Error("Expected logger to be set")
	}
	if config.LogPrefix != "### 💥 API" {
		t.Errorf("Expected log prefix '### 💥 API', got '%s'", config.LogPrefix)
	}
	if !config.LogErrors {
		t.Error("Expected LogErrors to be true by default")
	}

	// Test with custom options
	mockLogger := &MockLogger{output: &bytes.Buffer{}}

	config = NewProblemConfig(
		WithLogger(mockLogger),
		WithLogPrefix("[ERROR]"),
		WithLogErrors(false),
	)

	if config.Logger != mockLogger {
		t.Error("Expected custom logger to be set")
	}
	if config.LogPrefix != "[ERROR]" {
		t.Errorf("Expected log prefix '[ERROR]', got '%s'", config.LogPrefix)
	}
	if config.LogErrors {
		t.Error("Expected LogErrors to be false")
	}
}

// Legacy function tests (existing tests)
func TestNew(t *testing.T) {
	problem := New("test-type", "Test Title", 400, "Test detail", "test-instance")

	if problem.Type != "test-type" {
		t.Errorf("Expected type 'test-type', got '%s'", problem.Type)
	}
	if problem.Title != "Test Title" {
		t.Errorf("Expected title 'Test Title', got '%s'", problem.Title)
	}
	if problem.Status != 400 {
		t.Errorf("Expected status 400, got %d", problem.Status)
	}
	if problem.Detail != "Test detail" {
		t.Errorf("Expected detail 'Test detail', got '%s'", problem.Detail)
	}
	if problem.Instance != "test-instance" {
		t.Errorf("Expected instance 'test-instance', got '%s'", problem.Instance)
	}
}

func TestSend(t *testing.T) {
	problem := New("test-type", "Test Title", 400, "Test detail", "test-instance")

	w := httptest.NewRecorder()

	problem.Send(w)

	if w.Code != 400 {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/problem+json" {
		t.Errorf("Expected content type 'application/problem+json', got '%s'", contentType)
	}
}

func TestWrap(t *testing.T) {
	testError := errors.New("test error")
	problem := Wrap(500, "server-error", "test-instance", testError)

	if problem.Type != "server-error" {
		t.Errorf("Expected type 'server-error', got '%s'", problem.Type)
	}
	if problem.Status != 500 {
		t.Errorf("Expected status 500, got %d", problem.Status)
	}
	if problem.Detail != "test error" {
		t.Errorf("Expected detail 'test error', got '%s'", problem.Detail)
	}
	if problem.Instance != "test-instance" {
		t.Errorf("Expected instance 'test-instance', got '%s'", problem.Instance)
	}
}

func TestError(t *testing.T) {
	problem := New("test-type", "Test Title", 400, "Test detail", "test-instance")

	expected := "Problem: Type: 'test-type', Title: 'Test Title', Status: '400', " +
		"Detail: 'Test detail', Instance: 'test-instance'"
	if problem.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, problem.Error())
	}
}

func TestMyCaller(t *testing.T) {
	caller := MyCaller()
	if caller == "" {
		t.Error("Expected caller to not be empty")
	}
	if caller == "unknown" {
		t.Error("Expected caller to not be 'unknown'")
	}
}

func TestProblemWithMinimalFields(t *testing.T) {
	problem := New("test-type", "Test Title", 0, "", "")

	if problem.Type != "test-type" {
		t.Errorf("Expected type 'test-type', got '%s'", problem.Type)
	}
	if problem.Title != "Test Title" {
		t.Errorf("Expected title 'Test Title', got '%s'", problem.Title)
	}
	if problem.Status != 0 {
		t.Errorf("Expected status 0, got %d", problem.Status)
	}
	if problem.Detail != "" {
		t.Errorf("Expected empty detail, got '%s'", problem.Detail)
	}
	if problem.Instance != "" {
		t.Errorf("Expected empty instance, got '%s'", problem.Instance)
	}
}

func TestProblemJSONSerialization(t *testing.T) {
	problem := New("test-type", "Test Title", 400, "Test detail", "test-instance")

	w := httptest.NewRecorder()

	problem.Send(w)

	// Check that the response body contains JSON
	body := w.Body.String()
	if body == "" {
		t.Error("Expected non-empty response body")
	}

	// Basic JSON structure check
	if !bytes.Contains([]byte(body), []byte("test-type")) {
		t.Error("Expected response to contain 'test-type'")
	}
	if !bytes.Contains([]byte(body), []byte("Test Title")) {
		t.Error("Expected response to contain 'Test Title'")
	}
	if !bytes.Contains([]byte(body), []byte("400")) {
		t.Error("Expected response to contain '400'")
	}
}
