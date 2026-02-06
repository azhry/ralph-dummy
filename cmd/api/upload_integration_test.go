package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap/zaptest"
	"wedding-invitation-backend/internal/domain/models"
	"wedding-invitation-backend/internal/domain/repository"
	"wedding-invitation-backend/internal/handlers"
	"wedding-invitation-backend/internal/services"
)

// TestUploadIntegration tests the complete file upload integration
func TestUploadIntegration(t *testing.T) {
	// Setup test environment
	gin.SetMode(gin.TestMode)
	logger := zaptest.NewLogger(t)

	// Create mock services
	mockMediaService := new(MockMediaServiceForIntegration)
	uploadHandler := handlers.NewUploadHandler(mockMediaService, logger)

	// Create router with upload routes
	router := gin.New()

	// Mock auth middleware to set user ID
	router.Use(func(c *gin.Context) {
		userID := primitive.NewObjectID()
		c.Set("userID", userID.Hex())
		c.Next()
	})

	// Add upload routes
	v1 := router.Group("/api/v1")
	protected := v1.Group("/")
	{
		protected.POST("/upload/single", uploadHandler.HandleSingleUpload)
		protected.POST("/upload/presign", uploadHandler.HandlePresignURL)
		protected.POST("/upload/confirm", uploadHandler.HandleConfirmUpload)
		protected.GET("/media/:id", uploadHandler.HandleGetMedia)
		protected.GET("/media", uploadHandler.HandleListMedia)
		protected.DELETE("/media/:id", uploadHandler.HandleDeleteMedia)
	}

	userID := primitive.NewObjectID()
	testMedia := createTestMedia()
	testMedia.CreatedBy = userID

	t.Run("Complete upload workflow", func(t *testing.T) {
		// Step 1: Generate presigned URL
		presignedInfo := &services.PresignedUploadInfo{
			URL:    "https://storage.example.com/upload",
			Fields: map[string]string{"mediaId": testMedia.ID.Hex()},
			Key:    "uploads/test.jpg",
		}

		mockMediaService.On("GeneratePresignedUploadURL", mock.Anything, "test.jpg", "image/jpeg", int64(1024000), mock.AnythingOfType("primitive.ObjectID")).
			Return(presignedInfo, nil).Once()

		presignReq := handlers.PresignedURLRequest{
			Filename:    "test.jpg",
			ContentType: "image/jpeg",
			Size:        1024000,
		}

		presignBody, err := json.Marshal(presignReq)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/upload/presign", bytes.NewBuffer(presignBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var presignResp handlers.PresignedUploadResponse
		err = json.Unmarshal(w.Body.Bytes(), &presignResp)
		require.NoError(t, err)
		assert.Equal(t, presignedInfo.URL, presignResp.UploadURL)
		assert.Equal(t, testMedia.ID.Hex(), presignResp.MediaID)

		// Step 2: Process uploaded file (simulate direct upload completion)
		mockMediaService.On("ProcessUploadedFile", mock.Anything, mock.AnythingOfType("*services.PresignedUploadInfo"), userID).
			Return(testMedia, nil).Once()

		confirmReq := handlers.ConfirmUploadRequest{
			MediaID: testMedia.ID.Hex(),
			Key:     presignedInfo.Key,
		}

		confirmBody, err := json.Marshal(confirmReq)
		require.NoError(t, err)

		req = httptest.NewRequest(http.MethodPost, "/api/v1/upload/confirm", bytes.NewBuffer(confirmBody))
		req.Header.Set("Content-Type", "application/json")

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var mediaResp handlers.UploadResponse
		err = json.Unmarshal(w.Body.Bytes(), &mediaResp)
		require.NoError(t, err)
		assert.Equal(t, testMedia.ID.Hex(), mediaResp.ID)
		assert.Equal(t, testMedia.Filename, mediaResp.Filename)

		// Step 3: Retrieve media
		mockMediaService.On("GetMedia", mock.Anything, testMedia.ID).Return(testMedia, nil).Once()

		req = httptest.NewRequest(http.MethodGet, "/api/v1/media/"+testMedia.ID.Hex(), nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var getMediaResp handlers.UploadResponse
		err = json.Unmarshal(w.Body.Bytes(), &getMediaResp)
		require.NoError(t, err)
		assert.Equal(t, testMedia.ID.Hex(), getMediaResp.ID)

		// Step 4: List user media
		mockMediaService.On("GetUserMedia", mock.Anything, userID, 1, 10, mock.AnythingOfType("repository.MediaFilter")).
			Return([]*models.Media{testMedia}, int64(1), nil).Once()

		req = httptest.NewRequest(http.MethodGet, "/api/v1/media?page=1&pageSize=10", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var listResp handlers.MediaListResponse
		err = json.Unmarshal(w.Body.Bytes(), &listResp)
		require.NoError(t, err)
		assert.Len(t, listResp.Media, 1)
		assert.Equal(t, int64(1), listResp.Total)

		// Step 5: Delete media
		mockMediaService.On("DeleteMedia", mock.Anything, testMedia.ID, userID).Return(nil).Once()

		req = httptest.NewRequest(http.MethodDelete, "/api/v1/media/"+testMedia.ID.Hex(), nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		mockMediaService.AssertExpectations(t)
	})
}

// MockMediaServiceForIntegration is a simplified mock for integration tests
type MockMediaServiceForIntegration struct {
	mock.Mock
}

func (m *MockMediaServiceForIntegration) UploadFile(ctx context.Context, file io.Reader, header *multipart.FileHeader, userID primitive.ObjectID) (*models.Media, error) {
	args := m.Called(ctx, mock.AnythingOfType("io.Reader"), mock.AnythingOfType("*multipart.FileHeader"), userID)
	return args.Get(0).(*models.Media), args.Error(1)
}

func (m *MockMediaServiceForIntegration) UploadFiles(ctx context.Context, files map[string][]*multipart.FileHeader, userID primitive.ObjectID) ([]*models.Media, error) {
	args := m.Called(ctx, mock.AnythingOfType("map[string][]*multipart.FileHeader"), userID)
	return args.Get(0).([]*models.Media), args.Error(1)
}

func (m *MockMediaServiceForIntegration) GetMedia(ctx context.Context, mediaID primitive.ObjectID) (*models.Media, error) {
	args := m.Called(ctx, mediaID)
	return args.Get(0).(*models.Media), args.Error(1)
}

func (m *MockMediaServiceForIntegration) GetUserMedia(ctx context.Context, userID primitive.ObjectID, page, pageSize int, filters repository.MediaFilter) ([]*models.Media, int64, error) {
	args := m.Called(ctx, userID, page, pageSize, mock.AnythingOfType("repository.MediaFilter"))
	return args.Get(0).([]*models.Media), args.Get(1).(int64), args.Error(2)
}

func (m *MockMediaServiceForIntegration) DeleteMedia(ctx context.Context, mediaID, userID primitive.ObjectID) error {
	args := m.Called(ctx, mediaID, userID)
	return args.Error(0)
}

func (m *MockMediaServiceForIntegration) GeneratePresignedUploadURL(ctx context.Context, filename, contentType string, size int64, userID primitive.ObjectID) (*services.PresignedUploadInfo, error) {
	args := m.Called(ctx, filename, contentType, size, userID)
	return args.Get(0).(*services.PresignedUploadInfo), args.Error(1)
}

func (m *MockMediaServiceForIntegration) ProcessUploadedFile(ctx context.Context, presignedInfo *services.PresignedUploadInfo, userID primitive.ObjectID) (*models.Media, error) {
	args := m.Called(ctx, presignedInfo, userID)
	return args.Get(0).(*models.Media), args.Error(1)
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
