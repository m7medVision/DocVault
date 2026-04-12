// Package auth provides password hashing and validation for user authentication.
package auth

import (
	"fmt"
	"regexp"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

const (
	// Bcrypt cost factor (number of hashing rounds)
	// Cost 12 = 2^12 iterations (~250ms on modern hardware)
	// Higher is more secure but slower
	BcryptCost = 12

	// Password requirements
	MinPasswordLength     = 8
	MaxPasswordLength     = 128
	RequireUppercase      = true
	RequireLowercase      = true
	RequireDigit          = true
	RequireSpecialChar    = true
)

// HashPassword generates a bcrypt hash of the password.
// Returns the hash string which can be stored in the database.
func HashPassword(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	// Validate password before hashing
	if err := ValidatePasswordStrength(password); err != nil {
		return "", fmt.Errorf("password validation failed: %w", err)
	}

	// Generate bcrypt hash
	hash, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hash), nil
}

// ComparePassword compares a plaintext password with a bcrypt hash.
// Returns nil if they match, error otherwise.
func ComparePassword(hash, password string) error {
	if hash == "" || password == "" {
		return fmt.Errorf("hash and password cannot be empty")
	}

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return fmt.Errorf("invalid password")
		}
		return fmt.Errorf("password comparison failed: %w", err)
	}

	return nil
}

// ValidatePasswordStrength checks if a password meets complexity requirements.
func ValidatePasswordStrength(password string) error {
	if len(password) < MinPasswordLength {
		return fmt.Errorf("password must be at least %d characters long", MinPasswordLength)
	}

	if len(password) > MaxPasswordLength {
		return fmt.Errorf("password must be at most %d characters long", MaxPasswordLength)
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasDigit   bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if RequireUppercase && !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}

	if RequireLowercase && !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}

	if RequireDigit && !hasDigit {
		return fmt.Errorf("password must contain at least one digit")
	}

	if RequireSpecialChar && !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}

// ValidateEmail checks if an email address is valid using RFC 5322 regex.
func ValidateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email cannot be empty")
	}

	// RFC 5322 compliant email regex (simplified but robust)
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+/=?^_{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)

	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}

	if len(email) > 254 {
		return fmt.Errorf("email address too long (max 254 characters)")
	}

	return nil
}

// PasswordStrengthScore returns a score (0-100) indicating password strength.
// Higher is better. Useful for UI feedback.
func PasswordStrengthScore(password string) int {
	if len(password) == 0 {
		return 0
	}

	score := 0

	// Length score (up to 40 points)
	lengthScore := len(password) * 2
	if lengthScore > 40 {
		lengthScore = 40
	}
	score += lengthScore

	// Character variety (up to 60 points)
	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, char := range password {
		if unicode.IsUpper(char) {
			hasUpper = true
		}
		if unicode.IsLower(char) {
			hasLower = true
		}
		if unicode.IsDigit(char) {
			hasDigit = true
		}
		if unicode.IsPunct(char) || unicode.IsSymbol(char) {
			hasSpecial = true
		}
	}

	if hasUpper {
		score += 15
	}
	if hasLower {
		score += 15
	}
	if hasDigit {
		score += 15
	}
	if hasSpecial {
		score += 15
	}

	// Cap at 100
	if score > 100 {
		score = 100
	}

	return score
}
