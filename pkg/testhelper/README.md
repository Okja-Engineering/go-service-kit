# Test Helper Utilities

This package provides flexible HTTP endpoint testing utilities with table-driven test support.

## Features

- **Write fewer test boilerplate** - Define test cases in simple structs instead of repetitive test functions
- **Test multiple scenarios quickly** - Run dozens of endpoint tests with a single function call
- **Consistent test structure** - All your HTTP tests follow the same pattern
- **Easy to maintain** - Change test behavior globally without touching individual tests
- **Flexible configuration** - Customize logging, validation, and headers to match your needs

## Quick Start

```go
import "github.com/Okja-Engineering/go-service-kit/pkg/testhelper"

// Define your test cases
testCases := []testhelper.TestCase{
    {
        Name:           "Get users",
        URL:            "/api/users",
        Method:         "GET",
        CheckStatus:    200,
        CheckBody:      "users",
        CheckBodyCount: 1,
    },
    {
        Name:           "Create user",
        URL:            "/api/users",
        Method:         "POST",
        Body:           `{"name": "John"}`,
        CheckStatus:    201,
        CheckBody:      "id",
        CheckBodyCount: 1,
    },
}

// Run all tests
testhelper.Run(t, router, testCases)
```

That's it! No more writing repetitive test functions for each endpoint.

## Configuration Options

### Functional Options

```go
// WithLogger sets a custom logger
helper := testhelper.NewTestHelper(
    testhelper.WithLogger(customLogger),
)

// WithResponseValidator sets a custom response validator
helper := testhelper.NewTestHelper(
    testhelper.WithResponseValidator(customValidator),
)

// WithLogTestExecution enables/disables test execution logging
helper := testhelper.NewTestHelper(
    testhelper.WithLogTestExecution(false),
)

// WithDefaultHeaders sets default headers for all requests
helper := testhelper.NewTestHelper(
    testhelper.WithDefaultHeaders(map[string]string{
        "Content-Type": "application/json",
        "Authorization": "Bearer token",
    }),
)
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
helper := testhelper.NewTestHelper(testhelper.WithLogger(customLogger))
```

### Custom Response Validators

```go
// Implement the ResponseValidator interface
type CustomResponseValidator struct {
    validateCalled bool
}

func (v *CustomResponseValidator) Validate(t *testing.T, rec *httptest.ResponseRecorder, test *testhelper.TestCase) {
    v.validateCalled = true
    // Custom validation logic
}

// Use custom validator
customValidator := &CustomResponseValidator{}
helper := testhelper.NewTestHelper(testhelper.WithResponseValidator(customValidator))
```

## API Reference

### Core Functions

```go
func Run(t *testing.T, router chi.Router, testCases []TestCase)
```

### Interfaces

```go
type Logger interface {
    Printf(format string, v ...interface{})
}

type ResponseValidator interface {
    Validate(t *testing.T, rec *httptest.ResponseRecorder, test *TestCase)
}

type TestHelper struct {
    config *TestHelperConfig
}

func NewTestHelper(options ...TestHelperOption) *TestHelper
func (th *TestHelper) Run(t *testing.T, router chi.Router, testCases []TestCase)
```

### Configuration

```go
type TestHelperConfig struct {
    Logger           Logger
    ResponseValidator ResponseValidator
    LogTestExecution  bool
    DefaultHeaders    map[string]string
}

func DefaultTestHelperConfig() *TestHelperConfig
func NewTestHelperConfig(options ...TestHelperOption) *TestHelperConfig
```

### Test Case Structure

```go
type TestCase struct {
    Name           string            // Test case description
    URL            string            // Endpoint URL
    Method         string            // HTTP method
    Body           string            // Request body
    Headers        map[string]string // Request headers
    CheckBody      string            // Regex to match in response
    CheckBodyCount int               // Expected matches for CheckBody
    CheckStatus    int               // Expected HTTP status code
}
```

## Testing

### Using Mock Loggers

```go
func TestMyFunction(t *testing.T) {
    // Create mock logger
    mockLogger := &MockLogger{output: &bytes.Buffer{}}
    
    // Create test helper with mock
    helper := testhelper.NewTestHelper(testhelper.WithLogger(mockLogger))
    
    // Test your function
    helper.Run(t, router, testCases)
    
    // Check logger output
    if mockLogger.output.Len() == 0 {
        t.Error("Expected logger to be called")
    }
}
```

### Using Mock Response Validators

```go
func TestResponseValidation(t *testing.T) {
    // Create mock validator
    mockValidator := &MockResponseValidator{}
    
    // Create test helper with mock validator
    helper := testhelper.NewTestHelper(testhelper.WithResponseValidator(mockValidator))
    
    // Test validation behavior
    helper.Run(t, router, testCases)
    
    // Check validator was called
    if !mockValidator.validateCalled {
        t.Error("Expected validator to be called")
    }
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

type MockResponseValidator struct {
    validateCalled bool
    lastRecorder   *httptest.ResponseRecorder
    lastTestCase   *TestCase
}

func (m *MockResponseValidator) Validate(t *testing.T, rec *httptest.ResponseRecorder, test *TestCase) {
    m.validateCalled = true
    m.lastRecorder = rec
    m.lastTestCase = test
}
```

## Examples

### Basic Configuration

```go
// Simple usage with defaults
helper := testhelper.NewTestHelper()
helper.Run(t, router, testCases)
```

### Advanced Configuration

```go
// Custom configuration for API testing
helper := testhelper.NewTestHelper(
    testhelper.WithLogTestExecution(false),
    testhelper.WithDefaultHeaders(map[string]string{
        "Content-Type":  "application/json",
        "Authorization": "Bearer test-token",
        "X-API-Version": "v1",
    }),
)

testCases := []testhelper.TestCase{
    {
        Name:           "Get users",
        URL:            "/api/v1/users",
        Method:         "GET",
        CheckStatus:    200,
        CheckBody:      `"users"`,
        CheckBodyCount: 1,
    },
    {
        Name:           "Create user",
        URL:            "/api/v1/users",
        Method:         "POST",
        Body:           `{"name": "John", "email": "john@example.com"}`,
        CheckStatus:    201,
        CheckBody:      `"id"`,
        CheckBodyCount: 1,
    },
}

helper.Run(t, router, testCases)
```

### Custom Response Validation

```go
// Custom validator for specific response formats
type JSONResponseValidator struct{}

func (v *JSONResponseValidator) Validate(t *testing.T, rec *httptest.ResponseRecorder, test *testhelper.TestCase) {
    // Check status code
    if rec.Code != test.CheckStatus {
        t.Errorf("Expected status %d, got %d", test.CheckStatus, rec.Code)
    }
    
    // Check content type
    contentType := rec.Header().Get("Content-Type")
    if contentType != "application/json" {
        t.Errorf("Expected Content-Type application/json, got %s", contentType)
    }
    
    // Parse JSON response
    var response map[string]interface{}
    if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
        t.Errorf("Failed to parse JSON response: %v", err)
    }
}

// Use custom validator
validator := &JSONResponseValidator{}
helper := testhelper.NewTestHelper(testhelper.WithResponseValidator(validator))
helper.Run(t, router, testCases)
```

### Advanced Usage

```go
// For more control, use the test helper directly
helper := testhelper.NewTestHelper(
    testhelper.WithLogTestExecution(false),
    testhelper.WithDefaultHeaders(map[string]string{
        "Authorization": "Bearer token",
    }),
)
helper.Run(t, router, testCases)
```

## Best Practices

1. **Start simple** - Use the basic `Run()` function for most cases
2. **Use descriptive names** - Make test case names clear and descriptive
3. **Test edge cases** - Include tests for error conditions and edge cases
4. **Group related tests** - Keep test cases for the same endpoint together
5. **Use custom validators** - For complex response validation needs
