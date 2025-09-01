package env

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// EnvironmentProvider defines the interface for environment variable access
type EnvironmentProvider interface {
	Get(key string) string
	Lookup(key string) (string, bool)
}

// DefaultEnvironmentProvider implements EnvironmentProvider using os.LookupEnv
type DefaultEnvironmentProvider struct{}

// Get returns the environment variable value or empty string if not found
func (p *DefaultEnvironmentProvider) Get(key string) string {
	value, _ := os.LookupEnv(key)
	return value
}

// Lookup returns the environment variable value and whether it exists
func (p *DefaultEnvironmentProvider) Lookup(key string) (string, bool) {
	return os.LookupEnv(key)
}

// EnvironmentOption is a functional option for environment configuration
type EnvironmentOption func(*EnvironmentConfig)

// EnvironmentConfig holds configuration for environment variable handling
type EnvironmentConfig struct {
	Provider      EnvironmentProvider
	TrimSpaces    bool
	CaseSensitive bool
}

// DefaultEnvironmentConfig provides sensible defaults
func DefaultEnvironmentConfig() *EnvironmentConfig {
	return &EnvironmentConfig{
		Provider:      &DefaultEnvironmentProvider{},
		TrimSpaces:    true,
		CaseSensitive: true,
	}
}

// WithProvider sets a custom environment provider
func WithProvider(provider EnvironmentProvider) EnvironmentOption {
	return func(config *EnvironmentConfig) {
		config.Provider = provider
	}
}

// WithTrimSpaces enables/disables trimming of whitespace
func WithTrimSpaces(trim bool) EnvironmentOption {
	return func(config *EnvironmentConfig) {
		config.TrimSpaces = trim
	}
}

// WithCaseSensitive enables/disables case sensitivity
func WithCaseSensitive(sensitive bool) EnvironmentOption {
	return func(config *EnvironmentConfig) {
		config.CaseSensitive = sensitive
	}
}

// NewEnvironmentConfig creates a new environment config with options
func NewEnvironmentConfig(options ...EnvironmentOption) *EnvironmentConfig {
	config := DefaultEnvironmentConfig()
	for _, option := range options {
		option(config)
	}
	return config
}

// Environment handles environment variable access with configuration
type Environment struct {
	config *EnvironmentConfig
}

// NewEnvironment creates a new Environment instance with options
func NewEnvironment(options ...EnvironmentOption) *Environment {
	config := NewEnvironmentConfig(options...)
	return &Environment{config: config}
}

// getEnv gets an environment variable with the configured settings
func (e *Environment) getEnv(key, defaultVal string) string {
	value, exists := e.config.Provider.Lookup(key)
	if !exists {
		return defaultVal
	}

	if e.config.TrimSpaces {
		value = strings.TrimSpace(value)
	}

	if !e.config.CaseSensitive {
		value = strings.ToLower(value)
	}

	return value
}

// GetString gets a string environment variable
func (e *Environment) GetString(key, defaultVal string) string {
	return e.getEnv(key, defaultVal)
}

// GetInt gets an integer environment variable
func (e *Environment) GetInt(key string, defaultVal int) int {
	valueStr := e.getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultVal
}

// GetFloat gets a float environment variable
func (e *Environment) GetFloat(key string, defaultVal float64) float64 {
	valueStr := e.getEnv(key, "")
	if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
		return value
	}
	return defaultVal
}

// GetBool gets a boolean environment variable
func (e *Environment) GetBool(key string, defaultVal bool) bool {
	valueStr := e.getEnv(key, "")
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultVal
}

// GetDuration gets a duration environment variable
func (e *Environment) GetDuration(key string, defaultVal time.Duration) time.Duration {
	valueStr := e.getEnv(key, "")
	if value, err := time.ParseDuration(valueStr); err == nil {
		return value
	}
	return defaultVal
}

// Legacy functions for backward compatibility
func getEnv(key, defaultVal string) string {
	env := NewEnvironment()
	return env.getEnv(key, defaultVal)
}

func GetEnvString(key, defaultVal string) string {
	return getEnv(key, defaultVal)
}

func GetEnvInt(key string, defaultVal int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultVal
}

func GetEnvFloat(key string, defaultVal float64) float64 {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
		return value
	}
	return defaultVal
}

func GetEnvBool(key string, defaultVal bool) bool {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultVal
}

func GetEnvDuration(key string, defaultVal time.Duration) time.Duration {
	valueStr := getEnv(key, "")
	if value, err := time.ParseDuration(valueStr); err == nil {
		return value
	}
	return defaultVal
}
