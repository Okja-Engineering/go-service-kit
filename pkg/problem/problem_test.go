package problem

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func TestNew(t *testing.T) {
	problem := New("test-type", "Test Title", 400, "Test detail", "test-instance")

	if problem.Type != "test-type" {
		t.Errorf("Expected Type 'test-type', got '%s'", problem.Type)
	}

	if problem.Title != "Test Title" {
		t.Errorf("Expected Title 'Test Title', got '%s'", problem.Title)
	}

	if problem.Status != 400 {
		t.Errorf("Expected Status 400, got %d", problem.Status)
	}

	if problem.Detail != "Test detail" {
		t.Errorf("Expected Detail 'Test detail', got '%s'", problem.Detail)
	}

	if problem.Instance != "test-instance" {
		t.Errorf("Expected Instance 'test-instance', got '%s'", problem.Instance)
	}
}

func TestSend(t *testing.T) {
	problem := New("test-type", "Test Title", 400, "Test detail", "test-instance")

	w := httptest.NewRecorder()

	problem.Send(w)

	if w.Code != 400 {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", w.Header().Get("Content-Type"))
	}

	var response Problem
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Type != "test-type" {
		t.Errorf("Expected Type 'test-type', got '%s'", response.Type)
	}

	if response.Title != "Test Title" {
		t.Errorf("Expected Title 'Test Title', got '%s'", response.Title)
	}

	if response.Status != 400 {
		t.Errorf("Expected Status 400, got %d", response.Status)
	}

	if response.Detail != "Test detail" {
		t.Errorf("Expected Detail 'Test detail', got '%s'", response.Detail)
	}

	if response.Instance != "test-instance" {
		t.Errorf("Expected Instance 'test-instance', got '%s'", response.Instance)
	}
}

func TestWrap(t *testing.T) {
	// Test with error
	testError := &testError{message: "test error message"}
	problem := Wrap(500, "test-type", "test-instance", testError)

	if problem.Type != "test-type" {
		t.Errorf("Expected Type 'test-type', got '%s'", problem.Type)
	}

	if problem.Status != 500 {
		t.Errorf("Expected Status 500, got %d", problem.Status)
	}

	if problem.Detail != "test error message" {
		t.Errorf("Expected Detail 'test error message', got '%s'", problem.Detail)
	}

	if problem.Instance != "test-instance" {
		t.Errorf("Expected Instance 'test-instance', got '%s'", problem.Instance)
	}

	// Test without error
	problem = Wrap(400, "test-type", "test-instance", nil)

	if problem.Type != "test-type" {
		t.Errorf("Expected Type 'test-type', got '%s'", problem.Type)
	}

	if problem.Status != 400 {
		t.Errorf("Expected Status 400, got %d", problem.Status)
	}

	if problem.Detail != "Other error occurred" {
		t.Errorf("Expected Detail 'Other error occurred', got '%s'", problem.Detail)
	}

	if problem.Instance != "test-instance" {
		t.Errorf("Expected Instance 'test-instance', got '%s'", problem.Instance)
	}
}

func TestError(t *testing.T) {
	problem := New("test-type", "Test Title", 400, "Test detail", "test-instance")

	expected := "Problem: Type: 'test-type', Title: 'Test Title', Status: '400', " +
		"Detail: 'Test detail', Instance: 'test-instance'"
	if problem.Error() != expected {
		t.Errorf("Expected Error '%s', got '%s'", expected, problem.Error())
	}
}

func TestMyCaller(t *testing.T) {
	caller := MyCaller()

	// Should contain the function name
	if caller == "" {
		t.Error("Expected caller to not be empty")
	}

	// Should contain the package name
	if len(caller) < 10 {
		t.Error("Expected caller to be substantial")
	}
}

func TestProblemWithMinimalFields(t *testing.T) {
	// Test problem with minimal required fields
	problem := New("test-type", "Test Title", 0, "", "")

	if problem.Type != "test-type" {
		t.Errorf("Expected Type 'test-type', got '%s'", problem.Type)
	}

	if problem.Title != "Test Title" {
		t.Errorf("Expected Title 'Test Title', got '%s'", problem.Title)
	}

	if problem.Status != 0 {
		t.Errorf("Expected Status 0, got %d", problem.Status)
	}

	if problem.Detail != "" {
		t.Errorf("Expected Detail to be empty, got '%s'", problem.Detail)
	}

	if problem.Instance != "" {
		t.Errorf("Expected Instance to be empty, got '%s'", problem.Instance)
	}
}

func TestProblemJSONSerialization(t *testing.T) {
	problem := New("test-type", "Test Title", 400, "Test detail", "test-instance")

	data, err := json.Marshal(problem)
	if err != nil {
		t.Fatalf("Failed to marshal problem: %v", err)
	}

	var unmarshaled Problem
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal problem: %v", err)
	}

	if unmarshaled.Type != problem.Type {
		t.Errorf("Expected Type '%s', got '%s'", problem.Type, unmarshaled.Type)
	}

	if unmarshaled.Title != problem.Title {
		t.Errorf("Expected Title '%s', got '%s'", problem.Title, unmarshaled.Title)
	}

	if unmarshaled.Status != problem.Status {
		t.Errorf("Expected Status %d, got %d", problem.Status, unmarshaled.Status)
	}

	if unmarshaled.Detail != problem.Detail {
		t.Errorf("Expected Detail '%s', got '%s'", problem.Detail, unmarshaled.Detail)
	}

	if unmarshaled.Instance != problem.Instance {
		t.Errorf("Expected Instance '%s', got '%s'", problem.Instance, unmarshaled.Instance)
	}
}

// Helper type for testing
type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}
