package validator

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// PasswordStrength represents password strength requirements
type PasswordStrength struct {
	MinLength      int
	RequireUpper   bool
	RequireLower   bool
	RequireNumber  bool
	RequireSpecial bool
}

// DefaultPasswordStrength returns the default password requirements
func DefaultPasswordStrength() PasswordStrength {
	return PasswordStrength{
		MinLength:      8,
		RequireUpper:   true,
		RequireLower:   true,
		RequireNumber:  true,
		RequireSpecial: false, // Optional for better UX
	}
}

// ValidatePassword validates password against strength requirements
func ValidatePassword(password string, strength PasswordStrength) error {
	if len(password) < strength.MinLength {
		return fmt.Errorf("password must be at least %d characters long", strength.MinLength)
	}

	if len(password) > 128 {
		return fmt.Errorf("password must be less than 128 characters")
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if strength.RequireUpper && !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}

	if strength.RequireLower && !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}

	if strength.RequireNumber && !hasNumber {
		return fmt.Errorf("password must contain at least one number")
	}

	if strength.RequireSpecial && !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	// Check for common weak passwords
	if isCommonPassword(password) {
		return fmt.Errorf("password is too common, please choose a stronger password")
	}

	return nil
}

// isCommonPassword checks if password is in common passwords list
func isCommonPassword(password string) bool {
	// Top 100 most common passwords
	commonPasswords := []string{
		"password", "123456", "12345678", "qwerty", "abc123", "monkey",
		"1234567", "letmein", "trustno1", "dragon", "baseball", "111111",
		"iloveyou", "master", "sunshine", "ashley", "bailey", "passw0rd",
		"shadow", "123123", "654321", "superman", "qazwsx", "michael",
		"football", "password1", "welcome", "admin", "test", "guest",
	}

	lowerPassword := strings.ToLower(password)
	for _, common := range commonPasswords {
		if lowerPassword == common {
			return true
		}
	}

	// Check for simple patterns
	simplePatterns := []string{
		`^123+$`,      // 123123...
		`^abc+$`,      // abcabc...
		`^(.)\1{5,}$`, // aaaaaaa...
		`^[0-9]+$`,    // only numbers
		`^[a-z]+$`,    // only lowercase
		`^[A-Z]+$`,    // only uppercase
	}

	for _, pattern := range simplePatterns {
		if matched, _ := regexp.MatchString(pattern, password); matched {
			return true
		}
	}

	return false
}
