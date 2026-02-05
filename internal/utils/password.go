package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidPasswordFormat = errors.New("password must contain at least 8 characters, one uppercase, one lowercase, one digit, and one special character")
	ErrPasswordMismatch      = errors.New("passwords do not match")
	ErrWeakPassword          = errors.New("password is too weak or commonly used")
)

type PasswordValidator struct {
	minLength       int
	requireUpper    bool
	requireLower    bool
	requireDigit    bool
	requireSpecial  bool
	commonPasswords []string
}

func NewPasswordValidator() *PasswordValidator {
	return &PasswordValidator{
		minLength:      8,
		requireUpper:   true,
		requireLower:   true,
		requireDigit:   true,
		requireSpecial: true,
		commonPasswords: []string{
			"password", "123456", "123456789", "qwerty", "abc123",
			"password123", "admin", "letmein", "welcome", "monkey",
			"1234567890", "password1", "123123", "qwerty123", "password!",
		},
	}
}

func (pv *PasswordValidator) Validate(password string) error {
	// Check minimum length
	if len(password) < pv.minLength {
		return fmt.Errorf("password must be at least %d characters long", pv.minLength)
	}

	// Check for common passwords
	for _, common := range pv.commonPasswords {
		if subtle.ConstantTimeCompare([]byte(password), []byte(common)) == 1 {
			return ErrWeakPassword
		}
	}

	// Check character requirements
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password)

	if pv.requireUpper && !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}

	if pv.requireLower && !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}

	if pv.requireDigit && !hasDigit {
		return fmt.Errorf("password must contain at least one digit")
	}

	if pv.requireSpecial && !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedPassword), nil
}

func HashPasswordWithCost(password string, cost int) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedPassword), nil
}

func CheckPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func GenerateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate secure token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

func GenerateResetToken() (string, error) {
	return GenerateSecureToken(32)
}

func GenerateVerificationToken() (string, error) {
	return GenerateSecureToken(32)
}

func IsStrongPassword(password string) bool {
	validator := NewPasswordValidator()
	return validator.Validate(password) == nil
}

func ComparePasswords(password1, password2 string) bool {
	return subtle.ConstantTimeCompare([]byte(password1), []byte(password2)) == 1
}
