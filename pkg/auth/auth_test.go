package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestNewJWTValidator(t *testing.T) {
	// Test with valid configuration
	config := &JWTConfig{
		ClientID: "test-client",
		JWKSURL:  "https://invalid-jwks-url.com", // Will fail but should return error
		Scope:    "test-scope",
	}

	_, err := NewJWTValidator(config)
	if err == nil {
		t.Error("Expected error for invalid JWKS URL")
	}

	// Test with missing client ID
	config.ClientID = ""
	_, err = NewJWTValidator(config)
	if err == nil {
		t.Error("Expected error for missing client ID")
	}

	// Test with missing JWKS URL
	config.ClientID = "test-client"
	config.JWKSURL = ""
	_, err = NewJWTValidator(config)
	if err == nil {
		t.Error("Expected error for missing JWKS URL")
	}
}

func TestDefaultJWTConfig(t *testing.T) {
	config := DefaultJWTConfig()

	if config.ClientID != "" {
		t.Error("Expected empty client ID in default config")
	}

	if len(config.AllowedAlgs) == 0 {
		t.Error("Expected allowed algorithms in default config")
	}

	if config.CacheTTL == 0 {
		t.Error("Expected cache TTL in default config")
	}

	if config.RefreshInterval == 0 {
		t.Error("Expected refresh interval in default config")
	}
}

func TestPassthroughValidator(t *testing.T) {
	validator := NewPassthroughValidator()

	// Test middleware
	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	middleware := validator.Middleware(testHandler)
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	if !handlerCalled {
		t.Error("Expected handler to be called")
	}

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Test protect
	handlerCalled = false
	protected := validator.Protect(testHandler)
	protected(w, req)

	if !handlerCalled {
		t.Error("Expected protected handler to be called")
	}
}

func TestExtractToken(t *testing.T) {
	validator := &JWTValidator{}

	tests := []struct {
		name        string
		authHeader  string
		expected    string
		description string
	}{
		{
			name:        "valid bearer token",
			authHeader:  "Bearer test-token-123",
			expected:    "test-token-123",
			description: "Should extract token from valid bearer header",
		},
		{
			name:        "bearer token with spaces",
			authHeader:  "Bearer   test-token-456   ",
			expected:    "test-token-456",
			description: "Should handle extra spaces",
		},
		{
			name:        "missing auth header",
			authHeader:  "",
			expected:    "",
			description: "Should return empty string for missing header",
		},
		{
			name:        "invalid format",
			authHeader:  "Invalid",
			expected:    "",
			description: "Should return empty string for invalid format",
		},
		{
			name:        "basic auth",
			authHeader:  "Basic dXNlcjpwYXNz",
			expected:    "",
			description: "Should return empty string for basic auth",
		},
		{
			name:        "bearer without token",
			authHeader:  "Bearer",
			expected:    "",
			description: "Should return empty string for bearer without token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			result := validator.extractToken(req)
			if result != tt.expected {
				t.Errorf("%s: expected '%s', got '%s'", tt.description, tt.expected, result)
			}
		})
	}
}

func TestValidateClaims(t *testing.T) {
	validator := &JWTValidator{
		clientID: "test-client",
		scope:    "test-scope",
	}

	// Use timestamps relative to a fixed time for testing
	baseTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	exp := baseTime.Add(1 * time.Hour)
	iat := baseTime.Add(-1 * time.Hour)

	tests := []struct {
		name        string
		claims      jwt.MapClaims
		expectError bool
		description string
	}{
		{
			name: "valid claims",
			claims: jwt.MapClaims{
				"exp": exp.Unix(),
				"iat": iat.Unix(),
				"aud": "test-client",
				"scp": "test-scope",
			},
			expectError: false,
			description: "Should accept valid claims",
		},
		// Note: Time-based validation tests are skipped due to complexity
		// of mocking time.Now() in Go. These are tested in integration tests.
		{
			name: "invalid audience",
			claims: jwt.MapClaims{
				"exp": exp.Unix(),
				"iat": iat.Unix(),
				"aud": "wrong-client",
				"scp": "test-scope",
			},
			expectError: true,
			description: "Should reject token with wrong audience",
		},
		{
			name: "missing audience",
			claims: jwt.MapClaims{
				"exp": exp.Unix(),
				"iat": iat.Unix(),
				"scp": "test-scope",
			},
			expectError: true,
			description: "Should reject token without audience",
		},
		{
			name: "insufficient scope",
			claims: jwt.MapClaims{
				"exp": exp.Unix(),
				"iat": iat.Unix(),
				"aud": "test-client",
				"scp": "other-scope",
			},
			expectError: true,
			description: "Should reject token with insufficient scope",
		},
		{
			name: "missing scope",
			claims: jwt.MapClaims{
				"exp": exp.Unix(),
				"iat": iat.Unix(),
				"aud": "test-client",
			},
			expectError: true,
			description: "Should reject token without scope",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateClaims(tt.claims)
			if tt.expectError && err == nil {
				t.Errorf("%s: expected error but got none", tt.description)
			}
			if !tt.expectError && err != nil {
				t.Errorf("%s: unexpected error: %v", tt.description, err)
			}
		})
	}
}

func TestTokenRevocation(t *testing.T) {
	validator := &JWTValidator{
		revokedTokens: make(map[string]time.Time),
	}

	token := "test-token-123"

	// Initially not revoked
	if validator.isTokenRevoked(token) {
		t.Error("Token should not be revoked initially")
	}

	// Revoke token
	validator.RevokeToken(token)

	// Should be revoked
	if !validator.isTokenRevoked(token) {
		t.Error("Token should be revoked after revocation")
	}
}

func TestGetUserIDFromContext(t *testing.T) {
	tests := []struct {
		name     string
		claims   jwt.MapClaims
		expected string
		found    bool
	}{
		{
			name: "sub claim",
			claims: jwt.MapClaims{
				"sub": "user123",
			},
			expected: "user123",
			found:    true,
		},
		{
			name: "user_id claim",
			claims: jwt.MapClaims{
				"user_id": "user456",
			},
			expected: "user456",
			found:    true,
		},
		{
			name: "uid claim",
			claims: jwt.MapClaims{
				"uid": "user789",
			},
			expected: "user789",
			found:    true,
		},
		{
			name: "userid claim",
			claims: jwt.MapClaims{
				"userid": "user101",
			},
			expected: "user101",
			found:    true,
		},
		{
			name: "no user id claims",
			claims: jwt.MapClaims{
				"other": "value",
			},
			expected: "",
			found:    false,
		},
		{
			name:     "no claims",
			claims:   nil,
			expected: "",
			found:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.claims != nil {
				ctx = context.WithValue(ctx, JWTClaimsKey, tt.claims)
			}

			userID, found := GetUserIDFromContext(ctx)
			if found != tt.found {
				t.Errorf("Expected found=%v, got %v", tt.found, found)
			}
			if found && userID != tt.expected {
				t.Errorf("Expected user ID '%s', got '%s'", tt.expected, userID)
			}
		})
	}
}

func TestGetClaimsFromContext(t *testing.T) {
	claims := jwt.MapClaims{
		"sub": "user123",
		"aud": "test-client",
	}

	ctx := context.WithValue(context.Background(), JWTClaimsKey, claims)

	retrievedClaims, found := GetClaimsFromContext(ctx)
	if !found {
		t.Error("Expected to find claims in context")
	}

	if retrievedClaims["sub"] != "user123" {
		t.Error("Expected to retrieve correct claims")
	}

	// Test with no claims
	emptyCtx := context.Background()
	_, found = GetClaimsFromContext(emptyCtx)
	if found {
		t.Error("Expected not to find claims in empty context")
	}
}

func TestSendUnauthorizedResponse(t *testing.T) {
	validator := &JWTValidator{}

	w := httptest.NewRecorder()
	validator.sendUnauthorizedResponse(w, "TEST_ERROR", "Test error message")

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}

	wwwAuth := w.Header().Get("WWW-Authenticate")
	if wwwAuth != "Bearer error=\"TEST_ERROR\"" {
		t.Errorf("Expected WWW-Authenticate header, got %s", wwwAuth)
	}
}
