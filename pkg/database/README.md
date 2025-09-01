# Database Package

A production-ready PostgreSQL database package for Go services with connection pooling, health checks, migrations, and comprehensive observability.

## Features

- **Connection Management**: Efficient connection pooling with configurable limits
- **Health Checks**: Built-in health monitoring for load balancers and monitoring systems
- **Migration Support**: Schema versioning with automatic migration tracking
- **Production Ready**: Connection timeouts, retry logic, and graceful shutdown
- **Observability**: Connection pool statistics and metrics
- **Functional Options**: Clean, composable configuration
- **Thread Safe**: Full concurrency support with proper locking

## Quick Start

```go
package main

import (
    "log"
    "net/http"
    
    "github.com/go-chi/chi/v5"
    "github.com/Okja-Engineering/go-service-kit/pkg/database"
    "github.com/Okja-Engineering/go-service-kit/pkg/api"
)

func main() {
    // Create database connection
    db := database.NewPostgreSQLWithOptions(
        database.WithHost("localhost"),
        database.WithPort(5432),
        database.WithUser("postgres"),
        database.WithPassword("password"),
        database.WithDatabase("myapp"),
        database.WithSSLMode("require"),
    )

    // Connect to database
    if err := db.Connect(); err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    defer db.Close()

    // Run migrations
    migrations := []database.Migration{
        {
            Version:     1,
            Description: "Create users table",
            UpSQL:       "CREATE TABLE users (id SERIAL PRIMARY KEY, name TEXT, email TEXT UNIQUE)",
            DownSQL:     "DROP TABLE users",
        },
    }
    
    if err := db.Migrate(migrations); err != nil {
        log.Fatalf("Failed to run migrations: %v", err)
    }

    // Setup router with health check
    r := chi.NewRouter()
    
    // Add database health check endpoint
    r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
        if err := db.HealthCheck(); err != nil {
            http.Error(w, "Database unhealthy", http.StatusServiceUnavailable)
            return
        }
        w.WriteHeader(http.StatusOK)
    })

    // Start server
    log.Fatal(http.ListenAndServe(":8080", r))
}
```

## Configuration

### Default Configuration

The package provides secure defaults optimized for production:

```go
config := database.DefaultConfig()
// Host: "localhost"
// Port: 5432
// User: "postgres"
// Password: ""
// Database: "postgres"
// SSLMode: "require"
// MaxOpenConns: 25
// MaxIdleConns: 5
// ConnMaxLifetime: 5m
// ConnMaxIdleTime: 5m
// ConnectTimeout: 10s
// QueryTimeout: 30s
// RetryAttempts: 3
// RetryDelay: 1s
```

### Custom Configuration

Use functional options for clean, composable configuration:

```go
db := database.NewPostgreSQLWithOptions(
    database.WithHost("db.example.com"),
    database.WithPort(5432),
    database.WithUser("app_user"),
    database.WithPassword("secure_password"),
    database.WithDatabase("production_db"),
    database.WithSSLMode("require"),
    database.WithMaxOpenConns(100),
    database.WithMaxIdleConns(20),
    database.WithConnMaxLifetime(10 * time.Minute),
    database.WithConnMaxIdleTime(10 * time.Minute),
    database.WithConnectTimeout(5 * time.Second),
    database.WithQueryTimeout(60 * time.Second),
    database.WithRetryAttempts(5),
    database.WithRetryDelay(2 * time.Second),
)
```

## Usage Examples

### Basic Database Operations

```go
// Get the underlying sql.DB for direct operations
sqlDB := db.GetDB()

// Execute a query
rows, err := sqlDB.Query("SELECT id, name FROM users WHERE active = $1", true)
if err != nil {
    log.Printf("Query failed: %v", err)
    return
}
defer rows.Close()

// Process results
for rows.Next() {
    var id int
    var name string
    if err := rows.Scan(&id, &name); err != nil {
        log.Printf("Row scan failed: %v", err)
        continue
    }
    log.Printf("User %d: %s", id, name)
}
```

### Health Monitoring

```go
// Check database health
if err := db.HealthCheck(); err != nil {
    log.Printf("Database health check failed: %v", err)
    // Handle unhealthy state
}

// Get connection pool statistics
stats := db.GetStats()
log.Printf("Open connections: %d, In use: %d, Idle: %d", 
    stats.OpenConnections, stats.InUse, stats.Idle)
log.Printf("Wait count: %d, Wait duration: %v", 
    stats.WaitCount, stats.WaitDuration)
```

### Database Migrations

```go
migrations := []database.Migration{
    {
        Version:     1,
        Description: "Create users table",
        UpSQL:       "CREATE TABLE users (id SERIAL PRIMARY KEY, name TEXT, email TEXT UNIQUE)",
        DownSQL:     "DROP TABLE users",
    },
    {
        Version:     2,
        Description: "Add user roles",
        UpSQL:       "ALTER TABLE users ADD COLUMN role TEXT DEFAULT 'user'",
        DownSQL:     "ALTER TABLE users DROP COLUMN role",
    },
    {
        Version:     3,
        Description: "Create sessions table",
        UpSQL:       "CREATE TABLE sessions (id UUID PRIMARY KEY, user_id INTEGER REFERENCES users(id), expires_at TIMESTAMP)",
        DownSQL:     "DROP TABLE sessions",
    },
}

if err := db.Migrate(migrations); err != nil {
    log.Fatalf("Migration failed: %v", err)
}
```

### Integration with API Health Checks

```go
package main

import (
    "net/http"
    
    "github.com/go-chi/chi/v5"
    "github.com/Okja-Engineering/go-service-kit/pkg/api"
    "github.com/Okja-Engineering/go-service-kit/pkg/database"
)

func main() {
    db := database.NewPostgreSQLWithOptions(
        database.WithHost("localhost"),
        database.WithPort(5432),
        database.WithUser("postgres"),
        database.WithPassword("password"),
        database.WithDatabase("myapp"),
    )

    if err := db.Connect(); err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    defer db.Close()

    r := chi.NewRouter()
    
    // Add health endpoint with database check
    r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
        if err := db.HealthCheck(); err != nil {
            api.ReturnErrorJSON(w, "Database unhealthy", http.StatusServiceUnavailable)
            return
        }
        
        stats := db.GetStats()
        api.ReturnJSON(w, map[string]interface{}{
            "status": "healthy",
            "database": map[string]interface{}{
                "status": "connected",
                "connections": map[string]interface{}{
                    "open": stats.OpenConnections,
                    "in_use": stats.InUse,
                    "idle": stats.Idle,
                },
            },
        }, http.StatusOK)
    })

    log.Fatal(http.ListenAndServe(":8080", r))
}
```

### Graceful Shutdown

```go
package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
    
    "github.com/Okja-Engineering/go-service-kit/pkg/database"
)

func main() {
    db := database.NewPostgreSQLWithOptions(
        database.WithHost("localhost"),
        database.WithPort(5432),
        database.WithUser("postgres"),
        database.WithPassword("password"),
        database.WithDatabase("myapp"),
    )

    if err := db.Connect(); err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }

    server := &http.Server{
        Addr:    ":8080",
        Handler: setupRouter(db),
    }

    // Start server in goroutine
    go func() {
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Server failed: %v", err)
        }
    }()

    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("Shutting down server...")

    // Graceful shutdown with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := server.Shutdown(ctx); err != nil {
        log.Fatalf("Server forced to shutdown: %v", err)
    }

    // Close database connection
    if err := db.Close(); err != nil {
        log.Printf("Error closing database: %v", err)
    }

    log.Println("Server exited gracefully")
}
```

## Best Practices

### Connection Pool Sizing

- **MaxOpenConns**: Set to 2-4x the number of CPU cores for most applications
- **MaxIdleConns**: Keep some connections warm, typically 25-50% of MaxOpenConns
- **ConnMaxLifetime**: Rotate connections every 5-15 minutes to prevent stale connections

```go
db := database.NewPostgreSQLWithOptions(
    database.WithMaxOpenConns(50),    // 2-4x CPU cores
    database.WithMaxIdleConns(25),    // 50% of max open
    database.WithConnMaxLifetime(10 * time.Minute),
    database.WithConnMaxIdleTime(5 * time.Minute),
)
```

### SSL Configuration

- **Production**: Always use `require` or `verify-full`
- **Development**: Can use `disable` for local development
- **Testing**: Use `disable` for test databases

```go
// Production
database.WithSSLMode("require")

// Development
database.WithSSLMode("disable")
```

### Timeout Configuration

- **ConnectTimeout**: 5-10 seconds for most networks
- **QueryTimeout**: 30-60 seconds for complex queries
- **Context Timeouts**: Always use context with timeouts for queries

```go
db := database.NewPostgreSQLWithOptions(
    database.WithConnectTimeout(5 * time.Second),
    database.WithQueryTimeout(30 * time.Second),
)
```

### Migration Best Practices

1. **Version Numbers**: Use sequential integers starting from 1
2. **Descriptions**: Clear, concise descriptions of what the migration does
3. **Rollback**: Always provide DownSQL for rollback capability
4. **Testing**: Test migrations on staging before production
5. **Backup**: Always backup before running migrations in production

```go
migrations := []database.Migration{
    {
        Version:     1,
        Description: "Create users table with email and name",
        UpSQL:       "CREATE TABLE users (id SERIAL PRIMARY KEY, email TEXT UNIQUE NOT NULL, name TEXT NOT NULL)",
        DownSQL:     "DROP TABLE users",
    },
    {
        Version:     2,
        Description: "Add user roles and active status",
        UpSQL:       "ALTER TABLE users ADD COLUMN role TEXT DEFAULT 'user', ADD COLUMN active BOOLEAN DEFAULT true",
        DownSQL:     "ALTER TABLE users DROP COLUMN role, DROP COLUMN active",
    },
}
```

### Error Handling

Always check for errors and handle them appropriately:

```go
// Connect with retry logic
var err error
for i := 0; i < 3; i++ {
    if err = db.Connect(); err == nil {
        break
    }
    log.Printf("Connection attempt %d failed: %v", i+1, err)
    time.Sleep(time.Duration(i+1) * time.Second)
}

if err != nil {
    log.Fatalf("Failed to connect after retries: %v", err)
}

// Health check in health endpoint
if err := db.HealthCheck(); err != nil {
    log.Printf("Health check failed: %v", err)
    // Return unhealthy status
    return
}
```

## Monitoring and Observability

### Health Check Endpoint

```go
r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
    if err := db.HealthCheck(); err != nil {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusServiceUnavailable)
        json.NewEncoder(w).Encode(map[string]interface{}{
            "status": "unhealthy",
            "error":  err.Error(),
            "timestamp": time.Now().UTC(),
        })
        return
    }

    stats := db.GetStats()
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "healthy",
        "database": map[string]interface{}{
            "status": "connected",
            "connections": stats,
        },
        "timestamp": time.Now().UTC(),
    })
})
```

### Metrics Collection

```go
// Get connection pool statistics
stats := db.GetStats()

// Log metrics for monitoring
log.Printf("DB_METRICS open=%d in_use=%d idle=%d wait_count=%d wait_duration=%v",
    stats.OpenConnections, stats.InUse, stats.Idle, 
    stats.WaitCount, stats.WaitDuration)

// Send to monitoring system (Prometheus, DataDog, etc.)
// This is just an example - implement according to your monitoring stack
```

## Troubleshooting

### Common Issues

1. **Connection Refused**
   - Check if PostgreSQL is running
   - Verify host and port configuration
   - Check firewall settings

2. **Authentication Failed**
   - Verify username and password
   - Check pg_hba.conf configuration
   - Ensure user has proper permissions

3. **Connection Pool Exhausted**
   - Increase MaxOpenConns
   - Check for connection leaks in your application
   - Monitor connection usage patterns

4. **Migration Failures**
   - Check SQL syntax
   - Verify database permissions
   - Check for conflicting migrations

### Debug Mode

Enable debug logging for troubleshooting:

```go
// Set log level for debugging
log.SetFlags(log.LstdFlags | log.Lshortfile)

// Test connection with verbose logging
if err := db.Connect(); err != nil {
    log.Printf("Connection failed: %v", err)
    // Check configuration
    log.Printf("Config: %+v", db.config)
}
```

## API Reference

### Types

- `Database` - Main interface for database operations
- `PostgreSQL` - PostgreSQL implementation
- `Config` - Database configuration
- `ConnectionStats` - Connection pool statistics
- `Migration` - Database migration definition

### Functions

- `NewPostgreSQL(config *Config) *PostgreSQL` - Create new PostgreSQL instance
- `NewPostgreSQLWithOptions(options ...Option) *PostgreSQL` - Create with options
- `DefaultConfig() *Config` - Get secure default configuration
- `NewConfig(options ...Option) *Config` - Create configuration with options

### Options

- `WithHost(host string)` - Set database host
- `WithPort(port int)` - Set database port
- `WithUser(user string)` - Set database user
- `WithPassword(password string)` - Set database password
- `WithDatabase(database string)` - Set database name
- `WithSSLMode(sslMode string)` - Set SSL mode
- `WithMaxOpenConns(maxOpen int)` - Set maximum open connections
- `WithMaxIdleConns(maxIdle int)` - Set maximum idle connections
- `WithConnMaxLifetime(lifetime time.Duration)` - Set connection max lifetime
- `WithConnMaxIdleTime(idleTime time.Duration)` - Set connection max idle time
- `WithConnectTimeout(timeout time.Duration)` - Set connection timeout
- `WithQueryTimeout(timeout time.Duration)` - Set query timeout
- `WithRetryAttempts(attempts int)` - Set retry attempts
- `WithRetryDelay(delay time.Duration)` - Set retry delay

## Dependencies

- `github.com/lib/pq` - PostgreSQL driver
- Standard library packages: `database/sql`, `context`, `sync`, `time`

## License

MIT License - see LICENSE file for details.
