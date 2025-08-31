# Creating the API

To create an API instance, initialize a Base struct with the necessary information:

## Functions

`StartServer(port int, router chi.Router, timeout time.Duration)`

Starts the API server on the specified port, using the provided chi.Router to handle incoming requests. Configures the server with the specified timeout duration for read and write operations.

`ReturnJSON(resp http.ResponseWriter, data interface{})`

Serializes the provided data as JSON and writes it to the response.

`ReturnText(resp http.ResponseWriter, msg string)`

Writes the provided msg as plain text to the response.

`ReturnErrorJSON(resp http.ResponseWriter, err error)`

Returns an error JSON response with the specified err message.

`ReturnOKJSON(resp http.ResponseWriter)`

Returns an OK JSON response.

## Middleware

### Rate Limiting

The API package provides several rate limiting middleware options:

#### `RateLimitByIP(config *RateLimiterConfig) func(next http.Handler) http.Handler`

Rate limits requests by client IP address. Supports various IP detection methods including X-Forwarded-For, X-Real-IP, and X-Client-IP headers.

```go
// Use default configuration (10 req/sec, burst of 20)
router.Use(api.RateLimitByIP(nil))

// Use custom configuration
config := &api.RateLimiterConfig{
    RequestsPerSecond: 5.0,  // 5 requests per second
    Burst:             10,   // Allow 10 burst requests
    Window:            1 * time.Minute,
}
router.Use(api.RateLimitByIP(config))
```

#### `RateLimitByToken(config *RateLimiterConfig) func(next http.Handler) http.Handler`

Rate limits requests by JWT token or API key from the Authorization header. Requests without tokens pass through without rate limiting.

```go
router.Use(api.RateLimitByToken(nil))
```

#### `RateLimitByUserID(config *RateLimiterConfig) func(next http.Handler) http.Handler`

Rate limits requests by user ID extracted from JWT tokens. Looks for user ID in "sub", "user_id", or "uid" claims.

```go
router.Use(api.RateLimitByUserID(nil))
```

#### Rate Limiter Configuration

```go
type RateLimiterConfig struct {
    RequestsPerSecond float64       // Requests allowed per second
    Burst             int           // Maximum burst size
    Window            time.Duration // Time window for rate limiting
}

// Default configuration
func DefaultRateLimiterConfig() *RateLimiterConfig {
    return &RateLimiterConfig{
        RequestsPerSecond: 10.0,
        Burst:             20,
        Window:            1 * time.Minute,
    }
}
```

### CORS

`SimpleCORSMiddleware(next http.Handler) http.Handler`

Provides CORS support with permissive settings for development.

```go
router.Use(api.SimpleCORSMiddleware)
```

### JWT Enrichment

`JWTRequestEnricher(fieldName string, claim string) func(next http.Handler) http.Handler`

Extracts JWT claims and adds them to the request context.

```go
router.Use(api.JWTRequestEnricher("user_id", "sub"))
```

## Usage

```go
type MyAPI struct {
  *api.Base
  // Add extra fields as per your implementation
}

router := chi.NewRouter()
api := MyAPI{
  api.NewBase(serviceName, version, buildInfo, healthy),
}

// Add middleware
router.Use(api.SimpleCORSMiddleware)
router.Use(api.RateLimitByIP(nil))

api.AddHealthEndpoint(router, "health")
api.AddStatusEndpoint(router, "status")
```

## Rate Limit Headers

When rate limiting is active, the following headers are added to responses:

- `X-RateLimit-Limit`: Maximum requests allowed per window
- `X-RateLimit-Remaining`: Remaining requests in current window
- `X-RateLimit-Reset`: Time when the rate limit resets (RFC3339 format)

When rate limits are exceeded, a 429 Too Many Requests response is returned with a JSON error message.

// See tests for more usage patterns.
