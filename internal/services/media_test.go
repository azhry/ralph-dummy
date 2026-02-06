package services

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap/zaptest"
	"wedding-invitation-backend/internal/domain/models"
	"wedding-invitation-backend/internal/domain/repository"
)

// MockStorageService is a mock implementation of StorageService
type MockStorageService struct {
	mock.Mock
}

func (m *MockStorageService) Upload(ctx context.Context, key string, data []byte, contentType string, metadata map[string]string) (string, error) {
	args := m.Called(ctx, key, data, contentType, metadata)
	return args.String(0), args.Error(1)
}

func (m *MockStorageService) UploadStream(ctx context.Context, key string, reader io.Reader, contentType string, size int64, metadata map[string]string) (string, error) {
	args := m.Called(ctx, key, reader, contentType, size, metadata)
	return args.String(0), args.Error(1)
}

func (m *MockStorageService) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockStorageService) GetPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	args := m.Called(ctx, key, expiry)
	return args.String(0), args.Error(1)
}

func (m *MockStorageService) GeneratePresignedUploadURL(ctx context.Context, key string, contentType string, size int64, expiry time.Duration) (*PresignedUploadInfo, error) {
	args := m.Called(ctx, key, contentType, size, expiry)
	return args.Get(0).(*PresignedUploadInfo), args.Error(1)
}

func (m *MockStorageService) Exists(ctx context.Context, key string) (bool, error) {
	args := m.Called(ctx, key)
	return args.Bool(0), args.Error(1)
}

// MockMediaRepository is a mock implementation of MediaRepository
type MockMediaRepository struct {
	mock.Mock
}

func (m *MockMediaRepository) Create(ctx context.Context, media *models.Media) error {
	args := m.Called(ctx, media)
	return args.Error(0)
}

func (m *MockMediaRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.Media, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.Media), args.Error(1)
}

func (m *MockMediaRepository) GetByStorageKey(ctx context.Context, key string) (*models.Media, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(*models.Media), args.Error(1)
}

func (m *MockMediaRepository) List(ctx context.Context, filter repository.MediaFilter, opts repository.ListOptions) ([]*models.Media, int64, error) {
	args := m.Called(ctx, filter, opts)
	return args.Get(0).([]*models.Media), args.Get(1).(int64), args.Error(2)
}

func (m *MockMediaRepository) Update(ctx context.Context, media *models.Media) error {
	args := m.Called(ctx, media)
	return args.Error(0)
}

func (m *MockMediaRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockMediaRepository) SoftDelete(ctx context.Context, id primitive.ObjectID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockMediaRepository) GetOrphaned(ctx context.Context, before time.Time) ([]*models.Media, error) {
	args := m.Called(ctx, before)
	return args.Get(0).([]*models.Media), args.Error(1)
}

func (m *MockMediaRepository) GetByCreatedBy(ctx context.Context, userID primitive.ObjectID, opts repository.ListOptions) ([]*models.Media, int64, error) {
	args := m.Called(ctx, userID, opts)
	return args.Get(0).([]*models.Media), args.Get(1).(int64), args.Error(2)
}

func TestFileValidator_Validate(t *testing.T) {
	validator := NewFileValidator([]string{"image/jpeg", "image/png", "image/webp"}, 5*1024*1024)
	
	tests := []struct {
		name           string
		filename       string
		fileContent    []byte
		expectedError  string
		expectedResult *ValidationResult
	}{
		{
			name:           "valid JPEG file",
			filename:       "test.jpg",
			fileContent:    []byte{0xFF, 0xD8, 0xFF, 0xE0}, // JPEG magic number
			expectedResult: &ValidationResult{MimeType: "image/jpeg", Extension: "jpg", IsValid: true},
		},
		{
			name:           "valid PNG file",
			filename:       "test.png",
			fileContent:    []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, // PNG magic number
			expectedResult: &ValidationResult{MimeType: "image/png", Extension: "png", IsValid: true},
		},
		{
			name:          "no extension",
			filename:      "test",
			fileContent:   []byte{0xFF, 0xD8, 0xFF, 0xE0},
			expectedError: "file must have an extension",
		},
		{
			name:          "unsupported extension",
			filename:      "test.pdf",
			fileContent:   []byte{0xFF, 0xD8, 0xFF, 0xE0},
			expectedError: "unsupported file extension: pdf",
		},
		{
			name:          "invalid magic number",
			filename:      "test.jpg",
			fileContent:   []byte{0x00, 0x01, 0x02, 0x03},
			expectedError: "file content does not match extension: invalid magic number",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			reader := bytes.NewReader(tt.fileContent)
			header := &multipart.FileHeader{
				Filename: tt.filename,
				Size:     int64(len(tt.fileContent)),
			}

			result, err := validator.Validate(ctx, reader, header)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult.MimeType, result.MimeType)
				assert.Equal(t, tt.expectedResult.Extension, result.Extension)
				assert.Equal(t, tt.expectedResult.IsValid, result.IsValid)
			}
		})
	}
}

func TestMediaService_UploadFile(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)
	
	// Create mocks
	mockRepo := new(MockMediaRepository)
	mockStorage := new(MockStorageService)
	validator := NewFileValidator([]string{"image/jpeg", "image/png", "image/webp"}, 5*1024*1024)
	imageProcessor := NewImageProcessor([]ThumbnailSize{}, false) // No thumbnails for simplicity
	
	config := DefaultMediaServiceConfig()
	service := NewMediaService(mockRepo, mockStorage, validator, imageProcessor, logger, config)
	
	userID := primitive.NewObjectID()
	
	// Valid JPEG file content
	jpegContent := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01}
	
	tests := []struct {
		name        string
		filename    string
		fileSize    int64
		setupMocks  func()
		expectError bool
		errorMsg    string
	}{
		{
			name:     "successful upload",
			filename: "test.jpg",
			fileSize: int64(len(jpegContent)),
			setupMocks: func() {
				mockStorage.On("Upload", mock.Anything, mock.AnythingOfType("string"), 
					mock.AnythingOfType("[]uint8"), "image/jpeg", mock.AnythingOfType("map[string]string")).
					Return("http://example.com/uploads/test.jpg", nil)
				mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Media")).Return(nil)
			},
			expectError: false,
		},
		{
			name:        "file too large",
			filename:    "test.jpg",
			fileSize:    10 * 1024 * 1024, // 10MB
			setupMocks:  func() {},
			expectError: true,
			errorMsg:    "file size 10485760 exceeds maximum allowed size 5242880",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockRepo.ExpectedCalls = nil
			mockStorage.ExpectedCalls = nil
			
			tt.setupMocks()
			
			reader := bytes.NewReader(jpegContent)
			header := &multipart.FileHeader{
				Filename: tt.filename,
				Size:     tt.fileSize,
			}

			result, err := service.UploadFile(ctx, reader, header, userID)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				mockRepo.AssertExpectations(t)
				mockStorage.AssertExpectations(t)
			}
		})
	}
}

func TestMediaService_GetUserMedia(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)
	
	// Create mocks
	mockRepo := new(MockMediaRepository)
	mockStorage := new(MockStorageService)
	validator := NewFileValidator([]string{"image/jpeg", "image/png", "image/webp"}, 5*1024*1024)
	imageProcessor := NewImageProcessor([]ThumbnailSize{}, false)
	
	config := DefaultMediaServiceConfig()
	service := NewMediaService(mockRepo, mockStorage, validator, imageProcessor, logger, config)
	
	userID := primitive.NewObjectID()
	
	tests := []struct {
		name         string
		page         int
		pageSize     int
		filters      repository.MediaFilter
		setupMocks   func()
		expectError  bool
		expectedLen  int
		expectedTotal int64
	}{
		{
			name:     "successful retrieval",
			page:     1,
			pageSize: 10,
			filters:  repository.MediaFilter{},
			setupMocks: func() {
				mediaList := []*models.Media{
					{ID: primitive.NewObjectID(), Filename: "test1.jpg"},
					{ID: primitive.NewObjectID(), Filename: "test2.jpg"},
				}
				mockRepo.On("List", ctx, mock.MatchedBy(func(filter repository.MediaFilter) bool {
					return filter.CreatedBy != nil && *filter.CreatedBy == userID
				}), mock.AnythingOfType("repository.ListOptions")).Return(mediaList, int64(2), nil)
			},
			expectError:   false,
			expectedLen:  2,
			expectedTotal: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			
			tt.setupMocks()
			
			result, total, err := service.GetUserMedia(ctx, userID, tt.page, tt.pageSize, tt.filters)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, tt.expectedLen)
				assert.Equal(t, tt.expectedTotal, total)
				mockRepo.AssertExpectations(t)
			}
		})
	}
}

func TestMediaService_DeleteMedia(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)
	
	// Create mocks
	mockRepo := new(MockMediaRepository)
	mockStorage := new(MockStorageService)
	validator := NewFileValidator([]string{"image/jpeg", "image/png", "image/webp"}, 5*1024*1024)
	imageProcessor := NewImageProcessor([]ThumbnailSize{}, false)
	
	config := DefaultMediaServiceConfig()
	service := NewMediaService(mockRepo, mockStorage, validator, imageProcessor, logger, config)
	
	userID := primitive.NewObjectID()
	mediaID := primitive.NewObjectID()
	otherUserID := primitive.NewObjectID()

	tests := []struct {
		name        string
		mediaID     primitive.ObjectID
		userID      primitive.ObjectID
		setupMocks  func()
		expectError bool
		errorMsg    string
	}{
		{
			name:    "successful deletion",
			mediaID: mediaID,
			userID:  userID,
			setupMocks: func() {
				media := &models.Media{ID: mediaID, CreatedBy: userID}
				mockRepo.On("GetByID", ctx, mediaID).Return(media, nil)
				mockRepo.On("SoftDelete", ctx, mediaID).Return(nil)
			},
			expectError: false,
		},
		{
			name:    "media not found",
			mediaID: mediaID,
			userID:  userID,
			setupMocks: func() {
				mockRepo.On("GetByID", ctx, mediaID).Return(nil, fmt.Errorf("media not found"))
			},
			expectError: true,
			errorMsg:    "media not found",
		},
		{
			name:    "unauthorized deletion",
			mediaID: mediaID,
			userID:  otherUserID,
			setupMocks: func() {
				media := &models.Media{ID: mediaID, CreatedBy: userID}
				mockRepo.On("GetByID", ctx, mediaID).Return(media, nil)
			},
			expectError: true,
			errorMsg:    "unauthorized: you can only delete your own media",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			
			tt.setupMocks()
			
			err := service.DeleteMedia(ctx, tt.mediaID, tt.userID)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				mockRepo.AssertExpectations(t)
			}
		})
	}
}