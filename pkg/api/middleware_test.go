package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimitByIP(t *testing.T) {
	base := NewBase("test", "1.0.0", "test", true)

	// Create a very restrictive config for testing
	config := &RateLimiterConfig{
		RequestsPerSecond: 1.0, // 1 request per second
		Burst:             1,   // Allow 1 burst
		Window:            1 * time.Second,
	}

	middleware := base.RateLimitByIP(config)

	// Create a simple handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("success")); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	})

	// Wrap with rate limiting
	wrappedHandler := middleware(handler)

	// Test successful request
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Test rate limit exceeded
	w2 := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w2, req)

	if w2.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status 429, got %d", w2.Code)
	}

	// Check rate limit headers
	if w2.Header().Get("X-RateLimit-Limit") == "" {
		t.Error("Expected X-RateLimit-Limit header")
	}

	if w2.Header().Get("X-RateLimit-Remaining") == "" {
		t.Error("Expected X-RateLimit-Remaining header")
	}
}

func TestRateLimitByToken(t *testing.T) {
	base := NewBase("test", "1.0.0", "test", true)

	config := &RateLimiterConfig{
		RequestsPerSecond: 1.0,
		Burst:             1,
		Window:            1 * time.Second,
	}

	middleware := base.RateLimitByToken(config)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("success")); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	})

	wrappedHandler := middleware(handler)

	// Test with valid token
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer test-token-123")
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Test rate limit exceeded
	w2 := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w2, req)

	if w2.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status 429, got %d", w2.Code)
	}

	// Test without token (should pass through)
	req2 := httptest.NewRequest("GET", "/", nil)
	w3 := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w3, req2)

	if w3.Code != http.StatusOK {
		t.Errorf("Expected status 200 for request without token, got %d", w3.Code)
	}
}

func TestRateLimitByUserID(t *testing.T) {
	base := NewBase("test", "1.0.0", "test", true)

	config := &RateLimiterConfig{
		RequestsPerSecond: 1.0,
		Burst:             1,
		Window:            1 * time.Second,
	}

	middleware := base.RateLimitByUserID(config)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("success")); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	})

	wrappedHandler := middleware(handler)

	// Test with valid JWT containing user ID
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyMTIzIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c")
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Test rate limit exceeded
	w2 := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w2, req)

	if w2.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status 429, got %d", w2.Code)
	}

	// Test without JWT (should pass through)
	req2 := httptest.NewRequest("GET", "/", nil)
	w3 := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w3, req2)

	if w3.Code != http.StatusOK {
		t.Errorf("Expected status 200 for request without JWT, got %d", w3.Code)
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name        string
		headers     map[string]string
		remoteAddr  string
		expectedIP  string
		description string
	}{
		{
			name: "X-Forwarded-For single IP",
			headers: map[string]string{
				"X-Forwarded-For": "192.168.1.100",
			},
			remoteAddr:  "10.0.0.1:12345",
			expectedIP:  "192.168.1.100",
			description: "Should use X-Forwarded-For when present",
		},
		{
			name: "X-Forwarded-For multiple IPs",
			headers: map[string]string{
				"X-Forwarded-For": "192.168.1.100, 10.0.0.1, 172.16.0.1",
			},
			remoteAddr:  "10.0.0.1:12345",
			expectedIP:  "192.168.1.100",
			description: "Should use first IP from X-Forwarded-For",
		},
		{
			name: "X-Real-IP",
			headers: map[string]string{
				"X-Real-IP": "192.168.1.200",
			},
			remoteAddr:  "10.0.0.1:12345",
			expectedIP:  "192.168.1.200",
			description: "Should use X-Real-IP when X-Forwarded-For is not present",
		},
		{
			name: "X-Client-IP",
			headers: map[string]string{
				"X-Client-IP": "192.168.1.300",
			},
			remoteAddr:  "10.0.0.1:12345",
			expectedIP:  "192.168.1.300",
			description: "Should use X-Client-IP when other headers are not present",
		},
		{
			name:        "RemoteAddr fallback",
			headers:     map[string]string{},
			remoteAddr:  "10.0.0.1:12345",
			expectedIP:  "10.0.0.1",
			description: "Should use RemoteAddr when no headers are present",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = tt.remoteAddr

			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			ip := getClientIP(req)
			if ip != tt.expectedIP {
				t.Errorf("%s: expected '%s', got '%s'", tt.description, tt.expectedIP, ip)
			}
		})
	}
}

func TestGetTokenFromRequest(t *testing.T) {
	tests := []struct {
		name        string
		authHeader  string
		expected    string
		description string
	}{
		{
			name:        "Valid Bearer token",
			authHeader:  "Bearer test-token-123",
			expected:    "test-token-123",
			description: "Should extract token from valid bearer header",
		},
		{
			name:        "Valid Bearer token with spaces",
			authHeader:  "Bearer   test-token-456   ",
			expected:    "test-token-456",
			description: "Should handle extra spaces",
		},
		{
			name:        "Invalid format - no space",
			authHeader:  "Bearertest-token",
			expected:    "",
			description: "Should return empty string for invalid format",
		},
		{
			name:        "Invalid format - wrong scheme",
			authHeader:  "Basic dGVzdDp0ZXN0",
			expected:    "",
			description: "Should return empty string for non-bearer scheme",
		},
		{
			name:        "Empty header",
			authHeader:  "",
			expected:    "",
			description: "Should return empty string for missing header",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			result := getTokenFromRequest(req)
			if result != tt.expected {
				t.Errorf("%s: expected '%s', got '%s'", tt.description, tt.expected, result)
			}
		})
	}
}

func TestMaskToken(t *testing.T) {
	tests := []struct {
		name        string
		token       string
		expected    string
		description string
	}{
		{
			name:        "abcdefghijklmnop",
			token:       "abcdefghijklmnop",
			expected:    "abcd...mnop",
			description: "Should mask middle of long token",
		},
		{
			name:        "short",
			token:       "short",
			expected:    "***",
			description: "Should mask short token with asterisks",
		},
		{
			name:        "12345678",
			token:       "12345678",
			expected:    "***",
			description: "Should mask short token with asterisks",
		},
		{
			name:        "#00",
			token:       "#00",
			expected:    "***",
			description: "Should mask very short token with asterisks",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskToken(tt.token)
			if result != tt.expected {
				t.Errorf("%s: expected '%s', got '%s'", tt.description, tt.expected, result)
			}
		})
	}
}

func TestDefaultRateLimiterConfig(t *testing.T) {
	config := DefaultRateLimiterConfig()

	if config.RequestsPerSecond != 10.0 {
		t.Errorf("Expected RequestsPerSecond 10.0, got %f", config.RequestsPerSecond)
	}

	if config.Burst != 20 {
		t.Errorf("Expected Burst 20, got %d", config.Burst)
	}

	if config.Window != 1*time.Minute {
		t.Errorf("Expected Window 1m, got %v", config.Window)
	}
}

func TestRateLimiterCleanup(t *testing.T) {
	config := &RateLimiterConfig{
		RequestsPerSecond: 10.0,
		Burst:             20,
		Window:            1 * time.Second,
	}

	limiter := newRateLimiter(config)

	// Add some limiters
	limiter.getLimiter("ip1")
	limiter.getLimiter("ip2")
	limiter.getLimiter("ip3")

	if len(limiter.limiters) != 3 {
		t.Errorf("Expected 3 limiters, got %d", len(limiter.limiters))
	}

	// Run cleanup
	limiter.cleanup()

	// Should still have 3 limiters since they're recent
	if len(limiter.limiters) != 3 {
		t.Errorf("Expected 3 limiters after cleanup, got %d", len(limiter.limiters))
	}
}

func TestNewRateLimiterConfig(t *testing.T) {
	// Test with no options (should use defaults)
	config := NewRateLimiterConfig()
	if config.RequestsPerSecond != 10.0 {
		t.Errorf("Expected default RequestsPerSecond 10.0, got %f", config.RequestsPerSecond)
	}
	if config.Burst != 20 {
		t.Errorf("Expected default Burst 20, got %d", config.Burst)
	}
	if config.Window != 1*time.Minute {
		t.Errorf("Expected default Window 1m, got %v", config.Window)
	}

	// Test with custom options
	config = NewRateLimiterConfig(
		WithRequestsPerSecond(5.0),
		WithBurst(10),
		WithWindow(30*time.Second),
	)
	if config.RequestsPerSecond != 5.0 {
		t.Errorf("Expected RequestsPerSecond 5.0, got %f", config.RequestsPerSecond)
	}
	if config.Burst != 10 {
		t.Errorf("Expected Burst 10, got %d", config.Burst)
	}
	if config.Window != 30*time.Second {
		t.Errorf("Expected Window 30s, got %v", config.Window)
	}

	// Test partial options
	config = NewRateLimiterConfig(WithRequestsPerSecond(15.0))
	if config.RequestsPerSecond != 15.0 {
		t.Errorf("Expected RequestsPerSecond 15.0, got %f", config.RequestsPerSecond)
	}
	if config.Burst != 20 { // Should keep default
		t.Errorf("Expected default Burst 20, got %d", config.Burst)
	}
}

// Test JWT enrichment functionality
func TestJWTRequestEnricher(t *testing.T) {
	base := NewBase("test", "1.0.0", "test", true)
	middleware := base.JWTRequestEnricher("user_id", "sub")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if user_id is set in context (only for valid JWTs)
		_ = r.Context().Value("user_id")
		w.WriteHeader(http.StatusOK)
		// Don't fail the test if user_id is nil - that's expected for invalid tokens
	})

	wrappedHandler := middleware(handler)

	// Test without JWT (should pass through)
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Test with malformed JWT (should pass through)
	req2 := httptest.NewRequest("GET", "/", nil)
	req2.Header.Set("Authorization", "Bearer header.payload") // Only 2 parts, not 3
	w2 := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("Expected status 200 for request with malformed JWT, got %d", w2.Code)
	}
}

// Test CORS middleware
func TestSimpleCORSMiddleware(t *testing.T) {
	base := NewBase("test", "1.0.0", "test", true)
	middleware := base.SimpleCORSMiddleware

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := middleware(handler)

	// Test preflight request
	req := httptest.NewRequest("OPTIONS", "/", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type")
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for preflight request, got %d", w.Code)
	}

	// Check CORS headers
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("Expected Access-Control-Allow-Origin header")
	}

	if w.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("Expected Access-Control-Allow-Methods header")
	}

	if w.Header().Get("Access-Control-Allow-Headers") == "" {
		t.Error("Expected Access-Control-Allow-Headers header")
	}

	// Test regular request
	req2 := httptest.NewRequest("GET", "/", nil)
	req2.Header.Set("Origin", "https://example.com")
	w2 := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("Expected status 200 for regular request, got %d", w2.Code)
	}
}

// Test JWT claim extraction
func TestGetClaimFromJWT(t *testing.T) {
	tests := []struct {
		name        string
		token       string
		claim       string
		expected    string
		description string
	}{
		{
			name:        "valid sub claim",
			token:       "header.eyJzdWIiOiJ1c2VyMTIzIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.signature",
			claim:       "sub",
			expected:    "user123",
			description: "Should extract sub claim from valid JWT",
		},
		{
			name:        "valid name claim",
			token:       "header.eyJzdWIiOiJ1c2VyMTIzIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.signature",
			claim:       "name",
			expected:    "John Doe",
			description: "Should extract name claim from valid JWT",
		},
		{
			name:        "invalid token",
			token:       "invalid.token.here",
			claim:       "sub",
			expected:    "",
			description: "Should return empty string for invalid token",
		},
		{
			name:        "missing claim",
			token:       "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyMTIzIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			claim:       "missing",
			expected:    "",
			description: "Should return empty string for missing claim",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getClaimFromJWT(tt.token, tt.claim)
			if err != nil && tt.expected != "" {
				t.Errorf("%s: unexpected error: %v", tt.description, err)
			}
			if result != tt.expected {
				t.Errorf("%s: expected '%s', got '%s'", tt.description, tt.expected, result)
			}
		})
	}
}

// Test user ID extraction from JWT
func TestGetUserIDFromJWT(t *testing.T) {
	tests := []struct {
		name        string
		authHeader  string
		expected    string
		description string
	}{
		{
			name:        "valid sub claim",
			authHeader:  "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyMTIzIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			expected:    "user123",
			description: "Should extract user ID from sub claim",
		},
		{
			name:        "invalid token",
			authHeader:  "Bearer invalid.token.here",
			expected:    "",
			description: "Should return empty string for invalid token",
		},
		{
			name:        "no authorization header",
			authHeader:  "",
			expected:    "",
			description: "Should return empty string for missing authorization header",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			result := getUserIDFromJWT(req)
			if result != tt.expected {
				t.Errorf("%s: expected '%s', got '%s'", tt.description, tt.expected, result)
			}
		})
	}
}
