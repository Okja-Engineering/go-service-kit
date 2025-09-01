# go-service-kit

A minimal framework and starter kit for building API services in Go, focused on using the standard library and idiomatic Go patterns.

[![CI](https://github.com/Okja-Engineering/go-service-kit/actions/workflows/ci.yml/badge.svg)](https://github.com/Okja-Engineering/go-service-kit/actions/workflows/ci.yml)

## Features

- Simple environment variable helpers
- Structured logging with filtering
- REST API helpers (endpoints, middleware)
- Problem+JSON error responses (RFC-7807)
- JWT authentication utilities
- Built-in metrics and health endpoints
- Linting and CI setup

## Installation

### Latest Version
```bash
go get github.com/Okja-Engineering/go-service-kit
```

### Specific Version
```bash
# Latest stable release
go get github.com/Okja-Engineering/go-service-kit@latest

# Specific version (recommended for production)
go get github.com/Okja-Engineering/go-service-kit@v0.2.0

# Latest commit on main (development)
go get github.com/Okja-Engineering/go-service-kit@main
```

### In Your go.mod
```go
require (
    github.com/Okja-Engineering/go-service-kit v0.2.0
)
```

## Getting Started

This repository is a minimal framework and starter kit for building robust, maintainable API services in Go. It is not an application or server by itself, but a set of packages you can use in your own projects. The included tests serve as validation and usage examples for each package.

To validate the packages and see usage in action:

```sh
git clone https://github.com/Okja-Engineering/go-service-kit.git
cd go-service-kit
go mod tidy
make test   # Run all package tests
make lint   # Lint the codebase
```

## Project Structure

```
pkg/
├── api        # API helpers and endpoints ([docs](pkg/api/README.md))
├── auth       # JWT and auth middleware ([docs](pkg/auth/README.md))
├── crypto     # Password hashing and token management ([docs](pkg/crypto/README.md))
├── database   # PostgreSQL connection management ([docs](pkg/database/README.md))
├── env        # Environment variable helpers ([docs](pkg/env/README.md))
├── logging    # Logging utilities ([docs](pkg/logging/README.md))
├── problem    # Problem+JSON error responses ([docs](pkg/problem/README.md))
```

## Usage

This repository is intended as a minimal framework and starter kit for building robust, maintainable API services in Go, with a focus on the standard library and [chi](https://github.com/go-chi/chi) for routing. See each package’s README for details and examples:

- [API](pkg/api/README.md) - HTTP endpoints, middleware, rate limiting
- [Auth](pkg/auth/README.md) - JWT authentication and validation
- [Crypto](pkg/crypto/README.md) - Password hashing, token generation, and validation
- [Database](pkg/database/README.md) - PostgreSQL connection management and migrations
- [Env](pkg/env/README.md) - Environment variable helpers
- [Logging](pkg/logging/README.md) - Structured logging utilities
- [Problem](pkg/problem/README.md) - RFC-7807 Problem+JSON responses

### Quick Example
```go
package main

import (
    "log"
    "net/http"
    
    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    
    "github.com/Okja-Engineering/go-service-kit/pkg/api"
    "github.com/Okja-Engineering/go-service-kit/pkg/auth"
    "github.com/Okja-Engineering/go-service-kit/pkg/database"
    "github.com/Okja-Engineering/go-service-kit/pkg/env"
)

func main() {
    // Load configuration
    port := env.GetString("PORT", "8080")
    
    // Setup database
    db := database.NewPostgreSQLWithOptions(
        database.WithHost(env.GetString("DB_HOST", "localhost")),
        database.WithPort(env.GetInt("DB_PORT", 5432)),
        database.WithUser(env.GetString("DB_USER", "postgres")),
        database.WithPassword(env.GetString("DB_PASSWORD", "")),
        database.WithDatabase(env.GetString("DB_NAME", "myapp")),
    )
    
    if err := db.Connect(); err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    defer db.Close()
    
    // Setup router with middleware
    r := chi.NewRouter()
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    
    // Add health and metrics endpoints
    api.AddHealthEndpoints(r)
    api.AddMetricsEndpoints(r)
    
    // Add database health check
    r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
        if err := db.HealthCheck(); err != nil {
            api.ReturnErrorJSON(w, "Database unhealthy", http.StatusServiceUnavailable)
            return
        }
        api.ReturnOKJSON(w, "Service healthy")
    })
    
    // Start server
    log.Printf("Server starting on port %s", port)
    http.ListenAndServe(":"+port, r)
}
```

<!-- CONTRIBUTING -->

## Contributing

Please see our [contributing guide](/CONTRIBUTING.md).

## License

MIT (see LICENSE)
