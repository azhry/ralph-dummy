package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"wedding-invitation-backend/internal/domain/repository"
	"wedding-invitation-backend/internal/domain/models"
	"wedding-invitation-backend/internal/services"
	"wedding-invitation-backend/internal/utils"
)

// GuestHandler handles guest-related HTTP requests
type GuestHandler struct {
	guestService *services.GuestService
}

// NewGuestHandler creates a new guest handler
func NewGuestHandler(guestService *services.GuestService) *GuestHandler {
	return &GuestHandler{
		guestService: guestService,
	}
}

// CreateGuestRequest represents the request to create a guest
type CreateGuestRequest struct {
	FirstName       string             `json:"first_name" binding:"required,max=50"`
	LastName        string             `json:"last_name" binding:"required,max=50"`
	Email           string             `json:"email,omitempty" binding:"omitempty,email,max=100"`
	Phone           string             `json:"phone,omitempty"`
	Relationship    string             `json:"relationship,omitempty"`
	Side            string             `json:"side,omitempty" binding:"omitempty,oneof=bride groom both"`
	InvitedVia      string             `json:"invited_via,omitempty" binding:"omitempty,oneof=digital manual"`
	InvitationStatus string            `json:"invitation_status,omitempty" binding:"omitempty,oneof=pending sent delivered failed"`
	AllowPlusOne    bool               `json:"allow_plus_one"`
	MaxPlusOnes     int                `json:"max_plus_ones" binding:"omitempty,min=0,max=5"`
	VIP             bool               `json:"vip"`
	Notes           string             `json:"notes,omitempty"`
	Address         *models.Address    `json:"address,omitempty"`
}

// UpdateGuestRequest represents the request to update a guest
type UpdateGuestRequest struct {
	FirstName       string             `json:"first_name" binding:"omitempty,max=50"`
	LastName        string             `json:"last_name" binding:"omitempty,max=50"`
	Email           string             `json:"email,omitempty" binding:"omitempty,email,max=100"`
	Phone           string             `json:"phone,omitempty"`
	Address         *models.Address    `json:"address,omitempty"`
	Relationship    string             `json:"relationship,omitempty"`
	Side            string             `json:"side,omitempty" binding:"omitempty,oneof=bride groom both"`
	InvitedVia      string             `json:"invited_via,omitempty" binding:"omitempty,oneof=digital manual"`
	InvitationStatus string            `json:"invitation_status,omitempty" binding:"omitempty,oneof=pending sent delivered failed"`
	AllowPlusOne    bool               `json:"allow_plus_one"`
	MaxPlusOnes     int                `json:"max_plus_ones" binding:"omitempty,min=0,max=5"`
	RSVPStatus      string             `json:"rsvp_status,omitempty" binding:"omitempty,oneof=attending not-attending maybe pending"`
	DietaryNotes    string             `json:"dietary_notes,omitempty"`
	VIP             bool               `json:"vip"`
	Notes           string             `json:"notes,omitempty"`
}

// BulkCreateGuestsRequest represents the request to create multiple guests
type BulkCreateGuestsRequest struct {
	Guests []CreateGuestRequest `json:"guests" binding:"required,min=1,max=100"`
}

// GuestResponse represents the guest response
type GuestResponse struct {
	ID               primitive.ObjectID `json:"id"`
	FirstName        string             `json:"first_name"`
	LastName         string             `json:"last_name"`
	Email            string             `json:"email,omitempty"`
	Phone            string             `json:"phone,omitempty"`
	Address          *models.Address    `json:"address,omitempty"`
	Relationship     string             `json:"relationship,omitempty"`
	Side             string             `json:"side,omitempty"`
	InvitedVia       string             `json:"invited_via"`
	InvitationStatus string             `json:"invitation_status"`
	AllowPlusOne     bool               `json:"allow_plus_one"`
	MaxPlusOnes      int                `json:"max_plus_ones"`
	RSVPStatus       string             `json:"rsvp_status,omitempty"`
	RSVPID           *primitive.ObjectID `json:"rsvp_id,omitempty"`
	DietaryNotes     string             `json:"dietary_notes,omitempty"`
	VIP              bool               `json:"vip"`
	Notes            string             `json:"notes,omitempty"`
	ImportBatchID    string             `json:"import_batch_id,omitempty"`
	CreatedAt        string             `json:"created_at"`
	UpdatedAt        string             `json:"updated_at"`
}

// CreateGuest creates a new guest
// @Summary Create a guest
// @Description Create a new guest for a wedding
// @Tags Guests
// @Param wedding_id path string true "Wedding ID"
// @Param request body CreateGuestRequest true "Guest data"
// @Success 201 {object} utils.Response{data=GuestResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 403 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Router /weddings/{wedding_id}/guests [post]
func (h *GuestHandler) CreateGuest(c *gin.Context) {
	weddingID, err := primitive.ObjectIDFromHex(c.Param("wedding_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "Invalid wedding ID"})
		return
	}

	userID := utils.GetUserIDFromContext(c)
	if userID == nil {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{Error: "User not authenticated"})
		return
	}

	var req CreateGuestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "Invalid request data: " + err.Error()})
		return
	}

	// Create guest model
	guest := &models.Guest{
		FirstName:        req.FirstName,
		LastName:         req.LastName,
		Email:            req.Email,
		Phone:            req.Phone,
		Address:          req.Address,
		Relationship:     req.Relationship,
		Side:             req.Side,
		InvitedVia:       req.InvitedVia,
		InvitationStatus: req.InvitationStatus,
		AllowPlusOne:     req.AllowPlusOne,
		MaxPlusOnes:      req.MaxPlusOnes,
		VIP:              req.VIP,
		Notes:            req.Notes,
	}

	// Set defaults
	if guest.InvitedVia == "" {
		guest.InvitedVia = "digital"
	}
	if guest.InvitationStatus == "" {
		guest.InvitationStatus = "pending"
	}

	if err := h.guestService.CreateGuest(c.Request.Context(), weddingID, *userID, guest); err != nil {
		if errors.Is(err, services.ErrWeddingNotFound) {
			c.JSON(http.StatusNotFound, utils.ErrorResponse{Error: "Wedding not found"})
			return
		}
		if errors.Is(err, services.ErrUnauthorized) {
			c.JSON(http.StatusForbidden, utils.ErrorResponse{Error: "You don't own this wedding"})
			return
		}
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "Failed to create guest"})
		return
	}

	c.JSON(http.StatusCreated, utils.Response{
		Message: "Guest created successfully",
		Data:    h.convertToGuestResponse(guest),
	})
}

// GetGuest retrieves a guest by ID
// @Summary Get a guest
// @Description Retrieve a guest by ID
// @Tags Guests
// @Param id path string true "Guest ID"
// @Success 200 {object} utils.Response{data=GuestResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 403 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Router /guests/{id} [get]
func (h *GuestHandler) GetGuest(c *gin.Context) {
	guestID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "Invalid guest ID"})
		return
	}

	userID := utils.GetUserIDFromContext(c)
	if userID == nil {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{Error: "User not authenticated"})
		return
	}

	guest, err := h.guestService.GetGuestByID(c.Request.Context(), guestID, *userID)
	if err != nil {
		if errors.Is(err, services.ErrGuestNotFound) {
			c.JSON(http.StatusNotFound, utils.ErrorResponse{Error: "Guest not found"})
			return
		}
		if errors.Is(err, services.ErrUnauthorized) {
			c.JSON(http.StatusForbidden, utils.ErrorResponse{Error: "You don't have access to this guest"})
			return
		}
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "Failed to retrieve guest"})
		return
	}

	c.JSON(http.StatusOK, utils.Response{
		Data: h.convertToGuestResponse(guest),
	})
}

// ListGuests retrieves guests for a wedding
// @Summary List guests
// @Description List guests for a wedding with pagination and filtering
// @Tags Guests
// @Param wedding_id path string true "Wedding ID"
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 20, max: 100)"
// @Param search query string false "Search term"
// @Param side query string false "Filter by side (bride, groom, both)"
// @Param rsvp_status query string false "Filter by RSVP status"
// @Param relationship query string false "Filter by relationship"
// @Param vip query bool false "Filter by VIP status"
// @Param invitation_status query string false "Filter by invitation status"
// @Param invited_via query string false "Filter by invited via"
// @Param allow_plus_one query bool false "Filter by plus one allowed"
// @Success 200 {object} utils.Response{data=utils.PaginatedResponse{data=[]GuestResponse}}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 403 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Router /weddings/{wedding_id}/guests [get]
func (h *GuestHandler) ListGuests(c *gin.Context) {
	weddingID, err := primitive.ObjectIDFromHex(c.Param("wedding_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "Invalid wedding ID"})
		return
	}

	userID := utils.GetUserIDFromContext(c)
	if userID == nil {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{Error: "User not authenticated"})
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// Build filters
	filters := repository.GuestFilters{
		Search:            c.Query("search"),
		Side:              c.Query("side"),
		RSVPStatus:        c.Query("rsvp_status"),
		Relationship:      c.Query("relationship"),
		InvitationStatus:  c.Query("invitation_status"),
		InvitedVia:        c.Query("invited_via"),
	}

	// Parse boolean filters
	if vipStr := c.Query("vip"); vipStr != "" {
		vip := vipStr == "true"
		filters.VIP = &vip
	}

	if allowPlusOneStr := c.Query("allow_plus_one"); allowPlusOneStr != "" {
		allowPlusOne := allowPlusOneStr == "true"
		filters.AllowPlusOne = &allowPlusOne
	}

	guests, total, err := h.guestService.ListGuests(c.Request.Context(), weddingID, *userID, page, pageSize, filters)
	if err != nil {
		if errors.Is(err, services.ErrWeddingNotFound) {
			c.JSON(http.StatusNotFound, utils.ErrorResponse{Error: "Wedding not found"})
			return
		}
		if errors.Is(err, services.ErrUnauthorized) {
			c.JSON(http.StatusForbidden, utils.ErrorResponse{Error: "You don't own this wedding"})
			return
		}
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "Failed to retrieve guests"})
		return
	}

	// Convert to response format
	var guestResponses []GuestResponse
	for _, guest := range guests {
		guestResponses = append(guestResponses, *h.convertToGuestResponse(guest))
	}

	c.JSON(http.StatusOK, utils.Response{
		Data: utils.PaginatedResponse{
			Data:       guestResponses,
			Total:      total,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// UpdateGuest updates an existing guest
// @Summary Update a guest
// @Description Update an existing guest
// @Tags Guests
// @Param id path string true "Guest ID"
// @Param request body UpdateGuestRequest true "Guest data"
// @Success 200 {object} utils.Response{data=GuestResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 403 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Router /guests/{id} [put]
func (h *GuestHandler) UpdateGuest(c *gin.Context) {
	guestID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "Invalid guest ID"})
		return
	}

	userID := utils.GetUserIDFromContext(c)
	if userID == nil {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{Error: "User not authenticated"})
		return
	}

	var req UpdateGuestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "Invalid request data: " + err.Error()})
		return
	}

	// Create guest model
	guest := &models.Guest{
		ID:               guestID,
		FirstName:        req.FirstName,
		LastName:         req.LastName,
		Email:            req.Email,
		Phone:            req.Phone,
		Address:          req.Address,
		Relationship:     req.Relationship,
		Side:             req.Side,
		InvitedVia:       req.InvitedVia,
		InvitationStatus: req.InvitationStatus,
		AllowPlusOne:     req.AllowPlusOne,
		MaxPlusOnes:      req.MaxPlusOnes,
		RSVPStatus:       req.RSVPStatus,
		DietaryNotes:     req.DietaryNotes,
		VIP:              req.VIP,
		Notes:            req.Notes,
	}

	if err := h.guestService.UpdateGuest(c.Request.Context(), guestID, *userID, guest); err != nil {
		if errors.Is(err, services.ErrGuestNotFound) {
			c.JSON(http.StatusNotFound, utils.ErrorResponse{Error: "Guest not found"})
			return
		}
		if errors.Is(err, services.ErrUnauthorized) {
			c.JSON(http.StatusForbidden, utils.ErrorResponse{Error: "You don't have access to this guest"})
			return
		}
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "Failed to update guest"})
		return
	}

	c.JSON(http.StatusOK, utils.Response{
		Message: "Guest updated successfully",
		Data:    h.convertToGuestResponse(guest),
	})
}

// DeleteGuest deletes a guest
// @Summary Delete a guest
// @Description Delete a guest
// @Tags Guests
// @Param id path string true "Guest ID"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 403 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Router /guests/{id} [delete]
func (h *GuestHandler) DeleteGuest(c *gin.Context) {
	guestID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "Invalid guest ID"})
		return
	}

	userID := utils.GetUserIDFromContext(c)
	if userID == nil {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{Error: "User not authenticated"})
		return
	}

	if err := h.guestService.DeleteGuest(c.Request.Context(), guestID, *userID); err != nil {
		if errors.Is(err, services.ErrGuestNotFound) {
			c.JSON(http.StatusNotFound, utils.ErrorResponse{Error: "Guest not found"})
			return
		}
		if errors.Is(err, services.ErrUnauthorized) {
			c.JSON(http.StatusForbidden, utils.ErrorResponse{Error: "You don't have access to this guest"})
			return
		}
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "Failed to delete guest"})
		return
	}

	c.JSON(http.StatusOK, utils.Response{
		Message: "Guest deleted successfully",
	})
}

// BulkCreateGuests creates multiple guests at once
// @Summary Create multiple guests
// @Description Create multiple guests for a wedding at once
// @Tags Guests
// @Param wedding_id path string true "Wedding ID"
// @Param request body BulkCreateGuestsRequest true "Guests data"
// @Success 201 {object} utils.Response
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 403 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Router /weddings/{wedding_id}/guests/bulk [post]
func (h *GuestHandler) BulkCreateGuests(c *gin.Context) {
	weddingID, err := primitive.ObjectIDFromHex(c.Param("wedding_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "Invalid wedding ID"})
		return
	}

	userID := utils.GetUserIDFromContext(c)
	if userID == nil {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{Error: "User not authenticated"})
		return
	}

	var req BulkCreateGuestsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "Invalid request data: " + err.Error()})
		return
	}

	// Convert requests to guest models
	var guests []*models.Guest
	for _, guestReq := range req.Guests {
		guest := &models.Guest{
			FirstName:        guestReq.FirstName,
			LastName:         guestReq.LastName,
			Email:            guestReq.Email,
			Phone:            guestReq.Phone,
			Address:          guestReq.Address,
			Relationship:     guestReq.Relationship,
			Side:             guestReq.Side,
			InvitedVia:       guestReq.InvitedVia,
			InvitationStatus: guestReq.InvitationStatus,
			AllowPlusOne:     guestReq.AllowPlusOne,
			MaxPlusOnes:      guestReq.MaxPlusOnes,
			VIP:              guestReq.VIP,
			Notes:            guestReq.Notes,
		}

		// Set defaults
		if guest.InvitedVia == "" {
			guest.InvitedVia = "digital"
		}
		if guest.InvitationStatus == "" {
			guest.InvitationStatus = "pending"
		}

		guests = append(guests, guest)
	}

	if err := h.guestService.CreateManyGuests(c.Request.Context(), weddingID, *userID, guests); err != nil {
		if errors.Is(err, services.ErrWeddingNotFound) {
			c.JSON(http.StatusNotFound, utils.ErrorResponse{Error: "Wedding not found"})
			return
		}
		if errors.Is(err, services.ErrUnauthorized) {
			c.JSON(http.StatusForbidden, utils.ErrorResponse{Error: "You don't own this wedding"})
			return
		}
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "Failed to create guests"})
		return
	}

	c.JSON(http.StatusCreated, utils.Response{
		Message: "Guests created successfully",
		Data: map[string]interface{}{
			"created_count": len(guests),
		},
	})
}

// ImportGuestsCSV imports guests from a CSV file
// @Summary Import guests from CSV
// @Description Import guests from a CSV file for a wedding
// @Tags Guests
// @Param wedding_id path string true "Wedding ID"
// @Param file formData file true "CSV file"
// @Success 200 {object} utils.Response{data=models.GuestImportResult}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 403 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Router /weddings/{wedding_id}/guests/import [post]
func (h *GuestHandler) ImportGuestsCSV(c *gin.Context) {
	weddingID, err := primitive.ObjectIDFromHex(c.Param("wedding_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "Invalid wedding ID"})
		return
	}

	userID := utils.GetUserIDFromContext(c)
	if userID == nil {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{Error: "User not authenticated"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "No file uploaded or invalid file: " + err.Error()})
		return
	}
	defer file.Close()

	// Check file extension
	if !strings.HasSuffix(header.Filename, ".csv") {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "Only CSV files are allowed"})
		return
	}

	result, err := h.guestService.ImportGuestsFromCSV(c.Request.Context(), weddingID, *userID, file)
	if err != nil {
		if errors.Is(err, services.ErrWeddingNotFound) {
			c.JSON(http.StatusNotFound, utils.ErrorResponse{Error: "Wedding not found"})
			return
		}
		if errors.Is(err, services.ErrUnauthorized) {
			c.JSON(http.StatusForbidden, utils.ErrorResponse{Error: "You don't own this wedding"})
			return
		}
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "Failed to import guests: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, utils.Response{
		Message: "Guests imported successfully",
		Data:    result,
	})
}

// convertToGuestResponse converts a guest model to response format
func (h *GuestHandler) convertToGuestResponse(guest *models.Guest) *GuestResponse {
	return &GuestResponse{
		ID:               guest.ID,
		FirstName:        guest.FirstName,
		LastName:         guest.LastName,
		Email:            guest.Email,
		Phone:            guest.Phone,
		Address:          guest.Address,
		Relationship:     guest.Relationship,
		Side:             guest.Side,
		InvitedVia:       guest.InvitedVia,
		InvitationStatus: guest.InvitationStatus,
		AllowPlusOne:     guest.AllowPlusOne,
		MaxPlusOnes:      guest.MaxPlusOnes,
		RSVPStatus:       guest.RSVPStatus,
		RSVPID:           guest.RSVPID,
		DietaryNotes:     guest.DietaryNotes,
		VIP:              guest.VIP,
		Notes:            guest.Notes,
		ImportBatchID:    guest.ImportBatchID,
		CreatedAt:        guest.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:        guest.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}