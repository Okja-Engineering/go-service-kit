# Authentication Middleware

This package provides hardened JWT authentication middleware for Go HTTP servers with comprehensive security features, supporting both JWT validation and passthrough (no-op) validation for testing or open endpoints.

## Features

- **🔐 Comprehensive JWT Validation** - Full RFC 7519 compliance with signature verification
- **⏰ Time-based Security** - Expiration, issued-at, and not-before time validation
- **🎯 Audience & Scope Validation** - Configurable audience and scope checking
- **🚫 Token Revocation** - In-memory token blacklisting with automatic cleanup
- **⚡ Performance Caching** - Configurable token caching to reduce validation overhead
- **🛡️ Algorithm Validation** - Configurable allowed signing algorithms
- **📊 Detailed Error Responses** - Proper HTTP 401 responses with error codes
- **🔄 JWKS Auto-refresh** - Automatic JSON Web Key Set refresh with error handling
- **🔍 Context Integration** - Easy access to JWT claims in request handlers

## Configuration

### JWTConfig

```go
type JWTConfig struct {
    ClientID        string        // Required: Your application's client ID
    JWKSURL         string        // Required: URL to fetch JSON Web Key Set
    Scope           string        // Optional: Required scope for tokens
    AllowedAlgs     []string      // Optional: Allowed signing algorithms
    CacheTTL        time.Duration // Optional: Token cache TTL (default: 5m)
    RefreshInterval time.Duration // Optional: JWKS refresh interval (default: 1h)
}
```

### Default Configuration

```go
func DefaultJWTConfig() *JWTConfig {
    return &JWTConfig{
        AllowedAlgs:     []string{"RS256", "RS384", "RS512", "ES256", "ES384", "ES512"},
        CacheTTL:        5 * time.Minute,
        RefreshInterval: 1 * time.Hour,
    }
}
```

## Usage

### Basic Setup

```go
package main

import (
    "log"
    "net/http"
    "time"

    "github.com/Okja-Engineering/go-service-kit/pkg/auth"
)

func main() {
    // Create JWT configuration
    config := &auth.JWTConfig{
        ClientID: "your-client-id",
        JWKSURL:  "https://your-auth-server/.well-known/jwks.json",
        Scope:    "api:read",
    }

    // Create JWT validator
    validator, err := auth.NewJWTValidator(config)
    if err != nil {
        log.Fatalf("Failed to create JWT validator: %v", err)
    }

    // Create router
    mux := http.NewServeMux()

    // Protected endpoint with middleware
    mux.Handle("/api/protected", validator.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Get user ID from JWT claims
        userID, found := auth.GetUserIDFromContext(r.Context())
        if found {
            w.Write([]byte("Hello, " + userID + "!"))
        } else {
            w.Write([]byte("Hello, authenticated user!"))
        }
    })))

    // Protected endpoint with Protect wrapper
    mux.HandleFunc("/api/another", validator.Protect(func(w http.ResponseWriter, r *http.Request) {
        // Get all claims from context
        claims, found := auth.GetClaimsFromContext(r.Context())
        if found {
            w.Write([]byte("Authenticated with claims"))
        }
    }))

    log.Fatal(http.ListenAndServe(":8080", mux))
}
```

### Advanced Configuration

```go
config := &auth.JWTConfig{
    ClientID:        "my-app",
    JWKSURL:         "https://auth.example.com/.well-known/jwks.json",
    Scope:           "api:write",
    AllowedAlgs:     []string{"RS256", "ES256"},
    CacheTTL:        10 * time.Minute,
    RefreshInterval: 30 * time.Minute,
}

validator, err := auth.NewJWTValidator(config)
```

### Token Revocation

```go
// Revoke a specific token (e.g., on logout)
validator.RevokeToken("eyJhbGciOiJIUzI1NiIs...")

// The token will be rejected in subsequent requests
```

### Development/Testing

```go
// Use passthrough validator for development
validator := auth.NewPassthroughValidator()

// All requests will pass through without validation
```

## Security Features

### 1. Comprehensive Token Validation

- **Signature Verification** - Uses JWKS for public key validation
- **Algorithm Validation** - Configurable allowed signing algorithms
- **Time Validation** - Checks `exp`, `iat`, and `nbf` claims
- **Audience Validation** - Ensures token is intended for your application
- **Scope Validation** - Verifies required permissions

### 2. Token Revocation

- **In-memory Blacklist** - Immediate token revocation
- **Automatic Cleanup** - Old revoked tokens are automatically removed
- **Thread-safe** - Concurrent access is handled safely

### 3. Performance Optimization

- **Token Caching** - Validated tokens are cached to reduce validation overhead
- **JWKS Caching** - Public keys are cached and automatically refreshed
- **Configurable TTL** - Adjust cache duration based on your needs

### 4. Error Handling

- **Detailed Error Codes** - Specific error codes for different failure reasons
- **Proper HTTP Headers** - Includes `WWW-Authenticate` header with error details
- **JSON Error Responses** - Structured error responses for API clients

## Error Codes

| Code | Description |
|------|-------------|
| `MISSING_TOKEN` | Authorization header is missing or invalid |
| `TOKEN_REVOKED` | Token has been revoked |
| `INVALID_TOKEN` | Token signature or format is invalid |
| `INVALID_CLAIMS` | Token claims are invalid (expired, wrong audience, etc.) |

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

The `GetUserIDFromContext` function looks for user ID in these claim fields (in order):
1. `sub` (standard JWT subject claim)
2. `user_id`
3. `uid`
4. `userid`

## Integration with Rate Limiting

The JWT authentication works seamlessly with the rate limiting middleware:

```go
// Apply rate limiting by user ID (extracted from JWT)
router.Use(api.RateLimitByUserID(nil))

// Apply JWT authentication
router.Use(validator.Middleware)

// Both middlewares work together - rate limiting uses the user ID from JWT
```

## Best Practices

1. **Use HTTPS** - Always use HTTPS in production
2. **Configure Scopes** - Use specific scopes for different endpoints
3. **Set Appropriate TTLs** - Balance security with performance
4. **Monitor JWKS Refresh** - Ensure your auth server's JWKS is accessible
5. **Handle Token Revocation** - Implement logout by revoking tokens
6. **Use Strong Algorithms** - Prefer RS256/ES256 over HS256 for API tokens

## Testing

See the test files for comprehensive examples of:
- JWT validation scenarios
- Error handling
- Token revocation
- Context integration
- Configuration validation

// See tests for more usage patterns.
