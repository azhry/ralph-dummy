package middleware

import (
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ErrorConfig holds configuration for error handling
type ErrorConfig struct {
	// Whether to include stack traces in responses
	IncludeStackTrace bool
	// Whether to log detailed error information
	LogDetailedErrors bool
	// Custom error handlers
	CustomHandlers map[string]func(c *gin.Context, err error)
}

// DefaultErrorConfig returns default error configuration
func DefaultErrorConfig() ErrorConfig {
	return ErrorConfig{
		IncludeStackTrace: false,
		LogDetailedErrors: true,
		CustomHandlers:    make(map[string]func(c *gin.Context, err error)),
	}
}

// DevelopmentErrorConfig returns error configuration for development
func DevelopmentErrorConfig() ErrorConfig {
	return ErrorConfig{
		IncludeStackTrace: true,
		LogDetailedErrors: true,
		CustomHandlers:    make(map[string]func(c *gin.Context, err error)),
	}
}

// ErrorHandler provides comprehensive error handling
type ErrorHandler struct {
	logger *zap.Logger
	config ErrorConfig
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(logger *zap.Logger, config ErrorConfig) *ErrorHandler {
	return &ErrorHandler{
		logger: logger,
		config: config,
	}
}

// Middleware returns the Gin middleware for error handling
func (eh *ErrorHandler) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				eh.handlePanic(c, err)
			}
		}()

		c.Next()

		// Handle any errors that occurred during the request
		if len(c.Errors) > 0 {
			eh.handleRequestErrors(c)
		}
	}
}

// handlePanic handles panics that occurred during request processing
func (eh *ErrorHandler) handlePanic(c *gin.Context, recovered interface{}) {
	// Log the panic with full details
	eh.logger.Error("panic recovered",
		zap.Any("error", recovered),
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
		zap.String("client_ip", c.ClientIP()),
		zap.String("user_agent", c.GetHeader("User-Agent")),
		zap.String("stack", string(debug.Stack())),
	)

	// Prepare error response
	errorResponse := gin.H{
		"success": false,
		"error": gin.H{
			"code":    "INTERNAL_SERVER_ERROR",
			"message": "An unexpected error occurred",
		},
	}

	// Include stack trace in development mode
	if eh.config.IncludeStackTrace {
		errorResponse["error"].(gin.H)["stack_trace"] = string(debug.Stack())
		errorResponse["error"].(gin.H)["debug_info"] = gin.H{
			"recovered": recovered,
			"path":      c.Request.URL.Path,
			"method":    c.Request.Method,
		}
	}

	c.JSON(http.StatusInternalServerError, errorResponse)
	c.Abort()
}

// handleRequestErrors handles errors that occurred during request processing
func (eh *ErrorHandler) handleRequestErrors(c *gin.Context) {
	// Get the last error (most recent)
	err := c.Errors.Last()

	// Log the error
	if eh.config.LogDetailedErrors {
		eh.logger.Error("request error",
			zap.Error(err.Err),
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
			zap.String("client_ip", c.ClientIP()),
		)
	}

	// Check for custom error handlers
	if handler, exists := eh.config.CustomHandlers[string(rune(err.Type))]; exists {
		handler(c, err.Err)
		return
	}

	// Handle common error types
	eh.handleCommonError(c, err.Err)
}

// handleCommonError handles common error types with appropriate HTTP status codes
func (eh *ErrorHandler) handleCommonError(c *gin.Context, err error) {
	errorMessage := err.Error()

	// Determine HTTP status code and error code based on error message
	statusCode, errorCode, message := eh.categorizeError(errorMessage)

	c.JSON(statusCode, gin.H{
		"success": false,
		"error": gin.H{
			"code":    errorCode,
			"message": message,
		},
	})
}

// categorizeError categorizes errors and returns appropriate status codes
func (eh *ErrorHandler) categorizeError(errorMessage string) (int, string, string) {
	errorMessage = strings.ToLower(errorMessage)

	// Validation errors
	if strings.Contains(errorMessage, "validation") ||
		strings.Contains(errorMessage, "invalid") ||
		strings.Contains(errorMessage, "required") ||
		strings.Contains(errorMessage, "format") {
		return http.StatusBadRequest, "VALIDATION_ERROR", "Invalid input data"
	}

	// Authentication errors
	if strings.Contains(errorMessage, "unauthorized") ||
		strings.Contains(errorMessage, "authentication") ||
		strings.Contains(errorMessage, "token") ||
		strings.Contains(errorMessage, "login") {
		return http.StatusUnauthorized, "AUTHENTICATION_ERROR", "Authentication required"
	}

	// Authorization errors
	if strings.Contains(errorMessage, "forbidden") ||
		strings.Contains(errorMessage, "permission") ||
		strings.Contains(errorMessage, "access") {
		return http.StatusForbidden, "AUTHORIZATION_ERROR", "Insufficient permissions"
	}

	// Not found errors
	if strings.Contains(errorMessage, "not found") ||
		strings.Contains(errorMessage, "exists") == false && strings.Contains(errorMessage, "exist") {
		return http.StatusNotFound, "NOT_FOUND", "Resource not found"
	}

	// Conflict errors
	if strings.Contains(errorMessage, "conflict") ||
		strings.Contains(errorMessage, "duplicate") ||
		strings.Contains(errorMessage, "already") {
		return http.StatusConflict, "CONFLICT", "Resource conflict"
	}

	// Rate limiting errors
	if strings.Contains(errorMessage, "rate limit") ||
		strings.Contains(errorMessage, "too many") {
		return http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "Too many requests"
	}

	// Default internal server error
	return http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred"
}

// RegisterCustomHandler registers a custom error handler for a specific error type
func (eh *ErrorHandler) RegisterCustomHandler(errorType string, handler func(c *gin.Context, err error)) {
	eh.config.CustomHandlers[errorType] = handler
}

// APIError represents a structured API error
type APIError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// Error implements the error interface
func (e *APIError) Error() string {
	return e.Message
}

// NewAPIError creates a new API error
func NewAPIError(code, message string) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
	}
}

// NewAPIErrorWithDetails creates a new API error with details
func NewAPIErrorWithDetails(code, message string, details interface{}) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// Common API error constructors
var (
	ErrInvalidInput        = NewAPIError("INVALID_INPUT", "Invalid input data")
	ErrUnauthorized        = NewAPIError("UNAUTHORIZED", "Authentication required")
	ErrForbidden           = NewAPIError("FORBIDDEN", "Insufficient permissions")
	ErrNotFound            = NewAPIError("NOT_FOUND", "Resource not found")
	ErrConflict            = NewAPIError("CONFLICT", "Resource conflict")
	ErrRateLimitExceeded   = NewAPIError("RATE_LIMIT_EXCEEDED", "Too many requests")
	ErrInternalError       = NewAPIError("INTERNAL_ERROR", "An unexpected error occurred")
	ErrValidationFailed    = NewAPIError("VALIDATION_FAILED", "Request validation failed")
	ErrInvalidToken        = NewAPIError("INVALID_TOKEN", "Invalid or expired token")
	ErrTokenRevoked        = NewAPIError("TOKEN_REVOKED", "Token has been revoked")
	ErrInvalidCredentials  = NewAPIError("INVALID_CREDENTIALS", "Invalid email or password")
	ErrAccountLocked       = NewAPIError("ACCOUNT_LOCKED", "Account is temporarily locked")
	ErrEmailNotVerified    = NewAPIError("EMAIL_NOT_VERIFIED", "Email address not verified")
	ErrWeakPassword        = NewAPIError("WEAK_PASSWORD", "Password does not meet security requirements")
	ErrEmailAlreadyExists  = NewAPIError("EMAIL_ALREADY_EXISTS", "Email address already registered")
	ErrInvalidEmail        = NewAPIError("INVALID_EMAIL", "Invalid email address")
	ErrInvalidSlug         = NewAPIError("INVALID_SLUG", "Invalid slug format")
	ErrWeddingNotFound     = NewAPIError("WEDDING_NOT_FOUND", "Wedding not found")
	ErrRSVPNotFound        = NewAPIError("RSVP_NOT_FOUND", "RSVP not found")
	ErrGuestNotFound       = NewAPIError("GUEST_NOT_FOUND", "Guest not found")
	ErrWeddingNotPublished = NewAPIError("WEDDING_NOT_PUBLISHED", "Wedding is not published")
	ErrRSVPClosed          = NewAPIError("RSVP_CLOSED", "RSVP period is closed")
	ErrInvalidFileType     = NewAPIError("INVALID_FILE_TYPE", "Invalid file type")
	ErrFileTooLarge        = NewAPIError("FILE_TOO_LARGE", "File size exceeds limit")
	ErrUploadFailed        = NewAPIError("UPLOAD_FAILED", "File upload failed")
)

// HandleAPIError handles API errors and returns appropriate responses
func HandleAPIError(c *gin.Context, err *APIError) {
	statusCode := getStatusCodeForError(err.Code)

	c.JSON(statusCode, gin.H{
		"success": false,
		"error": gin.H{
			"code":    err.Code,
			"message": err.Message,
			"details": err.Details,
		},
	})
}

// getStatusCodeForError returns HTTP status code for error code
func getStatusCodeForError(errorCode string) int {
	switch errorCode {
	case "INVALID_INPUT", "VALIDATION_FAILED", "INVALID_EMAIL", "INVALID_SLUG", "WEAK_PASSWORD":
		return http.StatusBadRequest
	case "UNAUTHORIZED", "INVALID_TOKEN", "TOKEN_REVOKED", "INVALID_CREDENTIALS":
		return http.StatusUnauthorized
	case "FORBIDDEN", "EMAIL_NOT_VERIFIED", "ACCOUNT_LOCKED":
		return http.StatusForbidden
	case "NOT_FOUND", "WEDDING_NOT_FOUND", "RSVP_NOT_FOUND", "GUEST_NOT_FOUND":
		return http.StatusNotFound
	case "CONFLICT", "EMAIL_ALREADY_EXISTS":
		return http.StatusConflict
	case "RATE_LIMIT_EXCEEDED":
		return http.StatusTooManyRequests
	case "INVALID_FILE_TYPE", "FILE_TOO_LARGE", "UPLOAD_FAILED":
		return http.StatusBadRequest
	case "WEDDING_NOT_PUBLISHED", "RSVPClosed":
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
