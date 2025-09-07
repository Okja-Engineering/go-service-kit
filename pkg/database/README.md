# Database Package

A production-ready PostgreSQL database package for Go services with connection pooling, health checks, and RLS multitenancy support.

## Features

- **Connection Management**: Efficient connection pooling with configurable limits
- **Health Checks**: Built-in health monitoring for load balancers and monitoring systems
- **RLS Multitenancy**: Simple tenant context switching for Row Level Security
- **Production Ready**: Connection timeouts, SSL support, and graceful shutdown
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

    if err := db.Connect(); err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    defer db.Close()

    // Setup router with health check
    r := chi.NewRouter()
    
    // Add database health check endpoint
    r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
        if err := db.HealthCheck(); err != nil {
            http.Error(w, "Database unhealthy", http.StatusServiceUnavailable)
            return
        }
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
    })

    log.Println("Server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", r))
}
```

## Configuration

### Default Configuration

The package provides secure defaults:

```go
config := database.DefaultConfig()
// Host: localhost
// Port: 5432
// User: postgres
// Database: postgres
// SSLMode: require (secure default)
// MaxOpenConns: 25
// MaxIdleConns: 5
// ConnMaxLifetime: 5 minutes
// ConnMaxIdleTime: 5 minutes
// ConnectTimeout: 10 seconds
// QueryTimeout: 30 seconds
// RLSContextVarName: "app.current_tenant_id"
```

### Custom Configuration

Use functional options for clean configuration:

```go
db := database.NewPostgreSQLWithOptions(
    database.WithHost("db.example.com"),
    database.WithPort(5432),
    database.WithUser("app_user"),
    database.WithPassword("secure_password"),
    database.WithDatabase("production_db"),
    database.WithSSLMode("require"),
    database.WithMaxOpenConns(50),
    database.WithMaxIdleConns(10),
    database.WithConnMaxLifetime(10*time.Minute),
    database.WithConnMaxIdleTime(5*time.Minute),
    database.WithConnectTimeout(5*time.Second),
    database.WithQueryTimeout(60*time.Second),
    database.WithRLSContextVarName("app.tenant_id"),
)
```

## RLS Multitenancy Support

The database package provides simple Row Level Security (RLS) multitenancy support:

```go
db := database.NewPostgreSQLWithOptions(
    database.WithHost("localhost"),
    database.WithPort(5432),
    database.WithUser("postgres"),
    database.WithPassword("password"),
    database.WithDatabase("myapp"),
    database.WithRLSContextVarName("app.current_tenant_id"),
)

// Set tenant context for the current session
ctx := context.Background()
if err := db.SetTenantContext(ctx, "tenant123"); err != nil {
    log.Printf("Failed to set tenant context: %v", err)
}

// All subsequent queries respect RLS policies
rows, err := db.GetDB().Query("SELECT * FROM users")
// This will only return users for tenant123

// Clear tenant context when done
if err := db.ClearTenantContext(ctx); err != nil {
    log.Printf("Failed to clear tenant context: %v", err)
}
```

### RLS Setup

You'll need to set up RLS policies in your database. Here's an example:

```sql
-- Enable RLS on your tables
ALTER TABLE users ENABLE ROW LEVEL SECURITY;

-- Create a policy that filters by tenant
CREATE POLICY tenant_isolation ON users
    FOR ALL TO PUBLIC
    USING (tenant_id = current_setting('app.current_tenant_id', true));

-- Set the context variable name to match your configuration
-- The package will automatically set this when you call SetTenantContext()
```

## Connection Pool Statistics

Monitor your database connection usage:

```go
stats := db.GetStats()
log.Printf("Open connections: %d, In use: %d, Idle: %d", 
    stats.OpenConnections, stats.InUse, stats.Idle)
log.Printf("Wait count: %d, Wait duration: %v", 
    stats.WaitCount, stats.WaitDuration)
```

## Error Handling

Always check for errors and handle them appropriately:

```go
if err := db.Connect(); err != nil {
    log.Fatalf("Failed to connect to database: %v", err)
}

if err := db.HealthCheck(); err != nil {
    log.Printf("Database health check failed: %v", err)
    // Handle unhealthy database
}

if err := db.SetTenantContext(ctx, "tenant123"); err != nil {
    log.Printf("Failed to set tenant context: %v", err)
    // Handle tenant context error
}
```

## Best Practices

### 1. Connection Management
- Use connection pooling (enabled by default)
- Set appropriate timeouts for your use case
- Always close connections when done
- Monitor connection pool statistics

### 2. RLS Multitenancy
- Always set tenant context before queries
- Clear tenant context when done
- Use consistent tenant ID formats
- Test RLS policies thoroughly

### 3. Security
- Use SSL connections in production (`sslmode=require`)
- Don't hardcode credentials in code
- Use environment variables for configuration
- Validate tenant IDs before setting context

### 4. Performance
- Tune connection pool settings for your workload
- Use appropriate query timeouts
- Monitor connection pool statistics
- Consider connection lifetime settings

## API Reference

### Database Interface

- `Connect() error` - Establish database connection
- `Close() error` - Close database connection
- `GetDB() *sql.DB` - Get underlying sql.DB instance
- `HealthCheck() error` - Check database health
- `GetStats() ConnectionStats` - Get connection pool statistics
- `SetTenantContext(ctx context.Context, tenantID string) error` - Set tenant context for RLS
- `ClearTenantContext(ctx context.Context) error` - Clear tenant context

### Configuration Options

- `WithHost(host string)` - Set database host
- `WithPort(port int)` - Set database port
- `WithUser(user string)` - Set database user
- `WithPassword(password string)` - Set database password
- `WithDatabase(database string)` - Set database name
- `WithSSLMode(sslMode string)` - Set SSL mode
- `WithMaxOpenConns(maxOpenConns int)` - Set max open connections
- `WithMaxIdleConns(maxIdleConns int)` - Set max idle connections
- `WithConnMaxLifetime(connMaxLifetime time.Duration)` - Set connection max lifetime
- `WithConnMaxIdleTime(connMaxIdleTime time.Duration)` - Set connection max idle time
- `WithConnectTimeout(connectTimeout time.Duration)` - Set connection timeout
- `WithQueryTimeout(queryTimeout time.Duration)` - Set query timeout
- `WithRLSContextVarName(varName string)` - Set RLS context variable name

### Types

- `ConnectionStats` - Connection pool statistics
- `TenantContext` - Tenant context information
- `Config` - Database configuration

## Migration Strategy

This package focuses on connection management and RLS support. For database migrations, use PostgreSQL's native tools:

- **pg_migrate**: Popular migration tool for PostgreSQL
- **Flyway**: Database migration tool with PostgreSQL support
- **Raw SQL files**: Version your SQL files and apply them manually
- **Custom migration scripts**: Build your own migration system

The package provides a clean connection interface that works well with any migration strategy.

## Troubleshooting

### Common Issues

1. **Connection Refused**
   - Check host and port
   - Verify PostgreSQL is running
   - Check firewall settings

2. **Authentication Failed**
   - Verify username and password
   - Check PostgreSQL user permissions
   - Ensure user can connect to database

3. **SSL Connection Failed**
   - Check SSL mode setting
   - Verify SSL certificates
   - Consider using `sslmode=disable` for development

4. **Connection Pool Exhausted**
   - Increase MaxOpenConns
   - Check for connection leaks in your application
   - Monitor connection usage patterns

### Debug Mode

Enable debug logging for troubleshooting:

```go
// Set log level for debugging
log.SetLevel(log.DebugLevel)

// The package will log connection events
```

## License

This package is part of the go-service-kit and follows the same license terms.