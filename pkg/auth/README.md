# Authentication Middleware

This package provides authentication middleware for Go HTTP servers, supporting both JWT validation and passthrough (no-op) validation for testing or open endpoints.

## Functions

`NewJWTValidator(clientID string, jwksURL string, scope string, logger func(format string, v ...interface{})) (*JWTValidator, error)`

Creates a new JWT validator with the given client ID, JWKS URL, and scope. The logger function is used for logging events and errors.

`NewPassthroughValidator() PassthroughValidator`

Creates a new Passthrough validator for use in tests or when authentication is not required.

`Middleware(next http.Handler) http.Handler`

For both JWTValidator and PassthroughValidator:
- JWTValidator: Validates the request JWT token and passes it to the next handler on success.
- PassthroughValidator: Passes the request to the next handler without any validation.

## Usage

```go
func main() {
	logger := log.Printf

	validator, err := auth.NewJWTValidator(
		"yourClientID",
		"https://yourJwksURL.com",
		"yourScope",
		logger,
	)
	if err != nil {
		log.Fatalf("Failed to create JWT validator: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/secured", validator.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, secured world!"))
	})))

	log.Fatal(http.ListenAndServe(":8080", mux))
}
```

// See tests for more usage patterns.
