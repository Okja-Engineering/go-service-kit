# Problem Package

RFC-7807 Problem+JSON error response helpers with interface-based design and functional configuration.

## Features

- **RFC-7807 compliant** - Standard Problem+JSON error responses
- **Interface-based design** - Custom loggers for testing and flexibility
- **Functional configuration** - Clean configuration with functional option pattern
- **Error wrapping** - Wrap errors and send as structured JSON responses
- **Mock support** - Mock loggers for unit testing

## Quick Start

```go
package main

import (
    "net/http"
    
    "github.com/Okja-Engineering/go-service-kit/pkg/problem"
)

func handleUsers(w http.ResponseWriter, r *http.Request) {
    // Create problem manager
    pm := problem.NewProblemManager()
    
    // Send error response
    if err := someOperation(); err != nil {
        pm.Send(w, http.StatusBadRequest, "Invalid request", err)
        return
    }
    
    // Wrap and send error
    if err := anotherOperation(); err != nil {
        wrappedErr := pm.Wrap(err, "Failed to process request")
        pm.Send(w, http.StatusInternalServerError, "Internal error", wrappedErr)
        return
    }
}
```

## Configuration

### Functional Options

```go
// Basic configuration
pm := problem.NewProblemManager()

// Advanced configuration
pm := problem.NewProblemManager(
    problem.WithLogger(customLogger),
    problem.WithLogPrefix("[ERROR]"),
    problem.WithLogErrors(true),
)
```

### Available Options

```go
func WithLogger(logger Logger) ProblemOption
func WithLogPrefix(prefix string) ProblemOption
func WithLogErrors(log bool) ProblemOption
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

// Use custom logger
customLogger := &CustomLogger{output: &bytes.Buffer{}}
pm := problem.NewProblemManager(problem.WithLogger(customLogger))
```

## API Reference

### Core Interfaces

```go
type Logger interface {
    Printf(format string, v ...interface{})
}
```

### Problem Manager

```go
type ProblemManager struct {
    config *ProblemConfig
}

func NewProblemManager(options ...ProblemOption) *ProblemManager
func (pm *ProblemManager) New(status int, title string, detail string) *Problem
func (pm *ProblemManager) Send(w http.ResponseWriter, status int, title string, err error)
func (pm *ProblemManager) Wrap(err error, message string) error
```

### Configuration

```go
type ProblemConfig struct {
    Logger     Logger
    LogPrefix  string
    LogErrors  bool
}

func DefaultProblemConfig() *ProblemConfig
func NewProblemConfig(options ...ProblemOption) *ProblemConfig
```

### Problem Structure

```go
type Problem struct {
    Type     string `json:"type,omitempty"`
    Title    string `json:"title"`
    Status   int    `json:"status"`
    Detail   string `json:"detail,omitempty"`
    Instance string `json:"instance,omitempty"`
}
```

## Examples

### Basic Usage

```go
// Simple usage with defaults
pm := problem.NewProblemManager()

// Send error response
pm.Send(w, http.StatusBadRequest, "Invalid input", err)
```

### Advanced Configuration

```go
// Custom configuration for production
pm := problem.NewProblemManager(
    problem.WithLogger(log.New(os.Stderr, "[API] ", log.LstdFlags)),
    problem.WithLogPrefix("[ERROR]"),
    problem.WithLogErrors(true),
)

// Send error with logging
pm.Send(w, http.StatusInternalServerError, "Database error", err)
```

### Error Handling in HTTP Handlers

```go
func handleCreateUser(w http.ResponseWriter, r *http.Request) {
    pm := problem.NewProblemManager()
    
    // Parse request
    var user User
    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
        pm.Send(w, http.StatusBadRequest, "Invalid JSON", err)
        return
    }
    
    // Validate user
    if err := validateUser(user); err != nil {
        pm.Send(w, http.StatusUnprocessableEntity, "Validation failed", err)
        return
    }
    
    // Save user
    if err := saveUser(user); err != nil {
        wrappedErr := pm.Wrap(err, "Failed to save user")
        pm.Send(w, http.StatusInternalServerError, "Database error", wrappedErr)
        return
    }
    
    w.WriteHeader(http.StatusCreated)
}
```

### Custom Problem Creation

```go
func handleCustomError(w http.ResponseWriter, r *http.Request) {
    pm := problem.NewProblemManager()
    
    // Create custom problem
    prob := pm.New(
        http.StatusTooManyRequests,
        "Rate limit exceeded",
        "You have exceeded the rate limit for this endpoint",
    )
    
    // Add custom fields
    prob.Instance = r.URL.Path
    prob.Type = "https://api.example.com/errors/rate-limit"
    
    // Send response
    w.Header().Set("Content-Type", "application/problem+json")
    w.WriteHeader(prob.Status)
    json.NewEncoder(w).Encode(prob)
}
```

### Complete Error Handling Setup

```go
package main

import (
    "log"
    "net/http"
    "os"
    
    "github.com/go-chi/chi/v5"
    "github.com/Okja-Engineering/go-service-kit/pkg/problem"
)

func main() {
    router := chi.NewRouter()
    
    // Create problem manager with logging
    pm := problem.NewProblemManager(
        problem.WithLogger(log.New(os.Stderr, "[API] ", log.LstdFlags)),
        problem.WithLogPrefix("[ERROR]"),
        problem.WithLogErrors(true),
    )
    
    // Routes with error handling
    router.Get("/api/users", func(w http.ResponseWriter, r *http.Request) {
        if err := getUsers(); err != nil {
            pm.Send(w, http.StatusInternalServerError, "Failed to get users", err)
            return
        }
        // ... success response
    })
    
    router.Post("/api/users", func(w http.ResponseWriter, r *http.Request) {
        if err := createUser(r); err != nil {
            pm.Send(w, http.StatusBadRequest, "Failed to create user", err)
            return
        }
        // ... success response
    })
    
    http.ListenAndServe(":8080", router)
}
```

## Testing

### Using Mock Loggers

```go
func TestErrorHandling(t *testing.T) {
    // Create mock logger
    mockLogger := &MockLogger{output: &bytes.Buffer{}}
    
    // Create problem manager with mock
    pm := problem.NewProblemManager(problem.WithLogger(mockLogger))
    
    // Test error handling
    recorder := httptest.NewRecorder()
    pm.Send(recorder, http.StatusBadRequest, "Test error", errors.New("test"))
    
    // Check response
    if recorder.Code != http.StatusBadRequest {
        t.Errorf("Expected status %d, got %d", http.StatusBadRequest, recorder.Code)
    }
    
    // Check logger was called
    if mockLogger.output.Len() == 0 {
        t.Error("Expected logger to be called")
    }
}
```

### Mock Logger Implementation

```go
type MockLogger struct {
    output *bytes.Buffer
}

func (m *MockLogger) Printf(format string, v ...interface{}) {
    m.output.WriteString(fmt.Sprintf(format, v...))
}
```

## Best Practices

1. **Use consistent error responses** - Always use Problem+JSON format for API errors
2. **Log errors appropriately** - Use custom loggers to capture error details
3. **Provide meaningful titles** - Use clear, user-friendly error titles
4. **Include relevant details** - Add helpful error details when appropriate
5. **Use appropriate status codes** - Choose the correct HTTP status code for each error type
