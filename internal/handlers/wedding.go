package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"wedding-invitation-backend/internal/domain/models"
	"wedding-invitation-backend/internal/domain/repository"
	"wedding-invitation-backend/internal/services"
	"wedding-invitation-backend/internal/utils"
)

type WeddingHandler struct {
	weddingService *services.WeddingService
}

func NewWeddingHandler(weddingService *services.WeddingService) *WeddingHandler {
	return &WeddingHandler{
		weddingService: weddingService,
	}
}

// CreateWedding godoc
// @Summary Create a new wedding
// @Description Create a new wedding for the authenticated user
// @Tags weddings
// @Accept json
// @Produce json
// @Param wedding body models.Wedding true "Wedding data"
// @Success 201 {object} models.Wedding
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/weddings [post]
func (h *WeddingHandler) CreateWedding(c *gin.Context) {
	var wedding models.Wedding
	if err := c.ShouldBindJSON(&wedding); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body: " + err.Error()})
		return
	}

	// Validate wedding data
	if err := utils.ValidateStruct(&wedding); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	userID := c.GetString("userID")
	userOID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid user ID"})
		return
	}

	if err := h.weddingService.CreateWedding(c.Request.Context(), &wedding, userOID); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, wedding)
}

// GetWedding godoc
// @Summary Get a wedding by ID
// @Description Get a wedding by ID (accessible if public or owned)
// @Tags weddings
// @Produce json
// @Param id path string true "Wedding ID"
// @Success 200 {object} models.Wedding
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/weddings/{id} [get]
func (h *WeddingHandler) GetWedding(c *gin.Context) {
	weddingID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid wedding ID"})
		return
	}

	userID := c.GetString("userID")
	var userOID primitive.ObjectID
	if userID != "" {
		userOID, err = primitive.ObjectIDFromHex(userID)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid user ID"})
			return
		}
	} else {
		userOID = primitive.NilObjectID
	}

	wedding, err := h.weddingService.GetWeddingByID(c.Request.Context(), weddingID, userOID)
	if err != nil {
		if err.Error() == "wedding not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Wedding not found"})
			return
		}
		if err.Error() == "access denied" {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: "Access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, wedding)
}

// GetWeddingBySlug godoc
// @Summary Get a wedding by slug
// @Description Get a wedding by slug (accessible if public or owned)
// @Tags weddings
// @Produce json
// @Param slug path string true "Wedding slug"
// @Success 200 {object} models.Wedding
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/weddings/slug/{slug} [get]
func (h *WeddingHandler) GetWeddingBySlug(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Slug is required"})
		return
	}

	userID := c.GetString("userID")
	var userOID primitive.ObjectID
	if userID != "" {
		var err error
		userOID, err = primitive.ObjectIDFromHex(userID)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid user ID"})
			return
		}
	}

	wedding, err := h.weddingService.GetWeddingBySlug(c.Request.Context(), slug, userOID)
	if err != nil {
		if err.Error() == "wedding not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Wedding not found"})
			return
		}
		if err.Error() == "access denied" {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: "Access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, wedding)
}

// GetUserWeddings godoc
// @Summary Get user's weddings
// @Description Get all weddings for the authenticated user with pagination
// @Tags weddings
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 20)"
// @Param status query string false "Filter by status"
// @Param search query string false "Search term"
// @Success 200 {object} PaginatedWeddingsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/weddings [get]
func (h *WeddingHandler) GetUserWeddings(c *gin.Context) {
	userID := c.GetString("userID")
	userOID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid user ID"})
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// Build filters
	filters := repository.WeddingFilters{
		Status: c.Query("status"),
		Search: c.Query("search"),
	}

	// Parse date filters
	if createdAfter := c.Query("created_after"); createdAfter != "" {
		if t, err := time.Parse(time.RFC3339, createdAfter); err == nil {
			filters.CreatedAfter = &t
		}
	}

	if createdBefore := c.Query("created_before"); createdBefore != "" {
		if t, err := time.Parse(time.RFC3339, createdBefore); err == nil {
			filters.CreatedBefore = &t
		}
	}

	if eventDate := c.Query("event_date"); eventDate != "" {
		if t, err := time.Parse(time.RFC3339, eventDate); err == nil {
			filters.EventDate = &t
		}
	}

	weddings, total, err := h.weddingService.GetUserWeddings(c.Request.Context(), userOID, page, pageSize, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, PaginatedWeddingsResponse{
		Weddings: weddings,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// UpdateWedding godoc
// @Summary Update a wedding
// @Description Update an existing wedding (only owner can update)
// @Tags weddings
// @Accept json
// @Produce json
// @Param id path string true "Wedding ID"
// @Param wedding body models.Wedding true "Wedding data"
// @Success 200 {object} models.Wedding
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/weddings/{id} [put]
func (h *WeddingHandler) UpdateWedding(c *gin.Context) {
	weddingID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid wedding ID"})
		return
	}

	var wedding models.Wedding
	if err := c.ShouldBindJSON(&wedding); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body: " + err.Error()})
		return
	}

	// Set ID from URL parameter
	wedding.ID = weddingID

	// Validate wedding data
	if err := utils.ValidateStruct(&wedding); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	userID := c.GetString("userID")
	userOID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid user ID"})
		return
	}

	if err := h.weddingService.UpdateWedding(c.Request.Context(), &wedding, userOID); err != nil {
		if err.Error() == "wedding not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Wedding not found"})
			return
		}
		if err.Error() == "access denied" {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: "Access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, wedding)
}

// DeleteWedding godoc
// @Summary Delete a wedding
// @Description Delete a wedding (only owner can delete)
// @Tags weddings
// @Produce json
// @Param id path string true "Wedding ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/weddings/{id} [delete]
func (h *WeddingHandler) DeleteWedding(c *gin.Context) {
	weddingID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid wedding ID"})
		return
	}

	userID := c.GetString("userID")
	userOID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid user ID"})
		return
	}

	if err := h.weddingService.DeleteWedding(c.Request.Context(), weddingID, userOID); err != nil {
		if err.Error() == "wedding not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Wedding not found"})
			return
		}
		if err.Error() == "access denied" {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: "Access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Message: "Wedding deleted successfully"})
}

// PublishWedding godoc
// @Summary Publish a wedding
// @Description Publish a wedding to make it public (only owner can publish)
// @Tags weddings
// @Produce json
// @Param id path string true "Wedding ID"
// @Success 200 {object} models.Wedding
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/weddings/{id}/publish [post]
func (h *WeddingHandler) PublishWedding(c *gin.Context) {
	weddingID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid wedding ID"})
		return
	}

	userID := c.GetString("userID")
	userOID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid user ID"})
		return
	}

	// Get the wedding first to validate
	wedding, err := h.weddingService.GetWeddingByID(c.Request.Context(), weddingID, userOID)
	if err != nil {
		if err.Error() == "wedding not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Wedding not found"})
			return
		}
		if err.Error() == "access denied" {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: "Access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	// Check ownership
	if wedding.UserID != userOID {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "Access denied"})
		return
	}

	if err := h.weddingService.PublishWedding(c.Request.Context(), weddingID, userOID); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	// Get updated wedding
	updatedWedding, err := h.weddingService.GetWeddingByID(c.Request.Context(), weddingID, userOID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedWedding)
}

// ListPublicWeddings godoc
// @Summary List public weddings
// @Description Get a list of public weddings with pagination
// @Tags weddings
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 20)"
// @Param search query string false "Search term"
// @Param event_date query string false "Filter by event date (RFC3339 format)"
// @Success 200 {object} PaginatedWeddingsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/public/weddings [get]
func (h *WeddingHandler) ListPublicWeddings(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// Build filters
	filters := repository.PublicWeddingFilters{
		Search: c.Query("search"),
	}

	// Parse event date filter
	if eventDate := c.Query("event_date"); eventDate != "" {
		if t, err := time.Parse(time.RFC3339, eventDate); err == nil {
			filters.EventDate = &t
		}
	}

	weddings, total, err := h.weddingService.ListPublicWeddings(c.Request.Context(), page, pageSize, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, PaginatedWeddingsResponse{
		Weddings: weddings,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// Response types

type PaginatedWeddingsResponse struct {
	Weddings []*models.Wedding `json:"weddings"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
}
