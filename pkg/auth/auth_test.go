package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewJWTValidator(t *testing.T) {
	// Test with invalid JWKS URL (should not panic)
	validator := NewJWTValidator("test-client", "https://invalid-jwks-url.com", "test-scope")

	if validator.clientID != "test-client" {
		t.Errorf("Expected clientID 'test-client', got '%s'", validator.clientID)
	}

	if validator.scope != "test-scope" {
		t.Errorf("Expected scope 'test-scope', got '%s'", validator.scope)
	}

	// JWKS should be nil due to invalid URL
	if validator.jwks != nil {
		t.Error("Expected jwks to be nil for invalid URL")
	}
}

func TestNewPassthroughValidator(t *testing.T) {
	validator := NewPassthroughValidator()

	// PassthroughValidator is a struct, so it can't be nil
	// Just verify the function returns without error
	_ = validator
}

func TestPassthroughValidatorMiddleware(t *testing.T) {
	validator := NewPassthroughValidator()

	// Create a test handler
	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("test")); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	})

	// Wrap with middleware
	middleware := validator.Middleware(testHandler)

	// Test request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	if !handlerCalled {
		t.Error("Expected handler to be called")
	}

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "test" {
		t.Errorf("Expected body 'test', got '%s'", w.Body.String())
	}
}

func TestPassthroughValidatorProtect(t *testing.T) {
	validator := NewPassthroughValidator()

	// Create a test handler
	handlerCalled := false
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("test")); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}

	// Wrap with protect
	protected := validator.Protect(testHandler)

	// Test request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	protected(w, req)

	if !handlerCalled {
		t.Error("Expected handler to be called")
	}

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "test" {
		t.Errorf("Expected body 'test', got '%s'", w.Body.String())
	}
}

func TestJWTValidatorMiddlewareWithoutAuth(t *testing.T) {
	validator := NewJWTValidator("test-client", "https://invalid-jwks-url.com", "test-scope")

	// Create a test handler
	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("test")); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	})

	// Wrap with middleware
	middleware := validator.Middleware(testHandler)

	// Test request without auth header
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	// Should not call handler due to missing auth
	if handlerCalled {
		t.Error("Expected handler to not be called due to missing auth")
	}

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestJWTValidatorProtectWithoutAuth(t *testing.T) {
	validator := NewJWTValidator("test-client", "https://invalid-jwks-url.com", "test-scope")

	// Create a test handler
	handlerCalled := false
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("test")); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}

	// Wrap with protect
	protected := validator.Protect(testHandler)

	// Test request without auth header
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	protected(w, req)

	// Should not call handler due to missing auth
	if handlerCalled {
		t.Error("Expected handler to not be called due to missing auth")
	}

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestJWTValidatorWithInvalidAuthHeader(t *testing.T) {
	validator := NewJWTValidator("test-client", "https://invalid-jwks-url.com", "test-scope")

	// Create a test handler
	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("test")); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	})

	// Wrap with middleware
	middleware := validator.Middleware(testHandler)

	tests := []struct {
		name       string
		authHeader string
	}{
		{"invalid format", "Invalid"},
		{"basic auth", "Basic dXNlcjpwYXNz"},
		{"bearer without token", "Bearer"},
		{"bearer with invalid token", "Bearer invalid.token.here"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", tt.authHeader)
			w := httptest.NewRecorder()

			middleware.ServeHTTP(w, req)

			// Should not call handler due to invalid auth
			if handlerCalled {
				t.Error("Expected handler to not be called due to invalid auth")
			}

			if w.Code != http.StatusUnauthorized {
				t.Errorf("Expected status 401, got %d", w.Code)
			}
		})
	}
}
