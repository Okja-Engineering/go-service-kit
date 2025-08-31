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
	req3 := httptest.NewRequest("GET", "/", nil)
	w3 := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w3, req3)

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

	// Test with JWT containing user ID
	// This is a mock JWT with "sub": "user123" in the payload
	mockJWT := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9." +
		"eyJzdWIiOiJ1c2VyMTIzIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ." +
		"SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+mockJWT)
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
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name       string
		headers    map[string]string
		remoteAddr string
		expected   string
	}{
		{
			name: "X-Forwarded-For single IP",
			headers: map[string]string{
				"X-Forwarded-For": "192.168.1.1",
			},
			expected: "192.168.1.1",
		},
		{
			name: "X-Forwarded-For multiple IPs",
			headers: map[string]string{
				"X-Forwarded-For": "192.168.1.1, 10.0.0.1, 172.16.0.1",
			},
			expected: "192.168.1.1",
		},
		{
			name: "X-Real-IP",
			headers: map[string]string{
				"X-Real-IP": "192.168.1.2",
			},
			expected: "192.168.1.2",
		},
		{
			name: "X-Client-IP",
			headers: map[string]string{
				"X-Client-IP": "192.168.1.3",
			},
			expected: "192.168.1.3",
		},
		{
			name:       "RemoteAddr fallback",
			remoteAddr: "192.168.1.4:12345",
			expected:   "192.168.1.4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)

			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			if tt.remoteAddr != "" {
				req.RemoteAddr = tt.remoteAddr
			}

			result := getClientIP(req)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestGetTokenFromRequest(t *testing.T) {
	tests := []struct {
		name       string
		authHeader string
		expected   string
	}{
		{
			name:       "Valid Bearer token",
			authHeader: "Bearer test-token-123",
			expected:   "test-token-123",
		},
		{
			name:       "Valid Bearer token with spaces",
			authHeader: "Bearer   test-token-456   ",
			expected:   "test-token-456",
		},
		{
			name:       "Invalid format - no space",
			authHeader: "Bearertest-token",
			expected:   "",
		},
		{
			name:       "Invalid format - wrong scheme",
			authHeader: "Basic dGVzdDp0ZXN0",
			expected:   "",
		},
		{
			name:       "Empty header",
			authHeader: "",
			expected:   "",
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
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestMaskToken(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "abcdefghijklmnop",
			expected: "abcd...mnop",
		},
		{
			input:    "short",
			expected: "***",
		},
		{
			input:    "12345678",
			expected: "***",
		},
		{
			input:    "",
			expected: "***",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := maskToken(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestDefaultRateLimiterConfig(t *testing.T) {
	config := DefaultRateLimiterConfig()

	if config.RequestsPerSecond != 10.0 {
		t.Errorf("Expected 10.0 requests per second, got %f", config.RequestsPerSecond)
	}

	if config.Burst != 20 {
		t.Errorf("Expected burst of 20, got %d", config.Burst)
	}

	if config.Window != time.Minute {
		t.Errorf("Expected window of 1 minute, got %v", config.Window)
	}
}

func TestRateLimiterCleanup(t *testing.T) {
	config := &RateLimiterConfig{
		RequestsPerSecond: 10.0,
		Burst:             20,
		Window:            1 * time.Minute,
	}

	limiter := newRateLimiter(config)

	// Add some limiters
	limiter.getLimiter("ip1")
	limiter.getLimiter("ip2")
	limiter.getLimiter("ip3")

	if len(limiter.limiters) != 3 {
		t.Errorf("Expected 3 limiters, got %d", len(limiter.limiters))
	}

	// Test cleanup (should not trigger since we have < 1000 limiters)
	limiter.cleanup()

	if len(limiter.limiters) != 3 {
		t.Errorf("Expected 3 limiters after cleanup, got %d", len(limiter.limiters))
	}
}
