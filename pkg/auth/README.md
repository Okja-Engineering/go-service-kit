# Authentication Package

Secure JWT authentication middleware with functional configuration and interface-based design.

## Features

- **JWT validation** - RFC 7519 compliant with signature verification and JWKS support
- **Time-based security** - Expiration, issued-at, and not-before validation
- **Audience & scope validation** - Configurable audience and scope checking
- **Token revocation** - In-memory token blacklisting with automatic cleanup
- **Performance caching** - Configurable token caching to reduce validation overhead
- **Interface-based design** - Flexible interfaces for custom token extraction and validation
- **Functional configuration** - Clean configuration with functional option pattern
- **Middleware composition** - Chain and compose middleware for complex scenarios

## Quick Start

```go
package main

import (
    "net/http"
    
    "github.com/Okja-Engineering/go-service-kit/pkg/auth"
)

func main() {
    // Create JWT validator with functional options
    validator := auth.NewJWTValidator(
        auth.WithClientID("your-client-id"),
        auth.WithJWKSURL("https://your-auth-server/.well-known/jwks.json"),
        auth.WithScope("api:read"),
    )
    
    // Protected endpoint
    http.HandleFunc("/api/protected", validator.Protect(func(w http.ResponseWriter, r *http.Request) {
        // Get user ID from JWT claims
        userID, found := auth.GetUserIDFromContext(r.Context())
        if found {
            w.Write([]byte("Hello, " + userID + "!"))
        }
    }))
    
    http.ListenAndServe(":8080", nil)
}
```

## Configuration

### Functional Options

```go
// Basic configuration
validator := auth.NewJWTValidator(
    auth.WithClientID("my-app"),
    auth.WithJWKSURL("https://auth.example.com/.well-known/jwks.json"),
)

// Advanced configuration
validator := auth.NewJWTValidator(
    auth.WithClientID("my-app"),
    auth.WithJWKSURL("https://auth.example.com/.well-known/jwks.json"),
    auth.WithScope("api:write"),
    auth.WithAllowedAlgs([]string{"RS256", "ES256"}),
    auth.WithCacheTTL(10 * time.Minute),
)
```

### Available Options

```go
func WithClientID(clientID string) Option
func WithJWKSURL(jwksURL string) Option
func WithScope(scope string) Option
func WithAllowedAlgs(algs []string) Option
func WithCacheTTL(ttl time.Duration) Option
func WithTokenExtractor(extractor TokenExtractor) Option
func WithClaimsValidator(validator ClaimsValidator) Option
```

## Usage

### Basic Authentication

```go
// Create validator
validator := auth.NewJWTValidator(
    auth.WithClientID("my-app"),
    auth.WithJWKSURL("https://auth.example.com/.well-known/jwks.json"),
)

// Use as middleware
router.Use(validator.Middleware)

// Or protect individual handlers
router.Get("/api/users", validator.Protect(handleGetUsers))
```

### Token Revocation

```go
// Revoke a token (e.g., on logout)
validator.RevokeToken("eyJhbGciOiJIUzI1NiIs...")

// Subsequent requests with this token will be rejected
```

### Development/Testing

```go
// Use passthrough validator for development
validator := auth.NewPassthroughValidator()

// All requests pass through without validation
```

## Context Integration

### Getting User Information

```go
// Get user ID from JWT claims
userID, found := auth.GetUserIDFromContext(r.Context())
if found {
    // Use userID
}

// Get all claims
claims, found := auth.GetClaimsFromContext(r.Context())
if found {
    // Access any claim
    email := claims["email"].(string)
    roles := claims["roles"].([]interface{})
}
```

### Supported User ID Fields

The `GetUserIDFromContext` function looks for user ID in these claim fields:
1. `sub` (standard JWT subject claim)
2. `user_id`
3. `uid`
4. `userid`

## API Reference

### Core Interfaces

```go
type Validator interface {
    Middleware(next http.Handler) http.Handler
    Protect(handler http.HandlerFunc) http.HandlerFunc
    ValidateRequest(r *http.Request) (*jwt.Token, error)
}

type TokenExtractor interface {
    ExtractToken(r *http.Request) string
}

type ClaimsValidator interface {
    ValidateClaims(claims jwt.MapClaims) error
}
```

### Configuration

```go
type JWTConfig struct {
    ClientID        string
    JWKSURL         string
    Scope           string
    AllowedAlgs     []string
    CacheTTL        time.Duration
    RefreshInterval time.Duration
}

func DefaultJWTConfig() *JWTConfig
func NewJWTConfig(options ...Option) *JWTConfig
```

### Functions

```go
func NewJWTValidator(options ...Option) (Validator, error)
func NewPassthroughValidator() Validator
func GetClaimsFromContext(ctx context.Context) (jwt.MapClaims, bool)
func GetUserIDFromContext(ctx context.Context) (string, bool)
func Chain(middlewares ...func(http.Handler) http.Handler) func(http.Handler) http.Handler
func Compose(middlewares ...func(http.Handler) http.Handler) func(http.Handler) http.Handler
```

## Examples

### Complete Authentication Setup

```go
package main

import (
    "net/http"
    "time"
    
    "github.com/go-chi/chi/v5"
    "github.com/Okja-Engineering/go-service-kit/pkg/auth"
)

func main() {
    router := chi.NewRouter()
    
    // Create JWT validator
    validator := auth.NewJWTValidator(
        auth.WithClientID("my-api"),
        auth.WithJWKSURL("https://auth.example.com/.well-known/jwks.json"),
        auth.WithScope("api:read"),
        auth.WithAllowedAlgs([]string{"RS256", "ES256"}),
        auth.WithCacheTTL(5 * time.Minute),
    )
    
    // Public routes
    router.Get("/health", handleHealth)
    
    // Protected routes
    router.Group(func(r chi.Router) {
        r.Use(validator.Middleware)
        r.Get("/api/users", handleGetUsers)
        r.Post("/api/users", handleCreateUser)
        r.Get("/api/users/{id}", handleGetUser)
    })
    
    http.ListenAndServe(":8080", router)
}
```

### Custom Token Extraction

```go
// Custom token extractor for API keys
type APIKeyExtractor struct{}

func (e *APIKeyExtractor) ExtractToken(r *http.Request) string {
    return r.Header.Get("X-API-Key")
}

// Use custom extractor
validator := auth.NewJWTValidator(
    auth.WithClientID("my-app"),
    auth.WithJWKSURL("https://auth.example.com/.well-known/jwks.json"),
    auth.WithTokenExtractor(&APIKeyExtractor{}),
)
```

### Custom Claims Validation

```go
// Custom claims validator
type CustomValidator struct{}

func (v *CustomValidator) ValidateClaims(claims jwt.MapClaims) error {
    // Check for required role
    if roles, ok := claims["roles"].([]interface{}); ok {
        for _, role := range roles {
            if role == "admin" {
                return nil
            }
        }
    }
    return errors.New("admin role required")
}

// Use custom validator
validator := auth.NewJWTValidator(
    auth.WithClientID("my-app"),
    auth.WithJWKSURL("https://auth.example.com/.well-known/jwks.json"),
    auth.WithClaimsValidator(&CustomValidator{}),
)
```

### Middleware Composition

```go
// Chain multiple middleware
authChain := auth.Chain(
    validator.Middleware,
    rateLimiter.Middleware,
    logging.Middleware,
)

router.Use(authChain)
```

## Error Handling

### Error Types

```go
type ValidationError struct {
    Code    string
    Message string
}

type ConfigurationError struct {
    Message string
}

func IsValidationError(err error) bool
func IsConfigurationError(err error) bool
```

### Error Codes

| Code | Description |
|------|-------------|
| `MISSING_TOKEN` | Authorization header is missing or invalid |
| `TOKEN_REVOKED` | Token has been revoked |
| `INVALID_TOKEN` | Token signature or format is invalid |
| `INVALID_CLAIMS` | Token claims are invalid (expired, wrong audience, etc.) |

## Best Practices

1. **Use HTTPS** - Always use HTTPS in production
2. **Configure scopes** - Use specific scopes for different endpoints
3. **Set appropriate TTLs** - Balance security with performance
4. **Monitor JWKS refresh** - Ensure your auth server's JWKS is accessible
5. **Handle token revocation** - Implement logout by revoking tokens
6. **Use strong algorithms** - Prefer RS256/ES256 over HS256 for API tokens
7. **Compose middleware** - Use middleware composition for complex scenarios
