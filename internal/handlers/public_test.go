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
	"go.mongodb.org/mongo-driver/bson/primitive"

	"wedding-invitation-backend/internal/domain/models"
)

// MockWeddingServiceForPublic for testing public handler
type MockWeddingServiceForPublic struct {
	mock.Mock
}

func (m *MockWeddingServiceForPublic) GetWeddingBySlugForPublic(ctx context.Context, slug string) (*models.Wedding, error) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Wedding), args.Error(1)
}

// MockRSVPServiceForPublic for testing public handler
type MockRSVPServiceForPublic struct {
	mock.Mock
}

func (m *MockRSVPServiceForPublic) SubmitRSVP(ctx context.Context, rsvp *models.RSVP) error {
	args := m.Called(ctx, rsvp)
	return args.Error(0)
}

func setupPublicTestRouter(publicHandler *PublicHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	v1 := router.Group("/api/v1")
	public := v1.Group("/public")
	{
		public.GET("/weddings/:slug", publicHandler.GetWeddingBySlug)
		public.POST("/weddings/:slug/rsvp", publicHandler.SubmitRSVP)
	}
	
	return router
}

func TestPublicHandler_GetWeddingBySlug_Success(t *testing.T) {
	// Arrange
	mockWeddingService := new(MockWeddingServiceForPublic)
	mockRSVPService := new(MockRSVPServiceForPublic)
	publicHandler := NewPublicHandler(mockWeddingService, mockRSVPService)
	
	router := setupPublicTestRouter(publicHandler)
	
	wedding := &models.Wedding{
		ID:     primitive.NewObjectID(),
		Slug:   "john-jane-wedding",
		Status: string(models.WeddingStatusPublished),
		Couple: models.CoupleInfo{
			GroomName: "John",
			BrideName: "Jane",
		},
		EventDetails: models.EventDetails{
			WeddingDate: time.Now().AddDate(0, 6, 0),
			VenueName:   "Garden Pavilion",
		},
		Theme: models.Theme{
			ThemeID: "dark-romance",
		},
		RSVP: models.RSVPSettings{
			AllowPlusOne:   true,
			CollectDietary: true,
			Deadline:       time.Now().AddDate(0, 3, 0),
		},
		SEO: models.SEO{
			SiteTitle:       "John & Jane Wedding",
			MetaDescription: "Join us for our special day",
		},
	}
	
	mockWeddingService.On("GetWeddingBySlugForPublic", mock.Anything, "john-jane-wedding").Return(wedding, nil)
	
	// Act
	req, _ := http.NewRequest("GET", "/api/v1/public/weddings/john-jane-wedding", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response PublicWeddingResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "john-jane-wedding", response.Slug)
	assert.Equal(t, "dark-romance", response.Theme)
	assert.Equal(t, "John", response.GroomName)
	assert.Equal(t, "Jane", response.BrideName)
	
	mockWeddingService.AssertExpectations(t)
}

func TestPublicHandler_GetWeddingBySlug_NotFound(t *testing.T) {
	// Arrange
	mockWeddingService := new(MockWeddingServiceForPublic)
	mockRSVPService := new(MockRSVPServiceForPublic)
	publicHandler := NewPublicHandler(mockWeddingService, mockRSVPService)
	
	router := setupPublicTestRouter(publicHandler)
	
	mockWeddingService.On("GetWeddingBySlugForPublic", mock.Anything, "nonexistent").Return(nil, errors.New("wedding not found"))
	
	// Act
	req, _ := http.NewRequest("GET", "/api/v1/public/weddings/nonexistent", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	
	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response.Error, "not found")
	
	mockWeddingService.AssertExpectations(t)
}

func TestPublicHandler_SubmitRSVP_Success(t *testing.T) {
	// Arrange
	mockWeddingService := new(MockWeddingServiceForPublic)
	mockRSVPService := new(MockRSVPServiceForPublic)
	publicHandler := NewPublicHandler(mockWeddingService, mockRSVPService)
	
	router := setupPublicTestRouter(publicHandler)
	
	weddingID := primitive.NewObjectID()
	wedding := &models.Wedding{
		ID:     weddingID,
		Slug:   "john-jane-wedding",
		Status: string(models.WeddingStatusPublished),
		RSVP: models.RSVPSettings{
			AllowPlusOne:   true,
			CollectDietary: true,
			Deadline:       time.Now().AddDate(0, 3, 0),
		},
	}
	
	mockWeddingService.On("GetWeddingBySlugForPublic", mock.Anything, "john-jane-wedding").Return(wedding, nil)
	mockRSVPService.On("SubmitRSVP", mock.Anything, mock.AnythingOfType("*models.RSVP")).Return(nil)
	
	requestBody := PublicRSVPRequest{
		Name:           "Alice Smith",
		Email:          "alice@example.com",
		Attending:      true,
		NumberOfGuests: 2,
		PlusOneName:    "Bob Smith",
	}
	
	body, _ := json.Marshal(requestBody)
	
	// Act
	req, _ := http.NewRequest("POST", "/api/v1/public/weddings/john-jane-wedding/rsvp", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)
	
	var response PublicRSVPResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Alice Smith", response.Name)
	assert.Equal(t, "alice@example.com", response.Email)
	assert.Equal(t, true, response.Attending)
	assert.Equal(t, 2, response.NumberOfGuests)
	
	mockWeddingService.AssertExpectations(t)
	mockRSVPService.AssertExpectations(t)
}

func TestPublicHandler_SubmitRSVP_InvalidJSON(t *testing.T) {
	// Arrange
	mockWeddingService := new(MockWeddingServiceForPublic)
	mockRSVPService := new(MockRSVPServiceForPublic)
	publicHandler := NewPublicHandler(mockWeddingService, mockRSVPService)
	
	router := setupPublicTestRouter(publicHandler)
	
	// Act
	req, _ := http.NewRequest("POST", "/api/v1/public/weddings/john-jane-wedding/rsvp", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response.Error, "Invalid request")
}

func TestPublicHandler_convertToPublicResponse(t *testing.T) {
	// Arrange
	mockWeddingService := new(MockWeddingServiceForPublic)
	mockRSVPService := new(MockRSVPServiceForPublic)
	publicHandler := NewPublicHandler(mockWeddingService, mockRSVPService)
	
	wedding := &models.Wedding{
		Slug:   "test-wedding",
		Status: string(models.WeddingStatusPublished),
		Couple: models.CoupleInfo{
			GroomName: "John",
			BrideName: "Jane",
		},
		Theme: models.Theme{
			ThemeID: "romantic",
		},
		RSVP: models.RSVPSettings{
			AllowPlusOne:   true,
			CollectDietary: true,
			Deadline:       time.Now().AddDate(0, 3, 0),
		},
		SEO: models.SEO{
			SiteTitle:       "Test Wedding",
			MetaDescription: "Test description",
		},
	}
	
	// Act
	response := publicHandler.convertToPublicResponse(wedding)
	
	// Assert
	assert.Equal(t, "test-wedding", response.Slug)
	assert.Equal(t, "romantic", response.Theme)
	assert.Equal(t, "John", response.GroomName)
	assert.Equal(t, "Jane", response.BrideName)
	assert.Equal(t, "Test Wedding", response.SiteTitle)
	assert.Equal(t, "Test description", response.MetaDescription)
	assert.True(t, response.AllowPlusOne)
	assert.True(t, response.CollectDietary)
}