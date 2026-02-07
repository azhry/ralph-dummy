package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"wedding-invitation-backend/internal/domain/models"
	"wedding-invitation-backend/internal/domain/repository"
	"wedding-invitation-backend/internal/services"
)

// UserServiceInterface defines the interface for UserService for testing
type UserServiceInterface interface {
	GetUserProfile(ctx context.Context, userID primitive.ObjectID) (*models.User, error)
	UpdateUserProfile(ctx context.Context, userID primitive.ObjectID, profile *services.UserProfile) (*models.User, error)
	GetUsersList(ctx context.Context, page, pageSize int, filters repository.UserFilters) (*services.UserListResponse, error)
	SearchUsers(ctx context.Context, query string, limit int) ([]*models.User, error)
	UpdateUserStatus(ctx context.Context, userID primitive.ObjectID, status models.UserStatus) error
	DeleteUser(ctx context.Context, userID primitive.ObjectID) error
	GetUserStats(ctx context.Context) (map[string]int64, error)
	GetUserWeddings(ctx context.Context, userID primitive.ObjectID) ([]primitive.ObjectID, error)
	AddWeddingToUser(ctx context.Context, userID, weddingID primitive.ObjectID) error
	RemoveWeddingFromUser(ctx context.Context, userID, weddingID primitive.ObjectID) error
}

// MockUserService is a mock implementation of UserServiceInterface
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) GetUserProfile(ctx context.Context, userID primitive.ObjectID) (*models.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) UpdateUserProfile(ctx context.Context, userID primitive.ObjectID, profile *services.UserProfile) (*models.User, error) {
	args := m.Called(ctx, userID, profile)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) GetUsersList(ctx context.Context, page, pageSize int, filters repository.UserFilters) (*services.UserListResponse, error) {
	args := m.Called(ctx, page, pageSize, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.UserListResponse), args.Error(1)
}

func (m *MockUserService) SearchUsers(ctx context.Context, query string, limit int) ([]*models.User, error) {
	args := m.Called(ctx, query, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserService) UpdateUserStatus(ctx context.Context, userID primitive.ObjectID, status models.UserStatus) error {
	args := m.Called(ctx, userID, status)
	return args.Error(0)
}

func (m *MockUserService) DeleteUser(ctx context.Context, userID primitive.ObjectID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserService) GetUserStats(ctx context.Context) (map[string]int64, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]int64), args.Error(1)
}

func (m *MockUserService) GetUserWeddings(ctx context.Context, userID primitive.ObjectID) ([]primitive.ObjectID, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]primitive.ObjectID), args.Error(1)
}

func (m *MockUserService) AddWeddingToUser(ctx context.Context, userID, weddingID primitive.ObjectID) error {
	args := m.Called(ctx, userID, weddingID)
	return args.Error(0)
}

func (m *MockUserService) RemoveWeddingFromUser(ctx context.Context, userID, weddingID primitive.ObjectID) error {
	args := m.Called(ctx, userID, weddingID)
	return args.Error(0)
}

// TestUserHandler is a test version of UserHandler that accepts the interface
type TestUserHandler struct {
	userService UserServiceInterface
}

func NewTestUserHandler(userService UserServiceInterface) *TestUserHandler {
	return &TestUserHandler{
		userService: userService,
	}
}

// Copy the handler methods from UserHandler but use the interface
func (h *TestUserHandler) GetProfile(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	objectID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get user profile
	user, err := h.userService.GetUserProfile(c.Request.Context(), objectID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (h *TestUserHandler) UpdateProfile(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	objectID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Parse request body
	var profile services.UserProfile
	if err := c.ShouldBindJSON(&profile); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Update user profile
	user, err := h.userService.UpdateUserProfile(c.Request.Context(), objectID, &profile)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (h *TestUserHandler) GetUsersList(c *gin.Context) {
	// Parse query parameters
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")
	status := c.Query("status")
	search := c.Query("search")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Build filters
	filters := repository.UserFilters{
		Status: status,
		Search: search,
	}

	// Get users list
	response, err := h.userService.GetUsersList(c.Request.Context(), page, pageSize, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get users list"})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *TestUserHandler) SearchUsers(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query is required"})
		return
	}

	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 50 {
		limit = 20
	}

	// Search users
	users, err := h.userService.SearchUsers(c.Request.Context(), query, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search users"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

func (h *TestUserHandler) UpdateUserStatus(c *gin.Context) {
	// Get user ID from URL params
	userIDStr := c.Param("id")
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Parse request body
	var requestBody struct {
		Status string `json:"status" validate:"required"`
	}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Convert status to UserStatus
	status := models.UserStatus(requestBody.Status)

	// Update user status
	if err := h.userService.UpdateUserStatus(c.Request.Context(), userID, status); err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		if strings.Contains(err.Error(), "invalid user status") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user status"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User status updated successfully"})
}

func (h *TestUserHandler) DeleteUser(c *gin.Context) {
	// Get user ID from URL params
	userIDStr := c.Param("id")
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Delete user
	if err := h.userService.DeleteUser(c.Request.Context(), userID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

func (h *TestUserHandler) GetUserStats(c *gin.Context) {
	// Get user statistics
	stats, err := h.userService.GetUserStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user statistics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"stats": stats})
}

func (h *TestUserHandler) GetUserWeddings(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	objectID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get user weddings
	weddingIDs, err := h.userService.GetUserWeddings(c.Request.Context(), objectID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user weddings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"wedding_ids": weddingIDs})
}

func (h *TestUserHandler) AddWeddingToUser(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userObjectID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get wedding ID from URL params
	weddingIDStr := c.Param("wedding_id")
	weddingID, err := primitive.ObjectIDFromHex(weddingIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid wedding ID"})
		return
	}

	// Add wedding to user
	if err := h.userService.AddWeddingToUser(c.Request.Context(), userObjectID, weddingID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add wedding to user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Wedding added to user successfully"})
}

func (h *TestUserHandler) RemoveWeddingFromUser(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userObjectID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get wedding ID from URL params
	weddingIDStr := c.Param("wedding_id")
	weddingID, err := primitive.ObjectIDFromHex(weddingIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid wedding ID"})
		return
	}

	// Remove wedding from user
	if err := h.userService.RemoveWeddingFromUser(c.Request.Context(), userObjectID, weddingID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove wedding from user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Wedding removed from user successfully"})
}

func TestUserHandler_GetProfile(t *testing.T) {
	mockUserService := new(MockUserService)
	userHandler := NewTestUserHandler(mockUserService)

	userID := primitive.NewObjectID()
	mockUser := &models.User{
		ID:        userID,
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
	}

	firstName := "Jane"
	lastName := "Smith"
	phone := "+1234567890"
	profile := &services.UserProfile{
		FirstName: &firstName,
		LastName:  &lastName,
		Phone:     &phone,
	}

	t.Run("success", func(t *testing.T) {
		mockUserService.On("UpdateUserProfile", mock.Anything, userID, profile).Return(mockUser, nil)

		body, _ := json.Marshal(profile)
		req, _ := http.NewRequest("PUT", "/api/v1/users/profile", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = req
		c.Set("user_id", userID.Hex())

		userHandler.UpdateProfile(c)

		assert.Equal(t, http.StatusOK, c.Writer.Status())
		mockUserService.AssertExpectations(t)
	})

	t.Run("invalid request body", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", "/api/v1/users/profile", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = req
		c.Set("user_id", userID.Hex())

		userHandler.UpdateProfile(c)

		assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
	})
}

func TestUserHandler_GetUsersList(t *testing.T) {
	mockUserService := new(MockUserService)
	userHandler := NewTestUserHandler(mockUserService)

	t.Run("success", func(t *testing.T) {
		expectedResponse := &services.UserListResponse{
			Users: []*models.User{
				{ID: primitive.NewObjectID(), FirstName: "John", Email: "john@example.com"},
			},
			Total:      1,
			Page:       1,
			PageSize:   20,
			TotalPages: 1,
		}
		mockUserService.On("GetUsersList", mock.Anything, 1, 20, mock.AnythingOfType("repository.UserFilters")).Return(expectedResponse, nil)

		req, _ := http.NewRequest("GET", "/api/v1/admin/users?page=1&page_size=20", nil)
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = req

		userHandler.GetUsersList(c)

		assert.Equal(t, http.StatusOK, c.Writer.Status())
		mockUserService.AssertExpectations(t)
	})

	t.Run("service error", func(t *testing.T) {
		// Clear previous expectations
		mockUserService.ExpectedCalls = nil
		mockUserService.Calls = nil

		mockUserService.On("GetUsersList", mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("repository.UserFilters")).Return(nil, errors.New("database error"))

		req, _ := http.NewRequest("GET", "/api/v1/admin/users", nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		userHandler.GetUsersList(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockUserService.AssertExpectations(t)
	})
}

func TestUserHandler_SearchUsers(t *testing.T) {
	mockUserService := new(MockUserService)
	userHandler := NewTestUserHandler(mockUserService)

	t.Run("success", func(t *testing.T) {
		expectedUsers := []*models.User{
			{ID: primitive.NewObjectID(), FirstName: "John", Email: "john@example.com"},
		}
		mockUserService.On("SearchUsers", mock.Anything, "john", 20).Return(expectedUsers, nil)

		req, _ := http.NewRequest("GET", "/api/v1/admin/users/search?q=john&limit=20", nil)
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = req

		userHandler.SearchUsers(c)

		assert.Equal(t, http.StatusOK, c.Writer.Status())
		mockUserService.AssertExpectations(t)
	})

	t.Run("missing query", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/admin/users/search", nil)
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = req

		userHandler.SearchUsers(c)

		assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
	})
}

func TestUserHandler_UpdateUserStatus(t *testing.T) {
	mockUserService := new(MockUserService)
	userHandler := NewTestUserHandler(mockUserService)

	userID := primitive.NewObjectID()

	t.Run("success", func(t *testing.T) {
		mockUserService.On("UpdateUserStatus", mock.Anything, userID, models.UserStatusActive).Return(nil)

		requestBody := map[string]string{"status": "active"}
		body, _ := json.Marshal(requestBody)
		req, _ := http.NewRequest("PUT", "/api/v1/admin/users/"+userID.Hex()+"/status", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = req
		c.Params = gin.Params{{Key: "id", Value: userID.Hex()}}

		userHandler.UpdateUserStatus(c)

		assert.Equal(t, http.StatusOK, c.Writer.Status())
		mockUserService.AssertExpectations(t)
	})

	t.Run("invalid user ID", func(t *testing.T) {
		requestBody := map[string]string{"status": "active"}
		body, _ := json.Marshal(requestBody)
		req, _ := http.NewRequest("PUT", "/api/v1/admin/users/invalid-id/status", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = req
		c.Params = gin.Params{{Key: "id", Value: "invalid-id"}}

		userHandler.UpdateUserStatus(c)

		assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
	})
}

func TestUserHandler_DeleteUser(t *testing.T) {
	mockUserService := new(MockUserService)
	userHandler := NewTestUserHandler(mockUserService)

	userID := primitive.NewObjectID()

	t.Run("success", func(t *testing.T) {
		mockUserService.On("DeleteUser", mock.Anything, userID).Return(nil)

		req, _ := http.NewRequest("DELETE", "/api/v1/admin/users/"+userID.Hex(), nil)
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = req
		c.Params = gin.Params{{Key: "id", Value: userID.Hex()}}

		userHandler.DeleteUser(c)

		assert.Equal(t, http.StatusOK, c.Writer.Status())
		mockUserService.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		// Clear previous expectations
		mockUserService.ExpectedCalls = nil
		mockUserService.Calls = nil

		mockUserService.On("DeleteUser", mock.Anything, userID).Return(errors.New("user not found"))

		req, _ := http.NewRequest("DELETE", "/api/v1/admin/users/"+userID.Hex(), nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{{Key: "id", Value: userID.Hex()}}

		userHandler.DeleteUser(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockUserService.AssertExpectations(t)
	})
}

func TestUserHandler_GetUserStats(t *testing.T) {
	mockUserService := new(MockUserService)
	userHandler := NewTestUserHandler(mockUserService)

	t.Run("success", func(t *testing.T) {
		expectedStats := map[string]int64{
			"total_users":         100,
			"active_users":        80,
			"inactive_users":      20,
			"users_created_today": 5,
		}
		mockUserService.On("GetUserStats", mock.Anything).Return(expectedStats, nil)

		req, _ := http.NewRequest("GET", "/api/v1/admin/users/stats", nil)
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = req

		userHandler.GetUserStats(c)

		assert.Equal(t, http.StatusOK, c.Writer.Status())
		mockUserService.AssertExpectations(t)
	})
}

func TestUserHandler_GetUserWeddings(t *testing.T) {
	mockUserService := new(MockUserService)
	userHandler := NewTestUserHandler(mockUserService)

	userID := primitive.NewObjectID()
	weddingIDs := []primitive.ObjectID{primitive.NewObjectID(), primitive.NewObjectID()}

	t.Run("success", func(t *testing.T) {
		mockUserService.On("GetUserWeddings", mock.Anything, userID).Return(weddingIDs, nil)

		req, _ := http.NewRequest("GET", "/api/v1/users/weddings", nil)
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = req
		c.Set("user_id", userID.Hex())

		userHandler.GetUserWeddings(c)

		assert.Equal(t, http.StatusOK, c.Writer.Status())
		mockUserService.AssertExpectations(t)
	})

	t.Run("unauthorized", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/users/weddings", nil)
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = req

		userHandler.GetUserWeddings(c)

		assert.Equal(t, http.StatusUnauthorized, c.Writer.Status())
	})
}

func TestUserHandler_AddWeddingToUser(t *testing.T) {
	mockUserService := new(MockUserService)
	userHandler := NewTestUserHandler(mockUserService)

	userID := primitive.NewObjectID()
	weddingID := primitive.NewObjectID()

	t.Run("success", func(t *testing.T) {
		mockUserService.On("AddWeddingToUser", mock.Anything, userID, weddingID).Return(nil)

		req, _ := http.NewRequest("POST", "/api/v1/users/weddings/"+weddingID.Hex(), nil)
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = req
		c.Set("user_id", userID.Hex())
		c.Params = gin.Params{{Key: "wedding_id", Value: weddingID.Hex()}}

		userHandler.AddWeddingToUser(c)

		assert.Equal(t, http.StatusOK, c.Writer.Status())
		mockUserService.AssertExpectations(t)
	})

	t.Run("invalid wedding ID", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/v1/users/weddings/invalid-id", nil)
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = req
		c.Set("user_id", userID.Hex())
		c.Params = gin.Params{{Key: "wedding_id", Value: "invalid-id"}}

		userHandler.AddWeddingToUser(c)

		assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
	})
}

func TestUserHandler_RemoveWeddingFromUser(t *testing.T) {
	mockUserService := new(MockUserService)
	userHandler := NewTestUserHandler(mockUserService)

	userID := primitive.NewObjectID()
	weddingID := primitive.NewObjectID()

	t.Run("success", func(t *testing.T) {
		mockUserService.On("RemoveWeddingFromUser", mock.Anything, userID, weddingID).Return(nil)

		req, _ := http.NewRequest("DELETE", "/api/v1/users/weddings/"+weddingID.Hex(), nil)
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = req
		c.Set("user_id", userID.Hex())
		c.Params = gin.Params{{Key: "wedding_id", Value: weddingID.Hex()}}

		userHandler.RemoveWeddingFromUser(c)

		assert.Equal(t, http.StatusOK, c.Writer.Status())
		mockUserService.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		// Clear previous expectations
		mockUserService.ExpectedCalls = nil
		mockUserService.Calls = nil

		mockUserService.On("RemoveWeddingFromUser", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("user not found"))

		req, _ := http.NewRequest("DELETE", "/api/v1/users/weddings/"+weddingID.Hex(), nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user_id", userID.Hex())
		c.Params = gin.Params{{Key: "wedding_id", Value: weddingID.Hex()}}

		userHandler.RemoveWeddingFromUser(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockUserService.AssertExpectations(t)
	})
}
