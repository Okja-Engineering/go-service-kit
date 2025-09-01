package crypto

import (
	"fmt"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "mySecurePassword123!",
			wantErr:  false,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashed, err := HashPassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("HashPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if hashed == "" {
					t.Error("HashPassword() returned empty hash")
				}
				if hashed == tt.password {
					t.Error("HashPassword() returned unhashed password")
				}
			}
		})
	}
}

func TestHashPasswordWithCost(t *testing.T) {
	password := "mySecurePassword123!"

	// Test valid costs (use lower costs for faster testing)
	validCosts := []int{bcrypt.MinCost, 6, 8}
	for _, cost := range validCosts {
		t.Run(fmt.Sprintf("cost_%d", cost), func(t *testing.T) {
			hashed, err := HashPasswordWithCost(password, cost)
			if err != nil {
				t.Errorf("HashPasswordWithCost() error = %v", err)
				return
			}

			if hashed == "" {
				t.Error("HashPasswordWithCost() returned empty hash")
			}
		})
	}

	// Test invalid costs
	invalidCosts := []int{bcrypt.MinCost - 1, bcrypt.MaxCost + 1}
	for _, cost := range invalidCosts {
		t.Run(fmt.Sprintf("invalid_cost_%d", cost), func(t *testing.T) {
			_, err := HashPasswordWithCost(password, cost)
			if err == nil {
				t.Errorf("HashPasswordWithCost() should have failed for cost %d", cost)
			}
		})
	}
}

func TestVerifyPassword(t *testing.T) {
	password := "mySecurePassword123!"
	hashed, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password for test: %v", err)
	}

	tests := []struct {
		name           string
		hashedPassword string
		password       string
		wantErr        bool
	}{
		{
			name:           "correct password",
			hashedPassword: hashed,
			password:       password,
			wantErr:        false,
		},
		{
			name:           "incorrect password",
			hashedPassword: hashed,
			password:       "wrongPassword",
			wantErr:        true,
		},
		{
			name:           "empty hashed password",
			hashedPassword: "",
			password:       password,
			wantErr:        true,
		},
		{
			name:           "empty password",
			hashedPassword: hashed,
			password:       "",
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := VerifyPassword(tt.hashedPassword, tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("VerifyPassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenerateSecurePassword(t *testing.T) {
	tests := []struct {
		name    string
		length  int
		wantErr bool
	}{
		{
			name:    "default length",
			length:  16,
			wantErr: false,
		},
		{
			name:    "custom length",
			length:  20,
			wantErr: false,
		},
		{
			name:    "minimum length",
			length:  8,
			wantErr: false,
		},
		{
			name:    "too short",
			length:  7,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			password, err := GenerateSecurePassword(tt.length)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateSecurePassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(password) != tt.length {
					t.Errorf("GenerateSecurePassword() length = %d, want %d", len(password), tt.length)
				}

				// Check that password contains all character types
				hasLower := strings.ContainsAny(password, lowercase)
				hasUpper := strings.ContainsAny(password, uppercase)
				hasDigit := strings.ContainsAny(password, digits)
				hasSymbol := strings.ContainsAny(password, symbols)

				if !hasLower || !hasUpper || !hasDigit || !hasSymbol {
					t.Errorf("Password missing required character types: lower=%v, upper=%v, digit=%v, symbol=%v",
						hasLower, hasUpper, hasDigit, hasSymbol)
				}
			}
		})
	}
}

func TestGenerateSecurePasswordWithConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *PasswordConfig
		wantErr bool
	}{
		{
			name: "default config",
			config: &PasswordConfig{
				Length:     12,
				UseLower:   true,
				UseUpper:   true,
				UseDigits:  true,
				UseSymbols: true,
			},
			wantErr: false,
		},
		{
			name: "only lowercase",
			config: &PasswordConfig{
				Length:     10,
				UseLower:   true,
				UseUpper:   false,
				UseDigits:  false,
				UseSymbols: false,
			},
			wantErr: false,
		},
		{
			name: "no character sets",
			config: &PasswordConfig{
				Length:     10,
				UseLower:   false,
				UseUpper:   false,
				UseDigits:  false,
				UseSymbols: false,
			},
			wantErr: true,
		},
		{
			name:    "nil config",
			config:  nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			password, err := GenerateSecurePasswordWithConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateSecurePasswordWithConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.config != nil {
				if len(password) != tt.config.Length {
					t.Errorf("GenerateSecurePasswordWithConfig() length = %d, want %d", len(password), tt.config.Length)
				}
			}
		})
	}
}

func TestGenerateSecureToken(t *testing.T) {
	token, err := GenerateSecureToken()
	if err != nil {
		t.Fatalf("GenerateSecureToken() error = %v", err)
	}

	if len(token) != DefaultTokenLength*2 { // hex encoding doubles the length
		t.Errorf("GenerateSecureToken() length = %d, want %d", len(token), DefaultTokenLength*2)
	}

	// Test that tokens are different
	token2, err := GenerateSecureToken()
	if err != nil {
		t.Fatalf("GenerateSecureToken() error = %v", err)
	}

	if token == token2 {
		t.Error("GenerateSecureToken() generated identical tokens")
	}
}

func TestGenerateSecureTokenWithLength(t *testing.T) {
	tests := []struct {
		name    string
		length  int
		wantErr bool
	}{
		{
			name:    "valid length",
			length:  16,
			wantErr: false,
		},
		{
			name:    "zero length",
			length:  0,
			wantErr: true,
		},
		{
			name:    "negative length",
			length:  -1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateSecureTokenWithLength(tt.length)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateSecureTokenWithLength() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(token) != tt.length*2 { // hex encoding doubles the length
					t.Errorf("GenerateSecureTokenWithLength() length = %d, want %d", len(token), tt.length*2)
				}
			}
		})
	}
}

func TestHashToken(t *testing.T) {
	tests := []struct {
		name  string
		token string
		want  string
	}{
		{
			name:  "valid token",
			token: "my-secret-token",
			want:  "a1b2c3d4e5f6", // This will be different, just checking it's not empty
		},
		{
			name:  "empty token",
			token: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashed := HashToken(tt.token)
			if tt.token == "" && hashed != "" {
				t.Errorf("HashToken() for empty token = %v, want empty string", hashed)
			}
			if tt.token != "" && hashed == "" {
				t.Errorf("HashToken() returned empty hash for non-empty token")
			}
		})
	}

	// Test that same token produces same hash
	token := "test-token"
	hash1 := HashToken(token)
	hash2 := HashToken(token)
	if hash1 != hash2 {
		t.Errorf("HashToken() produced different hashes for same token: %v != %v", hash1, hash2)
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	token, err := GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}

	if len(token) != DefaultRefreshTokenLength*2 { // hex encoding doubles the length
		t.Errorf("GenerateRefreshToken() length = %d, want %d", len(token), DefaultRefreshTokenLength*2)
	}
}

func TestGenerateRefreshTokenWithLength(t *testing.T) {
	tests := []struct {
		name    string
		length  int
		wantErr bool
	}{
		{
			name:    "valid length",
			length:  64,
			wantErr: false,
		},
		{
			name:    "minimum length",
			length:  32,
			wantErr: false,
		},
		{
			name:    "too short",
			length:  31,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateRefreshTokenWithLength(tt.length)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateRefreshTokenWithLength() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(token) != tt.length*2 { // hex encoding doubles the length
					t.Errorf("GenerateRefreshTokenWithLength() length = %d, want %d", len(token), tt.length*2)
				}
			}
		})
	}
}

func TestValidatePasswordStrength(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "strong password",
			password: "MySecurePass123!",
			wantErr:  false,
		},
		{
			name:     "too short",
			password: "Short1!",
			wantErr:  true,
		},
		{
			name:     "no lowercase",
			password: "MYSECUREPASS123!",
			wantErr:  true,
		},
		{
			name:     "no uppercase",
			password: "mysecurepass123!",
			wantErr:  true,
		},
		{
			name:     "no digits",
			password: "MySecurePass!",
			wantErr:  true,
		},
		{
			name:     "no symbols",
			password: "MySecurePass123",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePasswordStrength(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePasswordStrength() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDefaultPasswordConfig(t *testing.T) {
	config := DefaultPasswordConfig()

	if config.Length != DefaultPasswordLength {
		t.Errorf("DefaultPasswordConfig().Length = %d, want %d", config.Length, DefaultPasswordLength)
	}

	if !config.UseLower || !config.UseUpper || !config.UseDigits || !config.UseSymbols {
		t.Error("DefaultPasswordConfig() should enable all character sets")
	}
}
