package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/MicahParks/keyfunc/v2"
	"github.com/golang-jwt/jwt/v5"
)

// ContextKey is a type-safe key for context values
type ContextKey string

const (
	// JWTClaimsKey is the context key for JWT claims
	JWTClaimsKey ContextKey = "jwt_claims"
)

// JWTValidator provides hardened JWT validation with comprehensive security checks
type JWTValidator struct {
	clientID        string
	scope           string
	jwks            *keyfunc.JWKS
	allowedAlgs     []string
	tokenCache      map[string]*CachedToken
	tokenCacheMutex sync.RWMutex
	cacheTTL        time.Duration
	revokedTokens   map[string]time.Time
	revokedMutex    sync.RWMutex
}

// CachedToken represents a cached validated token
type CachedToken struct {
	Claims    jwt.MapClaims
	ExpiresAt time.Time
	Validated time.Time
}

// ValidationResult provides detailed validation information
type ValidationResult struct {
	Valid     bool
	Claims    jwt.MapClaims
	Error     string
	ErrorCode string
}

// JWTConfig holds configuration for JWT validation
type JWTConfig struct {
	ClientID        string
	JWKSURL         string
	Scope           string
	AllowedAlgs     []string
	CacheTTL        time.Duration
	RefreshInterval time.Duration
}

// DefaultJWTConfig provides secure defaults
func DefaultJWTConfig() *JWTConfig {
	return &JWTConfig{
		AllowedAlgs:     []string{"RS256", "RS384", "RS512", "ES256", "ES384", "ES512"},
		CacheTTL:        5 * time.Minute,
		RefreshInterval: 1 * time.Hour,
	}
}

// NewJWTValidator creates a new hardened JWT validator
func NewJWTValidator(config *JWTConfig) (*JWTValidator, error) {
	if config == nil {
		config = DefaultJWTConfig()
	}

	// Validate required fields
	if config.ClientID == "" {
		return nil, fmt.Errorf("client ID is required")
	}
	if config.JWKSURL == "" {
		return nil, fmt.Errorf("JWKS URL is required")
	}

	// Fetch JWKS
	jwks, err := keyfunc.Get(config.JWKSURL, keyfunc.Options{
		RefreshInterval: config.RefreshInterval,
		RefreshErrorHandler: func(err error) {
			log.Printf("### 🔐 Auth: JWKS refresh error: %v", err)
		},
		RefreshUnknownKID: true,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}

	log.Printf("### 🔐 Auth: JWT validation enabled with JWKS from %s", config.JWKSURL)

	return &JWTValidator{
		clientID:      config.ClientID,
		scope:         config.Scope,
		jwks:          jwks,
		allowedAlgs:   config.AllowedAlgs,
		tokenCache:    make(map[string]*CachedToken),
		cacheTTL:      config.CacheTTL,
		revokedTokens: make(map[string]time.Time),
	}, nil
}

// Middleware returns a middleware function that validates JWT tokens
func (v *JWTValidator) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		result := v.ValidateRequest(r)
		if !result.Valid {
			v.sendUnauthorizedResponse(w, result.ErrorCode, result.Error)
			return
		}

		// Add claims to request context
		ctx := context.WithValue(r.Context(), JWTClaimsKey, result.Claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Protect wraps a handler function with JWT validation
func (v *JWTValidator) Protect(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result := v.ValidateRequest(r)
		if !result.Valid {
			v.sendUnauthorizedResponse(w, result.ErrorCode, result.Error)
			return
		}

		// Add claims to request context
		ctx := context.WithValue(r.Context(), JWTClaimsKey, result.Claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// ValidateRequest performs comprehensive JWT validation
func (v *JWTValidator) ValidateRequest(r *http.Request) ValidationResult {
	// Extract token from Authorization header
	tokenString := v.extractToken(r)
	if tokenString == "" {
		return ValidationResult{
			Valid:     false,
			ErrorCode: "MISSING_TOKEN",
			Error:     "Authorization header is required",
		}
	}

	// Check if token is revoked
	if v.isTokenRevoked(tokenString) {
		return ValidationResult{
			Valid:     false,
			ErrorCode: "TOKEN_REVOKED",
			Error:     "Token has been revoked",
		}
	}

	// Check cache first
	if cached := v.getCachedToken(tokenString); cached != nil {
		return ValidationResult{
			Valid:  true,
			Claims: cached.Claims,
		}
	}

	// Parse and validate token
	token, err := jwt.Parse(tokenString, v.jwks.Keyfunc, jwt.WithValidMethods(v.allowedAlgs))
	if err != nil {
		return ValidationResult{
			Valid:     false,
			ErrorCode: "INVALID_TOKEN",
			Error:     fmt.Sprintf("Token validation failed: %v", err),
		}
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return ValidationResult{
			Valid:     false,
			ErrorCode: "INVALID_CLAIMS",
			Error:     "Invalid token claims",
		}
	}

	// Validate claims
	if err := v.validateClaims(claims); err != nil {
		return ValidationResult{
			Valid:     false,
			ErrorCode: "INVALID_CLAIMS",
			Error:     err.Error(),
		}
	}

	// Cache the validated token
	v.cacheToken(tokenString, claims)

	return ValidationResult{
		Valid:  true,
		Claims: claims,
	}
}

// validateClaims performs comprehensive claim validation
func (v *JWTValidator) validateClaims(claims jwt.MapClaims) error {
	if err := v.validateTimeClaims(claims); err != nil {
		return err
	}

	if err := v.validateAudience(claims); err != nil {
		return err
	}

	if err := v.validateScope(claims); err != nil {
		return err
	}

	// Validate issuer (if configured)
	if iss, ok := claims["iss"]; ok {
		// You can add issuer validation here if needed
		_ = iss
	}

	return nil
}

// validateTimeClaims validates time-based claims (exp, iat, nbf)
func (v *JWTValidator) validateTimeClaims(claims jwt.MapClaims) error {
	now := time.Now()

	// Check expiration
	if exp, ok := claims["exp"]; ok {
		if expTime, ok := exp.(float64); ok {
			if time.Unix(int64(expTime), 0).Before(now) {
				return fmt.Errorf("token has expired")
			}
		}
	}

	// Check issued at time
	if iat, ok := claims["iat"]; ok {
		if iatTime, ok := iat.(float64); ok {
			issuedAt := time.Unix(int64(iatTime), 0)
			if issuedAt.After(now.Add(5 * time.Minute)) {
				return fmt.Errorf("token issued in the future")
			}
		}
	}

	// Check not before time
	if nbf, ok := claims["nbf"]; ok {
		if nbfTime, ok := nbf.(float64); ok {
			if time.Unix(int64(nbfTime), 0).After(now) {
				return fmt.Errorf("token not yet valid")
			}
		}
	}

	return nil
}

// validateAudience validates the audience claim
func (v *JWTValidator) validateAudience(claims jwt.MapClaims) error {
	if aud, ok := claims["aud"]; ok {
		audience := aud.(string)
		audience = strings.TrimPrefix(audience, "api://")
		if audience != v.clientID {
			return fmt.Errorf("invalid audience: expected %s, got %s", v.clientID, audience)
		}
		return nil
	}
	return fmt.Errorf("missing audience claim")
}

// validateScope validates the scope claim
func (v *JWTValidator) validateScope(claims jwt.MapClaims) error {
	if v.scope == "" {
		return nil
	}

	if scp, ok := claims["scp"]; ok {
		scope := scp.(string)
		if !strings.Contains(scope, v.scope) {
			return fmt.Errorf("insufficient scope: required %s, got %s", v.scope, scope)
		}
		return nil
	}
	return fmt.Errorf("missing scope claim")
}

// extractToken extracts the JWT token from the Authorization header
func (v *JWTValidator) extractToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	parts := strings.Fields(authHeader)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}

	return parts[1]
}

// sendUnauthorizedResponse sends a proper 401 response with error details
func (v *JWTValidator) sendUnauthorizedResponse(w http.ResponseWriter, errorCode, errorMsg string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("WWW-Authenticate", "Bearer error=\""+errorCode+"\"")
	w.WriteHeader(http.StatusUnauthorized)

	response := map[string]interface{}{
		"error": errorMsg,
		"code":  errorCode,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("### 🔐 Auth: Error encoding error response: %v", err)
	}
}

// cacheToken caches a validated token
func (v *JWTValidator) cacheToken(tokenString string, claims jwt.MapClaims) {
	v.tokenCacheMutex.Lock()
	defer v.tokenCacheMutex.Unlock()

	// Extract expiration time
	var expiresAt time.Time
	if exp, ok := claims["exp"]; ok {
		if expTime, ok := exp.(float64); ok {
			expiresAt = time.Unix(int64(expTime), 0)
		}
	}

	v.tokenCache[tokenString] = &CachedToken{
		Claims:    claims,
		ExpiresAt: expiresAt,
		Validated: time.Now(),
	}
}

// getCachedToken retrieves a cached token if it's still valid
func (v *JWTValidator) getCachedToken(tokenString string) *CachedToken {
	v.tokenCacheMutex.RLock()
	defer v.tokenCacheMutex.RUnlock()

	cached, exists := v.tokenCache[tokenString]
	if !exists {
		return nil
	}

	// Check if cache entry is still valid
	if time.Now().After(cached.Validated.Add(v.cacheTTL)) {
		return nil
	}

	// Check if token has expired
	if !cached.ExpiresAt.IsZero() && time.Now().After(cached.ExpiresAt) {
		return nil
	}

	return cached
}

// isTokenRevoked checks if a token has been revoked
func (v *JWTValidator) isTokenRevoked(tokenString string) bool {
	v.revokedMutex.RLock()
	defer v.revokedMutex.RUnlock()

	revokedAt, exists := v.revokedTokens[tokenString]
	if !exists {
		return false
	}

	// Clean up old revoked tokens (older than 24 hours)
	if time.Since(revokedAt) > 24*time.Hour {
		v.revokedMutex.RUnlock()
		v.revokedMutex.Lock()
		delete(v.revokedTokens, tokenString)
		v.revokedMutex.Unlock()
		v.revokedMutex.RLock()
		return false
	}

	return true
}

// RevokeToken marks a token as revoked
func (v *JWTValidator) RevokeToken(tokenString string) {
	v.revokedMutex.Lock()
	defer v.revokedMutex.Unlock()
	v.revokedTokens[tokenString] = time.Now()
}

// GetClaimsFromContext extracts JWT claims from request context
func GetClaimsFromContext(ctx context.Context) (jwt.MapClaims, bool) {
	claims, ok := ctx.Value(JWTClaimsKey).(jwt.MapClaims)
	return claims, ok
}

// GetUserIDFromContext extracts user ID from JWT claims in context
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	claims, ok := GetClaimsFromContext(ctx)
	if !ok {
		return "", false
	}

	// Try different claim names for user ID
	userIDFields := []string{"sub", "user_id", "uid", "userid"}
	for _, field := range userIDFields {
		if userID, ok := claims[field].(string); ok && userID != "" {
			return userID, true
		}
	}

	return "", false
}

// PassthroughValidator for testing/development
type PassthroughValidator struct{}

func NewPassthroughValidator() *PassthroughValidator {
	return &PassthroughValidator{}
}

func (v *PassthroughValidator) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func (v *PassthroughValidator) Protect(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	}
}
