package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"wedding-invitation-backend/internal/domain/models"
	"wedding-invitation-backend/internal/domain/repository"
	"wedding-invitation-backend/internal/services"
	"wedding-invitation-backend/internal/utils"
)

// GuestHandler handles guest-related HTTP requests
type GuestHandler struct {
	guestService services.GuestServiceInterface
}

// NewGuestHandler creates a new guest handler
func NewGuestHandler(guestService services.GuestServiceInterface) *GuestHandler {
	return &GuestHandler{
		guestService: guestService,
	}
}

// CreateGuestRequest represents a request to create a guest
type CreateGuestRequest struct {
	FirstName        string          `json:"first_name" validate:"required"`
	LastName         string          `json:"last_name" validate:"required"`
	Email            string          `json:"email,omitempty" validate:"omitempty,email"`
	Phone            string          `json:"phone,omitempty"`
	Address          *models.Address `json:"address,omitempty"`
	Relationship     string          `json:"relationship,omitempty"`
	Side             string          `json:"side,omitempty" validate:"oneof=bride groom both"`
	InvitedVia       string          `json:"invited_via,omitempty" validate:"oneof=digital manual"`
	InvitationStatus string          `json:"invitation_status,omitempty" validate:"oneof=pending sent delivered failed"`
	AllowPlusOne     bool            `json:"allow_plus_one,omitempty"`
	MaxPlusOnes      int             `json:"max_plus_ones,omitempty" validate:"min=0,max=5"`
	RSVPStatus       string          `json:"rsvp_status,omitempty" validate:"omitempty,oneof=attending not-attending maybe pending"`
	DietaryNotes     string          `json:"dietary_notes,omitempty"`
	VIP              bool            `json:"vip,omitempty"`
	Notes            string          `json:"notes,omitempty"`
}

// BulkCreateGuestsRequest represents a request to create multiple guests
type BulkCreateGuestsRequest struct {
	Guests []CreateGuestRequest `json:"guests" validate:"required,min=1,max=100"`
}

// UpdateGuestRequest represents a request to update a guest
type UpdateGuestRequest struct {
	FirstName        *string         `json:"first_name,omitempty"`
	LastName         *string         `json:"last_name,omitempty"`
	Email            *string         `json:"email,omitempty" validate:"omitempty,email"`
	Phone            *string         `json:"phone,omitempty"`
	Address          *models.Address `json:"address,omitempty"`
	Relationship     *string         `json:"relationship,omitempty"`
	Side             *string         `json:"side,omitempty" validate:"omitempty,oneof=bride groom both"`
	InvitedVia       *string         `json:"invited_via,omitempty" validate:"omitempty,oneof=digital manual"`
	InvitationStatus *string         `json:"invitation_status,omitempty" validate:"omitempty,oneof=pending sent delivered failed"`
	AllowPlusOne     *bool           `json:"allow_plus_one,omitempty"`
	MaxPlusOnes      *int            `json:"max_plus_ones,omitempty" validate:"omitempty,min=0,max=5"`
	RSVPStatus       *string         `json:"rsvp_status,omitempty" validate:"omitempty,oneof=attending not-attending maybe pending"`
	DietaryNotes     *string         `json:"dietary_notes,omitempty"`
	VIP              *bool           `json:"vip,omitempty"`
	Notes            *string         `json:"notes,omitempty"`
}

// GuestResponse represents a guest response
type GuestResponse struct {
	ID               primitive.ObjectID  `json:"id"`
	WeddingID        primitive.ObjectID  `json:"wedding_id"`
	FirstName        string              `json:"first_name"`
	LastName         string              `json:"last_name"`
	Email            string              `json:"email,omitempty"`
	Phone            string              `json:"phone,omitempty"`
	Address          *models.Address     `json:"address,omitempty"`
	Relationship     string              `json:"relationship,omitempty"`
	Side             string              `json:"side,omitempty"`
	InvitedVia       string              `json:"invited_via"`
	InvitationStatus string              `json:"invitation_status"`
	AllowPlusOne     bool                `json:"allow_plus_one"`
	MaxPlusOnes      int                 `json:"max_plus_ones"`
	RSVPStatus       string              `json:"rsvp_status,omitempty"`
	RSVPID           *primitive.ObjectID `json:"rsvp_id,omitempty"`
	DietaryNotes     string              `json:"dietary_notes,omitempty"`
	VIP              bool                `json:"vip"`
	Notes            string              `json:"notes,omitempty"`
	ImportBatchID    string              `json:"import_batch_id,omitempty"`
	CreatedBy        primitive.ObjectID  `json:"created_by"`
	CreatedAt        primitive.DateTime  `json:"created_at"`
	UpdatedAt        primitive.DateTime  `json:"updated_at"`
}

// GuestListResponse represents a list of guests with pagination
type GuestListResponse struct {
	Guests []GuestResponse `json:"guests"`
	Total  int64           `json:"total"`
	Page   int             `json:"page"`
	Size   int             `json:"size"`
}

// ImportResult represents the result of a CSV import
type ImportResult struct {
	SuccessCount int      `json:"success_count"`
	ErrorCount   int      `json:"error_count"`
	Errors       []string `json:"errors,omitempty"`
	BatchID      string   `json:"batch_id"`
}

// CreateGuest creates a new guest
func (h *GuestHandler) CreateGuest(c *gin.Context) {
	weddingID, err := primitive.ObjectIDFromHex(c.Param("wedding_id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid wedding ID")
		return
	}

	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req CreateGuestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request data: "+err.Error())
		return
	}

	// Validate request
	if err := utils.ValidateStruct(req); err != nil {
		if validationErrors, ok := err.(*utils.ValidationError); ok {
			errors := make(map[string]string)
			for i, errorMsg := range validationErrors.Errors {
				errors["field_"+strconv.Itoa(i+1)] = errorMsg
			}
			utils.ValidationErrorResponse(c, errors)
			return
		}
		utils.ErrorResponse(c, http.StatusBadRequest, "Validation failed")
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
		RSVPStatus:       req.RSVPStatus,
		DietaryNotes:     req.DietaryNotes,
		VIP:              req.VIP,
		Notes:            req.Notes,
	}

	// Set default status
	if guest.InvitationStatus == "" {
		guest.InvitationStatus = "pending"
	}
	if guest.InvitedVia == "" {
		guest.InvitedVia = "digital"
	}
	if guest.RSVPStatus == "" {
		guest.RSVPStatus = "pending"
	}

	if err := h.guestService.CreateGuest(c.Request.Context(), weddingID, userID, guest); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			utils.ErrorResponse(c, http.StatusNotFound, "Wedding not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create guest")
		return
	}

	c.JSON(http.StatusCreated, utils.APIResponse{
		Success: true,
		Message: "Guest created successfully",
		Data:    h.convertToGuestResponse(guest),
	})
}

// GetGuest retrieves a single guest by ID
func (h *GuestHandler) GetGuest(c *gin.Context) {
	guestID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid guest ID")
		return
	}

	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	guest, err := h.guestService.GetGuestByID(c.Request.Context(), guestID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			utils.ErrorResponse(c, http.StatusNotFound, "Guest not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get guest")
		return
	}

	utils.Response(c, http.StatusOK, h.convertToGuestResponse(guest))
}

// ListGuests retrieves guests for a wedding
func (h *GuestHandler) ListGuests(c *gin.Context) {
	weddingID, err := primitive.ObjectIDFromHex(c.Param("wedding_id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid wedding ID")
		return
	}

	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Parse pagination
	page, size := utils.ParsePaginationParams(c)

	// Parse filters
	filters := repository.GuestFilters{
		Search:           c.Query("search"),
		Side:             c.Query("side"),
		RSVPStatus:       c.Query("rsvp_status"),
		InvitationStatus: c.Query("invitation_status"),
	}

	guests, total, err := h.guestService.ListGuests(c.Request.Context(), weddingID, userID, page, size, filters)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			utils.ErrorResponse(c, http.StatusNotFound, "Wedding not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to list guests")
		return
	}

	// Convert to response format
	guestResponses := make([]GuestResponse, len(guests))
	for i, guest := range guests {
		guestResponses[i] = *h.convertToGuestResponse(guest)
	}

	utils.PaginatedResponse(c, http.StatusOK, guestResponses, int64(len(guestResponses)), total, page, size)
}

// UpdateGuest updates an existing guest
func (h *GuestHandler) UpdateGuest(c *gin.Context) {
	guestID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid guest ID")
		return
	}

	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req UpdateGuestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request data: "+err.Error())
		return
	}

	// Validate request
	if err := utils.ValidateStruct(req); err != nil {
		if validationErrors, ok := err.(*utils.ValidationError); ok {
			errors := make(map[string]string)
			for i, errorMsg := range validationErrors.Errors {
				errors["field_"+strconv.Itoa(i+1)] = errorMsg
			}
			utils.ValidationErrorResponse(c, errors)
			return
		}
		utils.ErrorResponse(c, http.StatusBadRequest, "Validation failed")
		return
	}

	// Get existing guest
	guest, err := h.guestService.GetGuestByID(c.Request.Context(), guestID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			utils.ErrorResponse(c, http.StatusNotFound, "Guest not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get guest")
		return
	}

	// Update fields if provided
	if req.FirstName != nil {
		guest.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		guest.LastName = *req.LastName
	}
	if req.Email != nil {
		guest.Email = *req.Email
	}
	if req.Phone != nil {
		guest.Phone = *req.Phone
	}
	if req.Address != nil {
		guest.Address = req.Address
	}
	if req.Relationship != nil {
		guest.Relationship = *req.Relationship
	}
	if req.Side != nil {
		guest.Side = *req.Side
	}
	if req.InvitedVia != nil {
		guest.InvitedVia = *req.InvitedVia
	}
	if req.InvitationStatus != nil {
		guest.InvitationStatus = *req.InvitationStatus
	}
	if req.AllowPlusOne != nil {
		guest.AllowPlusOne = *req.AllowPlusOne
	}
	if req.MaxPlusOnes != nil {
		guest.MaxPlusOnes = *req.MaxPlusOnes
	}
	if req.RSVPStatus != nil {
		guest.RSVPStatus = *req.RSVPStatus
	}
	if req.DietaryNotes != nil {
		guest.DietaryNotes = *req.DietaryNotes
	}
	if req.VIP != nil {
		guest.VIP = *req.VIP
	}
	if req.Notes != nil {
		guest.Notes = *req.Notes
	}

	if err := h.guestService.UpdateGuest(c.Request.Context(), guestID, userID, guest); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update guest")
		return
	}

	utils.Response(c, http.StatusOK, h.convertToGuestResponse(guest))
}

// DeleteGuest deletes a guest
func (h *GuestHandler) DeleteGuest(c *gin.Context) {
	guestID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid guest ID")
		return
	}

	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	if err := h.guestService.DeleteGuest(c.Request.Context(), guestID, userID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			utils.ErrorResponse(c, http.StatusNotFound, "Guest not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete guest")
		return
	}

	utils.SuccessResponse(c, "Guest deleted successfully")
}

// BulkCreateGuests creates multiple guests at once
func (h *GuestHandler) BulkCreateGuests(c *gin.Context) {
	weddingID, err := primitive.ObjectIDFromHex(c.Param("wedding_id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid wedding ID")
		return
	}

	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req BulkCreateGuestsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request data: "+err.Error())
		return
	}

	// Validate request
	if len(req.Guests) == 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "At least one guest is required")
		return
	}
	if len(req.Guests) > 100 {
		utils.ErrorResponse(c, http.StatusBadRequest, "Maximum 100 guests allowed per bulk creation")
		return
	}

	// Convert to guest models
	guests := make([]*models.Guest, len(req.Guests))
	for i, guestReq := range req.Guests {
		guests[i] = &models.Guest{
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
			RSVPStatus:       guestReq.RSVPStatus,
			DietaryNotes:     guestReq.DietaryNotes,
			VIP:              guestReq.VIP,
			Notes:            guestReq.Notes,
		}

		// Set defaults
		if guests[i].InvitationStatus == "" {
			guests[i].InvitationStatus = "pending"
		}
		if guests[i].InvitedVia == "" {
			guests[i].InvitedVia = "digital"
		}
		if guests[i].RSVPStatus == "" {
			guests[i].RSVPStatus = "pending"
		}
	}

	err = h.guestService.CreateManyGuests(c.Request.Context(), weddingID, userID, guests)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			utils.ErrorResponse(c, http.StatusNotFound, "Wedding not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create guests")
		return
	}

	// Convert to response format
	guestResponses := make([]GuestResponse, len(guests))
	for i, guest := range guests {
		guestResponses[i] = *h.convertToGuestResponse(guest)
	}

	c.JSON(http.StatusCreated, utils.APIResponse{
		Success: true,
		Message: "Guests created successfully",
		Data: GuestListResponse{
			Guests: guestResponses,
			Total:  int64(len(guests)),
		},
	})
}

// ImportGuestsCSV imports guests from a CSV file
func (h *GuestHandler) ImportGuestsCSV(c *gin.Context) {
	weddingID, err := primitive.ObjectIDFromHex(c.Param("wedding_id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid wedding ID")
		return
	}

	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	file, _, err := c.Request.FormFile("file")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "No file provided")
		return
	}
	defer file.Close()

	result, err := h.guestService.ImportGuestsFromCSV(c.Request.Context(), weddingID, userID, file)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to import guests: "+err.Error())
		return
	}

	utils.Response(c, http.StatusOK, result)
}

// Helper methods

func (h *GuestHandler) convertToGuestResponse(guest *models.Guest) *GuestResponse {
	return &GuestResponse{
		ID:               guest.ID,
		WeddingID:        guest.WeddingID,
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
		CreatedBy:        guest.CreatedBy,
		CreatedAt:        primitive.NewDateTimeFromTime(guest.CreatedAt),
		UpdatedAt:        primitive.NewDateTimeFromTime(guest.UpdatedAt),
	}
}
