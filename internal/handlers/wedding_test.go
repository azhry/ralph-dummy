package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"wedding-invitation-backend/internal/domain/models"
	"wedding-invitation-backend/internal/domain/repository"
)

// MockWeddingService is a mock implementation of WeddingService
type MockWeddingService struct {
	mock.Mock
}

func (m *MockWeddingService) CreateWedding(ctx context.Context, wedding *models.Wedding, userID primitive.ObjectID) error {
	args := m.Called(ctx, wedding, userID)
	return args.Error(0)
}

func (m *MockWeddingService) GetWeddingByID(ctx context.Context, id primitive.ObjectID, requestingUserID primitive.ObjectID) (*models.Wedding, error) {
	args := m.Called(ctx, id, requestingUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Wedding), args.Error(1)
}

func (m *MockWeddingService) GetWeddingBySlug(ctx context.Context, slug string, requestingUserID primitive.ObjectID) (*models.Wedding, error) {
	args := m.Called(ctx, slug, requestingUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Wedding), args.Error(1)
}

func (m *MockWeddingService) GetUserWeddings(ctx context.Context, userID primitive.ObjectID, page, pageSize int, filters repository.WeddingFilters) ([]*models.Wedding, int64, error) {
	args := m.Called(ctx, userID, page, pageSize, filters)
	return args.Get(0).([]*models.Wedding), args.Get(1).(int64), args.Error(2)
}

func (m *MockWeddingService) UpdateWedding(ctx context.Context, wedding *models.Wedding, requestingUserID primitive.ObjectID) error {
	args := m.Called(ctx, wedding, requestingUserID)
	return args.Error(0)
}

func (m *MockWeddingService) DeleteWedding(ctx context.Context, weddingID primitive.ObjectID, requestingUserID primitive.ObjectID) error {
	args := m.Called(ctx, weddingID, requestingUserID)
	return args.Error(0)
}

func (m *MockWeddingService) PublishWedding(ctx context.Context, weddingID primitive.ObjectID, requestingUserID primitive.ObjectID) error {
	args := m.Called(ctx, weddingID, requestingUserID)
	return args.Error(0)
}

func (m *MockWeddingService) ListPublicWeddings(ctx context.Context, page, pageSize int, filters repository.PublicWeddingFilters) ([]*models.Wedding, int64, error) {
	args := m.Called(ctx, page, pageSize, filters)
	return args.Get(0).([]*models.Wedding), args.Get(1).(int64), args.Error(2)
}

func createTestWedding() *models.Wedding {
	return &models.Wedding{
		ID:     primitive.NewObjectID(),
		UserID: primitive.NewObjectID(),
		Title:  "Test Wedding",
		Slug:   "test-wedding",
		Couple: models.CoupleInfo{
			Partner1: struct {
				FirstName   string            `bson:"first_name" json:"first_name" validate:"required"`
				LastName    string            `bson:"last_name" json:"last_name" validate:"required"`
				FullName    string            `bson:"full_name" json:"full_name"`
				PhotoURL    string            `bson:"photo_url,omitempty" json:"photo_url,omitempty"`
				SocialLinks map[string]string `bson:"social_links,omitempty" json:"social_links,omitempty"`
			}{
				FirstName: "John",
				LastName:  "Doe",
			},
			Partner2: struct {
				FirstName   string            `bson:"first_name" json:"first_name" validate:"required"`
				LastName    string            `bson:"last_name" json:"last_name" validate:"required"`
				FullName    string            `bson:"full_name" json:"full_name"`
				PhotoURL    string            `bson:"photo_url,omitempty" json:"photo_url,omitempty"`
				SocialLinks map[string]string `bson:"social_links,omitempty" json:"social_links,omitempty"`
			}{
				FirstName: "Jane",
				LastName:  "Smith",
			},
		},
		Event: models.EventDetails{
			Title:        "Wedding Ceremony",
			Date:         time.Now().AddDate(0, 6, 0),
			VenueName:    "Test Venue",
			VenueAddress: "123 Test St",
		},
		Theme: models.ThemeSettings{
			ThemeID: "default",
		},
		RSVP: models.RSVPSettings{
			Enabled:      true,
			AllowPlusOne: true,
			MaxPlusOnes:  2,
		},
		Status: string(models.WeddingStatusDraft),
	}
}

func setupTestRouter(mockService *MockWeddingService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	handler := NewWeddingHandler(mockService)

	v1 := router.Group("/api/v1")
	{
		v1.POST("/weddings", handler.CreateWedding)
		v1.GET("/weddings", handler.GetUserWeddings)
		v1.GET("/weddings/:id", handler.GetWedding)
		v1.PUT("/weddings/:id", handler.UpdateWedding)
		v1.DELETE("/weddings/:id", handler.DeleteWedding)
		v1.POST("/weddings/:id/publish", handler.PublishWedding)
		v1.GET("/weddings/slug/:slug", handler.GetWeddingBySlug)
		v1.GET("/public/weddings", handler.ListPublicWeddings)
	}

	return router
}

func TestWeddingHandler_CreateWedding(t *testing.T) {
	mockService := new(MockWeddingService)

	userID := primitive.NewObjectID()
	wedding := createTestWedding()
	wedding.ID = primitive.NilObjectID // Will be set by service

	weddingJSON, _ := json.Marshal(wedding)
	req, _ := http.NewRequest("POST", "/api/v1/weddings", bytes.NewBuffer(weddingJSON))
	req.Header.Set("Content-Type", "application/json")

	// Set user context
	mockService.On("CreateWedding", mock.Anything, mock.AnythingOfType("*models.Wedding"), userID).Return(nil)

	w := httptest.NewRecorder()

	// Create a gin context with user ID
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", userID.Hex())

	// Call the handler directly
	handler := NewWeddingHandler(mockService)
	handler.CreateWedding(c)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Wedding
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, wedding.Title, response.Title)

	mockService.AssertExpectations(t)
}

func TestWeddingHandler_CreateWedding_InvalidJSON(t *testing.T) {
	mockService := new(MockWeddingService)
	router := setupTestRouter(mockService)

	req, _ := http.NewRequest("POST", "/api/v1/weddings", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response.Error, "Invalid request body")
}

func TestWeddingHandler_GetWedding(t *testing.T) {
	mockService := new(MockWeddingService)
	_ = setupTestRouter(mockService)

	wedding := createTestWedding()
	userID := primitive.NewObjectID()

	mockService.On("GetWeddingByID", mock.Anything, wedding.ID, userID).Return(wedding, nil)

	req, _ := http.NewRequest("GET", "/api/v1/weddings/"+wedding.ID.Hex(), nil)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	// Create a gin context with user ID
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", userID.Hex())
	c.Params = gin.Params{{Key: "id", Value: wedding.ID.Hex()}}

	// Call the handler directly
	handler := NewWeddingHandler(mockService)
	handler.GetWedding(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Wedding
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, wedding.ID, response.ID)

	mockService.AssertExpectations(t)
}

func TestWeddingHandler_GetWedding_InvalidID(t *testing.T) {
	mockService := new(MockWeddingService)
	router := setupTestRouter(mockService)

	req, _ := http.NewRequest("GET", "/api/v1/weddings/invalid-id", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response.Error, "Invalid wedding ID")
}

func TestWeddingHandler_GetWedding_NotFound(t *testing.T) {
	mockService := new(MockWeddingService)
	_ = setupTestRouter(mockService)

	weddingID := primitive.NewObjectID()
	userID := primitive.NewObjectID()

	mockService.On("GetWeddingByID", mock.Anything, weddingID, userID).Return(nil, errors.New("wedding not found"))

	req, _ := http.NewRequest("GET", "/api/v1/weddings/"+weddingID.Hex(), nil)

	w := httptest.NewRecorder()

	// Create a gin context with user ID
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", userID.Hex())
	c.Params = gin.Params{{Key: "id", Value: weddingID.Hex()}}

	// Call the handler directly
	handler := NewWeddingHandler(mockService)
	handler.GetWedding(c)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response.Error, "Wedding not found")

	mockService.AssertExpectations(t)
}

func TestWeddingHandler_GetWeddingBySlug(t *testing.T) {
	mockService := new(MockWeddingService)
	_ = setupTestRouter(mockService)

	wedding := createTestWedding()
	userID := primitive.NewObjectID()

	mockService.On("GetWeddingBySlug", mock.Anything, wedding.Slug, userID).Return(wedding, nil)

	req, _ := http.NewRequest("GET", "/api/v1/weddings/slug/"+wedding.Slug, nil)

	w := httptest.NewRecorder()

	// Create a gin context with user ID
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", userID.Hex())
	c.Params = gin.Params{{Key: "slug", Value: wedding.Slug}}

	// Call the handler directly
	handler := NewWeddingHandler(mockService)
	handler.GetWeddingBySlug(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Wedding
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, wedding.Slug, response.Slug)

	mockService.AssertExpectations(t)
}

func TestWeddingHandler_GetUserWeddings(t *testing.T) {
	mockService := new(MockWeddingService)
	_ = setupTestRouter(mockService)

	userID := primitive.NewObjectID()
	weddings := []*models.Wedding{createTestWedding()}
	filters := repository.WeddingFilters{}

	mockService.On("GetUserWeddings", mock.Anything, userID, 1, 20, filters).Return(weddings, int64(1), nil)

	req, _ := http.NewRequest("GET", "/api/v1/weddings?page=1&page_size=20", nil)

	w := httptest.NewRecorder()

	// Create a gin context with user ID
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", userID.Hex())
	c.Request = req

	// Call the handler directly
	handler := NewWeddingHandler(mockService)
	handler.GetUserWeddings(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response PaginatedWeddingsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, int64(1), response.Total)
	assert.Equal(t, 1, response.Page)
	assert.Equal(t, 20, response.PageSize)
	assert.Len(t, response.Weddings, 1)

	mockService.AssertExpectations(t)
}

func TestWeddingHandler_UpdateWedding(t *testing.T) {
	mockService := new(MockWeddingService)
	_ = setupTestRouter(mockService)

	userID := primitive.NewObjectID()
	wedding := createTestWedding()

	weddingJSON, _ := json.Marshal(wedding)
	req, _ := http.NewRequest("PUT", "/api/v1/weddings/"+wedding.ID.Hex(), bytes.NewBuffer(weddingJSON))
	req.Header.Set("Content-Type", "application/json")

	mockService.On("UpdateWedding", mock.Anything, mock.AnythingOfType("*models.Wedding"), userID).Return(nil)

	w := httptest.NewRecorder()

	// Create a gin context with user ID
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", userID.Hex())
	c.Params = gin.Params{{Key: "id", Value: wedding.ID.Hex()}}

	// Call the handler directly
	handler := NewWeddingHandler(mockService)
	handler.UpdateWedding(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Wedding
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, wedding.Title, response.Title)

	mockService.AssertExpectations(t)
}

func TestWeddingHandler_DeleteWedding(t *testing.T) {
	mockService := new(MockWeddingService)
	_ = setupTestRouter(mockService)

	userID := primitive.NewObjectID()
	weddingID := primitive.NewObjectID()

	mockService.On("DeleteWedding", mock.Anything, weddingID, userID).Return(nil)

	req, _ := http.NewRequest("DELETE", "/api/v1/weddings/"+weddingID.Hex(), nil)

	w := httptest.NewRecorder()

	// Create a gin context with user ID
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", userID.Hex())
	c.Params = gin.Params{{Key: "id", Value: weddingID.Hex()}}

	// Call the handler directly
	handler := NewWeddingHandler(mockService)
	handler.DeleteWedding(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response SuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response.Message, "deleted successfully")

	mockService.AssertExpectations(t)
}

func TestWeddingHandler_PublishWedding(t *testing.T) {
	mockService := new(MockWeddingService)
	_ = setupTestRouter(mockService)

	userID := primitive.NewObjectID()
	wedding := createTestWedding()

	// First call for permission check
	mockService.On("GetWeddingByID", mock.Anything, wedding.ID, userID).Return(wedding, nil)
	// Second call for publish
	mockService.On("PublishWedding", mock.Anything, wedding.ID, userID).Return(nil)
	// Third call to get updated wedding
	mockService.On("GetWeddingByID", mock.Anything, wedding.ID, userID).Return(wedding, nil)

	req, _ := http.NewRequest("POST", "/api/v1/weddings/"+wedding.ID.Hex()+"/publish", nil)

	w := httptest.NewRecorder()

	// Create a gin context with user ID
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", userID.Hex())
	c.Params = gin.Params{{Key: "id", Value: wedding.ID.Hex()}}

	// Call the handler directly
	handler := NewWeddingHandler(mockService)
	handler.PublishWedding(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Wedding
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, wedding.ID, response.ID)

	mockService.AssertExpectations(t)
}

func TestWeddingHandler_ListPublicWeddings(t *testing.T) {
	mockService := new(MockWeddingService)
	router := setupTestRouter(mockService)

	weddings := []*models.Wedding{createTestWedding()}
	filters := repository.PublicWeddingFilters{}

	mockService.On("ListPublicWeddings", mock.Anything, 1, 20, filters).Return(weddings, int64(1), nil)

	req, _ := http.NewRequest("GET", "/api/v1/public/weddings?page=1&page_size=20", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response PaginatedWeddingsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, int64(1), response.Total)
	assert.Equal(t, 1, response.Page)
	assert.Equal(t, 20, response.PageSize)
	assert.Len(t, response.Weddings, 1)

	mockService.AssertExpectations(t)
}

func TestWeddingHandler_UpdateWedding_AccessDenied(t *testing.T) {
	mockService := new(MockWeddingService)
	_ = setupTestRouter(mockService)

	userID := primitive.NewObjectID()
	wedding := createTestWedding()

	weddingJSON, _ := json.Marshal(wedding)
	req, _ := http.NewRequest("PUT", "/api/v1/weddings/"+wedding.ID.Hex(), bytes.NewBuffer(weddingJSON))
	req.Header.Set("Content-Type", "application/json")

	mockService.On("UpdateWedding", mock.Anything, mock.AnythingOfType("*models.Wedding"), userID).Return(errors.New("access denied"))

	w := httptest.NewRecorder()

	// Create a gin context with user ID
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", userID.Hex())
	c.Params = gin.Params{{Key: "id", Value: wedding.ID.Hex()}}

	// Call the handler directly
	handler := NewWeddingHandler(mockService)
	handler.UpdateWedding(c)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response.Error, "Access denied")

	mockService.AssertExpectations(t)
}

func TestWeddingHandler_PublishWedding_OwnerCheck(t *testing.T) {
	mockService := new(MockWeddingService)
	_ = setupTestRouter(mockService)

	userID := primitive.NewObjectID()
	otherUserID := primitive.NewObjectID()
	wedding := createTestWedding()
	wedding.UserID = otherUserID // Different from requesting user

	mockService.On("GetWeddingByID", mock.Anything, wedding.ID, userID).Return(wedding, nil)

	req, _ := http.NewRequest("POST", "/api/v1/weddings/"+wedding.ID.Hex()+"/publish", nil)

	w := httptest.NewRecorder()

	// Create a gin context with user ID
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", userID.Hex())
	c.Params = gin.Params{{Key: "id", Value: wedding.ID.Hex()}}

	// Call the handler directly
	handler := NewWeddingHandler(mockService)
	handler.PublishWedding(c)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response.Error, "Access denied")

	mockService.AssertExpectations(t)
}
