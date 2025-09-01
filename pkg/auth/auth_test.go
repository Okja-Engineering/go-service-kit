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
		expectError bool
	}{
		{
			name:        "valid bearer token",
			authHeader:  "Bearer test-token",
			expected:    "test-token",
			expectError: false,
		},
		{
			name:        "bearer token with spaces",
			authHeader:  "Bearer  test-token  ",
			expected:    "test-token",
			expectError: false,
		},
		{
			name:        "missing auth header",
			authHeader:  "",
			expected:    "",
			expectError: true,
		},
		{
			name:        "invalid format",
			authHeader:  "Basic dGVzdDp0ZXN0",
			expected:    "",
			expectError: true,
		},
		{
			name:        "bearer without token",
			authHeader:  "Bearer",
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			token := validator.extractToken(req)
			if tt.expectError && token != "" {
				t.Errorf("Expected empty token for error case")
			}
			if !tt.expectError && token == "" {
				t.Errorf("Expected non-empty token")
			}
			if token != tt.expected {
				t.Errorf("Expected token %s, got %s", tt.expected, token)
			}
		})
	}
}

func TestValidateClaims(t *testing.T) {
	validator := &JWTValidator{
		clientID: "test-client",
		scope:    "test-scope",
	}

	tests := []struct {
		name        string
		claims      jwt.MapClaims
		expectError bool
	}{
		{
			name: "valid claims",
			claims: jwt.MapClaims{
				"aud": "test-client",
				"scp": "test-scope",
				"exp": float64(time.Now().Add(1 * time.Hour).Unix()),
				"iat": float64(time.Now().Unix()),
				"nbf": float64(time.Now().Unix()),
			},
			expectError: false,
		},
		{
			name: "invalid audience",
			claims: jwt.MapClaims{
				"aud": "wrong-client",
				"scp": "test-scope",
				"exp": float64(time.Now().Add(1 * time.Hour).Unix()),
			},
			expectError: true,
		},
		{
			name: "missing audience",
			claims: jwt.MapClaims{
				"scp": "test-scope",
				"exp": float64(time.Now().Add(1 * time.Hour).Unix()),
			},
			expectError: true,
		},
		{
			name: "insufficient scope",
			claims: jwt.MapClaims{
				"aud": "test-client",
				"scp": "insufficient-scope",
				"exp": float64(time.Now().Add(1 * time.Hour).Unix()),
			},
			expectError: true,
		},
		{
			name: "missing scope",
			claims: jwt.MapClaims{
				"aud": "test-client",
				"exp": float64(time.Now().Add(1 * time.Hour).Unix()),
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateClaims(tt.claims)
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestTokenRevocation(t *testing.T) {
	validator := &JWTValidator{
		revokedTokens: make(map[string]time.Time),
	}

	token := "test-token"

	// Test token is not revoked initially
	if validator.isTokenRevoked(token) {
		t.Error("Token should not be revoked initially")
	}

	// Revoke token
	validator.RevokeToken(token)

	// Test token is now revoked
	if !validator.isTokenRevoked(token) {
		t.Error("Token should be revoked after revocation")
	}
}

func TestGetUserIDFromContext(t *testing.T) {
	tests := []struct {
		name     string
		claims   jwt.MapClaims
		expected string
	}{
		{
			name: "sub claim",
			claims: jwt.MapClaims{
				"sub": "user123",
			},
			expected: "user123",
		},
		{
			name: "user_id claim",
			claims: jwt.MapClaims{
				"user_id": "user456",
			},
			expected: "user456",
		},
		{
			name: "uid claim",
			claims: jwt.MapClaims{
				"uid": "user789",
			},
			expected: "user789",
		},
		{
			name: "userid claim",
			claims: jwt.MapClaims{
				"userid": "user101",
			},
			expected: "user101",
		},
		{
			name: "no user id claims",
			claims: jwt.MapClaims{
				"other": "value",
			},
			expected: "",
		},
		{
			name:     "no claims",
			claims:   nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.claims != nil {
				ctx = context.WithValue(ctx, JWTClaimsKey, tt.claims)
			}

			userID, _ := GetUserIDFromContext(ctx)
			if userID != tt.expected {
				t.Errorf("Expected user ID %s, got %s", tt.expected, userID)
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
		t.Error("Expected claims to be found in context")
	}

	if retrievedClaims["sub"] != "user123" {
		t.Error("Expected sub claim to match")
	}
}

func TestSendUnauthorizedResponse(t *testing.T) {
	validator := &JWTValidator{}
	w := httptest.NewRecorder()

	validator.sendUnauthorizedResponse(w, "TEST_ERROR", "test error")

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}

	// Check that response contains error message
	body := w.Body.String()
	if body == "" {
		t.Error("Expected non-empty response body")
	}
}

// Test middleware functionality
func TestJWTValidatorMiddleware(t *testing.T) {
	validator := &JWTValidator{
		clientID: "test-client",
		scope:    "test-scope",
	}

	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	middleware := validator.Middleware(testHandler)

	// Test without token
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}

	if handlerCalled {
		t.Error("Handler should not be called without valid token")
	}
}

func TestJWTValidatorProtect(t *testing.T) {
	validator := &JWTValidator{
		clientID: "test-client",
		scope:    "test-scope",
	}

	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	protected := validator.Protect(testHandler)

	// Test without token
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	protected(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}

	if handlerCalled {
		t.Error("Handler should not be called without valid token")
	}
}

func TestJWTValidatorValidateRequest(t *testing.T) {
	validator := &JWTValidator{
		clientID: "test-client",
		scope:    "test-scope",
	}

	req := httptest.NewRequest("GET", "/test", nil)

	result := validator.ValidateRequest(req)

	if result.Valid {
		t.Error("Expected invalid result for request without token")
	}
}

// Test caching functionality
func TestTokenCaching(t *testing.T) {
	validator := &JWTValidator{
		tokenCache: make(map[string]*CachedToken),
		cacheTTL:   5 * time.Minute,
	}

	token := "test-token"
	claims := jwt.MapClaims{"sub": "user123"}

	// Test caching token
	validator.cacheToken(token, claims)

	// Test retrieving cached token
	cachedToken := validator.getCachedToken(token)
	if cachedToken == nil {
		t.Error("Expected cached token to be retrieved")
		return
	}

	if cachedToken.Claims["sub"] != "user123" {
		t.Error("Expected cached claims to match original")
	}

	// Test retrieving non-existent token
	nonExistentToken := validator.getCachedToken("non-existent")
	if nonExistentToken != nil {
		t.Error("Expected nil for non-existent token")
	}
}

// Test error types
func TestValidationError(t *testing.T) {
	err := &ValidationError{Message: "test error"}

	if err.Error() != "validation error []: test error" {
		t.Errorf("Expected error message 'validation error []: test error', got '%s'", err.Error())
	}

	if !IsValidationError(err) {
		t.Error("Expected IsValidationError to return true")
	}
}

func TestConfigurationError(t *testing.T) {
	err := &ConfigurationError{Message: "config error"}

	if err.Error() != "configuration error in : config error" {
		t.Errorf("Expected error message 'configuration error in : config error', got '%s'", err.Error())
	}

	if !IsConfigurationError(err) {
		t.Error("Expected IsConfigurationError to return true")
	}
}

// Test middleware composition
func TestChain(t *testing.T) {
	middleware1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Middleware1", "true")
			next.ServeHTTP(w, r)
		})
	}

	middleware2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Middleware2", "true")
			next.ServeHTTP(w, r)
		})
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	chained := Chain(middleware1, middleware2)(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	chained.ServeHTTP(w, req)

	if w.Header().Get("X-Middleware1") != "true" {
		t.Error("Expected middleware1 header to be set")
	}

	if w.Header().Get("X-Middleware2") != "true" {
		t.Error("Expected middleware2 header to be set")
	}
}

func TestCompose(t *testing.T) {
	handler1 := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Handler1", "true")
			next(w, r)
		}
	}

	handler2 := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Handler2", "true")
			next(w, r)
		}
	}

	composed := Compose(handler1, handler2)

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	composed(handler)(w, req)

	if w.Header().Get("X-Handler1") != "true" {
		t.Error("Expected handler1 header to be set")
	}

	if w.Header().Get("X-Handler2") != "true" {
		t.Error("Expected handler2 header to be set")
	}
}
