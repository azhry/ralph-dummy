package services

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
	"wedding-invitation-backend/internal/domain/models"
	"wedding-invitation-backend/internal/domain/repository"
)

// MediaService handles media operations
type MediaService interface {
	UploadFile(ctx context.Context, file io.Reader, header *multipart.FileHeader, userID primitive.ObjectID) (*models.Media, error)
	UploadFiles(ctx context.Context, files map[string][]*multipart.FileHeader, userID primitive.ObjectID) ([]*models.Media, error)
	GetMedia(ctx context.Context, mediaID primitive.ObjectID) (*models.Media, error)
	GetUserMedia(ctx context.Context, userID primitive.ObjectID, page, pageSize int, filters repository.MediaFilter) ([]*models.Media, int64, error)
	DeleteMedia(ctx context.Context, mediaID, userID primitive.ObjectID) error
	GeneratePresignedUploadURL(ctx context.Context, filename, contentType string, size int64, userID primitive.ObjectID) (*PresignedUploadInfo, error)
	ProcessUploadedFile(ctx context.Context, presignedInfo *PresignedUploadInfo, userID primitive.ObjectID) (*models.Media, error)
}

type mediaService struct {
	mediaRepo      repository.MediaRepository
	storageService StorageService
	validator      FileValidator
	imageProcessor ImageProcessor
	logger         *zap.Logger
	config         *MediaServiceConfig
}

// MediaServiceConfig contains configuration for media service
type MediaServiceConfig struct {
	MaxFileSize    int64           `json:"maxFileSize"`
	MaxTotalSize   int64           `json:"maxTotalSize"`
	MaxFiles       int             `json:"maxFiles"`
	AllowedTypes   []string        `json:"allowedTypes"`
	ThumbnailSizes []ThumbnailSize `json:"thumbnailSizes"`
	EnableWebP     bool            `json:"enableWebP"`
	PresignExpiry  time.Duration   `json:"presignExpiry"`
	BaseURL        string          `json:"baseUrl"`
}

// DefaultMediaServiceConfig returns default configuration
func DefaultMediaServiceConfig() *MediaServiceConfig {
	return &MediaServiceConfig{
		MaxFileSize:  5 * 1024 * 1024,  // 5MB
		MaxTotalSize: 20 * 1024 * 1024, // 20MB
		MaxFiles:     10,
		AllowedTypes: []string{"image/jpeg", "image/png", "image/webp"},
		ThumbnailSizes: []ThumbnailSize{
			{Name: "small", Width: 150, Height: 150},
			{Name: "medium", Width: 400, Height: 400},
			{Name: "large", Width: 800, Height: 800},
		},
		EnableWebP:    true,
		PresignExpiry: 15 * time.Minute,
		BaseURL:       "http://localhost:8080/uploads",
	}
}

// NewMediaService creates a new media service
func NewMediaService(
	mediaRepo repository.MediaRepository,
	storageService StorageService,
	validator FileValidator,
	imageProcessor ImageProcessor,
	logger *zap.Logger,
	config *MediaServiceConfig,
) MediaService {
	if config == nil {
		config = DefaultMediaServiceConfig()
	}

	return &mediaService{
		mediaRepo:      mediaRepo,
		storageService: storageService,
		validator:      validator,
		imageProcessor: imageProcessor,
		logger:         logger,
		config:         config,
	}
}

// UploadFile handles single file upload
func (s *mediaService) UploadFile(ctx context.Context, file io.Reader, header *multipart.FileHeader, userID primitive.ObjectID) (*models.Media, error) {
	// Validate file size
	if header.Size > s.config.MaxFileSize {
		return nil, fmt.Errorf("file size %d exceeds maximum allowed size %d", header.Size, s.config.MaxFileSize)
	}

	// Validate file type and content
	validationResult, err := s.validator.Validate(ctx, file, header)
	if err != nil {
		return nil, fmt.Errorf("file validation failed: %w", err)
	}

	// Reset file pointer for processing
	if seeker, ok := file.(io.Seeker); ok {
		_, err = seeker.Seek(0, io.SeekStart)
		if err != nil {
			return nil, fmt.Errorf("failed to reset file pointer: %w", err)
		}
	}

	// Process image (generate thumbnails, extract metadata)
	processed, err := s.imageProcessor.Process(ctx, file, validationResult.MimeType)
	if err != nil {
		return nil, fmt.Errorf("failed to process image: %w", err)
	}

	// Generate storage key
	mediaID := primitive.NewObjectID()
	storageKey := s.generateStorageKey(mediaID, validationResult.Extension)

	// Upload original file
	originalURL, err := s.storageService.Upload(ctx, storageKey, processed.OriginalData,
		validationResult.MimeType, s.buildMetadata(processed.Metadata))
	if err != nil {
		return nil, fmt.Errorf("failed to upload original file: %w", err)
	}

	// Upload thumbnails
	thumbnails := make(map[string]string)
	for name, thumbData := range processed.Thumbnails {
		thumbKey := s.generateThumbnailKey(mediaID, name, validationResult.Extension)
		thumbURL, err := s.storageService.Upload(ctx, thumbKey, thumbData,
			validationResult.MimeType, nil)
		if err != nil {
			s.logger.Warn("Failed to upload thumbnail",
				zap.String("thumbnail", name),
				zap.Error(err))
			continue
		}
		thumbnails[name] = thumbURL
	}

	// Create media record
	media := &models.Media{
		ID:          mediaID,
		Filename:    header.Filename,
		OriginalURL: originalURL,
		Thumbnails:  thumbnails,
		Size:        header.Size,
		MimeType:    validationResult.MimeType,
		Width:       processed.Metadata.Width,
		Height:      processed.Metadata.Height,
		Format:      processed.Metadata.Format,
		EXIF:        processed.Metadata.EXIF,
		StorageKey:  storageKey,
		CreatedBy:   userID,
	}

	if err := s.mediaRepo.Create(ctx, media); err != nil {
		// Clean up uploaded files if database insert fails
		s.cleanupFailedUpload(ctx, storageKey, thumbnails)
		return nil, fmt.Errorf("failed to create media record: %w", err)
	}

	return media, nil
}

// UploadFiles handles multiple file uploads
func (s *mediaService) UploadFiles(ctx context.Context, files map[string][]*multipart.FileHeader, userID primitive.ObjectID) ([]*models.Media, error) {
	var allFiles []*multipart.FileHeader

	// Flatten files map
	for _, fileHeaders := range files {
		allFiles = append(allFiles, fileHeaders...)
	}

	if len(allFiles) == 0 {
		return nil, fmt.Errorf("no files provided")
	}

	if len(allFiles) > s.config.MaxFiles {
		return nil, fmt.Errorf("maximum %d files allowed per request", s.config.MaxFiles)
	}

	// Check total size
	var totalSize int64
	for _, fileHeader := range allFiles {
		totalSize += fileHeader.Size
	}

	if totalSize > s.config.MaxTotalSize {
		return nil, fmt.Errorf("total upload size %d exceeds maximum allowed size %d",
			totalSize, s.config.MaxTotalSize)
	}

	var mediaFiles []*models.Media

	for _, fileHeader := range allFiles {
		file, err := fileHeader.Open()
		if err != nil {
			s.logger.Error("Failed to open uploaded file",
				zap.String("filename", fileHeader.Filename),
				zap.Error(err))
			continue
		}
		defer file.Close()

		media, err := s.UploadFile(ctx, file, fileHeader, userID)
		if err != nil {
			s.logger.Error("Failed to upload file",
				zap.String("filename", fileHeader.Filename),
				zap.Error(err))
			continue
		}

		mediaFiles = append(mediaFiles, media)
	}

	if len(mediaFiles) == 0 {
		return nil, fmt.Errorf("failed to upload any files")
	}

	return mediaFiles, nil
}

// GetMedia retrieves a media file by ID
func (s *mediaService) GetMedia(ctx context.Context, mediaID primitive.ObjectID) (*models.Media, error) {
	return s.mediaRepo.GetByID(ctx, mediaID)
}

// GetUserMedia retrieves media files uploaded by a user
func (s *mediaService) GetUserMedia(ctx context.Context, userID primitive.ObjectID, page, pageSize int, filters repository.MediaFilter) ([]*models.Media, int64, error) {
	// Set created by filter if not already set
	filters.CreatedBy = &userID

	offset := int64((page - 1) * pageSize)
	limit := int64(pageSize)

	opts := repository.ListOptions{
		Limit:  limit,
		Offset: offset,
		Sort:   primitive.D{{Key: "createdAt", Value: -1}},
	}

	return s.mediaRepo.List(ctx, filters, opts)
}

// DeleteMedia deletes a media file (soft delete)
func (s *mediaService) DeleteMedia(ctx context.Context, mediaID, userID primitive.ObjectID) error {
	// Get media to verify ownership
	media, err := s.mediaRepo.GetByID(ctx, mediaID)
	if err != nil {
		return fmt.Errorf("media not found: %w", err)
	}

	// Check ownership
	if media.CreatedBy != userID {
		return fmt.Errorf("unauthorized: you can only delete your own media")
	}

	// Soft delete the media record
	return s.mediaRepo.SoftDelete(ctx, mediaID)
}

// GeneratePresignedUploadURL generates a pre-signed URL for direct upload
func (s *mediaService) GeneratePresignedUploadURL(ctx context.Context, filename, contentType string, size int64, userID primitive.ObjectID) (*PresignedUploadInfo, error) {
	// Validate file size
	if size > s.config.MaxFileSize {
		return nil, fmt.Errorf("file size %d exceeds maximum allowed size %d", size, s.config.MaxFileSize)
	}

	// Extract file extension
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(filename), "."))
	if ext == "" {
		return nil, fmt.Errorf("file must have an extension")
	}

	// Map extension to MIME type
	mimeType := s.extensionToMimeType(ext)
	if mimeType == "" {
		return nil, fmt.Errorf("unsupported file extension: %s", ext)
	}

	// Check if MIME type is allowed
	allowed := false
	for _, allowedType := range s.config.AllowedTypes {
		if allowedType == mimeType {
			allowed = true
			break
		}
	}
	if !allowed {
		return nil, fmt.Errorf("file type not allowed: %s", mimeType)
	}

	// Generate unique storage key
	mediaID := primitive.NewObjectID()
	storageKey := s.generateStorageKey(mediaID, ext)

	// Generate presigned upload URL
	presignedInfo, err := s.storageService.GeneratePresignedUploadURL(
		ctx, storageKey, contentType, size, s.config.PresignExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate presigned upload URL: %w", err)
	}

	// Store media ID in presigned info for later processing
	presignedInfo.Key = storageKey
	presignedInfo.Fields["mediaId"] = mediaID.Hex()

	return presignedInfo, nil
}

// ProcessUploadedFile processes a file uploaded via pre-signed URL
func (s *mediaService) ProcessUploadedFile(ctx context.Context, presignedInfo *PresignedUploadInfo, userID primitive.ObjectID) (*models.Media, error) {
	// In a real implementation, this would:
	// 1. Verify the file exists in storage
	// 2. Download the file for processing
	// 3. Process the image (thumbnails, metadata)
	// 4. Upload thumbnails
	// 5. Create media record in database

	// For now, return a placeholder
	mediaID := primitive.NewObjectID()

	return &models.Media{
		ID:          mediaID,
		Filename:    "uploaded-file.jpg",
		OriginalURL: fmt.Sprintf("%s/%s", s.config.BaseURL, presignedInfo.Key),
		Size:        1024000,
		MimeType:    "image/jpeg",
		Width:       1920,
		Height:      1080,
		Format:      "jpeg",
		StorageKey:  presignedInfo.Key,
		CreatedBy:   userID,
	}, nil
}

// Helper functions

func (s *mediaService) generateStorageKey(mediaID primitive.ObjectID, ext string) string {
	date := time.Now().Format("2006/01/02")
	return fmt.Sprintf("uploads/%s/%s/original.%s", date, mediaID.Hex(), ext)
}

func (s *mediaService) generateThumbnailKey(mediaID primitive.ObjectID, name, ext string) string {
	date := time.Now().Format("2006/01/02")
	return fmt.Sprintf("uploads/%s/%s/%s.%s", date, mediaID.Hex(), name, ext)
}

func (s *mediaService) buildMetadata(metadata *ImageMetadata) map[string]string {
	result := map[string]string{
		"width":  fmt.Sprintf("%d", metadata.Width),
		"height": fmt.Sprintf("%d", metadata.Height),
		"format": metadata.Format,
	}

	if metadata.EXIF != nil {
		for key, value := range metadata.EXIF {
			result["exif_"+key] = fmt.Sprintf("%v", value)
		}
	}

	return result
}

func (s *mediaService) cleanupFailedUpload(ctx context.Context, originalKey string, thumbnails map[string]string) {
	// Delete original file
	s.storageService.Delete(ctx, originalKey)

	// Delete thumbnails
	for _, url := range thumbnails {
		// Extract key from URL (simple implementation)
		key := strings.TrimPrefix(url, s.config.BaseURL+"/")
		s.storageService.Delete(ctx, key)
	}
}

func (s *mediaService) extensionToMimeType(ext string) string {
	switch ext {
	case "jpg", "jpeg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "webp":
		return "image/webp"
	default:
		return ""
	}
}
