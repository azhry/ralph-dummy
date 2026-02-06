package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"wedding-invitation-backend/internal/domain/models"
	"wedding-invitation-backend/internal/services"
	"wedding-invitation-backend/internal/utils"
)

// AnalyticsHandler handles analytics-related requests
type AnalyticsHandler struct {
	analyticsService services.AnalyticsService
	weddingService   services.WeddingService
}

// NewAnalyticsHandler creates a new analytics handler
func NewAnalyticsHandler(analyticsService services.AnalyticsService, weddingService services.WeddingService) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: analyticsService,
		weddingService:   weddingService,
	}
}

// TrackPageViewRequest represents a page view tracking request
type TrackPageViewRequest struct {
	WeddingID string `json:"wedding_id" binding:"required"`
	SessionID string `json:"session_id" binding:"required"`
	Page      string `json:"page" binding:"required"`
}

// TrackConversionRequest represents a conversion tracking request
type TrackConversionRequest struct {
	WeddingID  string                 `json:"wedding_id" binding:"required"`
	SessionID  string                 `json:"session_id" binding:"required"`
	Event      string                 `json:"event" binding:"required"`
	Value      float64                `json:"value"`
	Properties map[string]interface{} `json:"properties"`
}

// TrackRSVPSubmissionRequest represents an RSVP submission tracking request
type TrackRSVPSubmissionRequest struct {
	WeddingID      string `json:"wedding_id" binding:"required"`
	RSVPID         string `json:"rsvp_id" binding:"required"`
	SessionID      string `json:"session_id" binding:"required"`
	Source         string `json:"source" binding:"required"`
	TimeToComplete int64  `json:"time_to_complete"`
}

// TrackRSVPAbandonmentRequest represents an RSVP abandonment tracking request
type TrackRSVPAbandonmentRequest struct {
	WeddingID     string   `json:"wedding_id" binding:"required"`
	SessionID     string   `json:"session_id" binding:"required"`
	AbandonedStep string   `json:"abandoned_step" binding:"required"`
	FormErrors    []string `json:"form_errors"`
}

// AnalyticsFilterRequest represents analytics filter parameters
type AnalyticsFilterRequest struct {
	StartDate *time.Time `json:"start_date"`
	EndDate   *time.Time `json:"end_date"`
	Device    string     `json:"device"`
	Source    string     `json:"source"`
	Page      string     `json:"page"`
	Event     string     `json:"event"`
	Limit     int        `json:"limit"`
	Offset    int        `json:"offset"`
}

// TrackPageView tracks a page view event
// @Summary Track page view
// @Description Track a page view for analytics (public endpoint)
// @Tags Analytics
// @Accept json
// @Produce json
// @Param request body TrackPageViewRequest true "Page view data"
// @Success 201 {object} gin.H
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /analytics/track/page-view [post]
func (h *AnalyticsHandler) TrackPageView(c *gin.Context) {
	var req TrackPageViewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request data: " + err.Error()})
		return
	}

	// Validate wedding ID
	weddingID, err := primitive.ObjectIDFromHex(req.WeddingID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid wedding ID"})
		return
	}

	// Validate page
	if !h.analyticsService.IsValidPage(req.Page) {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid page name"})
		return
	}

	// Track page view
	err = h.analyticsService.TrackPageView(c.Request.Context(), weddingID, req.SessionID, req.Page, c.Request)
	if err != nil {
		if err.Error() == "wedding not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Wedding not found"})
			return
		}
		if err.Error() == "cannot track analytics for unpublished wedding" {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Wedding is not published"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to track page view"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{Message: "Page view tracked successfully"})
}

// TrackRSVPSubmission tracks an RSVP submission event
// @Summary Track RSVP submission
// @Description Track an RSVP submission for analytics
// @Tags Analytics
// @Accept json
// @Produce json
// @Param request body TrackRSVPSubmissionRequest true "RSVP submission data"
// @Success 201 {object} gin.H
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /analytics/track/rsvp-submission [post]
func (h *AnalyticsHandler) TrackRSVPSubmission(c *gin.Context) {
	var req TrackRSVPSubmissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request data: " + err.Error()})
		return
	}

	// Validate wedding ID
	weddingID, err := primitive.ObjectIDFromHex(req.WeddingID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid wedding ID"})
		return
	}

	// Validate RSVP ID
	rsvpID, err := primitive.ObjectIDFromHex(req.RSVPID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid RSVP ID"})
		return
	}

	// Validate source
	validSources := []string{"web", "direct_link", "qr_code", "manual"}
	isValidSource := false
	for _, source := range validSources {
		if req.Source == source {
			isValidSource = true
			break
		}
	}
	if !isValidSource {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid source"})
		return
	}

	// Track RSVP submission
	err = h.analyticsService.TrackRSVPSubmission(c.Request.Context(), weddingID, rsvpID, req.SessionID, req.Source, req.TimeToComplete, c.Request)
	if err != nil {
		if err.Error() == "wedding not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Wedding not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to track RSVP submission"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{Message: "RSVP submission tracked successfully"})
}

// TrackRSVPAbandonment tracks an RSVP abandonment event
// @Summary Track RSVP abandonment
// @Description Track an RSVP abandonment for analytics
// @Tags Analytics
// @Accept json
// @Produce json
// @Param request body TrackRSVPAbandonmentRequest true "RSVP abandonment data"
// @Success 201 {object} gin.H
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /analytics/track/rsvp-abandonment [post]
func (h *AnalyticsHandler) TrackRSVPAbandonment(c *gin.Context) {
	var req TrackRSVPAbandonmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request data: " + err.Error()})
		return
	}

	// Validate wedding ID
	weddingID, err := primitive.ObjectIDFromHex(req.WeddingID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid wedding ID"})
		return
	}

	// Validate abandoned step
	validSteps := []string{"personal_info", "attending_status", "guest_count", "dietary_restrictions", "confirmation"}
	isValidStep := false
	for _, step := range validSteps {
		if req.AbandonedStep == step {
			isValidStep = true
			break
		}
	}
	if !isValidStep {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid abandoned step"})
		return
	}

	// Track RSVP abandonment
	err = h.analyticsService.TrackRSVPAbandonment(c.Request.Context(), weddingID, req.SessionID, req.AbandonedStep, req.FormErrors, c.Request)
	if err != nil {
		if err.Error() == "wedding not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Wedding not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to track RSVP abandonment"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{Message: "RSVP abandonment tracked successfully"})
}

// TrackConversion tracks a conversion event
// @Summary Track conversion
// @Description Track a conversion event for analytics
// @Tags Analytics
// @Accept json
// @Produce json
// @Param request body TrackConversionRequest true "Conversion data"
// @Success 201 {object} gin.H
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /analytics/track/conversion [post]
func (h *AnalyticsHandler) TrackConversion(c *gin.Context) {
	var req TrackConversionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request data: " + err.Error()})
		return
	}

	// Validate wedding ID
	weddingID, err := primitive.ObjectIDFromHex(req.WeddingID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid wedding ID"})
		return
	}

	// Validate event
	if !h.analyticsService.IsValidEvent(req.Event) {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid event"})
		return
	}

	// Sanitize properties
	if req.Properties != nil {
		req.Properties = h.analyticsService.SanitizeCustomData(req.Properties)
	}

	// Track conversion
	err = h.analyticsService.TrackConversion(c.Request.Context(), weddingID, req.SessionID, req.Event, req.Value, req.Properties)
	if err != nil {
		if err.Error() == "wedding not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Wedding not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to track conversion"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{Message: "Conversion tracked successfully"})
}

// GetWeddingAnalytics retrieves wedding analytics
// @Summary Get wedding analytics
// @Description Retrieve analytics for a specific wedding
// @Tags Analytics
// @Param id path string true "Wedding ID"
// @Success 200 {object} gin.H{data=models.WeddingAnalytics}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Router /weddings/{id}/analytics [get]
func (h *AnalyticsHandler) GetWeddingAnalytics(c *gin.Context) {
	weddingIDStr := c.Param("id")
	weddingID, err := primitive.ObjectIDFromHex(weddingIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid wedding ID"})
		return
	}

	// Get user ID from context (would be set by auth middleware)
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid user ID"})
		return
	}

	// Verify wedding ownership
	wedding, err := h.weddingService.GetWeddingByID(c.Request.Context(), weddingID)
	if err != nil {
		if err.Error() == "wedding not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Wedding not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to retrieve wedding"})
		return
	}

	if wedding.UserID != userID {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "Access denied"})
		return
	}

	// Get analytics
	analytics, err := h.analyticsService.GetWeddingAnalytics(c.Request.Context(), weddingID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to retrieve analytics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{Data: analytics})
}

// GetAnalyticsSummary retrieves analytics summary
// @Summary Get analytics summary
// @Description Retrieve analytics summary for a wedding with specified period
// @Tags Analytics
// @Param id path string true "Wedding ID"
// @Param period query string false "Period (daily, weekly, monthly)" default(daily)
// @Success 200 {object} gin.H{data=models.AnalyticsSummary}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Router /weddings/{id}/analytics/summary [get]
func (h *AnalyticsHandler) GetAnalyticsSummary(c *gin.Context) {
	weddingIDStr := c.Param("id")
	weddingID, err := primitive.ObjectIDFromHex(weddingIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid wedding ID"})
		return
	}

	// Get user ID from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid user ID"})
		return
	}

	// Verify wedding ownership
	wedding, err := h.weddingService.GetWeddingByID(c.Request.Context(), weddingID)
	if err != nil {
		if err.Error() == "wedding not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Wedding not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to retrieve wedding"})
		return
	}

	if wedding.UserID != userID {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "Access denied"})
		return
	}

	// Get period from query
	period := c.DefaultQuery("period", "daily")
	if !h.analyticsService.ValidatePeriod(period) {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid period"})
		return
	}

	// Get analytics summary
	summary, err := h.analyticsService.GetAnalyticsSummary(c.Request.Context(), weddingID, period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to retrieve analytics summary"})
		return
	}

	c.JSON(http.StatusOK, gin.H{Data: summary})
}

// GetPageViews retrieves page views with filtering
// @Summary Get page views
// @Description Retrieve page views for a wedding with filtering
// @Tags Analytics
// @Param id path string true "Wedding ID"
// @Param start_date query string false "Start date (RFC3339)"
// @Param end_date query string false "End date (RFC3339)"
// @Param device query string false "Device filter"
// @Param page query string false "Page filter"
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} gin.H{data=PageViewsResponse}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Router /weddings/{id}/analytics/page-views [get]
func (h *AnalyticsHandler) GetPageViews(c *gin.Context) {
	weddingIDStr := c.Param("id")
	weddingID, err := primitive.ObjectIDFromHex(weddingIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid wedding ID"})
		return
	}

	// Get user ID from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid user ID"})
		return
	}

	// Verify wedding ownership
	wedding, err := h.weddingService.GetWeddingByID(c.Request.Context(), weddingID)
	if err != nil {
		if err.Error() == "wedding not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Wedding not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to retrieve wedding"})
		return
	}

	if wedding.UserID != userID {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "Access denied"})
		return
	}

	// Parse filter parameters
	filter := &models.AnalyticsFilter{}
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		startDate, err := time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid start date format"})
			return
		}
		filter.StartDate = &startDate
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		endDate, err := time.Parse(time.RFC3339, endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid end date format"})
			return
		}
		filter.EndDate = &endDate
	}

	filter.Device = c.Query("device")
	filter.Page = c.Query("page")

	if limitStr := c.Query("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 0 {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid limit"})
			return
		}
		filter.Limit = limit
	} else {
		filter.Limit = 50
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid offset"})
			return
		}
		filter.Offset = offset
	} else {
		filter.Offset = 0
	}

	// Get page views
	pageViews, total, err := h.analyticsService.GetPageViews(c.Request.Context(), weddingID, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to retrieve page views"})
		return
	}

	response := PageViewsResponse{
		PageViews: pageViews,
		Total:     total,
		Limit:     filter.Limit,
		Offset:    filter.Offset,
	}

	c.JSON(http.StatusOK, gin.H{Data: response})
}

// GetPopularPages retrieves popular pages for a wedding
// @Summary Get popular pages
// @Description Retrieve most popular pages for a wedding
// @Tags Analytics
// @Param id path string true "Wedding ID"
// @Param limit query int false "Limit" default(10)
// @Success 200 {object} gin.H{data=[]models.PageStats}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Router /weddings/{id}/analytics/popular-pages [get]
func (h *AnalyticsHandler) GetPopularPages(c *gin.Context) {
	weddingIDStr := c.Param("id")
	weddingID, err := primitive.ObjectIDFromHex(weddingIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid wedding ID"})
		return
	}

	// Get user ID from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid user ID"})
		return
	}

	// Verify wedding ownership
	wedding, err := h.weddingService.GetWeddingByID(c.Request.Context(), weddingID)
	if err != nil {
		if err.Error() == "wedding not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Wedding not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to retrieve wedding"})
		return
	}

	if wedding.UserID != userID {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "Access denied"})
		return
	}

	// Get limit from query
	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil || parsedLimit < 1 || parsedLimit > 100 {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid limit (must be 1-100)"})
			return
		}
		limit = parsedLimit
	}

	// Get popular pages
	pages, err := h.analyticsService.GetPopularPages(c.Request.Context(), weddingID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to retrieve popular pages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{Data: pages})
}

// GetSystemAnalytics retrieves system-wide analytics
// @Summary Get system analytics
// @Description Retrieve system-wide analytics (admin only)
// @Tags Analytics
// @Success 200 {object} gin.H{data=models.SystemAnalytics}
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /admin/analytics/system [get]
func (h *AnalyticsHandler) GetSystemAnalytics(c *gin.Context) {
	// Check if user is admin (would be set by auth middleware)
	isAdmin, exists := c.Get("is_admin")
	if !exists || !isAdmin.(bool) {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "Admin access required"})
		return
	}

	// Get system analytics
	analytics, err := h.analyticsService.GetSystemAnalytics(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to retrieve system analytics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{Data: analytics})
}

// RefreshAnalytics refreshes analytics data
// @Summary Refresh analytics
// @Description Force refresh of analytics data
// @Tags Analytics
// @Param id path string true "Wedding ID"
// @Success 200 {object} gin.H
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /weddings/{id}/analytics/refresh [post]
func (h *AnalyticsHandler) RefreshAnalytics(c *gin.Context) {
	weddingIDStr := c.Param("id")
	weddingID, err := primitive.ObjectIDFromHex(weddingIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid wedding ID"})
		return
	}

	// Get user ID from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid user ID"})
		return
	}

	// Verify wedding ownership
	wedding, err := h.weddingService.GetWeddingByID(c.Request.Context(), weddingID)
	if err != nil {
		if err.Error() == "wedding not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Wedding not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to retrieve wedding"})
		return
	}

	if wedding.UserID != userID {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "Access denied"})
		return
	}

	// Refresh analytics
	err = h.analyticsService.RefreshWeddingAnalytics(c.Request.Context(), weddingID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to refresh analytics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{Message: "Analytics refreshed successfully"})
}

// RefreshSystemAnalytics refreshes system analytics
// @Summary Refresh system analytics
// @Description Force refresh of system analytics (admin only)
// @Tags Analytics
// @Success 200 {object} gin.H
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /admin/analytics/refresh [post]
func (h *AnalyticsHandler) RefreshSystemAnalytics(c *gin.Context) {
	// Check if user is admin
	isAdmin, exists := c.Get("is_admin")
	if !exists || !isAdmin.(bool) {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "Admin access required"})
		return
	}

	// Refresh system analytics
	err := h.analyticsService.RefreshSystemAnalytics(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to refresh system analytics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{Message: "System analytics refreshed successfully"})
}

// Response types

type PageViewsResponse struct {
	PageViews []*models.PageView `json:"page_views"`
	Total     int64              `json:"total"`
	Limit     int                `json:"limit"`
	Offset    int                `json:"offset"`
}

// Helper validation functions

func (h *AnalyticsHandler) validateAnalyticsFilter(req *AnalyticsFilterRequest) error {
	if req.Limit < 0 || req.Limit > 100 {
		return errors.New("limit must be between 0 and 100")
	}

	if req.Offset < 0 {
		return errors.New("offset must be non-negative")
	}

	if req.StartDate != nil && req.EndDate != nil && req.StartDate.After(*req.EndDate) {
		return errors.New("start date must be before end date")
	}

	return nil
}
