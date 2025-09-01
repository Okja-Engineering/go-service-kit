package env

import (
	"os"
	"testing"
	"time"
)

// MockEnvironmentProvider for testing
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

func TestNewEnvironment(t *testing.T) {
	// Test default environment
	env := NewEnvironment()
	if env.config.Provider == nil {
		t.Error("Expected provider to be set")
	}
	if !env.config.TrimSpaces {
		t.Error("Expected TrimSpaces to be true by default")
	}
	if !env.config.CaseSensitive {
		t.Error("Expected CaseSensitive to be true by default")
	}
}

func TestNewEnvironmentWithOptions(t *testing.T) {
	mockProvider := &MockEnvironmentProvider{
		values: map[string]string{
			"TEST_KEY": "test_value",
		},
	}

	env := NewEnvironment(
		WithProvider(mockProvider),
		WithTrimSpaces(false),
		WithCaseSensitive(false),
	)

	if env.config.Provider != mockProvider {
		t.Error("Expected custom provider to be set")
	}
	if env.config.TrimSpaces {
		t.Error("Expected TrimSpaces to be false")
	}
	if env.config.CaseSensitive {
		t.Error("Expected CaseSensitive to be false")
	}
}

func TestEnvironmentGetString(t *testing.T) {
	mockProvider := &MockEnvironmentProvider{
		values: map[string]string{
			"EXISTING_KEY": "existing_value",
		},
	}

	env := NewEnvironment(WithProvider(mockProvider))

	// Test existing key
	result := env.GetString("EXISTING_KEY", "default")
	if result != "existing_value" {
		t.Errorf("Expected 'existing_value', got '%s'", result)
	}

	// Test non-existing key
	result = env.GetString("NON_EXISTING_KEY", "default")
	if result != "default" {
		t.Errorf("Expected 'default', got '%s'", result)
	}
}

func TestEnvironmentGetStringWithTrimSpaces(t *testing.T) {
	mockProvider := &MockEnvironmentProvider{
		values: map[string]string{
			"TRIMMED_KEY": "  value with spaces  ",
		},
	}

	// Test with trim spaces enabled (default)
	env := NewEnvironment(WithProvider(mockProvider))
	result := env.GetString("TRIMMED_KEY", "default")
	if result != "value with spaces" {
		t.Errorf("Expected 'value with spaces', got '%s'", result)
	}

	// Test with trim spaces disabled
	env = NewEnvironment(
		WithProvider(mockProvider),
		WithTrimSpaces(false),
	)
	result = env.GetString("TRIMMED_KEY", "default")
	if result != "  value with spaces  " {
		t.Errorf("Expected '  value with spaces  ', got '%s'", result)
	}
}

func TestEnvironmentGetStringWithCaseSensitive(t *testing.T) {
	mockProvider := &MockEnvironmentProvider{
		values: map[string]string{
			"UPPERCASE_KEY": "UPPERCASE_VALUE",
		},
	}

	// Test with case sensitive enabled (default)
	env := NewEnvironment(WithProvider(mockProvider))
	result := env.GetString("UPPERCASE_KEY", "default")
	if result != "UPPERCASE_VALUE" {
		t.Errorf("Expected 'UPPERCASE_VALUE', got '%s'", result)
	}

	// Test with case sensitive disabled
	env = NewEnvironment(
		WithProvider(mockProvider),
		WithCaseSensitive(false),
	)
	result = env.GetString("UPPERCASE_KEY", "default")
	if result != "uppercase_value" {
		t.Errorf("Expected 'uppercase_value', got '%s'", result)
	}
}

func TestEnvironmentGetInt(t *testing.T) {
	mockProvider := &MockEnvironmentProvider{
		values: map[string]string{
			"VALID_INT":   "42",
			"INVALID_INT": "not_a_number",
		},
	}

	env := NewEnvironment(WithProvider(mockProvider))

	// Test valid integer
	result := env.GetInt("VALID_INT", 0)
	if result != 42 {
		t.Errorf("Expected 42, got %d", result)
	}

	// Test invalid integer
	result = env.GetInt("INVALID_INT", 100)
	if result != 100 {
		t.Errorf("Expected 100, got %d", result)
	}

	// Test non-existing key
	result = env.GetInt("NON_EXISTING", 200)
	if result != 200 {
		t.Errorf("Expected 200, got %d", result)
	}
}

func TestEnvironmentGetFloat(t *testing.T) {
	mockProvider := &MockEnvironmentProvider{
		values: map[string]string{
			"VALID_FLOAT":   "3.14",
			"INVALID_FLOAT": "not_a_number",
		},
	}

	env := NewEnvironment(WithProvider(mockProvider))

	// Test valid float
	result := env.GetFloat("VALID_FLOAT", 0.0)
	if result != 3.14 {
		t.Errorf("Expected 3.14, got %f", result)
	}

	// Test invalid float
	result = env.GetFloat("INVALID_FLOAT", 2.71)
	if result != 2.71 {
		t.Errorf("Expected 2.71, got %f", result)
	}
}

func TestEnvironmentGetBool(t *testing.T) {
	mockProvider := &MockEnvironmentProvider{
		values: map[string]string{
			"TRUE_VALUE":   "true",
			"FALSE_VALUE":  "false",
			"INVALID_BOOL": "not_a_bool",
		},
	}

	env := NewEnvironment(WithProvider(mockProvider))

	// Test true value
	result := env.GetBool("TRUE_VALUE", false)
	if !result {
		t.Error("Expected true, got false")
	}

	// Test false value
	result = env.GetBool("FALSE_VALUE", true)
	if result {
		t.Error("Expected false, got true")
	}

	// Test invalid bool
	result = env.GetBool("INVALID_BOOL", true)
	if !result {
		t.Error("Expected true (default), got false")
	}
}

func TestEnvironmentGetDuration(t *testing.T) {
	mockProvider := &MockEnvironmentProvider{
		values: map[string]string{
			"VALID_DURATION":   "30s",
			"INVALID_DURATION": "not_a_duration",
		},
	}

	env := NewEnvironment(WithProvider(mockProvider))

	// Test valid duration
	result := env.GetDuration("VALID_DURATION", time.Minute)
	if result != 30*time.Second {
		t.Errorf("Expected 30s, got %v", result)
	}

	// Test invalid duration
	result = env.GetDuration("INVALID_DURATION", time.Hour)
	if result != time.Hour {
		t.Errorf("Expected 1h, got %v", result)
	}
}

// Legacy function tests (existing tests)
func TestGetEnvString(t *testing.T) {
	os.Setenv("TEST_STRING", "test_value")
	defer os.Unsetenv("TEST_STRING")

	result := GetEnvString("TEST_STRING", "default")
	if result != "test_value" {
		t.Errorf("Expected 'test_value', got '%s'", result)
	}

	result = GetEnvString("NON_EXISTENT", "default")
	if result != "default" {
		t.Errorf("Expected 'default', got '%s'", result)
	}
}

func TestGetEnvInt(t *testing.T) {
	os.Setenv("TEST_INT", "42")
	defer os.Unsetenv("TEST_INT")

	result := GetEnvInt("TEST_INT", 0)
	if result != 42 {
		t.Errorf("Expected 42, got %d", result)
	}

	result = GetEnvInt("NON_EXISTENT", 100)
	if result != 100 {
		t.Errorf("Expected 100, got %d", result)
	}
}

func TestGetEnvFloat(t *testing.T) {
	os.Setenv("TEST_FLOAT", "3.14")
	defer os.Unsetenv("TEST_FLOAT")

	result := GetEnvFloat("TEST_FLOAT", 0.0)
	if result != 3.14 {
		t.Errorf("Expected 3.14, got %f", result)
	}

	result = GetEnvFloat("NON_EXISTENT", 2.71)
	if result != 2.71 {
		t.Errorf("Expected 2.71, got %f", result)
	}
}

func TestGetEnvBool(t *testing.T) {
	os.Setenv("TEST_BOOL", "true")
	defer os.Unsetenv("TEST_BOOL")

	result := GetEnvBool("TEST_BOOL", false)
	if !result {
		t.Error("Expected true, got false")
	}

	result = GetEnvBool("NON_EXISTENT", true)
	if !result {
		t.Error("Expected true, got false")
	}
}

func TestGetEnvDuration(t *testing.T) {
	os.Setenv("TEST_DURATION", "30s")
	defer os.Unsetenv("TEST_DURATION")

	result := GetEnvDuration("TEST_DURATION", time.Minute)
	if result != 30*time.Second {
		t.Errorf("Expected 30s, got %v", result)
	}

	result = GetEnvDuration("NON_EXISTENT", time.Hour)
	if result != time.Hour {
		t.Errorf("Expected 1h, got %v", result)
	}
}
