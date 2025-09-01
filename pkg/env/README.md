# Environment Package

Flexible environment variable access with interface-based design and functional configuration.

## Features

- **Interface-based design** - Custom environment providers for testing and flexibility
- **Functional configuration** - Clean configuration with functional option pattern
- **Type safety** - Strongly typed environment variable access
- **Mock support** - Mock providers for unit testing
- **Performance** - Efficient string parsing and caching

## Quick Start

```go
package main

import (
    "time"
    
    "github.com/Okja-Engineering/go-service-kit/pkg/env"
)

func main() {
    // Create environment instance
    env := env.NewEnvironment()
    
    // Get environment variables with defaults
    port := env.GetString("PORT", "8080")
    timeout := env.GetDuration("TIMEOUT", 30*time.Second)
    debug := env.GetBool("DEBUG", false)
    
    // Use in your application
    config := &Config{
        Port:     env.GetInt("PORT", 8080),
        Timeout:  env.GetDuration("TIMEOUT", 30*time.Second),
        Debug:    env.GetBool("DEBUG", false),
        LogLevel: env.GetString("LOG_LEVEL", "info"),
    }
}
```

## Configuration

### Functional Options

```go
// Basic configuration
env := env.NewEnvironment()

// Advanced configuration
env := env.NewEnvironment(
    env.WithProvider(customProvider),
    env.WithTrimSpaces(true),
    env.WithCaseSensitive(false),
)
```

### Available Options

```go
func WithProvider(provider EnvironmentProvider) EnvironmentOption
func WithTrimSpaces(trim bool) EnvironmentOption
func WithCaseSensitive(sensitive bool) EnvironmentOption
```

### Custom Environment Providers

```go
// Implement the EnvironmentProvider interface
type CustomProvider struct {
    values map[string]string
}

func (p *CustomProvider) Get(key string) string {
    return p.values[key]
}

func (p *CustomProvider) Lookup(key string) (string, bool) {
    value, exists := p.values[key]
    return value, exists
}

// Use custom provider
provider := &CustomProvider{
    values: map[string]string{
        "PORT": "9090",
        "DEBUG": "true",
    },
}

env := env.NewEnvironment(env.WithProvider(provider))
port := env.GetString("PORT", "8080") // Returns "9090"
```

## API Reference

### Core Interfaces

```go
type EnvironmentProvider interface {
    Get(key string) string
    Lookup(key string) (string, bool)
}
```

### Environment Instance

```go
type Environment struct {
    config *EnvironmentConfig
}

func NewEnvironment(options ...EnvironmentOption) *Environment
func (e *Environment) GetString(key, defaultVal string) string
func (e *Environment) GetInt(key string, defaultVal int) int
func (e *Environment) GetFloat(key string, defaultVal float64) float64
func (e *Environment) GetBool(key string, defaultVal bool) bool
func (e *Environment) GetDuration(key string, defaultVal time.Duration) time.Duration
```

### Configuration

```go
type EnvironmentConfig struct {
    Provider     EnvironmentProvider
    TrimSpaces   bool
    CaseSensitive bool
}

func DefaultEnvironmentConfig() *EnvironmentConfig
func NewEnvironmentConfig(options ...EnvironmentOption) *EnvironmentConfig
```

## Examples

### Basic Usage

```go
// Simple usage with defaults
env := env.NewEnvironment()
port := env.GetString("PORT", "8080")
timeout := env.GetDuration("TIMEOUT", 30*time.Second)
debug := env.GetBool("DEBUG", false)
```

### Advanced Configuration

```go
// Custom configuration for testing
env := env.NewEnvironment(
    env.WithProvider(mockProvider),
    env.WithTrimSpaces(false),
    env.WithCaseSensitive(false),
)

// Use in your application
config := &Config{
    Port:     env.GetInt("PORT", 8080),
    Timeout:  env.GetDuration("TIMEOUT", 30*time.Second),
    Debug:    env.GetBool("DEBUG", false),
    LogLevel: env.GetString("LOG_LEVEL", "info"),
}
```

### Application Configuration

```go
package main

import (
    "time"
    
    "github.com/Okja-Engineering/go-service-kit/pkg/env"
)

type Config struct {
    Port     int
    Timeout  time.Duration
    Debug    bool
    LogLevel string
    Database DatabaseConfig
}

type DatabaseConfig struct {
    Host     string
    Port     int
    Username string
    Password string
}

func loadConfig() *Config {
    env := env.NewEnvironment()
    
    return &Config{
        Port:     env.GetInt("PORT", 8080),
        Timeout:  env.GetDuration("TIMEOUT", 30*time.Second),
        Debug:    env.GetBool("DEBUG", false),
        LogLevel: env.GetString("LOG_LEVEL", "info"),
        Database: DatabaseConfig{
            Host:     env.GetString("DB_HOST", "localhost"),
            Port:     env.GetInt("DB_PORT", 5432),
            Username: env.GetString("DB_USERNAME", "postgres"),
            Password: env.GetString("DB_PASSWORD", ""),
        },
    }
}
```

## Testing

### Using Mock Providers

```go
func TestMyFunction(t *testing.T) {
    // Create mock provider
    mockProvider := &MockEnvironmentProvider{
        values: map[string]string{
            "TEST_PORT": "9090",
            "TEST_DEBUG": "true",
        },
    }

    // Create environment with mock
    env := env.NewEnvironment(env.WithProvider(mockProvider))

    // Test your function
    result := myFunction(env)
    // ... assertions
}
```

### Mock Provider Implementation

```go
type MockEnvironmentProvider struct {
    values map[string]string
}

func (m *MockEnvironmentProvider) Get(key string) string {
    return m.values[key]
}

func (m *MockEnvironmentProvider) Lookup(key string) (string, bool) {
    value, exists := m.values[key]
    return value, exists
}
```

## Best Practices

1. **Create once, reuse** - Create environment instance once and reuse throughout your application
2. **Provide sensible defaults** - Always provide default values for environment variables
3. **Use mock providers for testing** - Test your configuration loading with mock providers
4. **Handle errors gracefully** - Invalid values fall back to defaults automatically
5. **Group related configuration** - Use structs to group related environment variables
