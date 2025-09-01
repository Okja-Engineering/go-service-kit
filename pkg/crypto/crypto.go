package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const (
	// DefaultPasswordLength is the default length for generated passwords
	DefaultPasswordLength = 16

	// DefaultTokenLength is the default length for generated tokens
	DefaultTokenLength = 32

	// DefaultRefreshTokenLength is the default length for refresh tokens
	DefaultRefreshTokenLength = 64

	// Password character sets
	lowercase = "abcdefghijklmnopqrstuvwxyz"
	uppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digits    = "0123456789"
	symbols   = "!@#$%^&*()_+-=[]{}|;:,.<>?"
)

// PasswordConfig holds configuration for password generation
type PasswordConfig struct {
	Length     int
	UseLower   bool
	UseUpper   bool
	UseDigits  bool
	UseSymbols bool
}

// DefaultPasswordConfig returns a secure default password configuration
func DefaultPasswordConfig() *PasswordConfig {
	return &PasswordConfig{
		Length:     DefaultPasswordLength,
		UseLower:   true,
		UseUpper:   true,
		UseDigits:  true,
		UseSymbols: true,
	}
}

// HashPassword hashes a password using bcrypt with the default cost
func HashPassword(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hashedBytes), nil
}

// HashPasswordWithCost hashes a password using bcrypt with a specified cost
func HashPasswordWithCost(password string, cost int) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		return "", fmt.Errorf("bcrypt cost must be between %d and %d", bcrypt.MinCost, bcrypt.MaxCost)
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hashedBytes), nil
}

// VerifyPassword verifies a password against its hash
func VerifyPassword(hashedPassword, password string) error {
	if hashedPassword == "" {
		return fmt.Errorf("hashed password cannot be empty")
	}

	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return fmt.Errorf("password verification failed: %w", err)
	}

	return nil
}

// GenerateSecurePassword generates a cryptographically secure random password
func GenerateSecurePassword(length int) (string, error) {
	return GenerateSecurePasswordWithConfig(&PasswordConfig{
		Length:     length,
		UseLower:   true,
		UseUpper:   true,
		UseDigits:  true,
		UseSymbols: true,
	})
}

// buildCharset creates a character set based on configuration
func buildCharset(config *PasswordConfig) (string, error) {
	var charset strings.Builder
	if config.UseLower {
		charset.WriteString(lowercase)
	}
	if config.UseUpper {
		charset.WriteString(uppercase)
	}
	if config.UseDigits {
		charset.WriteString(digits)
	}
	if config.UseSymbols {
		charset.WriteString(symbols)
	}

	if charset.Len() == 0 {
		return "", fmt.Errorf("at least one character set must be enabled")
	}

	return charset.String(), nil
}

// ensureRequiredCharacters adds at least one character from each enabled set
func ensureRequiredCharacters(password []byte, config *PasswordConfig) (int, error) {
	position := 0
	if config.UseLower {
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(lowercase))))
		if err != nil {
			return 0, fmt.Errorf("failed to generate random character: %w", err)
		}
		password[position] = lowercase[randomIndex.Int64()]
		position++
	}
	if config.UseUpper {
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(uppercase))))
		if err != nil {
			return 0, fmt.Errorf("failed to generate random character: %w", err)
		}
		password[position] = uppercase[randomIndex.Int64()]
		position++
	}
	if config.UseDigits {
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return 0, fmt.Errorf("failed to generate random character: %w", err)
		}
		password[position] = digits[randomIndex.Int64()]
		position++
	}
	if config.UseSymbols {
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(symbols))))
		if err != nil {
			return 0, fmt.Errorf("failed to generate random character: %w", err)
		}
		password[position] = symbols[randomIndex.Int64()]
		position++
	}
	return position, nil
}

// fillRemainingCharacters fills the rest of the password with random characters
func fillRemainingCharacters(password []byte, startPos int, charset string) error {
	for i := startPos; i < len(password); i++ {
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return fmt.Errorf("failed to generate random character: %w", err)
		}
		password[i] = charset[randomIndex.Int64()]
	}
	return nil
}

// shufflePassword randomizes the order of characters to avoid predictable patterns
func shufflePassword(password []byte) error {
	for i := len(password) - 1; i > 0; i-- {
		j, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return fmt.Errorf("failed to shuffle password: %w", err)
		}
		password[i], password[j.Int64()] = password[j.Int64()], password[i]
	}
	return nil
}

// GenerateSecurePasswordWithConfig generates a password with custom configuration
func GenerateSecurePasswordWithConfig(config *PasswordConfig) (string, error) {
	if config == nil {
		config = DefaultPasswordConfig()
	}

	if config.Length < 8 {
		return "", fmt.Errorf("password length must be at least 8 characters")
	}

	charset, err := buildCharset(config)
	if err != nil {
		return "", err
	}

	password := make([]byte, config.Length)

	position, err := ensureRequiredCharacters(password, config)
	if err != nil {
		return "", err
	}

	if err := fillRemainingCharacters(password, position, charset); err != nil {
		return "", err
	}

	if err := shufflePassword(password); err != nil {
		return "", err
	}

	return string(password), nil
}

// GenerateSecureToken generates a cryptographically secure random token
func GenerateSecureToken() (string, error) {
	return GenerateSecureTokenWithLength(DefaultTokenLength)
}

// GenerateSecureTokenWithLength generates a token with specified length
func GenerateSecureTokenWithLength(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("token length must be positive")
	}

	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate secure token: %w", err)
	}

	return hex.EncodeToString(bytes), nil
}

// HashToken creates a SHA-256 hash of a token for secure storage
func HashToken(token string) string {
	if token == "" {
		return ""
	}

	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// GenerateRefreshToken generates a cryptographically secure refresh token
func GenerateRefreshToken() (string, error) {
	return GenerateSecureTokenWithLength(DefaultRefreshTokenLength)
}

// GenerateRefreshTokenWithLength generates a refresh token with specified length
func GenerateRefreshTokenWithLength(length int) (string, error) {
	if length < 32 {
		return "", fmt.Errorf("refresh token length must be at least 32 characters")
	}

	return GenerateSecureTokenWithLength(length)
}

// checkPasswordLength validates the minimum password length
func checkPasswordLength(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}
	return nil
}

// analyzePasswordCharacters analyzes password for required character types
func analyzePasswordCharacters(password string) (hasLower, hasUpper, hasDigit, hasSymbol bool) {
	for _, char := range password {
		switch {
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= '0' && char <= '9':
			hasDigit = true
		case strings.ContainsRune(symbols, char):
			hasSymbol = true
		}
	}
	return
}

// validateCharacterRequirements checks if password meets character type requirements
func validateCharacterRequirements(hasLower, hasUpper, hasDigit, hasSymbol bool) error {
	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if !hasDigit {
		return fmt.Errorf("password must contain at least one digit")
	}
	if !hasSymbol {
		return fmt.Errorf("password must contain at least one special character")
	}
	return nil
}

// ValidatePasswordStrength checks if a password meets minimum security requirements
func ValidatePasswordStrength(password string) error {
	if err := checkPasswordLength(password); err != nil {
		return err
	}

	hasLower, hasUpper, hasDigit, hasSymbol := analyzePasswordCharacters(password)
	return validateCharacterRequirements(hasLower, hasUpper, hasDigit, hasSymbol)
}
