package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"wedding-invitation-backend/internal/domain/models"
	"wedding-invitation-backend/internal/domain/repository"
	"wedding-invitation-backend/internal/services"
)

// MockRSVPService for handler testing
type MockRSVPService struct {
	rsvps     map[primitive.ObjectID]*models.RSVP
	createErr error
	getErr    error
}

func NewMockRSVPService() *MockRSVPService {
	return &MockRSVPService{
		rsvps: make(map[primitive.ObjectID]*models.RSVP),
	}
}

func (m *MockRSVPService) SubmitRSVP(ctx context.Context, weddingID primitive.ObjectID, req services.SubmitRSVPRequest) (*models.RSVP, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}

	rsvp := &models.RSVP{
		ID:              primitive.NewObjectID(),
		WeddingID:       weddingID,
		FirstName:       req.FirstName,
		LastName:        req.LastName,
		Email:           req.Email,
		Status:          req.Status,
		AttendanceCount: req.AttendanceCount,
		PlusOnes:        req.PlusOnes,
		PlusOneCount:    len(req.PlusOnes),
		Source:          req.Source,
	}
	m.rsvps[rsvp.ID] = rsvp
	return rsvp, nil
}

func (m *MockRSVPService) GetRSVPByID(ctx context.Context, id primitive.ObjectID) (*models.RSVP, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if rsvp, exists := m.rsvps[id]; exists {
		return rsvp, nil
	}
	return nil, services.ErrRSVPNotFound
}

func (m *MockRSVPService) UpdateRSVP(ctx context.Context, id primitive.ObjectID, req services.UpdateRSVPRequest) (*models.RSVP, error) {
	rsvp, exists := m.rsvps[id]
	if !exists {
		return nil, services.ErrRSVPNotFound
	}

	// Update fields if provided
	if req.Status != nil {
		rsvp.Status = *req.Status
	}
	if req.AttendanceCount != nil {
		rsvp.AttendanceCount = *req.AttendanceCount
	}

	return rsvp, nil
}

func (m *MockRSVPService) DeleteRSVP(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) error {
	if _, exists := m.rsvps[id]; !exists {
		return services.ErrRSVPNotFound
	}
	delete(m.rsvps, id)
	return nil
}

func (m *MockRSVPService) ListRSVPs(ctx context.Context, weddingID primitive.ObjectID, userID primitive.ObjectID, page, pageSize int, filters repository.RSVPFilters) ([]*models.RSVP, int64, error) {
	var results []*models.RSVP
	for _, rsvp := range m.rsvps {
		if rsvp.WeddingID == weddingID {
			results = append(results, rsvp)
		}
	}
	return results, int64(len(results)), nil
}

func (m *MockRSVPService) GetRSVPStatistics(ctx context.Context, weddingID primitive.ObjectID, userID primitive.ObjectID) (*models.RSVPStatistics, error) {
	stats := &models.RSVPStatistics{
		TotalResponses:    len(m.rsvps),
		DietaryCounts:     make(map[string]int),
		SubmissionTrend:    []models.DailyCount{},
	}
	return stats, nil
}

func (m *MockRSVPService) ExportRSVPs(ctx context.Context, weddingID primitive.ObjectID, userID primitive.ObjectID) ([]*models.RSVP, error) {
	var results []*models.RSVP
	for _, rsvp := range m.rsvps {
		if rsvp.WeddingID == weddingID {
			results = append(results, rsvp)
		}
	}
	return results, nil
}

func setupRSVPRouter() (*gin.Engine, *MockRSVPService) {
	gin.SetMode(gin.TestMode)
	mockService := NewMockRSVPService()
	handler := NewRSVPHandler(mockService)

	router := gin.New()

	// Mock auth middleware
	router.Use(func(c *gin.Context) {
		c.Set("user_id", primitive.NewObjectID().Hex())
		c.Next()
	})

	// Setup routes
	v1 := router.Group("/api/v1")
	{
		// Public routes
		public := v1.Group("/public")
		{
			public.POST("/weddings/:id/rsvp", handler.SubmitRSVP)
		}

		// Protected routes
		v1.GET("/weddings/:id/rsvps", handler.GetRSVPs)
		v1.GET("/weddings/:id/rsvps/statistics", handler.GetRSVPStatistics)
		v1.GET("/weddings/:id/rsvps/export", handler.ExportRSVPs)
		v1.PUT("/rsvps/:id", handler.UpdateRSVP)
		v1.DELETE("/rsvps/:id", handler.DeleteRSVP)
	}

	return router, mockService
}

func TestRSVPHandler_SubmitRSVP(t *testing.T) {
	router, mockService := setupRSVPRouter()

	weddingID := primitive.NewObjectID()
	reqBody := services.SubmitRSVPRequest{
		FirstName:       "John",
		LastName:        "Doe",
		Email:           "john.doe@example.com",
		Status:          "attending",
		AttendanceCount: 2,
		PlusOnes: []models.PlusOneInfo{
			{FirstName: "Jane", LastName: "Doe"},
		},
		Source: "web",
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/public/weddings/"+weddingID.Hex()+"/rsvp", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "test-agent")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	data, exists := response["data"]
	assert.True(t, exists)
	assert.NotNil(t, data)
}

func TestRSVPHandler_SubmitRSVP_InvalidID(t *testing.T) {
	router, _ := setupRSVPRouter()

	reqBody := services.SubmitRSVPRequest{
		FirstName:       "John",
		LastName:        "Doe",
		Status:          "attending",
		AttendanceCount: 1,
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/public/weddings/invalid-id/rsvp", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	errorMsg, exists := response["error"]
	assert.True(t, exists)
	assert.Contains(t, errorMsg.(string), "Invalid wedding ID")
}

func TestRSVPHandler_SubmitRSVP_InvalidBody(t *testing.T) {
	router, _ := setupRSVPRouter()

	weddingID := primitive.NewObjectID()
	reqBody := `{"invalid": "json"}`

	req, _ := http.NewRequest("POST", "/api/v1/public/weddings/"+weddingID.Hex()+"/rsvp", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	errorMsg, exists := response["error"]
	assert.True(t, exists)
	assert.Contains(t, errorMsg.(string), "Invalid request body")
}

func TestRSVPHandler_SubmitRSVP_ServiceError(t *testing.T) {
	router, mockService := setupRSVPRouter()

	// Set service to return error
	mockService.createErr = services.ErrRSVPClosed

	weddingID := primitive.NewObjectID()
	reqBody := services.SubmitRSVPRequest{
		FirstName:       "John",
		LastName:        "Doe",
		Status:          "attending",
		AttendanceCount: 1,
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/public/weddings/"+weddingID.Hex()+"/rsvp", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	errorMsg, exists := response["error"]
	assert.True(t, exists)
	assert.Contains(t, errorMsg.(string), "RSVP is not open")
}

func TestRSVPHandler_GetRSVPs(t *testing.T) {
	router, mockService := setupRSVPRouter()

	weddingID := primitive.NewObjectID()

	// Create test RSVP
	rsvp := &models.RSVP{
		ID:              primitive.NewObjectID(),
		WeddingID:       weddingID,
		FirstName:       "John",
		LastName:        "Doe",
		Status:          "attending",
		AttendanceCount: 1,
	}
	mockService.rsvps[rsvp.ID] = rsvp

	req, _ := http.NewRequest("GET", "/api/v1/weddings/"+weddingID.Hex()+"/rsvps", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	data, exists := response["data"]
	assert.True(t, exists)
	assert.NotNil(t, data)

	dataArray := data.([]interface{})
	assert.Len(t, dataArray, 1)
}

func TestRSVPHandler_GetRSVPs_InvalidID(t *testing.T) {
	router, _ := setupRSVPRouter()

	req, _ := http.NewRequest("GET", "/api/v1/weddings/invalid-id/rsvps", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRSVPHandler_GetRSVPStatistics(t *testing.T) {
	router, mockService := setupRSVPRouter()

	weddingID := primitive.NewObjectID()

	req, _ := http.NewRequest("GET", "/api/v1/weddings/"+weddingID.Hex()+"/rsvps/statistics", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	data, exists := response["data"]
	assert.True(t, exists)
	assert.NotNil(t, data)
}

func TestRSVPHandler_UpdateRSVP(t *testing.T) {
	router, mockService := setupRSVPRouter()

	rsvpID := primitive.NewObjectID()

	// Create test RSVP
	rsvp := &models.RSVP{
		ID:              rsvpID,
		WeddingID:       primitive.NewObjectID(),
		FirstName:       "John",
		LastName:        "Doe",
		Status:          "attending",
		AttendanceCount: 1,
		SubmittedAt:     time.Now().Add(-1 * time.Hour),
	}
	mockService.rsvps[rsvpID] = rsvp

	newStatus := "not-attending"
	reqBody := services.UpdateRSVPRequest{
		Status:          &newStatus,
		AttendanceCount: intPtr(2),
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PUT", "/api/v1/rsvps/"+rsvpID.Hex(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	data, exists := response["data"]
	assert.True(t, exists)
	assert.NotNil(t, data)
}

func TestRSVPHandler_DeleteRSVP(t *testing.T) {
	router, mockService := setupRSVPRouter()

	rsvpID := primitive.NewObjectID()

	// Create test RSVP
	rsvp := &models.RSVP{
		ID:        rsvpID,
		WeddingID: primitive.NewObjectID(),
		FirstName: "John",
		LastName:  "Doe",
		Status:    "attending",
	}
	mockService.rsvps[rsvpID] = rsvp

	req, _ := http.NewRequest("DELETE", "/api/v1/rsvps/"+rsvpID.Hex(), nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	message, exists := response["message"]
	assert.True(t, exists)
	assert.Equal(t, "RSVP deleted successfully", message)
}

func TestRSVPHandler_ExportRSVPs(t *testing.T) {
	router, mockService := setupRSVPRouter()

	weddingID := primitive.NewObjectID()

	// Create test RSVP
	rsvp := &models.RSVP{
		ID:        primitive.NewObjectID(),
		WeddingID: weddingID,
		FirstName: "John",
		LastName:  "Doe",
		Status:    "attending",
	}
	mockService.rsvps[rsvp.ID] = rsvp

	req, _ := http.NewRequest("GET", "/api/v1/weddings/"+weddingID.Hex()+"/rsvps/export", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	data, exists := response["data"]
	assert.True(t, exists)
	assert.NotNil(t, data)

	dataArray := data.([]interface{})
	assert.Len(t, dataArray, 1)
}

// Helper functions
func intPtr(i int) *int {
	return &i
}