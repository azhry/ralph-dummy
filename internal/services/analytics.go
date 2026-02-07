package services

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"wedding-invitation-backend/internal/domain/models"
	"wedding-invitation-backend/internal/domain/repository"
)

// AnalyticsService represents the analytics service interface
type AnalyticsService interface {
	// Page View Tracking
	TrackPageView(ctx context.Context, weddingID primitive.ObjectID, sessionID, page string, req *http.Request) error
	GetPageViews(ctx context.Context, weddingID primitive.ObjectID, filter *models.AnalyticsFilter) ([]*models.PageView, int64, error)

	// RSVP Analytics
	TrackRSVPSubmission(ctx context.Context, weddingID, rsvpID primitive.ObjectID, sessionID, source string, timeToComplete int64, req *http.Request) error
	TrackRSVPAbandonment(ctx context.Context, weddingID primitive.ObjectID, sessionID, abandonedStep string, formErrors []string, req *http.Request) error
	GetRSVPAnalytics(ctx context.Context, weddingID primitive.ObjectID, filter *models.AnalyticsFilter) ([]*models.RSVPAnalytics, int64, error)

	// Conversion Tracking
	TrackConversion(ctx context.Context, weddingID primitive.ObjectID, sessionID, event string, value float64, properties map[string]interface{}) error
	GetConversions(ctx context.Context, weddingID primitive.ObjectID, filter *models.AnalyticsFilter) ([]*models.ConversionEvent, int64, error)

	// Analytics Data
	GetWeddingAnalytics(ctx context.Context, weddingID primitive.ObjectID) (*models.WeddingAnalytics, error)
	GetSystemAnalytics(ctx context.Context) (*models.SystemAnalytics, error)
	GetAnalyticsSummary(ctx context.Context, weddingID primitive.ObjectID, period string) (*models.AnalyticsSummary, error)

	// Reports
	GetPopularPages(ctx context.Context, weddingID primitive.ObjectID, limit int) ([]models.PageStats, error)
	GetTrafficSources(ctx context.Context, weddingID primitive.ObjectID, limit int) ([]models.TrafficSourceStats, error)
	GetDailyMetrics(ctx context.Context, weddingID primitive.ObjectID, startDate, endDate time.Time) ([]models.DailyMetrics, error)

	// Management
	RefreshWeddingAnalytics(ctx context.Context, weddingID primitive.ObjectID) error
	RefreshSystemAnalytics(ctx context.Context) error
	CleanupOldAnalytics(ctx context.Context, olderThan time.Time) error

	// Validation
	IsValidPage(page string) bool
	IsValidEvent(event string) bool
	ValidatePeriod(period string) bool
	SanitizeCustomData(data map[string]interface{}) map[string]interface{}
}

type analyticsService struct {
	analyticsRepo repository.AnalyticsRepository
	weddingRepo   repository.WeddingRepository
	logger        *zap.Logger
}

// NewAnalyticsService creates a new analytics service
func NewAnalyticsService(analyticsRepo repository.AnalyticsRepository, weddingRepo repository.WeddingRepository, logger *zap.Logger) AnalyticsService {
	return &analyticsService{
		analyticsRepo: analyticsRepo,
		weddingRepo:   weddingRepo,
		logger:        logger,
	}
}

// TrackPageView tracks a page view event
func (s *analyticsService) TrackPageView(ctx context.Context, weddingID primitive.ObjectID, sessionID, page string, req *http.Request) error {
	// Validate that wedding exists and is published
	wedding, err := s.weddingRepo.GetByID(ctx, weddingID)
	if err != nil {
		return fmt.Errorf("wedding not found: %w", err)
	}

	if wedding.Status != string(models.WeddingStatusPublished) {
		return fmt.Errorf("cannot track analytics for unpublished wedding")
	}

	// Extract user agent and IP address
	userAgent := ""
	if req != nil {
		userAgent = req.Header.Get("User-Agent")
	}

	ipAddress := ""
	if req != nil {
		ipAddress = s.getClientIP(req)
	}

	// Parse device, browser, and OS from user agent
	device, browser, os := s.parseUserAgent(userAgent)

	// Get referrer
	referrer := ""
	if req != nil {
		referrer = req.Header.Get("Referer")
		if referrer == "" {
			referrer = req.URL.Query().Get("ref")
		}
	}

	// Geolocation (placeholder - would integrate with IP geolocation service in production)
	country, city := "", ""
	if ipAddress != "" {
		country, city = s.getGeoLocation(ipAddress)
	}

	pageView := &models.PageView{
		WeddingID: weddingID,
		SessionID: sessionID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Referrer:  referrer,
		Page:      page,
		Timestamp: time.Now(),
		Device:    device,
		Browser:   browser,
		OS:        os,
		Country:   country,
		City:      city,
		Metadata:  make(map[string]interface{}),
	}

	err = s.analyticsRepo.TrackPageView(ctx, pageView)
	if err != nil {
		s.logger.Error("Failed to track page view",
			zap.Error(err),
			zap.String("wedding_id", weddingID.Hex()),
			zap.String("page", page))
		return fmt.Errorf("failed to track page view: %w", err)
	}

	s.logger.Debug("Tracked page view",
		zap.String("wedding_id", weddingID.Hex()),
		zap.String("session_id", sessionID),
		zap.String("page", page))

	return nil
}

// GetPageViews retrieves page views with filtering
func (s *analyticsService) GetPageViews(ctx context.Context, weddingID primitive.ObjectID, filter *models.AnalyticsFilter) ([]*models.PageView, int64, error) {
	return s.analyticsRepo.GetPageViews(ctx, weddingID, filter)
}

// TrackRSVPSubmission tracks an RSVP submission event
func (s *analyticsService) TrackRSVPSubmission(ctx context.Context, weddingID, rsvpID primitive.ObjectID, sessionID, source string, timeToComplete int64, req *http.Request) error {
	// Validate that wedding exists
	_, err := s.weddingRepo.GetByID(ctx, weddingID)
	if err != nil {
		return fmt.Errorf("wedding not found: %w", err)
	}

	// Extract user agent and device info
	userAgent := ""
	device := "unknown"
	browser := "unknown"
	if req != nil {
		userAgent = req.Header.Get("User-Agent")
		device, browser, _ = s.parseUserAgent(userAgent)
	}

	// Get referrer
	referrer := ""
	if req != nil {
		referrer = req.Header.Get("Referer")
		if referrer == "" {
			referrer = req.URL.Query().Get("ref")
		}
	}

	event := &models.RSVPAnalytics{
		WeddingID:      weddingID,
		RSVPID:         rsvpID,
		SessionID:      sessionID,
		TimeToComplete: timeToComplete,
		Source:         source,
		Device:         device,
		Browser:        browser,
		Referrer:       referrer,
		Timestamp:      time.Now(),
	}

	err = s.analyticsRepo.TrackRSVPEvent(ctx, event)
	if err != nil {
		s.logger.Error("Failed to track RSVP submission",
			zap.Error(err),
			zap.String("wedding_id", weddingID.Hex()),
			zap.String("rsvp_id", rsvpID.Hex()))
		return fmt.Errorf("failed to track RSVP submission: %w", err)
	}

	// Track conversion
	err = s.TrackConversion(ctx, weddingID, sessionID, "rsvp_completed", 1, map[string]interface{}{
		"source":           source,
		"time_to_complete": timeToComplete,
	})
	if err != nil {
		s.logger.Warn("Failed to track RSVP conversion", zap.Error(err))
	}

	s.logger.Debug("Tracked RSVP submission",
		zap.String("wedding_id", weddingID.Hex()),
		zap.String("rsvp_id", rsvpID.Hex()),
		zap.String("source", source))

	return nil
}

// TrackRSVPAbandonment tracks an RSVP abandonment event
func (s *analyticsService) TrackRSVPAbandonment(ctx context.Context, weddingID primitive.ObjectID, sessionID, abandonedStep string, formErrors []string, req *http.Request) error {
	// Validate that wedding exists
	_, err := s.weddingRepo.GetByID(ctx, weddingID)
	if err != nil {
		return fmt.Errorf("wedding not found: %w", err)
	}

	// Extract device info
	device := "unknown"
	browser := "unknown"
	if req != nil {
		userAgent := req.Header.Get("User-Agent")
		device, browser, _ = s.parseUserAgent(userAgent)
	}

	// Get referrer
	referrer := ""
	if req != nil {
		referrer = req.Header.Get("Referer")
	}

	event := &models.RSVPAnalytics{
		WeddingID:     weddingID,
		SessionID:     sessionID,
		Device:        device,
		Browser:       browser,
		Referrer:      referrer,
		Timestamp:     time.Now(),
		AbandonedStep: abandonedStep,
		FormErrors:    formErrors,
	}

	err = s.analyticsRepo.TrackRSVPEvent(ctx, event)
	if err != nil {
		s.logger.Error("Failed to track RSVP abandonment",
			zap.Error(err),
			zap.String("wedding_id", weddingID.Hex()),
			zap.String("abandoned_step", abandonedStep))
		return fmt.Errorf("failed to track RSVP abandonment: %w", err)
	}

	// Track conversion funnel
	err = s.TrackConversion(ctx, weddingID, sessionID, "rsvp_abandoned", 0, map[string]interface{}{
		"step":        abandonedStep,
		"form_errors": len(formErrors),
	})
	if err != nil {
		s.logger.Warn("Failed to track RSVP abandonment conversion", zap.Error(err))
	}

	s.logger.Debug("Tracked RSVP abandonment",
		zap.String("wedding_id", weddingID.Hex()),
		zap.String("abandoned_step", abandonedStep))

	return nil
}

// GetRSVPAnalytics retrieves RSVP analytics with filtering
func (s *analyticsService) GetRSVPAnalytics(ctx context.Context, weddingID primitive.ObjectID, filter *models.AnalyticsFilter) ([]*models.RSVPAnalytics, int64, error) {
	return s.analyticsRepo.GetRSVPAnalytics(ctx, weddingID, filter)
}

// TrackConversion tracks a conversion event
func (s *analyticsService) TrackConversion(ctx context.Context, weddingID primitive.ObjectID, sessionID, event string, value float64, properties map[string]interface{}) error {
	// Validate that wedding exists
	_, err := s.weddingRepo.GetByID(ctx, weddingID)
	if err != nil {
		return fmt.Errorf("wedding not found: %w", err)
	}

	conversionEvent := &models.ConversionEvent{
		WeddingID:  weddingID,
		SessionID:  sessionID,
		Event:      event,
		Value:      value,
		Timestamp:  time.Now(),
		Properties: properties,
	}

	err = s.analyticsRepo.TrackConversion(ctx, conversionEvent)
	if err != nil {
		s.logger.Error("Failed to track conversion",
			zap.Error(err),
			zap.String("wedding_id", weddingID.Hex()),
			zap.String("event", event))
		return fmt.Errorf("failed to track conversion: %w", err)
	}

	return nil
}

// GetConversions retrieves conversion events with filtering
func (s *analyticsService) GetConversions(ctx context.Context, weddingID primitive.ObjectID, filter *models.AnalyticsFilter) ([]*models.ConversionEvent, int64, error) {
	return s.analyticsRepo.GetConversions(ctx, weddingID, filter)
}

// GetWeddingAnalytics retrieves aggregated analytics for a wedding
func (s *analyticsService) GetWeddingAnalytics(ctx context.Context, weddingID primitive.ObjectID) (*models.WeddingAnalytics, error) {
	// Verify wedding ownership would be handled at the handler level
	return s.analyticsRepo.GetWeddingAnalytics(ctx, weddingID)
}

// GetSystemAnalytics retrieves system-wide analytics
func (s *analyticsService) GetSystemAnalytics(ctx context.Context) (*models.SystemAnalytics, error) {
	return s.analyticsRepo.GetSystemAnalytics(ctx)
}

// GetAnalyticsSummary generates a summary report for a wedding
func (s *analyticsService) GetAnalyticsSummary(ctx context.Context, weddingID primitive.ObjectID, period string) (*models.AnalyticsSummary, error) {
	// Verify wedding ownership would be handled at the handler level
	return s.analyticsRepo.GetAnalyticsSummary(ctx, weddingID, period)
}

// GetPopularPages returns the most popular pages for a wedding
func (s *analyticsService) GetPopularPages(ctx context.Context, weddingID primitive.ObjectID, limit int) ([]models.PageStats, error) {
	return s.analyticsRepo.GetPopularPages(ctx, weddingID, limit)
}

// GetTrafficSources returns traffic sources for a wedding
func (s *analyticsService) GetTrafficSources(ctx context.Context, weddingID primitive.ObjectID, limit int) ([]models.TrafficSourceStats, error) {
	return s.analyticsRepo.GetTrafficSources(ctx, weddingID, limit)
}

// GetDailyMetrics returns daily metrics for a date range
func (s *analyticsService) GetDailyMetrics(ctx context.Context, weddingID primitive.ObjectID, startDate, endDate time.Time) ([]models.DailyMetrics, error) {
	return s.analyticsRepo.GetDailyMetrics(ctx, weddingID, startDate, endDate)
}

// RefreshWeddingAnalytics forces a refresh of wedding analytics
func (s *analyticsService) RefreshWeddingAnalytics(ctx context.Context, weddingID primitive.ObjectID) error {
	err := s.analyticsRepo.UpdateWeddingAnalytics(ctx, weddingID)
	if err != nil {
		s.logger.Error("Failed to refresh wedding analytics",
			zap.Error(err),
			zap.String("wedding_id", weddingID.Hex()))
		return fmt.Errorf("failed to refresh wedding analytics: %w", err)
	}

	s.logger.Info("Wedding analytics refreshed",
		zap.String("wedding_id", weddingID.Hex()))

	return nil
}

// RefreshSystemAnalytics forces a refresh of system analytics
func (s *analyticsService) RefreshSystemAnalytics(ctx context.Context) error {
	err := s.analyticsRepo.UpdateSystemAnalytics(ctx)
	if err != nil {
		s.logger.Error("Failed to refresh system analytics", zap.Error(err))
		return fmt.Errorf("failed to refresh system analytics: %w", err)
	}

	s.logger.Info("System analytics refreshed")
	return nil
}

// CleanupOldAnalytics removes old analytics data
func (s *analyticsService) CleanupOldAnalytics(ctx context.Context, olderThan time.Time) error {
	err := s.analyticsRepo.CleanupOldAnalytics(ctx, olderThan)
	if err != nil {
		s.logger.Error("Failed to cleanup old analytics", zap.Error(err))
		return fmt.Errorf("failed to cleanup old analytics: %w", err)
	}

	s.logger.Info("Old analytics data cleaned up",
		zap.Time("older_than", olderThan))

	return nil
}

// Helper methods

// getClientIP extracts the real client IP address
func (s *analyticsService) getClientIP(req *http.Request) string {
	// Check X-Forwarded-For header
	xForwardedFor := req.Header.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(xForwardedFor, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if net.ParseIP(ip) != nil {
				return ip
			}
		}
	}

	// Check X-Real-IP header
	xRealIP := req.Header.Get("X-Real-IP")
	if xRealIP != "" {
		if net.ParseIP(xRealIP) != nil {
			return xRealIP
		}
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return req.RemoteAddr
	}

	return ip
}

// parseUserAgent extracts device, browser, and OS from user agent string
func (s *analyticsService) parseUserAgent(userAgent string) (device, browser, os string) {
	if userAgent == "" {
		return "unknown", "unknown", "unknown"
	}

	userAgent = strings.ToLower(userAgent)

	// Detect device type - check for tablet first as iPad also contains "mobile"
	if strings.Contains(userAgent, "ipad") {
		device = "tablet"
	} else if strings.Contains(userAgent, "tablet") {
		device = "tablet"
	} else if strings.Contains(userAgent, "mobile") || strings.Contains(userAgent, "android") || strings.Contains(userAgent, "iphone") {
		device = "mobile"
	} else {
		device = "desktop"
	}

	// Detect browser
	if strings.Contains(userAgent, "chrome") && !strings.Contains(userAgent, "edg") {
		browser = "chrome"
	} else if strings.Contains(userAgent, "firefox") {
		browser = "firefox"
	} else if strings.Contains(userAgent, "safari") && !strings.Contains(userAgent, "chrome") {
		browser = "safari"
	} else if strings.Contains(userAgent, "edg") {
		browser = "edge"
	} else if strings.Contains(userAgent, "opera") {
		browser = "opera"
	} else {
		browser = "other"
	}

	// Detect OS - check for iOS first as it includes both iPhone and iPad
	if strings.Contains(userAgent, "iphone") || strings.Contains(userAgent, "ipad") || strings.Contains(userAgent, "ios") {
		os = "ios"
	} else if strings.Contains(userAgent, "mac os") {
		os = "macos"
	} else if strings.Contains(userAgent, "windows") {
		os = "windows"
	} else if strings.Contains(userAgent, "linux") {
		os = "linux"
	} else if strings.Contains(userAgent, "android") {
		os = "android"
	} else {
		os = "other"
	}

	return device, browser, os
}

// getGeoLocation extracts geolocation from IP address (placeholder implementation)
func (s *analyticsService) getGeoLocation(ipAddress string) (country, city string) {
	// This is a placeholder implementation
	// In production, you would integrate with a geolocation service like:
	// - MaxMind GeoIP2
	// - IPinfo
	// - IP-API
	// - Abstract API

	if ipAddress == "" || ipAddress == "127.0.0.1" || ipAddress == "::1" {
		return "", ""
	}

	// For demonstration, return empty values
	// In production, you would make an API call to get real geolocation data
	return "", ""
}

// GenerateSessionID generates a unique session ID
func (s *analyticsService) GenerateSessionID() string {
	return primitive.NewObjectID().Hex()
}

// IsValidPage validates that a page name is allowed
func (s *analyticsService) IsValidPage(page string) bool {
	validPages := []string{
		"invitation",
		"rsvp",
		"gallery",
		"details",
		"location",
		"schedule",
		"contact",
		"home",
	}

	for _, validPage := range validPages {
		if page == validPage {
			return true
		}
	}

	return false
}

// IsValidEvent validates that a conversion event is allowed
func (s *analyticsService) IsValidEvent(event string) bool {
	validEvents := []string{
		"rsvp_started",
		"rsvp_completed",
		"rsvp_abandoned",
		"share_clicked",
		"gallery_viewed",
		"map_opened",
		"contact_clicked",
		"page_viewed",
	}

	for _, validEvent := range validEvents {
		if event == validEvent {
			return true
		}
	}

	return false
}

// SanitizeReferrer sanitizes the referrer URL
func (s *analyticsService) SanitizeReferrer(referrer string) string {
	if referrer == "" {
		return ""
	}

	// Remove query parameters and fragments
	if idx := strings.Index(referrer, "?"); idx != -1 {
		referrer = referrer[:idx]
	}
	if idx := strings.Index(referrer, "#"); idx != -1 {
		referrer = referrer[:idx]
	}

	// Limit length to 500 characters
	if len(referrer) > 500 {
		referrer = referrer[:500]
	}

	return referrer
}

// ExtractSourceFromReferrer extracts traffic source from referrer
func (s *analyticsService) ExtractSourceFromReferrer(referrer string) string {
	if referrer == "" {
		return "direct"
	}

	// Check for common sources
	if strings.Contains(referrer, "google.com") {
		return "google"
	} else if strings.Contains(referrer, "facebook.com") {
		return "facebook"
	} else if strings.Contains(referrer, "instagram.com") {
		return "instagram"
	} else if strings.Contains(referrer, "twitter.com") {
		return "twitter"
	} else if strings.Contains(referrer, "linkedin.com") {
		return "linkedin"
	} else if strings.Contains(referrer, "pinterest.com") {
		return "pinterest"
	} else if strings.Contains(referrer, "youtube.com") {
		return "youtube"
	} else {
		return "referral"
	}
}

// ValidatePeriod validates analytics period
func (s *analyticsService) ValidatePeriod(period string) bool {
	validPeriods := []string{"daily", "weekly", "monthly", "yearly"}

	for _, validPeriod := range validPeriods {
		if period == validPeriod {
			return true
		}
	}

	return false
}

// SanitizeCustomData sanitizes custom data for analytics
func (s *analyticsService) SanitizeCustomData(data map[string]interface{}) map[string]interface{} {
	sanitized := make(map[string]interface{})

	for key, value := range data {
		// Sanitize key first (only alphanumeric and underscore)
		reg := regexp.MustCompile(`[^a-zA-Z0-9_]`)
		key = reg.ReplaceAllString(key, "_")

		// Limit key length after sanitization
		if len(key) > 50 {
			key = key[:50]
		}

		// Limit value size
		strValue := fmt.Sprintf("%v", value)
		if len(strValue) > 200 {
			sanitized[key] = strValue[:200]
		} else {
			sanitized[key] = value
		}
	}

	return sanitized
}
