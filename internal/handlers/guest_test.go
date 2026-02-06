package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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

// MockGuestService for testing
type MockGuestService struct {
	guests          map[primitive.ObjectID]*models.Guest
	createError     error
	getError        error
	updateError     error
	deleteError     error
	listError       error
	bulkCreateError error
	importError     error
}

func NewMockGuestService() *MockGuestService {
	return &MockGuestService{
		guests: make(map[primitive.ObjectID]*models.Guest),
	}
}

func (m *MockGuestService) CreateGuest(ctx context.Context, weddingID, userID primitive.ObjectID, guest *models.Guest) error {
	if m.createError != nil {
		return m.createError
	}

	id := primitive.NewObjectID()
	guest.ID = id
	guest.WeddingID = weddingID
	guest.CreatedBy = userID
	guest.CreatedAt = time.Now()
	guest.UpdatedAt = time.Now()
	m.guests[id] = guest
	return nil
}

func (m *MockGuestService) GetGuestByID(ctx context.Context, guestID, userID primitive.ObjectID) (*models.Guest, error) {
	if m.getError != nil {
		return nil, m.getError
	}

	guest, exists := m.guests[guestID]
	if !exists {
		return nil, services.ErrGuestNotFound
	}

	// Check ownership (simplified for test)
	if guest.CreatedBy != userID {
		return nil, services.ErrUnauthorized
	}

	return guest, nil
}

func (m *MockGuestService) ListGuests(ctx context.Context, weddingID, userID primitive.ObjectID, page, pageSize int, filters repository.GuestFilters) ([]*models.Guest, int64, error) {
	if m.listError != nil {
		return nil, 0, m.listError
	}

	var guests []*models.Guest
	for _, guest := range m.guests {
		if guest.WeddingID == weddingID && guest.CreatedBy == userID {
			// Apply simple filtering (for test purposes)
			if filters.Search == "" || guest.FirstName == filters.Search || guest.LastName == filters.Search {
				guests = append(guests, guest)
			}
		}
	}

	return guests, int64(len(guests)), nil
}

func (m *MockGuestService) UpdateGuest(ctx context.Context, guestID, userID primitive.ObjectID, guest *models.Guest) error {
	if m.updateError != nil {
		return m.updateError
	}

	existing, exists := m.guests[guestID]
	if !exists {
		return services.ErrGuestNotFound
	}

	if existing.CreatedBy != userID {
		return services.ErrUnauthorized
	}

	guest.ID = guestID
	guest.WeddingID = existing.WeddingID
	guest.CreatedBy = existing.CreatedBy
	guest.CreatedAt = existing.CreatedAt
	guest.UpdatedAt = time.Now()
	m.guests[guestID] = guest
	return nil
}

func (m *MockGuestService) DeleteGuest(ctx context.Context, guestID, userID primitive.ObjectID) error {
	if m.deleteError != nil {
		return m.deleteError
	}

	guest, exists := m.guests[guestID]
	if !exists {
		return services.ErrGuestNotFound
	}

	if guest.CreatedBy != userID {
		return services.ErrUnauthorized
	}

	delete(m.guests, guestID)
	return nil
}

func (m *MockGuestService) CreateManyGuests(ctx context.Context, weddingID, userID primitive.ObjectID, guests []*models.Guest) error {
	if m.bulkCreateError != nil {
		return m.bulkCreateError
	}

	for _, guest := range guests {
		id := primitive.NewObjectID()
		guest.ID = id
		guest.WeddingID = weddingID
		guest.CreatedBy = userID
		guest.CreatedAt = time.Now()
		guest.UpdatedAt = time.Now()
		m.guests[id] = guest
	}

	return nil
}

func (m *MockGuestService) ImportGuestsFromCSV(ctx context.Context, weddingID, userID primitive.ObjectID, csvData io.Reader) (*models.GuestImportResult, error) {
	if m.importError != nil {
		return nil, m.importError
	}

	return &models.GuestImportResult{
		SuccessCount: 2,
		ErrorCount:   0,
		Errors:       []string{},
		BatchID:      "test_batch_123",
	}, nil
}

func (m *MockGuestService) GetImportBatch(ctx context.Context, weddingID, userID primitive.ObjectID, batchID string) ([]*models.Guest, error) {
	var guests []*models.Guest
	for _, guest := range m.guests {
		if guest.ImportBatchID == batchID && guest.CreatedBy == userID {
			guests = append(guests, guest)
		}
	}
	return guests, nil
}

func setupGuestTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestGuestHandler_CreateGuest(t *testing.T) {
	mockService := NewMockGuestService()
	handler := NewGuestHandler(mockService)
	router := setupGuestTestRouter()

	weddingID := primitive.NewObjectID()
	userID := primitive.NewObjectID()

	// Mock user context
	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})

	router.POST("/weddings/:wedding_id/guests", handler.CreateGuest)

	// Test data
	req := CreateGuestRequest{
		FirstName:    "John",
		LastName:     "Doe",
		Email:        "john@example.com",
		Phone:        "+1234567890",
		Relationship: "Friend",
		Side:         "groom",
		AllowPlusOne: true,
		MaxPlusOnes:  1,
		VIP:          false,
	}

	reqBody, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	reqHTTP, _ := http.NewRequest("POST", fmt.Sprintf("/weddings/%s/guests", weddingID.Hex()), bytes.NewBuffer(reqBody))
	reqHTTP.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, reqHTTP)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Guest created successfully", response["message"])
	assert.NotNil(t, response["data"])
}

func TestGuestHandler_CreateGuest_ValidationError(t *testing.T) {
	mockService := NewMockGuestService()
	handler := NewGuestHandler(mockService)
	router := setupGuestTestRouter()

	weddingID := primitive.NewObjectID()
	userID := primitive.NewObjectID()

	// Mock user context
	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})

	router.POST("/weddings/:wedding_id/guests", handler.CreateGuest)

	// Test data with validation error (missing first_name)
	req := CreateGuestRequest{
		LastName: "Doe",
	}

	reqBody, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	reqHTTP, _ := http.NewRequest("POST", fmt.Sprintf("/weddings/%s/guests", weddingID.Hex()), bytes.NewBuffer(reqBody))
	reqHTTP.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, reqHTTP)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "required")
}

func TestGuestHandler_GetGuest(t *testing.T) {
	mockService := NewMockGuestService()
	handler := NewGuestHandler(mockService)
	router := setupGuestTestRouter()

	weddingID := primitive.NewObjectID()
	userID := primitive.NewObjectID()

	// Create test guest
	guest := &models.Guest{
		WeddingID: weddingID,
		FirstName: "John",
		LastName:  "Doe",
		CreatedBy: userID,
	}
	mockService.CreateGuest(context.Background(), weddingID, userID, guest)

	// Mock user context
	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})

	router.GET("/guests/:id", handler.GetGuest)

	w := httptest.NewRecorder()
	reqHTTP, _ := http.NewRequest("GET", fmt.Sprintf("/guests/%s", guest.ID.Hex()), nil)

	router.ServeHTTP(w, reqHTTP)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.NotNil(t, response["data"])

	data := response["data"].(map[string]interface{})
	assert.Equal(t, "John", data["first_name"])
	assert.Equal(t, "Doe", data["last_name"])
}

func TestGuestHandler_GetGuest_NotFound(t *testing.T) {
	mockService := NewMockGuestService()
	handler := NewGuestHandler(mockService)
	router := setupGuestTestRouter()

	userID := primitive.NewObjectID()

	// Mock user context
	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})

	router.GET("/guests/:id", handler.GetGuest)

	w := httptest.NewRecorder()
	reqHTTP, _ := http.NewRequest("GET", "/guests/"+primitive.NewObjectID().Hex(), nil)

	router.ServeHTTP(w, reqHTTP)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "Guest not found")
}

func TestGuestHandler_ListGuests(t *testing.T) {
	mockService := NewMockGuestService()
	handler := NewGuestHandler(mockService)
	router := setupGuestTestRouter()

	weddingID := primitive.NewObjectID()
	userID := primitive.NewObjectID()

	// Create test guests
	guests := []*models.Guest{
		{FirstName: "John", LastName: "Doe", WeddingID: weddingID, CreatedBy: userID},
		{FirstName: "Jane", LastName: "Smith", WeddingID: weddingID, CreatedBy: userID},
	}

	for _, guest := range guests {
		mockService.CreateGuest(context.Background(), weddingID, userID, guest)
	}

	// Mock user context
	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})

	router.GET("/weddings/:wedding_id/guests", handler.ListGuests)

	w := httptest.NewRecorder()
	reqHTTP, _ := http.NewRequest("GET", fmt.Sprintf("/weddings/%s/guests?page=1&page_size=10", weddingID.Hex()), nil)

	router.ServeHTTP(w, reqHTTP)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.NotNil(t, response["data"])

	data := response["data"].(map[string]interface{})
	assert.Equal(t, float64(2), data["total"])
	assert.Equal(t, float64(1), data["page"])
	assert.Equal(t, float64(10), data["page_size"])
}

func TestGuestHandler_UpdateGuest(t *testing.T) {
	mockService := NewMockGuestService()
	handler := NewGuestHandler(mockService)
	router := setupGuestTestRouter()

	weddingID := primitive.NewObjectID()
	userID := primitive.NewObjectID()

	// Create test guest
	guest := &models.Guest{
		WeddingID: weddingID,
		FirstName: "John",
		LastName:  "Doe",
		CreatedBy: userID,
	}
	mockService.CreateGuest(context.Background(), weddingID, userID, guest)

	// Mock user context
	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})

	router.PUT("/guests/:id", handler.UpdateGuest)

	// Update request
	updated := "Updated"
	name := "Name"
	req := UpdateGuestRequest{
		FirstName: &updated,
		LastName:  &name,
	}

	reqBody, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	reqHTTP, _ := http.NewRequest("PUT", fmt.Sprintf("/guests/%s", guest.ID.Hex()), bytes.NewBuffer(reqBody))
	reqHTTP.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, reqHTTP)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Guest updated successfully", response["message"])
}

func TestGuestHandler_DeleteGuest(t *testing.T) {
	mockService := NewMockGuestService()
	handler := NewGuestHandler(mockService)
	router := setupGuestTestRouter()

	weddingID := primitive.NewObjectID()
	userID := primitive.NewObjectID()

	// Create test guest
	guest := &models.Guest{
		WeddingID: weddingID,
		FirstName: "John",
		LastName:  "Doe",
		CreatedBy: userID,
	}
	mockService.CreateGuest(context.Background(), weddingID, userID, guest)

	// Mock user context
	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})

	router.DELETE("/guests/:id", handler.DeleteGuest)

	w := httptest.NewRecorder()
	reqHTTP, _ := http.NewRequest("DELETE", fmt.Sprintf("/guests/%s", guest.ID.Hex()), nil)

	router.ServeHTTP(w, reqHTTP)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Guest deleted successfully", response["message"])
}

func TestGuestHandler_BulkCreateGuests(t *testing.T) {
	mockService := NewMockGuestService()
	handler := NewGuestHandler(mockService)
	router := setupGuestTestRouter()

	weddingID := primitive.NewObjectID()
	userID := primitive.NewObjectID()

	// Mock user context
	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})

	router.POST("/weddings/:wedding_id/guests/bulk", handler.BulkCreateGuests)

	// Test data
	req := BulkCreateGuestsRequest{
		Guests: []CreateGuestRequest{
			{FirstName: "Guest1", LastName: "Test1"},
			{FirstName: "Guest2", LastName: "Test2"},
		},
	}

	reqBody, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	reqHTTP, _ := http.NewRequest("POST", fmt.Sprintf("/weddings/%s/guests/bulk", weddingID.Hex()), bytes.NewBuffer(reqBody))
	reqHTTP.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, reqHTTP)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Guests created successfully", response["message"])

	data := response["data"].(map[string]interface{})
	assert.Equal(t, float64(2), data["created_count"])
}
