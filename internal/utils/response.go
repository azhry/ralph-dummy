package utils

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"wedding-invitation-backend/internal/domain/repository"
)

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// APIErrorResponse represents a detailed error response
type APIErrorResponse struct {
	Success bool                   `json:"success"`
	Error   string                 `json:"error"`
	Details map[string]interface{} `json:"details,omitempty"`
	Code    string                 `json:"code,omitempty"`
}

// PaginationResponse represents a paginated response
type PaginationResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Count   int64       `json:"count"`
	Page    int         `json:"page"`
	Size    int         `json:"size"`
	Total   int64       `json:"total"`
}

// Response sends a standard JSON response
func Response(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, APIResponse{
		Success: statusCode < 400,
		Data:    data,
	})
}

// SuccessResponse sends a success response
func SuccessResponse(c *gin.Context, message string) {
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: message,
	})
}

// ErrorResponse sends an error response
func ErrorResponse(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, APIResponse{
		Success: false,
		Error:   message,
	})
}

// ValidationErrorResponse sends a validation error response
func ValidationErrorResponse(c *gin.Context, validationErrors map[string]string) {
	details := make(map[string]interface{})
	for key, value := range validationErrors {
		details[key] = value
	}
	c.JSON(http.StatusBadRequest, APIErrorResponse{
		Success: false,
		Error:   "Validation failed",
		Details: details,
		Code:    "validation_error",
	})
}

// PaginatedResponse sends a paginated response
func PaginatedResponse(c *gin.Context, statusCode int, data interface{}, count, total int64, page, size int) {
	c.JSON(statusCode, PaginationResponse{
		Success: true,
		Data:    data,
		Count:   count,
		Page:    page,
		Size:    size,
		Total:   total,
	})
}

// GetUserIDFromContext extracts user ID from Gin context
func GetUserIDFromContext(c *gin.Context) (primitive.ObjectID, error) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		return primitive.NilObjectID, errors.New("user ID not found in context")
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		return primitive.NilObjectID, errors.New("invalid user ID format")
	}

	return userID, nil
}

// ParsePaginationParams extracts pagination parameters from query string
func ParsePaginationParams(c *gin.Context) (int, int) {
	// Default page and size
	page := 1
	size := 20

	// Parse page
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	// Parse size with maximum limit
	if sizeStr := c.Query("size"); sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 && s <= 100 {
			size = s
		} else if s > 100 {
			size = 100 // Cap at 100
		}
	}

	return page, size
}

// GetPaginationHeaders returns pagination headers for list responses
func GetPaginationHeaders(count, total int64, page, size int) map[string]string {
	return map[string]string{
		"X-Total-Count":  strconv.FormatInt(total, 10),
		"X-Page-Count":   strconv.FormatInt((total+int64(size)-1)/int64(size), 10),
		"X-Current-Page": strconv.Itoa(page),
		"X-Page-Size":    strconv.Itoa(size),
	}
}

// SetPaginationHeaders sets pagination headers in Gin context
func SetPaginationHeaders(c *gin.Context, count, total int64, page, size int) {
	headers := GetPaginationHeaders(count, total, page, size)
	for key, value := range headers {
		c.Header(key, value)
	}
}

// HandleError handles different types of errors and returns appropriate responses
func HandleError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	// Handle validation errors
	if validationErr, ok := err.(*ValidationError); ok {
		details := make(map[string]string)
		for i, errorMsg := range validationErr.Errors {
			details["field_"+strconv.Itoa(i+1)] = errorMsg
		}
		ValidationErrorResponse(c, details)
		return
	}

	// Handle specific known errors
	switch err {
	case repository.ErrNotFound:
		ErrorResponse(c, http.StatusNotFound, "Resource not found")
	default:
		// Default internal server error
		ErrorResponse(c, http.StatusInternalServerError, "Internal server error")
	}
}
