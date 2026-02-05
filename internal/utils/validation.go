package utils

import (
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func init() {
	// Register custom validators if needed
	validate.RegisterValidation("custom", validateCustom)
}

// ValidateStruct validates a struct and returns validation errors
func ValidateStruct(s interface{}) error {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	// Collect all validation errors
	var errors []string
	for _, err := range err.(validator.ValidationErrors) {
		errors = append(errors, getErrorMessage(err))
	}

	return &ValidationError{Errors: errors}
}

// ValidateVar validates a single variable
func ValidateVar(field interface{}, tag string) error {
	err := validate.Var(field, tag)
	if err == nil {
		return nil
	}

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		return &ValidationError{Errors: []string{getErrorMessage(validationErrors[0])}}
	}

	return err
}

// ValidationError represents validation errors
type ValidationError struct {
	Errors []string
}

func (e *ValidationError) Error() string {
	return strings.Join(e.Errors, "; ")
}

// getErrorMessage converts validation error to human-readable message
func getErrorMessage(err validator.FieldError) string {
	field := err.Field()
	tag := err.Tag()
	param := err.Param()

	switch tag {
	case "required":
		return field + " is required"
	case "min":
		if field == "Password" || strings.Contains(strings.ToLower(field), "password") {
			return field + " must be at least " + param + " characters long"
		}
		return field + " must be at least " + param
	case "max":
		if field == "Password" || strings.Contains(strings.ToLower(field), "password") {
			return field + " must be at most " + param + " characters long"
		}
		return field + " must be at most " + param
	case "email":
		return field + " must be a valid email address"
	case "len":
		return field + " must be " + param + " characters long"
	case "numeric":
		return field + " must be numeric"
	case "alphanum":
		return field + " must contain only alphanumeric characters"
	case "alpha":
		return field + " must contain only letters"
	case "oneof":
		return field + " must be one of: " + param
	case "e164":
		return field + " must be a valid phone number in E.164 format"
	case "url":
		return field + " must be a valid URL"
	case "uuid":
		return field + " must be a valid UUID"
	case "datetime":
		return field + " must be a valid datetime"
	case "date":
		return field + " must be a valid date"
	case "gte":
		return field + " must be greater than or equal to " + param
	case "gt":
		return field + " must be greater than " + param
	case "lte":
		return field + " must be less than or equal to " + param
	case "lt":
		return field + " must be less than " + param
	default:
		return field + " is invalid"
	}
}

// validateCustom is a placeholder for custom validation logic
func validateCustom(fl validator.FieldLevel) bool {
	// Add custom validation logic here
	return true
}

// SanitizeString sanitizes a string by trimming spaces and converting to lowercase if needed
func SanitizeString(input string, toLower bool) string {
	result := strings.TrimSpace(input)
	if toLower {
		result = strings.ToLower(result)
	}
	return result
}

// SanitizeEmail sanitizes an email address
func SanitizeEmail(email string) string {
	return SanitizeString(strings.ToLower(email), false)
}

// IsEmpty checks if a value is empty or zero
func IsEmpty(value interface{}) bool {
	if value == nil {
		return true
	}

	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v) == ""
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(value).Int() == 0
	case uint, uint8, uint16, uint32, uint64:
		return reflect.ValueOf(value).Uint() == 0
	case float32, float64:
		return reflect.ValueOf(value).Float() == 0
	case bool:
		return !v
	case []interface{}:
		return len(v) == 0
	case map[string]interface{}:
		return len(v) == 0
	default:
		return reflect.ValueOf(value).IsZero()
	}
}

// ValidatePasswordStrength checks password strength based on common criteria
func ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return &ValidationError{Errors: []string{"Password must be at least 8 characters long"}}
	}

	if len(password) > 72 {
		return &ValidationError{Errors: []string{"Password must be at most 72 characters long"}}
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasNumber = true
		case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;:,.<>?", char):
			hasSpecial = true
		}
	}

	var errors []string
	if !hasUpper {
		errors = append(errors, "Password must contain at least one uppercase letter")
	}
	if !hasLower {
		errors = append(errors, "Password must contain at least one lowercase letter")
	}
	if !hasNumber {
		errors = append(errors, "Password must contain at least one number")
	}
	if !hasSpecial {
		errors = append(errors, "Password must contain at least one special character")
	}

	if len(errors) > 0 {
		return &ValidationError{Errors: errors}
	}

	return nil
}

// ValidateSlug validates a slug string
func ValidateSlug(slug string) error {
	if slug == "" {
		return &ValidationError{Errors: []string{"Slug cannot be empty"}}
	}

	if len(slug) < 3 {
		return &ValidationError{Errors: []string{"Slug must be at least 3 characters long"}}
	}

	if len(slug) > 50 {
		return &ValidationError{Errors: []string{"Slug must be at most 50 characters long"}}
	}

	// Slug should only contain alphanumeric characters and hyphens
	slugRegex := regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	if !slugRegex.MatchString(slug) {
		return &ValidationError{Errors: []string{"Slug can only contain lowercase letters, numbers, and hyphens"}}
	}

	return nil
}

// ValidateHexColor validates a hex color string
func ValidateHexColor(color string) error {
	if color == "" {
		return nil // Empty color is allowed
	}

	// Support both #RGB and #RRGGBB formats
	hexColorRegex := regexp.MustCompile(`^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$`)
	if !hexColorRegex.MatchString(color) {
		return &ValidationError{Errors: []string{"Invalid hex color format. Use #RGB or #RRGGBB"}}
	}

	return nil
}

// SanitizeSlug sanitizes a string to be used as a slug
func SanitizeSlug(input string) string {
	// Convert to lowercase
	result := strings.ToLower(input)

	// Replace spaces and special characters with hyphens
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	result = reg.ReplaceAllString(result, "-")

	// Remove leading and trailing hyphens
	result = strings.Trim(result, "-")

	// Remove multiple consecutive hyphens
	reg = regexp.MustCompile(`-+`)
	result = reg.ReplaceAllString(result, "-")

	return result
}

// Contains checks if a slice contains a specific string
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
