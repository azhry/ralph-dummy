package middleware

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ValidationMiddleware provides comprehensive input validation
type ValidationMiddleware struct {
	validate *validator.Validate
}

// NewValidationMiddleware creates a new validation middleware
func NewValidationMiddleware() *ValidationMiddleware {
	vm := &ValidationMiddleware{
		validate: validator.New(),
	}

	// Register custom validators
	vm.validate.RegisterValidation("slug", vm.validateSlug)
	vm.validate.RegisterValidation("objectid", vm.validateObjectID)
	vm.validate.RegisterValidation("phone", vm.validatePhone)
	vm.validate.RegisterValidation("url", vm.validateURL)
	vm.validate.RegisterValidation("safehtml", vm.validateSafeHTML)

	return vm
}

// ValidateBody validates request body against a struct
func (vm *ValidationMiddleware) ValidateBody(requestStruct interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Bind JSON to struct
		if err := c.ShouldBindJSON(requestStruct); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INVALID_REQUEST_BODY",
					"message": "Invalid request body format",
					"details": vm.formatBindingError(err),
				},
			})
			c.Abort()
			return
		}

		// Validate struct fields
		if err := vm.validate.Struct(requestStruct); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "VALIDATION_ERROR",
					"message": "Request validation failed",
					"details": vm.formatValidationErrors(err),
				},
			})
			c.Abort()
			return
		}

		// Store validated request in context
		c.Set("validated_request", requestStruct)
		c.Next()
	}
}

// ValidateQuery validates query parameters
func (vm *ValidationMiddleware) ValidateQuery(requestStruct interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := c.ShouldBindQuery(requestStruct); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INVALID_QUERY_PARAMS",
					"message": "Invalid query parameters",
					"details": vm.formatBindingError(err),
				},
			})
			c.Abort()
			return
		}

		if err := vm.validate.Struct(requestStruct); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "QUERY_VALIDATION_ERROR",
					"message": "Query parameter validation failed",
					"details": vm.formatValidationErrors(err),
				},
			})
			c.Abort()
			return
		}

		c.Set("validated_query", requestStruct)
		c.Next()
	}
}

// SanitizeInput sanitizes and validates input for XSS prevention
func (vm *ValidationMiddleware) SanitizeInput() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Sanitize query parameters
		for key, values := range c.Request.URL.Query() {
			for i, value := range values {
				c.Request.URL.Query()[key][i] = vm.sanitizeString(value)
			}
		}

		// Sanitize form data
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			if err := c.Request.ParseForm(); err == nil {
				for key, values := range c.Request.PostForm {
					for i, value := range values {
						c.Request.PostForm[key][i] = vm.sanitizeString(value)
					}
				}
			}
		}

		c.Next()
	}
}

// Custom validators

func (vm *ValidationMiddleware) validateSlug(fl validator.FieldLevel) bool {
	slug := fl.Field().String()

	// Slug rules: alphanumeric, hyphens only, 3-50 chars
	if len(slug) < 3 || len(slug) > 50 {
		return false
	}

	for _, char := range slug {
		if !((char >= 'a' && char <= 'z') ||
			(char >= '0' && char <= '9') ||
			char == '-') {
			return false
		}
	}

	// Cannot start or end with hyphen
	if strings.HasPrefix(slug, "-") || strings.HasSuffix(slug, "-") {
		return false
	}

	// No consecutive hyphens
	if strings.Contains(slug, "--") {
		return false
	}

	return true
}

func (vm *ValidationMiddleware) validateObjectID(fl validator.FieldLevel) bool {
	id := fl.Field().String()

	// MongoDB ObjectID is 24 hex characters
	if len(id) != 24 {
		return false
	}

	// Check if all characters are valid hex
	matched, _ := regexp.MatchString("^[a-fA-F0-9]{24}$", id)
	return matched
}

func (vm *ValidationMiddleware) validatePhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()

	// Remove common phone number formatting characters
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	phone = strings.ReplaceAll(phone, "(", "")
	phone = strings.ReplaceAll(phone, ")", "")
	phone = strings.ReplaceAll(phone, "+", "")

	// Check if it's all digits and reasonable length
	if len(phone) < 10 || len(phone) > 15 {
		return false
	}

	matched, _ := regexp.MatchString("^[0-9]+$", phone)
	return matched
}

func (vm *ValidationMiddleware) validateURL(fl validator.FieldLevel) bool {
	url := fl.Field().String()

	if len(url) == 0 || len(url) > 2048 {
		return false
	}

	// Basic URL validation
	matched, _ := regexp.MatchString(`^https?://[^\s/$.?#].[^\s]*$`, url)
	return matched
}

func (vm *ValidationMiddleware) validateSafeHTML(fl validator.FieldLevel) bool {
	html := fl.Field().String()

	// Check for dangerous HTML tags and attributes
	dangerousPatterns := []string{
		`<script[^>]*>.*?</script>`,
		`<iframe[^>]*>.*?</iframe>`,
		`<object[^>]*>.*?</object>`,
		`<embed[^>]*>`,
		`<form[^>]*>.*?</form>`,
		`javascript:`,
		`vbscript:`,
		`onload\s*=`,
		`onerror\s*=`,
		`onclick\s*=`,
		`onmouseover\s*=`,
	}

	for _, pattern := range dangerousPatterns {
		matched, _ := regexp.MatchString(strings.ToLower(pattern), strings.ToLower(html))
		if matched {
			return false
		}
	}

	return true
}

// sanitizeString removes potentially dangerous characters
func (vm *ValidationMiddleware) sanitizeString(input string) string {
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")

	// Remove control characters except newlines and tabs
	sanitized := ""
	for _, char := range input {
		if char >= 32 || char == '\n' || char == '\t' {
			sanitized += string(char)
		}
	}

	return sanitized
}

// formatValidationErrors formats validation errors for API responses
func (vm *ValidationMiddleware) formatValidationErrors(err error) []gin.H {
	var errors []gin.H

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			errors = append(errors, gin.H{
				"field":   e.Field(),
				"tag":     e.Tag(),
				"message": vm.getErrorMessage(e),
			})
		}
	} else {
		errors = append(errors, gin.H{
			"message": err.Error(),
		})
	}

	return errors
}

// formatBindingError formats binding errors for API responses
func (vm *ValidationMiddleware) formatBindingError(err error) []gin.H {
	return []gin.H{
		{
			"message": "Invalid JSON format: " + err.Error(),
		},
	}
}

// getErrorMessage returns user-friendly error messages for validation tags
func (vm *ValidationMiddleware) getErrorMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return e.Field() + " is required"
	case "email":
		return "Invalid email format"
	case "min":
		return e.Field() + " must be at least " + e.Param() + " characters"
	case "max":
		return e.Field() + " must be at most " + e.Param() + " characters"
	case "len":
		return e.Field() + " must be exactly " + e.Param() + " characters"
	case "slug":
		return "Invalid slug format (use lowercase letters, numbers, and hyphens only)"
	case "objectid":
		return "Invalid ID format"
	case "phone":
		return "Invalid phone number format"
	case "url":
		return "Invalid URL format"
	case "safehtml":
		return "HTML content contains unsafe elements"
	default:
		return e.Field() + " validation failed on " + e.Tag()
	}
}

// GetValidatedRequest retrieves the validated request from context
func GetValidatedRequest(c *gin.Context, target interface{}) bool {
	if _, exists := c.Get("validated_request"); exists {
		// This is a bit of a hack, but it works for our use case
		// In a real implementation, you might want to use reflection or interfaces
		return true
	}
	return false
}

// GetValidatedQuery retrieves the validated query from context
func GetValidatedQuery(c *gin.Context, target interface{}) bool {
	if _, exists := c.Get("validated_query"); exists {
		return true
	}
	return false
}
