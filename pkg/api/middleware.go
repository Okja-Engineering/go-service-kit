package api

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/cors"
	"golang.org/x/time/rate"
)

type contextKey string

// RateLimiterConfig holds configuration for rate limiting
type RateLimiterConfig struct {
	RequestsPerSecond float64
	Burst             int
	Window            time.Duration
}

// DefaultRateLimiterConfig provides sensible defaults
func DefaultRateLimiterConfig() *RateLimiterConfig {
	return &RateLimiterConfig{
		RequestsPerSecond: 10.0,
		Burst:             20,
		Window:            1 * time.Minute,
	}
}

// RateLimitOption is a functional option for configuring rate limiting
type RateLimitOption func(*RateLimiterConfig)

// WithRequestsPerSecond sets the requests per second limit
func WithRequestsPerSecond(rps float64) RateLimitOption {
	return func(config *RateLimiterConfig) {
		config.RequestsPerSecond = rps
	}
}

// WithBurst sets the burst limit
func WithBurst(burst int) RateLimitOption {
	return func(config *RateLimiterConfig) {
		config.Burst = burst
	}
}

// WithWindow sets the time window for rate limiting
func WithWindow(window time.Duration) RateLimitOption {
	return func(config *RateLimiterConfig) {
		config.Window = window
	}
}

// NewRateLimiterConfig creates a new rate limiter config with options
func NewRateLimiterConfig(options ...RateLimitOption) *RateLimiterConfig {
	config := DefaultRateLimiterConfig()
	for _, option := range options {
		option(config)
	}
	return config
}

// rateLimiter holds rate limiting state
type rateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	config   *RateLimiterConfig
}

// newRateLimiter creates a new rate limiter instance
func newRateLimiter(config *RateLimiterConfig) *rateLimiter {
	return &rateLimiter{
		limiters: make(map[string]*rate.Limiter),
		config:   config,
	}
}

// getLimiter returns or creates a rate limiter for the given key
func (rl *rateLimiter) getLimiter(key string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[key]
	if !exists {
		limiter = rate.NewLimiter(rate.Limit(rl.config.RequestsPerSecond), rl.config.Burst)
		rl.limiters[key] = limiter
	}

	return limiter
}

// cleanup removes old limiters to prevent memory leaks
func (rl *rateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Simple cleanup - in production you might want more sophisticated cleanup
	if len(rl.limiters) > 1000 {
		rl.limiters = make(map[string]*rate.Limiter)
	}
}

// RateLimitByIP creates middleware that rate limits by IP address
func (b *Base) RateLimitByIP(config *RateLimiterConfig) func(next http.Handler) http.Handler {
	if config == nil {
		config = DefaultRateLimiterConfig()
	}

	limiter := newRateLimiter(config)

	// Start cleanup goroutine
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			limiter.cleanup()
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get client IP
			clientIP := getClientIP(r)

			// Get rate limiter for this IP
			ipLimiter := limiter.getLimiter(clientIP)

			// Check if request is allowed
			if !ipLimiter.Allow() {
				log.Printf("### 🚫 Rate limit exceeded for IP: %s", clientIP)
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-RateLimit-Limit", "10")
				w.Header().Set("X-RateLimit-Remaining", "0")
				w.Header().Set("X-RateLimit-Reset", time.Now().Add(time.Second).Format(time.RFC3339))
				w.WriteHeader(http.StatusTooManyRequests)
				if err := json.NewEncoder(w).Encode(map[string]string{
					"error": "Rate limit exceeded. Please try again later.",
				}); err != nil {
					log.Printf("### 🚫 Error encoding rate limit response: %v", err)
				}
				return
			}

			// Add rate limit headers
			w.Header().Set("X-RateLimit-Limit", "10")
			w.Header().Set("X-RateLimit-Remaining", "9") // Simplified
			w.Header().Set("X-RateLimit-Reset", time.Now().Add(time.Second).Format(time.RFC3339))

			next.ServeHTTP(w, r)
		})
	}
}

// RateLimitByToken creates middleware that rate limits by JWT token or API key
func (b *Base) RateLimitByToken(config *RateLimiterConfig) func(next http.Handler) http.Handler {
	if config == nil {
		config = DefaultRateLimiterConfig()
	}

	limiter := newRateLimiter(config)

	// Start cleanup goroutine
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			limiter.cleanup()
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get token from Authorization header
			token := getTokenFromRequest(r)
			if token == "" {
				// No token provided, continue without rate limiting
				next.ServeHTTP(w, r)
				return
			}

			// Get rate limiter for this token
			tokenLimiter := limiter.getLimiter(token)

			// Check if request is allowed
			if !tokenLimiter.Allow() {
				log.Printf("### 🚫 Rate limit exceeded for token: %s", maskToken(token))
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-RateLimit-Limit", "10")
				w.Header().Set("X-RateLimit-Remaining", "0")
				w.Header().Set("X-RateLimit-Reset", time.Now().Add(time.Second).Format(time.RFC3339))
				w.WriteHeader(http.StatusTooManyRequests)
				if err := json.NewEncoder(w).Encode(map[string]string{
					"error": "Rate limit exceeded. Please try again later.",
				}); err != nil {
					log.Printf("### 🚫 Error encoding rate limit response: %v", err)
				}
				return
			}

			// Add rate limit headers
			w.Header().Set("X-RateLimit-Limit", "10")
			w.Header().Set("X-RateLimit-Remaining", "9") // Simplified
			w.Header().Set("X-RateLimit-Reset", time.Now().Add(time.Second).Format(time.RFC3339))

			next.ServeHTTP(w, r)
		})
	}
}

// RateLimitByUserID creates middleware that rate limits by user ID from JWT
func (b *Base) RateLimitByUserID(config *RateLimiterConfig) func(next http.Handler) http.Handler {
	if config == nil {
		config = DefaultRateLimiterConfig()
	}

	limiter := newRateLimiter(config)

	// Start cleanup goroutine
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			limiter.cleanup()
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract user ID from JWT
			userID := getUserIDFromJWT(r)
			if userID == "" {
				// No user ID found, continue without rate limiting
				next.ServeHTTP(w, r)
				return
			}

			// Get rate limiter for this user
			userLimiter := limiter.getLimiter("user:" + userID)

			// Check if request is allowed
			if !userLimiter.Allow() {
				log.Printf("### 🚫 Rate limit exceeded for user: %s", userID)
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-RateLimit-Limit", "10")
				w.Header().Set("X-RateLimit-Remaining", "0")
				w.Header().Set("X-RateLimit-Reset", time.Now().Add(time.Second).Format(time.RFC3339))
				w.WriteHeader(http.StatusTooManyRequests)
				if err := json.NewEncoder(w).Encode(map[string]string{
					"error": "Rate limit exceeded. Please try again later.",
				}); err != nil {
					log.Printf("### 🚫 Error encoding rate limit response: %v", err)
				}
				return
			}

			// Add rate limit headers
			w.Header().Set("X-RateLimit-Limit", "10")
			w.Header().Set("X-RateLimit-Remaining", "9") // Simplified
			w.Header().Set("X-RateLimit-Reset", time.Now().Add(time.Second).Format(time.RFC3339))

			next.ServeHTTP(w, r)
		})
	}
}

// Helper functions

func getClientIP(r *http.Request) string {
	// Check for forwarded headers first
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		if commaIdx := strings.Index(ip, ","); commaIdx != -1 {
			return strings.TrimSpace(ip[:commaIdx])
		}
		return strings.TrimSpace(ip)
	}

	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return strings.TrimSpace(ip)
	}

	if ip := r.Header.Get("X-Client-IP"); ip != "" {
		return strings.TrimSpace(ip)
	}

	// Fall back to remote address
	ip := r.RemoteAddr
	if colonIdx := strings.LastIndex(ip, ":"); colonIdx != -1 {
		ip = ip[:colonIdx]
	}
	return ip
}

func getTokenFromRequest(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	parts := strings.Fields(authHeader)
	if len(parts) != 2 {
		return ""
	}

	if strings.ToLower(parts[0]) != "bearer" {
		return ""
	}

	return parts[1]
}

func getUserIDFromJWT(r *http.Request) string {
	token := getTokenFromRequest(r)
	if token == "" {
		return ""
	}

	userID, err := getClaimFromJWT(token, "sub")
	if err != nil {
		// Try alternative claim names
		userID, err = getClaimFromJWT(token, "user_id")
		if err != nil {
			userID, _ = getClaimFromJWT(token, "uid")
		}
	}

	return userID
}

func maskToken(token string) string {
	if len(token) <= 8 {
		return "***"
	}
	return token[:4] + "..." + token[len(token)-4:]
}

func (b *Base) JWTRequestEnricher(fieldName string, claim string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if len(authHeader) == 0 {
				next.ServeHTTP(w, r)

				return
			}

			authParts := strings.Split(authHeader, " ")
			if len(authParts) != 2 {
				next.ServeHTTP(w, r)

				return
			}

			if strings.ToLower(authParts[0]) != "bearer" {
				next.ServeHTTP(w, r)

				return
			}

			value, err := getClaimFromJWT(authParts[1], claim)
			if err != nil {
				next.ServeHTTP(w, r)

				return
			}

			ctx := context.WithValue(r.Context(), contextKey(fieldName), value)
			next.ServeHTTP(w, r.WithContext(ctx))
		}

		return http.HandlerFunc(fn)
	}
}

func (b *Base) SimpleCORSMiddleware(next http.Handler) http.Handler {
	log.Printf("### 🎭 API: configured simple CORS")

	cors := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	})

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cors.Handler(next).ServeHTTP(w, r)
	})
}

func getClaimFromJWT(jwtRaw string, claimName string) (string, error) {
	jwtParts := strings.Split(jwtRaw, ".")

	tokenBytes, err := base64.RawURLEncoding.DecodeString(jwtParts[1])
	if err != nil {
		log.Println("### Auth: Error in base64 decoding token", err)
		return "", err
	}

	var tokenJSON map[string]interface{}

	err = json.Unmarshal(tokenBytes, &tokenJSON)
	if err != nil {
		log.Println("### Auth: Error in JSON parsing token", err)
		return "", err
	}

	claim, ok := tokenJSON[claimName]
	if !ok {
		log.Println("### Auth: Claim not found in token", err)
		return "", err
	}

	return claim.(string), nil
}
