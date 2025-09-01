# Logging Package

Flexible HTTP request logging middleware with URL filtering and interface-based design.

## Features

- **Interface-based design** - Custom loggers and URL filters for testing and flexibility
- **Functional configuration** - Clean configuration with functional option pattern
- **URL filtering** - Filter out specific URLs from logging
- **Customizable** - Custom loggers, formatters, and output writers
- **Mock support** - Mock loggers and filters for unit testing

## Quick Start

```go
package main

import (
    "regexp"
    
    "github.com/go-chi/chi/v5"
    "github.com/Okja-Engineering/go-service-kit/pkg/logging"
)

func main() {
    router := chi.NewRouter()
    
    // Create logger with URL filtering
    logger := logging.NewRequestLogger(
        logging.WithRegexFilter(regexp.MustCompile(`/(health|metrics)`)),
        logging.WithNoColor(true),
    )
    
    // Use in your router
    router.Use(logger.Middleware())
    
    // Your routes
    router.Get("/api/users", handleGetUsers)
    router.Get("/health", handleHealth) // This won't be logged
}
```

## Configuration

### Functional Options

```go
// Basic configuration
logger := logging.NewRequestLogger()

// Advanced configuration
logger := logging.NewRequestLogger(
    logging.WithLogger(customLogger),
    logging.WithFormatter(customFormatter),
    logging.WithRegexFilter(regexp.MustCompile(`/health`)),
    logging.WithNoColor(true),
    logging.WithOutput(os.Stderr),
)
```

### Available Options

```go
func WithLogger(logger Logger) LoggingOption
func WithFormatter(formatter middleware.LogFormatter) LoggingOption
func WithURLFilter(filter URLFilter) LoggingOption
func WithRegexFilter(pattern *regexp.Regexp) LoggingOption
func WithNoColor(noColor bool) LoggingOption
func WithOutput(output io.Writer) LoggingOption
```

### Custom Loggers

```go
// Implement the Logger interface
type CustomLogger struct {
    output *bytes.Buffer
}

func (l *CustomLogger) Printf(format string, v ...interface{}) {
    l.output.WriteString(fmt.Sprintf(format, v...))
}

func (l *CustomLogger) Println(v ...interface{}) {
    l.output.WriteString(fmt.Sprintln(v...))
}

// Use custom logger
customLogger := &CustomLogger{output: &bytes.Buffer{}}
logger := logging.NewRequestLogger(logging.WithLogger(customLogger))
```

### Custom URL Filters

```go
// Implement the URLFilter interface
type CustomURLFilter struct {
    excludedPaths []string
}

func (f *CustomURLFilter) ShouldFilter(url string) bool {
    for _, path := range f.excludedPaths {
        if strings.HasPrefix(url, path) {
            return true
        }
    }
    return false
}

// Use custom filter
filter := &CustomURLFilter{excludedPaths: []string{"/health", "/metrics"}}
logger := logging.NewRequestLogger(logging.WithURLFilter(filter))
```

## API Reference

### Core Interfaces

```go
type Logger interface {
    Printf(format string, v ...interface{})
    Println(v ...interface{})
}

type URLFilter interface {
    ShouldFilter(url string) bool
}
```

### Request Logger

```go
type RequestLogger struct {
    config *LoggingConfig
}

func NewRequestLogger(options ...LoggingOption) *RequestLogger
func (rl *RequestLogger) Middleware() func(next http.Handler) http.Handler
```

### Configuration

```go
type LoggingConfig struct {
    Logger     Logger
    Formatter  middleware.LogFormatter
    URLFilter  URLFilter
    NoColor    bool
    Output     io.Writer
}

func DefaultLoggingConfig() *LoggingConfig
func NewLoggingConfig(options ...LoggingOption) *LoggingConfig
```

### Built-in Implementations

```go
// RegexURLFilter implements URLFilter using regex patterns
type RegexURLFilter struct {
    pattern *regexp.Regexp
}

func (f *RegexURLFilter) ShouldFilter(url string) bool
```

## Examples

### Basic Usage

```go
// Simple usage with defaults
logger := logging.NewRequestLogger()
router.Use(logger.Middleware())
```

### Advanced Configuration

```go
// Custom configuration for production
logger := logging.NewRequestLogger(
    logging.WithRegexFilter(regexp.MustCompile(`/(health|metrics|ready)`)),
    logging.WithNoColor(true),
    logging.WithOutput(os.Stderr),
)

router.Use(logger.Middleware())
```

### Custom Formatter

```go
// Custom log formatter
formatter := &middleware.DefaultLogFormatter{
    Logger:  log.New(os.Stdout, "[API] ", log.LstdFlags),
    NoColor: true,
}

logger := logging.NewRequestLogger(
    logging.WithFormatter(formatter),
    logging.WithRegexFilter(regexp.MustCompile(`/health`)),
)

router.Use(logger.Middleware())
```

### Complete Logging Setup

```go
package main

import (
    "log"
    "os"
    "regexp"
    
    "github.com/go-chi/chi/v5"
    "github.com/Okja-Engineering/go-service-kit/pkg/logging"
)

func main() {
    router := chi.NewRouter()
    
    // Create custom formatter
    formatter := &middleware.DefaultLogFormatter{
        Logger:  log.New(os.Stdout, "[API] ", log.LstdFlags),
        NoColor: false,
    }
    
    // Create logger with filtering
    logger := logging.NewRequestLogger(
        logging.WithFormatter(formatter),
        logging.WithRegexFilter(regexp.MustCompile(`/(health|metrics|ready)`)),
        logging.WithNoColor(false),
    )
    
    // Use in router
    router.Use(logger.Middleware())
    
    // Routes
    router.Get("/health", handleHealth)     // Filtered out
    router.Get("/metrics", handleMetrics)   // Filtered out
    router.Get("/api/users", handleUsers)   // Logged
    router.Post("/api/users", handleCreate) // Logged
}
```

## Testing

### Using Mock Loggers

```go
func TestMyFunction(t *testing.T) {
    // Create mock logger
    mockLogger := &MockLogger{output: &bytes.Buffer{}}
    
    // Create logger with mock
    logger := logging.NewRequestLogger(logging.WithLogger(mockLogger))
    
    // Test your function
    middleware := logger.Middleware()
    // ... test middleware
}
```

### Using Mock URL Filters

```go
func TestURLFiltering(t *testing.T) {
    // Create mock filter
    mockFilter := &MockURLFilter{shouldFilter: true}
    
    // Create logger with mock filter
    logger := logging.NewRequestLogger(logging.WithURLFilter(mockFilter))
    
    // Test filtering behavior
    middleware := logger.Middleware()
    // ... test middleware
}
```

### Mock Implementations

```go
type MockLogger struct {
    output *bytes.Buffer
}

func (m *MockLogger) Printf(format string, v ...interface{}) {
    m.output.WriteString(format)
}

func (m *MockLogger) Println(v ...interface{}) {
    m.output.WriteString("log")
}

type MockURLFilter struct {
    shouldFilter bool
}

func (m *MockURLFilter) ShouldFilter(url string) bool {
    return m.shouldFilter
}
```

## Best Practices

1. **Filter noisy endpoints** - Filter out health checks, metrics, and other noisy endpoints
2. **Use appropriate output** - Use stderr for logs in production
3. **Create once, reuse** - Create logger instance once and reuse throughout your application
4. **Use mock loggers for testing** - Test your logging behavior with mock loggers
5. **Configure formatters appropriately** - Use structured logging for production environments
