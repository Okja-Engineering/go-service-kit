# env

Helpers for reading environment variables with default fallbacks.

## Features
- GetEnvString, GetEnvInt, GetEnvFloat, GetEnvBool, GetEnvDuration
- Always returns a value (never panics)

## Functions

`GetEnvString(key string, defaultVal string) string`

Returns the value of the environment variable key as a string, or defaultVal if key is not set.

`GetEnvInt(key string, defaultVal int) int`

Returns the value of the environment variable key as an int, or defaultVal if key is not set or cannot be converted to an int.

`GetEnvFloat(key string, defaultVal float64) float64`

Returns the value of the environment variable key as a float64, or defaultVal if key is not set or cannot be converted to a float64.

`GetEnvBool(key string, defaultVal bool) bool`

Returns the value of the environment variable key as a boolean, or defaultVal if key is not set or cannot be converted to a boolean.

`GetEnvDuration(key string, defaultVal time.Duration) time.Duration`

Returns the value of the environment variable key as a time.Duration, or defaultVal if key is not set or cannot be parsed as a duration.

## Usage Example
```go
import "github.com/Okja-Engineering/go-service-kit/pkg/env"

func main() {
	// Get environment variables or use default values
	dbUser := env.GetEnvString("DB_USER", "defaultUser")
	dbPort := env.GetEnvInt("DB_PORT", 5432)
	taxRate := env.GetEnvFloat("TAX_RATE", 7.5)
	isDebugMode := env.GetEnvBool("DEBUG", false)

	fmt.Println("Database User:", dbUser)
	fmt.Println("Database Port:", dbPort)
	fmt.Println("Tax Rate:", taxRate)
	fmt.Println("Is Debug Mode:", isDebugMode)
}
```

## Tests
Run `make test` from the repo root to validate this package.

See tests for more usage patterns.
