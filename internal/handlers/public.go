package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"wedding-invitation-backend/internal/domain/models"
	"wedding-invitation-backend/internal/services"
	"wedding-invitation-backend/internal/utils"
)

// PublicHandler handles public wedding operations
type PublicHandler struct {
	weddingService *services.WeddingService
	rsvpService    *services.RSVPService
}

// NewPublicHandler creates a new public handler
func NewPublicHandler(weddingService *services.WeddingService, rsvpService *services.RSVPService) *PublicHandler {
	return &PublicHandler{
		weddingService: weddingService,
		rsvpService:    rsvpService,
	}
}

// PublicWeddingResponse represents the public wedding view response
type PublicWeddingResponse struct {
	Slug            string                  `json:"slug"`
	Theme           string                  `json:"theme"`
	GroomName       string                  `json:"groom_name"`
	BrideName       string                  `json:"bride_name"`
	GroomRole       string                  `json:"groom_role"`
	BrideRole       string                  `json:"bride_role"`
	GroomBio        string                  `json:"groom_bio"`
	BrideBio        string                  `json:"bride_bio"`
	GroomPhotoURL   string                  `json:"groom_photo_url"`
	BridePhotoURL   string                  `json:"bride_photo_url"`
	LoveStory       string                  `json:"love_story"`
	WeddingDate     time.Time               `json:"wedding_date"`
	VenueName       string                  `json:"venue_name"`
	VenueAddress    string                  `json:"venue_address"`
	VenueMapURL     string                  `json:"venue_map_url"`
	ContactEmail    string                  `json:"contact_email"`
	SiteTitle       string                  `json:"site_title"`
	MetaDescription string                  `json:"meta_description"`
	Events          []models.Event          `json:"events"`
	GalleryImages   []string                `json:"gallery_images"`
	AllowPlusOne    bool                    `json:"allow_plus_one"`
	CollectDietary  bool                    `json:"collect_dietary"`
	CustomQuestions []models.CustomQuestion `json:"custom_questions"`
	RSVPDeadline    time.Time               `json:"rsvp_deadline"`
	RSVPStatus      string                  `json:"rsvp_status"`
}

// PublicRSVPRequest represents the public RSVP submission request
type PublicRSVPRequest struct {
	Name                string            `json:"name" binding:"required,min=1,max=100"`
	Email               string            `json:"email" binding:"email"`
	Phone               string            `json:"phone"`
	Attending           bool              `json:"attending" binding:"required"`
	NumberOfGuests      int               `json:"number_of_guests" binding:"required,min=1,max=10"`
	PlusOneName         string            `json:"plus_one_name"`
	DietaryRestrictions string            `json:"dietary_restrictions" binding:"max=500"`
	Message             string            `json:"message" binding:"max=1000"`
	CustomAnswers       map[string]string `json:"custom_answers"`
}

// PublicRSVPResponse represents the public RSVP submission response
type PublicRSVPResponse struct {
	ID               primitive.ObjectID `json:"id"`
	WeddingID        primitive.ObjectID `json:"wedding_id"`
	Name             string             `json:"name"`
	Email            string             `json:"email"`
	Attending        bool               `json:"attending"`
	NumberOfGuests   int                `json:"number_of_guests"`
	PlusOneName      string             `json:"plus_one_name"`
	SubmittedAt      time.Time          `json:"submitted_at"`
	ConfirmationSent bool               `json:"confirmation_sent"`
}

// GetWeddingBySlug retrieves a public wedding by slug
// @Summary Get wedding by slug (public)
// @Description View a public wedding invitation (no authentication required)
// @Tags Public
// @Param slug path string true "Wedding URL slug"
// @Success 200 {object} utils.Response{data=PublicWeddingResponse}
// @Failure 404 {object} utils.ErrorResponse
// @Failure 403 {object} utils.ErrorResponse
// @Router /public/weddings/{slug} [get]
func (h *PublicHandler) GetWeddingBySlug(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Slug is required"})
		return
	}

	// Get wedding by slug (public access - no user ID)
	wedding, err := h.weddingService.GetWeddingBySlugForPublic(c.Request.Context(), slug)
	if err != nil {
		if err.Error() == "wedding not found" || err.Error() == "wedding not published" {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Wedding not found or not yet published"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to retrieve wedding"})
		return
	}

	// Check if wedding is password protected
	if wedding.PasswordHash != "" {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "This wedding is password protected"})
		return
	}

	// Convert to public response
	response := h.convertToPublicResponse(wedding)

	c.JSON(http.StatusOK, response)
}

// SubmitRSVP submits an RSVP for a public wedding
// @Summary Submit RSVP for public wedding
// @Description Submit an RSVP for a public wedding (no authentication required)
// @Tags Public
// @Param slug path string true "Wedding URL slug"
// @Param request body PublicRSVPRequest true "RSVP data"
// @Success 201 {object} utils.Response{data=PublicRSVPResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 409 {object} utils.ErrorResponse
// @Router /public/weddings/{slug}/rsvp [post]
func (h *PublicHandler) SubmitRSVP(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Slug is required"})
		return
	}

	var req PublicRSVPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request data: " + err.Error()})
		return
	}

	// Validate RSVP request
	if err := h.validatePublicRSVPRequest(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Get wedding by slug to verify it exists and is published
	wedding, err := h.weddingService.GetWeddingBySlugForPublic(c.Request.Context(), slug)
	if err != nil {
		if err.Error() == "wedding not found" || err.Error() == "wedding not published" {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Wedding not found or not yet published"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to retrieve wedding"})
		return
	}

	// Check if wedding is password protected
	if wedding.PasswordHash != "" {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "This wedding is password protected"})
		return
	}

	// Create RSVP model
	rsvp := &models.RSVP{
		WeddingID:           wedding.ID,
		Name:                req.Name,
		Email:               req.Email,
		Phone:               req.Phone,
		Attending:           req.Attending,
		NumberOfGuests:      req.NumberOfGuests,
		PlusOneName:         req.PlusOneName,
		DietaryRestrictions: req.DietaryRestrictions,
		Message:             req.Message,
		CustomAnswers:       req.CustomAnswers,
		Source:              models.RSVPSourceWeb,
		IPAddress:           c.ClientIP(),
		UserAgent:           c.GetHeader("User-Agent"),
		SubmittedAt:         time.Now(),
	}

	// Submit RSVP
	if err := h.rsvpService.SubmitRSVP(c.Request.Context(), rsvp); err != nil {
		if err.Error() == "RSVP period is not open" || err.Error() == "RSVP deadline has passed" {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "RSVP period is not open"})
			return
		}
		if err.Error() == "email already exists" {
			c.JSON(http.StatusConflict, ErrorResponse{Error: "An RSVP with this email already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to submit RSVP"})
		return
	}

	// Convert to response
	response := &PublicRSVPResponse{
		ID:               rsvp.ID,
		WeddingID:        rsvp.WeddingID,
		Name:             rsvp.Name,
		Email:            rsvp.Email,
		Attending:        rsvp.Attending,
		NumberOfGuests:   rsvp.NumberOfGuests,
		PlusOneName:      rsvp.PlusOneName,
		SubmittedAt:      rsvp.SubmittedAt,
		ConfirmationSent: rsvp.ConfirmationSent,
	}

	c.JSON(http.StatusCreated, response)
}

// convertToPublicResponse converts a wedding model to public response
func (h *PublicHandler) convertToPublicResponse(wedding *models.Wedding) *PublicWeddingResponse {
	return &PublicWeddingResponse{
		Slug:            wedding.Slug,
		Theme:           wedding.Theme.ThemeID,
		GroomName:       wedding.Couple.GroomName,
		BrideName:       wedding.Couple.BrideName,
		GroomRole:       wedding.Couple.GroomRole,
		BrideRole:       wedding.Couple.BrideRole,
		GroomBio:        wedding.Couple.GroomBio,
		BrideBio:        wedding.Couple.BrideBio,
		GroomPhotoURL:   wedding.Couple.GroomPhotoURL,
		BridePhotoURL:   wedding.Couple.BridePhotoURL,
		LoveStory:       wedding.Couple.LoveStory,
		WeddingDate:     wedding.EventDetails.WeddingDate,
		VenueName:       wedding.EventDetails.VenueName,
		VenueAddress:    wedding.EventDetails.VenueAddress,
		VenueMapURL:     wedding.EventDetails.VenueMapURL,
		ContactEmail:    wedding.EventDetails.ContactEmail,
		SiteTitle:       wedding.SEO.SiteTitle,
		MetaDescription: wedding.SEO.MetaDescription,
		Events:          wedding.EventDetails.Events,
		GalleryImages:   wedding.Gallery.Images,
		AllowPlusOne:    wedding.RSVP.AllowPlusOne,
		CollectDietary:  wedding.RSVP.CollectDietary,
		CustomQuestions: wedding.RSVP.CustomQuestions,
		RSVPDeadline:    wedding.RSVP.Deadline,
		RSVPStatus:      h.getRSVPStatus(wedding),
	}
}

// getRSVPStatus determines the current RSVP status
func (h *PublicHandler) getRSVPStatus(wedding *models.Wedding) string {
	now := time.Now()

	// Check if wedding is published
	if wedding.Status != string(models.WeddingStatusPublished) {
		return "closed"
	}

	// Check if RSVP period is open
	if wedding.RSVP.OpenDate.After(now) {
		return "upcoming"
	}

	// Check if RSVP deadline has passed
	if wedding.RSVP.Deadline.Before(now) {
		return "closed"
	}

	return "open"
}

// validatePublicRSVPRequest validates the public RSVP request
func (h *PublicHandler) validatePublicRSVPRequest(req *PublicRSVPRequest) error {
	// Validate name
	if req.Name == "" {
		return errors.New("name is required")
	}
	if len(req.Name) > 100 {
		return errors.New("name must be 100 characters or less")
	}

	// Validate email if provided
	if req.Email != "" {
		if !isValidEmail(req.Email) {
			return errors.New("invalid email format")
		}
	}

	// Validate phone if provided
	if req.Phone != "" {
		if !isValidPhone(req.Phone) {
			return errors.New("invalid phone format")
		}
	}

	// Validate number of guests
	if req.NumberOfGuests < 1 || req.NumberOfGuests > 10 {
		return errors.New("number of guests must be between 1 and 10")
	}

	// Validate plus one name if more than 1 guest
	if req.NumberOfGuests > 1 && req.PlusOneName == "" {
		return errors.New("plus one name is required when bringing more than 1 guest")
	}

	// Validate dietary restrictions length
	if len(req.DietaryRestrictions) > 500 {
		return errors.New("dietary restrictions must be 500 characters or less")
	}

	// Validate message length
	if len(req.Message) > 1000 {
		return errors.New("message must be 1000 characters or less")
	}

	return nil
}

// Helper functions for validation
func isValidEmail(email string) bool {
	// Simple email validation - in production, use a more robust validator
	return len(email) > 3 && strings.Contains(email, "@") && strings.Contains(email, ".")
}

func isValidPhone(phone string) bool {
	// Simple phone validation - in production, use a more robust validator
	return len(phone) > 10
}
