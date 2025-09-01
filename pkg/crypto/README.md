# Crypto Package

Cryptographic utilities for authentication services with secure password hashing, token generation, and validation.

## Features

- **Password Management** - Secure bcrypt hashing with configurable cost
- **Token Generation** - Cryptographically secure random tokens and refresh tokens
- **Password Validation** - Strength checking with configurable requirements
- **Production Ready** - Designed for auth services like auth.okja.dev

## Quick Start

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/Okja-Engineering/go-service-kit/pkg/crypto"
)

func main() {
    // Hash a password
    hashedPassword, err := crypto.HashPassword("mySecurePassword123!")
    if err != nil {
        log.Fatal(err)
    }
    
    // Verify a password
    err = crypto.VerifyPassword(hashedPassword, "mySecurePassword123!")
    if err != nil {
        log.Fatal("Password verification failed")
    }
    
    // Generate a secure token
    token, err := crypto.GenerateSecureToken()
    if err != nil {
        log.Fatal(err)
    }
    
    // Hash token for storage
    hashedToken := crypto.HashToken(token)
    
    fmt.Printf("Hashed password: %s\n", hashedPassword)
    fmt.Printf("Token: %s\n", token)
    fmt.Printf("Hashed token: %s\n", hashedToken)
}
```

## Password Management

### Hashing Passwords

```go
// Basic password hashing with default cost
hashed, err := crypto.HashPassword("myPassword123!")
if err != nil {
    // Handle error
}

// Custom bcrypt cost (higher = more secure but slower)
hashed, err := crypto.HashPasswordWithCost("myPassword123!", 12)
```

### Verifying Passwords

```go
// Verify against stored hash
err := crypto.VerifyPassword(storedHash, "myPassword123!")
if err != nil {
    // Password is incorrect
}
```

### Generating Secure Passwords

```go
// Generate a 16-character password with all character types
password, err := crypto.GenerateSecurePassword(16)

// Custom configuration
config := &crypto.PasswordConfig{
    Length:     20,
    UseLower:   true,
    UseUpper:   true,
    UseDigits:  true,
    UseSymbols: false, // No symbols
}
password, err := crypto.GenerateSecurePasswordWithConfig(config)
```

### Password Validation

```go
// Check password strength
err := crypto.ValidatePasswordStrength("MySecurePass123!")
if err != nil {
    // Password doesn't meet requirements
    fmt.Println(err.Error())
}
```

## Token Management

### Generating Tokens

```go
// Generate a 32-byte secure token
token, err := crypto.GenerateSecureToken()

// Custom length
token, err := crypto.GenerateSecureTokenWithLength(64)

// Generate refresh token (64 bytes by default)
refreshToken, err := crypto.GenerateRefreshToken()

// Custom refresh token length
refreshToken, err := crypto.GenerateRefreshTokenWithLength(128)
```

### Token Hashing

```go
// Hash token for secure storage (never store plain tokens)
hashedToken := crypto.HashToken(token)

// Verify token by hashing and comparing
if crypto.HashToken(providedToken) == storedHashedToken {
    // Token is valid
}
```

## API Reference

### Password Functions

```go
func HashPassword(password string) (string, error)
func HashPasswordWithCost(password string, cost int) (string, error)
func VerifyPassword(hashedPassword, password string) error
func GenerateSecurePassword(length int) (string, error)
func GenerateSecurePasswordWithConfig(config *PasswordConfig) (string, error)
func ValidatePasswordStrength(password string) error
```

### Token Functions

```go
func GenerateSecureToken() (string, error)
func GenerateSecureTokenWithLength(length int) (string, error)
func HashToken(token string) string
func GenerateRefreshToken() (string, error)
func GenerateRefreshTokenWithLength(length int) (string, error)
```

### Configuration

```go
type PasswordConfig struct {
    Length     int
    UseLower   bool
    UseUpper   bool
    UseDigits  bool
    UseSymbols bool
}

func DefaultPasswordConfig() *PasswordConfig
```

### Constants

```go
const (
    DefaultPasswordLength     = 16
    DefaultTokenLength        = 32
    DefaultRefreshTokenLength = 64
)
```

## Examples

### User Registration

```go
func registerUser(username, password string) error {
    // Validate password strength
    if err := crypto.ValidatePasswordStrength(password); err != nil {
        return fmt.Errorf("password too weak: %w", err)
    }
    
    // Hash password for storage
    hashedPassword, err := crypto.HashPassword(password)
    if err != nil {
        return fmt.Errorf("failed to hash password: %w", err)
    }
    
    // Store user with hashed password
    user := User{
        Username: username,
        Password: hashedPassword,
    }
    
    return db.CreateUser(user)
}
```

### User Authentication

```go
func authenticateUser(username, password string) (*User, error) {
    // Get user from database
    user, err := db.GetUserByUsername(username)
    if err != nil {
        return nil, err
    }
    
    // Verify password
    if err := crypto.VerifyPassword(user.Password, password); err != nil {
        return nil, fmt.Errorf("invalid credentials")
    }
    
    return user, nil
}
```

### API Token Management

```go
func createAPIToken(userID string) (*APIToken, error) {
    // Generate secure token
    token, err := crypto.GenerateSecureToken()
    if err != nil {
        return nil, fmt.Errorf("failed to generate token: %w", err)
    }
    
    // Hash token for storage
    hashedToken := crypto.HashToken(token)
    
    // Store in database
    apiToken := &APIToken{
        UserID:      userID,
        TokenHash:   hashedToken,
        CreatedAt:   time.Now(),
        ExpiresAt:   time.Now().Add(24 * time.Hour),
    }
    
    if err := db.CreateAPIToken(apiToken); err != nil {
        return nil, err
    }
    
    // Return token to user (only time it's in plain text)
    apiToken.Token = token
    return apiToken, nil
}

func validateAPIToken(token string) (*APIToken, error) {
    // Hash provided token
    hashedToken := crypto.HashToken(token)
    
    // Find token in database
    apiToken, err := db.GetAPITokenByHash(hashedToken)
    if err != nil {
        return nil, fmt.Errorf("invalid token")
    }
    
    // Check expiration
    if time.Now().After(apiToken.ExpiresAt) {
        return nil, fmt.Errorf("token expired")
    }
    
    return apiToken, nil
}
```

### Refresh Token Flow

```go
func createRefreshToken(userID string) (*RefreshToken, error) {
    // Generate refresh token
    refreshToken, err := crypto.GenerateRefreshToken()
    if err != nil {
        return nil, fmt.Errorf("failed to generate refresh token: %w", err)
    }
    
    // Hash for storage
    hashedToken := crypto.HashToken(refreshToken)
    
    // Store in database
    token := &RefreshToken{
        UserID:    userID,
        TokenHash: hashedToken,
        CreatedAt: time.Now(),
        ExpiresAt: time.Now().AddDate(0, 1, 0), // 1 month
    }
    
    if err := db.CreateRefreshToken(token); err != nil {
        return nil, err
    }
    
    token.Token = refreshToken
    return token, nil
}
```

## Best Practices

### Password Security

1. **Always hash passwords** - Never store plain text passwords
2. **Use appropriate bcrypt cost** - Balance security with performance
3. **Validate password strength** - Enforce minimum requirements
4. **Generate secure passwords** - Use for temporary passwords or resets

### Token Security

1. **Hash tokens for storage** - Never store plain tokens in database
2. **Use sufficient token length** - 32+ bytes for API tokens, 64+ for refresh tokens
3. **Implement token expiration** - Set reasonable expiration times
4. **Secure token transmission** - Use HTTPS and secure headers

### General Security

1. **Use cryptographically secure randomness** - Always use `crypto/rand`
2. **Handle errors properly** - Don't expose sensitive information in errors
3. **Validate inputs** - Check for empty strings and invalid parameters
4. **Follow OWASP guidelines** - Implement security best practices

### Performance Considerations

1. **bcrypt cost tuning** - Higher cost = more secure but slower
2. **Token generation** - Use appropriate lengths for your use case
3. **Database indexing** - Index hashed token columns for fast lookups
4. **Caching** - Consider caching frequently accessed token validations

## Error Handling

The package provides descriptive error messages for common issues:

```go
// Password errors
"password cannot be empty"
"bcrypt cost must be between 4 and 31"
"password verification failed"

// Token errors
"token length must be positive"
"failed to generate secure token"
"refresh token length must be at least 32 characters"

// Password generation errors
"password length must be at least 8 characters"
"at least one character set must be enabled"
"failed to generate random character"

// Password validation errors
"password must be at least 8 characters long"
"password must contain at least one lowercase letter"
"password must contain at least one uppercase letter"
"password must contain at least one digit"
"password must contain at least one special character"
```
