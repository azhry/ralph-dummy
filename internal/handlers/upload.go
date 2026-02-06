package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
	"wedding-invitation-backend/internal/domain/models"
	"wedding-invitation-backend/internal/domain/repository"
	"wedding-invitation-backend/internal/services"
)

// UploadHandler handles file upload requests
type UploadHandler struct {
	mediaService services.MediaService
	logger       *zap.Logger
}

// NewUploadHandler creates a new upload handler
func NewUploadHandler(mediaService services.MediaService, logger *zap.Logger) *UploadHandler {
	return &UploadHandler{
		mediaService: mediaService,
		logger:       logger,
	}
}

// UploadRequest represents a file upload request
type UploadRequest struct {
	Files map[string][]*FileMetadata `json:"files" binding:"required"`
}

// FileMetadata represents file metadata
type FileMetadata struct {
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	MimeType string `json:"mimeType"`
}

// UploadResponse represents the response for a successful upload
type UploadResponse struct {
	ID          string                      `json:"id"`
	Filename    string                      `json:"filename"`
	OriginalURL string                      `json:"originalUrl"`
	Thumbnails  map[string]string           `json:"thumbnails,omitempty"`
	Size        int64                       `json:"size"`
	MimeType    string                      `json:"mimeType"`
	Width       int                         `json:"width,omitempty"`
	Height      int                         `json:"height,omitempty"`
	Format      string                      `json:"format,omitempty"`
	EXIF        map[string]interface{}       `json:"exif,omitempty"`
	CreatedAt   string                      `json:"createdAt"`
}

// PresignedURLRequest represents a request for pre-signed upload URL
type PresignedURLRequest struct {
	Filename   string `json:"filename" binding:"required"`
	ContentType string `json:"contentType" binding:"required"`
	Size       int64  `json:"size" binding:"required"`
}

// PresignedUploadResponse represents a pre-signed upload URL response
type PresignedUploadResponse struct {
	UploadURL string            `json:"uploadUrl"`
	Fields    map[string]string `json:"fields,omitempty"`
	Key       string            `json:"key"`
	MediaID   string            `json:"mediaId"`
	ExpiresAt string            `json:"expiresAt"`
}

// MediaListResponse represents a paginated list of media files
type MediaListResponse struct {
	Media      []*UploadResponse `json:"media"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"pageSize"`
	TotalPages int               `json:"totalPages"`
}

// HandleUpload handles multipart file uploads
// @Summary Upload files
// @Description Upload one or more files with validation and processing
// @Tags upload
// @Accept multipart/form-data
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param files formData file true "Files to upload"
// @Success 200 {array} UploadResponse
// @Failure 400 {object} ErrorResponse
// @Failure 413 {object} ErrorResponse
// @Failure 415 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/upload [post]
func (h *UploadHandler) HandleUpload(c *gin.Context) {
	ctx := c.Request.Context()
	userID := h.getUserIDFromContext(c)
	if userID == nil {
		respondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse multipart form
	err := c.Request.ParseMultipartForm(32 << 20) // 32MB max memory
	if err != nil {
		h.logger.Error("Failed to parse multipart form", zap.Error(err))
		respondWithError(c, http.StatusBadRequest, "Invalid form data")
		return
	}
	defer c.Request.MultipartForm.RemoveAll()

	// Get uploaded files
	files := c.Request.MultipartForm.File["files"]
	if len(files) == 0 {
		respondWithError(c, http.StatusBadRequest, "No files provided")
		return
	}

	// Convert to map for service
	filesMap := make(map[string][]*multipart.FileHeader)
	for i, fileHeader := range files {
		key := fmt.Sprintf("file_%d", i)
		filesMap[key] = []*multipart.FileHeader{fileHeader}
	}

	// Upload files
	mediaFiles, err := h.mediaService.UploadFiles(ctx, filesMap, *userID)
	if err != nil {
		h.logger.Error("Failed to upload files", zap.Error(err))
		respondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Convert to response format
	response := make([]*UploadResponse, len(mediaFiles))
	for i, media := range mediaFiles {
		response[i] = h.convertMediaToResponse(media)
	}

	respondWithJSON(c, http.StatusOK, response)
}

// HandleSingleUpload handles single file upload for simplicity
// @Summary Upload single file
// @Description Upload a single file with validation and processing
// @Tags upload
// @Accept multipart/form-data
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param file formData file true "File to upload"
// @Success 200 {object} UploadResponse
// @Failure 400 {object} ErrorResponse
// @Failure 413 {object} ErrorResponse
// @Failure 415 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/upload/single [post]
func (h *UploadHandler) HandleSingleUpload(c *gin.Context) {
	ctx := c.Request.Context()
	userID := h.getUserIDFromContext(c)
	if userID == nil {
		respondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get uploaded file
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		h.logger.Error("Failed to get uploaded file", zap.Error(err))
		respondWithError(c, http.StatusBadRequest, "No file provided")
		return
	}
	defer file.Close()

	// Upload file
	media, err := h.mediaService.UploadFile(ctx, file, header, *userID)
	if err != nil {
		h.logger.Error("Failed to upload file", zap.Error(err))
		respondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	response := h.convertMediaToResponse(media)
	respondWithJSON(c, http.StatusOK, response)
}

// HandlePresignURL generates a pre-signed URL for direct upload
// @Summary Generate pre-signed upload URL
// @Description Generate a pre-signed URL for direct file upload to storage
// @Tags upload
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param request body PresignedURLRequest true "Presigned URL request"
// @Success 200 {object} PresignedUploadResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/upload/presign [post]
func (h *UploadHandler) HandlePresignURL(c *gin.Context) {
	ctx := c.Request.Context()
	userID := h.getUserIDFromContext(c)
	if userID == nil {
		respondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req PresignedURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		respondWithError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if req.Filename == "" {
		respondWithError(c, http.StatusBadRequest, "Filename is required")
		return
	}

	if req.ContentType == "" {
		respondWithError(c, http.StatusBadRequest, "Content type is required")
		return
	}

	if req.Size <= 0 {
		respondWithError(c, http.StatusBadRequest, "File size must be greater than 0")
		return
	}

	// Generate presigned upload URL
	presignedInfo, err := h.mediaService.GeneratePresignedUploadURL(
		ctx, req.Filename, req.ContentType, req.Size, *userID)
	if err != nil {
		h.logger.Error("Failed to generate presigned upload URL", zap.Error(err))
		respondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Extract media ID from fields
	mediaID := ""
	if presignedInfo.Fields != nil {
		mediaID = presignedInfo.Fields["mediaId"]
	}

	response := PresignedUploadResponse{
		UploadURL: presignedInfo.URL,
		Fields:    presignedInfo.Fields,
		Key:       presignedInfo.Key,
		MediaID:   mediaID,
		ExpiresAt: "", // Calculate from expiry time
	}

	respondWithJSON(c, http.StatusOK, response)
}

// HandleConfirmUpload confirms a file uploaded via pre-signed URL
// @Summary Confirm pre-signed upload
// @Description Confirm a file that was uploaded using a pre-signed URL
// @Tags upload
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param request body ConfirmUploadRequest true "Confirm upload request"
// @Success 200 {object} UploadResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/upload/confirm [post]
func (h *UploadHandler) HandleConfirmUpload(c *gin.Context) {
	ctx := c.Request.Context()
	userID := h.getUserIDFromContext(c)
	if userID == nil {
		respondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req ConfirmUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		respondWithError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if req.MediaID == "" {
		respondWithError(c, http.StatusBadRequest, "Media ID is required")
		return
	}

	mediaID, err := primitive.ObjectIDFromHex(req.MediaID)
	if err != nil {
		respondWithError(c, http.StatusBadRequest, "Invalid media ID format")
		return
	}

	// Create presigned info from request
	presignedInfo := &services.PresignedUploadInfo{
		Key: req.Key,
		Fields: map[string]string{
			"mediaId": req.MediaID,
		},
	}

	// Process uploaded file
	media, err := h.mediaService.ProcessUploadedFile(ctx, presignedInfo, *userID)
	if err != nil {
		h.logger.Error("Failed to process uploaded file", zap.Error(err))
		respondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	response := h.convertMediaToResponse(media)
	respondWithJSON(c, http.StatusOK, response)
}

// HandleGetMedia retrieves a media file by ID
// @Summary Get media file
// @Description Retrieve media file metadata by ID
// @Tags upload
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Media ID"
// @Success 200 {object} UploadResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/media/{id} [get]
func (h *UploadHandler) HandleGetMedia(c *gin.Context) {
	ctx := c.Request.Context()
	userID := h.getUserIDFromContext(c)
	if userID == nil {
		respondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	mediaIDStr := c.Param("id")
	mediaID, err := primitive.ObjectIDFromHex(mediaIDStr)
	if err != nil {
		respondWithError(c, http.StatusBadRequest, "Invalid media ID format")
		return
	}

	media, err := h.mediaService.GetMedia(ctx, mediaID)
	if err != nil {
		h.logger.Error("Failed to get media", zap.String("mediaID", mediaIDStr), zap.Error(err))
		respondWithError(c, http.StatusNotFound, "Media not found")
		return
	}

	// Check ownership (optional: you might want to allow public access to certain media)
	if media.CreatedBy != *userID {
		respondWithError(c, http.StatusForbidden, "Access denied")
		return
	}

	response := h.convertMediaToResponse(media)
	respondWithJSON(c, http.StatusOK, response)
}

// HandleListMedia retrieves paginated list of user's media files
// @Summary List user media files
// @Description Retrieve a paginated list of media files uploaded by the user
// @Tags upload
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(10)
// @Param mimeType query string false "Filter by MIME type"
// @Success 200 {object} MediaListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/media [get]
func (h *UploadHandler) HandleListMedia(c *gin.Context) {
	ctx := c.Request.Context()
	userID := h.getUserIDFromContext(c)
	if userID == nil {
		respondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// Parse filters
	filters := repository.MediaFilter{
		MimeType: c.Query("mimeType"),
	}

	mediaList, total, err := h.mediaService.GetUserMedia(ctx, *userID, page, pageSize, filters)
	if err != nil {
		h.logger.Error("Failed to get user media", zap.Error(err))
		respondWithError(c, http.StatusInternalServerError, "Failed to retrieve media")
		return
	}

	// Convert to response format
	response := MediaListResponse{
		Media:      make([]*UploadResponse, len(mediaList)),
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int((total + int64(pageSize) - 1) / int64(pageSize)),
	}

	for i, media := range mediaList {
		response.Media[i] = h.convertMediaToResponse(media)
	}

	respondWithJSON(c, http.StatusOK, response)
}

// HandleDeleteMedia deletes a media file
// @Summary Delete media file
// @Description Delete a media file by ID
// @Tags upload
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Media ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/media/{id} [delete]
func (h *UploadHandler) HandleDeleteMedia(c *gin.Context) {
	ctx := c.Request.Context()
	userID := h.getUserIDFromContext(c)
	if userID == nil {
		respondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	mediaIDStr := c.Param("id")
	mediaID, err := primitive.ObjectIDFromHex(mediaIDStr)
	if err != nil {
		respondWithError(c, http.StatusBadRequest, "Invalid media ID format")
		return
	}

	err = h.mediaService.DeleteMedia(ctx, mediaID, *userID)
	if err != nil {
		h.logger.Error("Failed to delete media", zap.String("mediaID", mediaIDStr), zap.Error(err))
		respondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(c, http.StatusOK, gin.H{"message": "Media deleted successfully"})
}

// Helper functions

// ConfirmUploadRequest represents a request to confirm a pre-signed upload
type ConfirmUploadRequest struct {
	MediaID string `json:"mediaId" binding:"required"`
	Key     string `json:"key" binding:"required"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Message string `json:"message"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// getUserIDFromContext extracts user ID from JWT token context
func (h *UploadHandler) getUserIDFromContext(c *gin.Context) *primitive.ObjectID {
	// This should extract user ID from JWT token set by auth middleware
	// For now, we'll use a placeholder implementation
	userIDStr, exists := c.Get("userID")
	if !exists {
		return nil
	}

	userIDStrStr, ok := userIDStr.(string)
	if !ok {
		return nil
	}

	userID, err := primitive.ObjectIDFromHex(userIDStrStr)
	if err != nil {
		return nil
	}

	return &userID
}

// convertMediaToResponse converts a media model to upload response
func (h *UploadHandler) convertMediaToResponse(media *models.Media) *UploadResponse {
	return &UploadResponse{
		ID:          media.ID.Hex(),
		Filename:    media.Filename,
		OriginalURL: media.OriginalURL,
		Thumbnails:  media.Thumbnails,
		Size:        media.Size,
		MimeType:    media.MimeType,
		Width:       media.Width,
		Height:      media.Height,
		Format:      media.Format,
		EXIF:        media.EXIF,
		CreatedAt:   media.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// respondWithJSON sends a JSON response
func respondWithJSON(c *gin.Context, code int, payload interface{}) {
	c.JSON(code, payload)
}

// respondWithError sends an error response
func respondWithError(c *gin.Context, code int, message string) {
	c.JSON(code, ErrorResponse{Error: message})
}