package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"wedding-invitation-backend/internal/domain/repository"
	"wedding-invitation-backend/internal/services"
	"wedding-invitation-backend/internal/utils"
)

type RSVPHandler struct {
	rsvpService *services.RSVPService
}

func NewRSVPHandler(rsvpService *services.RSVPService) *RSVPHandler {
	return &RSVPHandler{
		rsvpService: rsvpService,
	}
}

// SubmitRSVP godoc
// @Summary Submit a new RSVP
// @Description Submit a new RSVP for a wedding (public endpoint)
// @Tags rsvp
// @Accept json
// @Produce json
// @Param id path string true "Wedding ID"
// @Param rsvp body services.SubmitRSVPRequest true "RSVP data"
// @Success 201 {object} models.RSVP
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 422 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/public/weddings/{id}/rsvp [post]
func (h *RSVPHandler) SubmitRSVP(c *gin.Context) {
	weddingID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid wedding ID")
		return
	}

	var req services.SubmitRSVPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	// Validate request
	if err := utils.ValidateStruct(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Set client info
	req.IPAddress = c.ClientIP()
	req.UserAgent = c.GetHeader("User-Agent")

	// Submit RSVP
	rsvp, err := h.rsvpService.SubmitRSVP(c.Request.Context(), weddingID, req)
	if err != nil {
		switch err {
		case services.ErrWeddingNotFound:
			utils.ErrorResponse(c, http.StatusNotFound, "Wedding not found")
			return
		case services.ErrRSVPClosed:
			utils.ErrorResponse(c, http.StatusUnprocessableEntity, "RSVP is not open for this wedding")
			return
		case services.ErrInvalidRSVPStatus:
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid RSVP status")
			return
		case services.ErrDuplicateRSVP:
			utils.ErrorResponse(c, http.StatusConflict, "RSVP already submitted for this email")
			return
		case services.ErrTooManyPlusOnes:
			utils.ErrorResponse(c, http.StatusBadRequest, "Too many plus ones")
			return
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to submit RSVP")
			return
		}
	}

	utils.Response(c, http.StatusCreated, rsvp)
}

// GetRSVPs godoc
// @Summary Get RSVPs for a wedding
// @Description Get paginated list of RSVPs for a wedding (owner only)
// @Tags rsvp
// @Produce json
// @Param id path string true "Wedding ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param status query string false "Filter by status"
// @Param search query string false "Search by name or email"
// @Param source query string false "Filter by source"
// @Success 200 {object} PaginatedResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/weddings/{id}/rsvps [get]
func (h *RSVPHandler) GetRSVPs(c *gin.Context) {
	weddingID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid wedding ID")
		return
	}

	// Get user ID from context (should be set by auth middleware)
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Build filters
	filters := repository.RSVPFilters{
		Status: c.Query("status"),
		Search: c.Query("search"),
		Source: c.Query("source"),
	}

	rsvps, total, err := h.rsvpService.ListRSVPs(c.Request.Context(), weddingID, userID, page, pageSize, filters)
	if err != nil {
		switch err {
		case services.ErrWeddingNotFound:
			utils.ErrorResponse(c, http.StatusNotFound, "Wedding not found")
			return
		case services.ErrUnauthorized:
			utils.ErrorResponse(c, http.StatusForbidden, "Not authorized to view RSVPs for this wedding")
			return
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get RSVPs")
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        rsvps,
		"page":        page,
		"page_size":   pageSize,
		"total":       total,
		"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// GetRSVPStatistics godoc
// @Summary Get RSVP statistics for a wedding
// @Description Get RSVP statistics including counts, dietary restrictions, and trends (owner only)
// @Tags rsvp
// @Produce json
// @Param id path string true "Wedding ID"
// @Success 200 {object} models.RSVPStatistics
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/weddings/{id}/rsvps/statistics [get]
func (h *RSVPHandler) GetRSVPStatistics(c *gin.Context) {
	weddingID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid wedding ID")
		return
	}

	// Get user ID from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid user ID")
		return
	}

	stats, err := h.rsvpService.GetRSVPStatistics(c.Request.Context(), weddingID, userID)
	if err != nil {
		switch err {
		case services.ErrWeddingNotFound:
			utils.ErrorResponse(c, http.StatusNotFound, "Wedding not found")
			return
		case services.ErrUnauthorized:
			utils.ErrorResponse(c, http.StatusForbidden, "Not authorized to view statistics for this wedding")
			return
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get RSVP statistics")
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": stats})
}

// UpdateRSVP godoc
// @Summary Update an RSVP
// @Description Update an existing RSVP (owner only, within 24 hours of submission)
// @Tags rsvp
// @Accept json
// @Produce json
// @Param id path string true "RSVP ID"
// @Param rsvp body services.UpdateRSVPRequest true "Updated RSVP data"
// @Success 200 {object} models.RSVP
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 422 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/rsvps/{id} [put]
func (h *RSVPHandler) UpdateRSVP(c *gin.Context) {
	rsvpID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid RSVP ID")
		return
	}

	var req services.UpdateRSVPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	// Validate request
	if err := utils.ValidateStruct(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	rsvp, err := h.rsvpService.UpdateRSVP(c.Request.Context(), rsvpID, req)
	if err != nil {
		switch err {
		case services.ErrRSVPNotFound:
			utils.ErrorResponse(c, http.StatusNotFound, "RSVP not found")
			return
		case services.ErrRSVPCannotModify:
			utils.ErrorResponse(c, http.StatusUnprocessableEntity, "RSVP cannot be modified after 24 hours")
			return
		case services.ErrInvalidRSVPStatus:
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid RSVP status")
			return
		case services.ErrTooManyPlusOnes:
			utils.ErrorResponse(c, http.StatusBadRequest, "Too many plus ones")
			return
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update RSVP")
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": rsvp})
}

// DeleteRSVP godoc
// @Summary Delete an RSVP
// @Description Delete an RSVP (wedding owner only)
// @Tags rsvp
// @Produce json
// @Param id path string true "RSVP ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/rsvps/{id} [delete]
func (h *RSVPHandler) DeleteRSVP(c *gin.Context) {
	rsvpID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid RSVP ID")
		return
	}

	// Get user ID from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid user ID")
		return
	}

	err = h.rsvpService.DeleteRSVP(c.Request.Context(), rsvpID, userID)
	if err != nil {
		switch err {
		case services.ErrRSVPNotFound:
			utils.ErrorResponse(c, http.StatusNotFound, "RSVP not found")
			return
		case services.ErrUnauthorized:
			utils.ErrorResponse(c, http.StatusForbidden, "Not authorized to delete this RSVP")
			return
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete RSVP")
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "RSVP deleted successfully"})
}

// ExportRSVPs godoc
// @Summary Export RSVPs
// @Description Export all RSVPs for a wedding (owner only, for CSV download)
// @Tags rsvp
// @Produce json
// @Param id path string true "Wedding ID"
// @Success 200 {array} models.RSVP
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/weddings/{id}/rsvps/export [get]
func (h *RSVPHandler) ExportRSVPs(c *gin.Context) {
	weddingID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid wedding ID")
		return
	}

	// Get user ID from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid user ID")
		return
	}

	rsvps, err := h.rsvpService.ExportRSVPs(c.Request.Context(), weddingID, userID)
	if err != nil {
		switch err {
		case services.ErrWeddingNotFound:
			utils.ErrorResponse(c, http.StatusNotFound, "Wedding not found")
			return
		case services.ErrUnauthorized:
			utils.ErrorResponse(c, http.StatusForbidden, "Not authorized to export RSVPs for this wedding")
			return
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to export RSVPs")
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": rsvps})
}
