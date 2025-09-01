# API Package

Build REST APIs and microservices with flexible middleware and functional configuration.

## Features

- **Rate limiting** - IP, token, and user-based rate limiting with configurable limits
- **CORS support** - Simple CORS middleware for cross-origin requests
- **JWT enrichment** - Extract and inject JWT claims into request context
- **Health endpoints** - Built-in health and status endpoints
- **Functional configuration** - Clean, composable configuration with functional options

## Quick Start

```go
package main

import (
    "net/http"
    
    "github.com/go-chi/chi/v5"
    "github.com/Okja-Engineering/go-service-kit/pkg/api"
)

func main() {
    router := chi.NewRouter()
    
    // Add rate limiting
    config := api.NewRateLimiterConfig(
        api.WithRequestsPerSecond(10.0),
        api.WithBurst(20),
    )
    router.Use(api.RateLimitByIP(config))
    
    // Add CORS
    router.Use(api.SimpleCORSMiddleware)
    
    // Add health endpoints
    api.AddHealthEndpoints(router)
    api.AddMetricsEndpoints(router)
    
    // Your routes
    router.Get("/api/users", handleGetUsers)
    
    http.ListenAndServe(":8080", router)
}
```

## Rate Limiting

### Configuration

```go
// Create rate limiter config with functional options
config := api.NewRateLimiterConfig(
    api.WithRequestsPerSecond(5.0),    // 5 requests per second
    api.WithBurst(10),                 // Allow bursts of 10
    api.WithWindow(1 * time.Minute),   // 1-minute window
)
```

### Rate Limiting Strategies

#### By IP Address
```go
router.Use(api.RateLimitByIP(config))
```

#### By JWT Token
```go
router.Use(api.RateLimitByToken(config))
```

#### By User ID
```go
router.Use(api.RateLimitByUserID(config))
```

### Rate Limit Headers

Responses include rate limit information:
- `X-RateLimit-Limit`: Maximum requests per window
- `X-RateLimit-Remaining`: Remaining requests
- `X-RateLimit-Reset`: Reset time (RFC3339)

## Middleware

### CORS
```go
router.Use(api.SimpleCORSMiddleware)
```

### JWT Enrichment
```go
// Extract user_id from JWT sub claim
router.Use(api.JWTRequestEnricher("user_id", "sub"))
```

## Health Endpoints

```go
// Add standard health and metrics endpoints
api.AddHealthEndpoints(router)
api.AddMetricsEndpoints(router)
```

## API Reference

### Rate Limiting

```go
type RateLimiterConfig struct {
    RequestsPerSecond float64
    Burst             int
    Window            time.Duration
}

type RateLimitOption func(*RateLimiterConfig)

func NewRateLimiterConfig(options ...RateLimitOption) *RateLimiterConfig
func WithRequestsPerSecond(rps float64) RateLimitOption
func WithBurst(burst int) RateLimitOption
func WithWindow(window time.Duration) RateLimitOption
```

### Middleware Functions

```go
func RateLimitByIP(config *RateLimiterConfig) func(next http.Handler) http.Handler
func RateLimitByToken(config *RateLimiterConfig) func(next http.Handler) http.Handler
func RateLimitByUserID(config *RateLimiterConfig) func(next http.Handler) http.Handler
func SimpleCORSMiddleware(next http.Handler) http.Handler
func JWTRequestEnricher(fieldName string, claim string) func(next http.Handler) http.Handler
```

### Endpoint Functions

```go
func AddHealthEndpoints(router chi.Router)
func AddMetricsEndpoints(router chi.Router)
```

## Examples

### Complete Microservice Setup

```go
package main

import (
    "net/http"
    "time"
    
    "github.com/go-chi/chi/v5"
    "github.com/Okja-Engineering/go-service-kit/pkg/api"
)

func main() {
    router := chi.NewRouter()
    
    // Configure rate limiting
    rateLimitConfig := api.NewRateLimiterConfig(
        api.WithRequestsPerSecond(100.0),
        api.WithBurst(200),
        api.WithWindow(1 * time.Minute),
    )
    
    // Add middleware
    router.Use(api.RateLimitByIP(rateLimitConfig))
    router.Use(api.SimpleCORSMiddleware)
    router.Use(api.JWTRequestEnricher("user_id", "sub"))
    
    // Add health endpoints
    api.AddHealthEndpoints(router)
    api.AddMetricsEndpoints(router)
    
    // API routes
    router.Route("/api/v1", func(r chi.Router) {
        r.Get("/users", handleGetUsers)
        r.Post("/users", handleCreateUser)
        r.Get("/users/{id}", handleGetUser)
    })
    
    http.ListenAndServe(":8080", router)
}
```

### Custom Rate Limiting

```go
// Different limits for different endpoints
publicConfig := api.NewRateLimiterConfig(
    api.WithRequestsPerSecond(50.0),
    api.WithBurst(100),
)

privateConfig := api.NewRateLimiterConfig(
    api.WithRequestsPerSecond(10.0),
    api.WithBurst(20),
)

// Public endpoints
router.Group(func(r chi.Router) {
    r.Use(api.RateLimitByIP(publicConfig))
    r.Get("/public/data", handlePublicData)
})

// Private endpoints
router.Group(func(r chi.Router) {
    r.Use(api.RateLimitByUserID(privateConfig))
    r.Get("/private/profile", handlePrivateProfile)
})
```

## Best Practices

1. **Configure rate limits appropriately** - Balance security with usability
2. **Use user-based limiting for authenticated endpoints** - More precise than IP-based
3. **Add health endpoints** - Essential for monitoring and load balancers
4. **Group related middleware** - Apply different rate limits to different endpoint groups
5. **Monitor rate limit headers** - Use them for client-side rate limiting
