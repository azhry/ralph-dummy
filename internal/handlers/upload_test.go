package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap/zaptest"
	"wedding-invitation-backend/internal/domain/models"
)

// MockMediaService is a mock implementation of MediaService
type MockMediaService struct {
	mock.Mock
}

func (m *MockMediaService) UploadFile(ctx context.Context, file io.Reader, header *multipart.FileHeader, userID primitive.ObjectID) (*models.Media, error) {
	args := m.Called(ctx, mock.AnythingOfType("io.Reader"), mock.AnythingOfType("*multipart.FileHeader"), userID)
	return args.Get(0).(*models.Media), args.Error(1)
}

func (m *MockMediaService) UploadFiles(ctx context.Context, files map[string][]*multipart.FileHeader, userID primitive.ObjectID) ([]*models.Media, error) {
	args := m.Called(ctx, mock.AnythingOfType("map[string][]*multipart.FileHeader"), userID)
	return args.Get(0).([]*models.Media), args.Error(1)
}

func (m *MockMediaService) GetMedia(ctx context.Context, mediaID primitive.ObjectID) (*models.Media, error) {
	args := m.Called(ctx, mediaID)
	return args.Get(0).(*models.Media), args.Error(1)
}

func (m *MockMediaService) GetUserMedia(ctx context.Context, userID primitive.ObjectID, page, pageSize int, filters repository.MediaFilter) ([]*models.Media, int64, error) {
	args := m.Called(ctx, userID, page, pageSize, filters)
	return args.Get(0).([]*models.Media), args.Get(1).(int64), args.Error(2)
}

func (m *MockMediaService) DeleteMedia(ctx context.Context, mediaID, userID primitive.ObjectID) error {
	args := m.Called(ctx, mediaID, userID)
	return args.Error(0)
}

func (m *MockMediaService) GeneratePresignedUploadURL(ctx context.Context, filename, contentType string, size int64, userID primitive.ObjectID) (*services.PresignedUploadInfo, error) {
	args := m.Called(ctx, filename, contentType, size, userID)
	return args.Get(0).(*services.PresignedUploadInfo), args.Error(1)
}

func (m *MockMediaService) ProcessUploadedFile(ctx context.Context, presignedInfo *services.PresignedUploadInfo, userID primitive.ObjectID) (*models.Media, error) {
	args := m.Called(ctx, presignedInfo, userID)
	return args.Get(0).(*models.Media), args.Error(1)
}

func setupTestRouter(handler *UploadHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Mock auth middleware to set user ID
	router.Use(func(c *gin.Context) {
		userID := primitive.NewObjectID()
		c.Set("userID", userID.Hex())
		c.Next()
	})
	
	v1 := router.Group("/api/v1")
	{
		v1.POST("/upload", handler.HandleUpload)
		v1.POST("/upload/single", handler.HandleSingleUpload)
		v1.POST("/upload/presign", handler.HandlePresignURL)
		v1.POST("/upload/confirm", handler.HandleConfirmUpload)
		v1.GET("/media/:id", handler.HandleGetMedia)
		v1.GET("/media", handler.HandleListMedia)
		v1.DELETE("/media/:id", handler.HandleDeleteMedia)
	}
	
	return router
}

func createTestMedia() *models.Media {
	userID := primitive.NewObjectID()
	return &models.Media{
		ID:          primitive.NewObjectID(),
		Filename:    "test-photo.jpg",
		OriginalURL: "http://example.com/test-photo.jpg",
		Thumbnails: map[string]string{
			"small":  "http://example.com/test-photo-small.jpg",
			"medium": "http://example.com/test-photo-medium.jpg",
		},
		Size:      1024000,
		MimeType:  "image/jpeg",
		Width:     1920,
		Height:    1080,
		Format:    "jpeg",
		CreatedBy: userID,
	}
}

func createMultipartFile(t *testing.T, filename string, content []byte) (*bytes.Buffer, *multipart.Writer) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	
	part, err := writer.CreateFormFile("files", filename)
	require.NoError(t, err)
	
	_, err = part.Write(content)
	require.NoError(t, err)
	
	err = writer.Close()
	require.NoError(t, err)
	
	return body, writer
}

func TestUploadHandler_HandleSingleUpload(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockService := new(MockMediaService)
	handler := NewUploadHandler(mockService, logger)
	router := setupTestRouter(handler)

	userID := primitive.NewObjectID()
	testMedia := createTestMedia()
	testMedia.CreatedBy = userID

	tests := []struct {
		name           string
		filename       string
		fileContent    []byte
		setupMocks     func()
		expectedStatus int
		expectError    bool
	}{
		{
			name:        "successful upload",
			filename:    "test.jpg",
			fileContent: []byte("test image content"),
			setupMocks: func() {
				mockService.On("UploadFile", mock.Anything, mock.AnythingOfType("*multipart.FileHeader"), userID).Return(testMedia, nil)
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "upload failure",
			filename:       "test.jpg",
			fileContent:    []byte("test image content"),
			setupMocks:     func() { mockService.On("UploadFile", mock.Anything, mock.Anything, userID).Return(nil, fmt.Errorf("upload failed")) },
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.ExpectedCalls = nil
			tt.setupMocks()

			// Create multipart form
			body, writer := createMultipartFile(t, tt.filename, tt.fileContent)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/upload/single", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectError {
				var response ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.NotEmpty(t, response.Error)
			} else {
				var response UploadResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, testMedia.ID.Hex(), response.ID)
				assert.Equal(t, testMedia.Filename, response.Filename)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestUploadHandler_HandlePresignURL(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockService := new(MockMediaService)
	handler := NewUploadHandler(mockService, logger)
	router := setupTestRouter(handler)

	userID := primitive.NewObjectID()
	presignedInfo := &services.PresignedUploadInfo{
		URL:    "https://storage.example.com/upload",
		Fields: map[string]string{"mediaId": "test-media-id"},
		Key:    "uploads/test.jpg",
	}

	tests := []struct {
		name           string
		requestBody    interface{}
		setupMocks     func()
		expectedStatus int
		expectError    bool
	}{
		{
			name: "successful presign",
			requestBody: PresignedURLRequest{
				Filename:    "test.jpg",
				ContentType: "image/jpeg",
				Size:        1024000,
			},
			setupMocks: func() {
				mockService.On("GeneratePresignedUploadURL", mock.Anything, "test.jpg", "image/jpeg", int64(1024000), userID).Return(presignedInfo, nil)
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "invalid request body",
			requestBody:    "invalid json",
			setupMocks:     func() {},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "missing filename",
			requestBody: PresignedURLRequest{
				ContentType: "image/jpeg",
				Size:        1024000,
			},
			setupMocks:     func() {},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.ExpectedCalls = nil
			tt.setupMocks()

			var reqBody []byte
			var err error

			if str, ok := tt.requestBody.(string); ok {
				reqBody = []byte(str)
			} else {
				reqBody, err = json.Marshal(tt.requestBody)
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/upload/presign", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectError {
				var response ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.NotEmpty(t, response.Error)
			} else {
				var response PresignedUploadResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, presignedInfo.URL, response.UploadURL)
				assert.Equal(t, presignedInfo.Key, response.Key)
			}

			if tt.name == "successful presign" {
				mockService.AssertExpectations(t)
			}
		})
	}
}

func TestUploadHandler_HandleGetMedia(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockService := new(MockMediaService)
	handler := NewUploadHandler(mockService, logger)
	router := setupTestRouter(handler)

	userID := primitive.NewObjectID()
	testMedia := createTestMedia()
	testMedia.CreatedBy = userID

	tests := []struct {
		name           string
		mediaID        string
		setupMocks     func()
		expectedStatus int
		expectError    bool
	}{
		{
			name:    "successful retrieval",
			mediaID: testMedia.ID.Hex(),
			setupMocks: func() {
				mockService.On("GetMedia", mock.Anything, testMedia.ID).Return(testMedia, nil)
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "invalid media ID format",
			mediaID:        "invalid-id",
			setupMocks:     func() {},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:    "media not found",
			mediaID: testMedia.ID.Hex(),
			setupMocks: func() {
				mockService.On("GetMedia", mock.Anything, testMedia.ID).Return(nil, fmt.Errorf("media not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.ExpectedCalls = nil
			tt.setupMocks()

			req := httptest.NewRequest(http.MethodGet, "/api/v1/media/"+tt.mediaID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectError {
				var response ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.NotEmpty(t, response.Error)
			} else {
				var response UploadResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, testMedia.ID.Hex(), response.ID)
				assert.Equal(t, testMedia.Filename, response.Filename)
			}

			if tt.name == "successful retrieval" {
				mockService.AssertExpectations(t)
			}
		})
	}
}

func TestUploadHandler_HandleListMedia(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockService := new(MockMediaService)
	handler := NewUploadHandler(mockService, logger)
	router := setupTestRouter(handler)

	userID := primitive.NewObjectID()
	testMedia1 := createTestMedia()
	testMedia1.CreatedBy = userID
	testMedia2 := createTestMedia()
	testMedia2.CreatedBy = userID
	testMedia2.ID = primitive.NewObjectID()
	testMedia2.Filename = "test2.jpg"

	tests := []struct {
		name           string
		queryParams    string
		setupMocks     func()
		expectedStatus int
		expectedTotal  int64
	}{
		{
			name:        "successful listing",
			queryParams: "?page=1&pageSize=10",
			setupMocks: func() {
				mediaList := []*models.Media{testMedia1, testMedia2}
				mockService.On("GetUserMedia", mock.Anything, userID, 1, 10, mock.AnythingOfType("repository.MediaFilter")).Return(mediaList, int64(2), nil)
			},
			expectedStatus: http.StatusOK,
			expectedTotal:  2,
		},
		{
			name:        "listing with filters",
			queryParams: "?page=1&pageSize=5&mimeType=image/jpeg",
			setupMocks: func() {
				mediaList := []*models.Media{testMedia1}
				mockService.On("GetUserMedia", mock.Anything, userID, 1, 5, mock.AnythingOfType("repository.MediaFilter")).Return(mediaList, int64(1), nil)
			},
			expectedStatus: http.StatusOK,
			expectedTotal:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.ExpectedCalls = nil
			tt.setupMocks()

			req := httptest.NewRequest(http.MethodGet, "/api/v1/media"+tt.queryParams, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response MediaListResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedTotal, response.Total)
			assert.Equal(t, 1, response.Page)

			mockService.AssertExpectations(t)
		})
	}
}

func TestUploadHandler_HandleDeleteMedia(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockService := new(MockMediaService)
	handler := NewUploadHandler(mockService, logger)
	router := setupTestRouter(handler)

	userID := primitive.NewObjectID()
	testMedia := createTestMedia()
	testMedia.CreatedBy = userID

	tests := []struct {
		name           string
		mediaID        string
		setupMocks     func()
		expectedStatus int
		expectError    bool
	}{
		{
			name:    "successful deletion",
			mediaID: testMedia.ID.Hex(),
			setupMocks: func() {
				mockService.On("DeleteMedia", mock.Anything, testMedia.ID, userID).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "invalid media ID format",
			mediaID:        "invalid-id",
			setupMocks:     func() {},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:    "deletion failure",
			mediaID: testMedia.ID.Hex(),
			setupMocks: func() {
				mockService.On("DeleteMedia", mock.Anything, testMedia.ID, userID).Return(fmt.Errorf("deletion failed"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.ExpectedCalls = nil
			tt.setupMocks()

			req := httptest.NewRequest(http.MethodDelete, "/api/v1/media/"+tt.mediaID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectError {
				var response ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.NotEmpty(t, response.Error)
			} else {
				var response SuccessResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, "Media deleted successfully", response.Message)
			}

			if tt.name == "successful deletion" {
				mockService.AssertExpectations(t)
			}
		})
	}
}