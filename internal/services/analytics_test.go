package services

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap/zaptest"

	"wedding-invitation-backend/internal/domain/models"
)

func TestAnalyticsService_TrackPageView(t *testing.T) {
	analyticsRepo := &MockAnalyticsRepository{}
	weddingRepo := &MockWeddingRepository{}
	logger := zaptest.NewLogger(t)

	service := NewAnalyticsService(analyticsRepo, weddingRepo, logger)

	ctx := context.Background()
	weddingID := primitive.NewObjectID()
	sessionID := "test-session-123"
	page := "invitation"

	// Create a mock HTTP request
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("X-Forwarded-For", "192.168.1.1")
	req.Header.Set("Referer", "https://google.com")

	t.Run("Success - published wedding", func(t *testing.T) {
		// Mock wedding exists and is published
		wedding := &models.Wedding{
			ID:     weddingID,
			Status: string(models.WeddingStatusPublished),
		}
		weddingRepo.On("GetByID", ctx, weddingID).Return(wedding, nil)

		// Mock successful page view tracking
		analyticsRepo.On("TrackPageView", ctx, mock.AnythingOfType("*models.PageView")).Return(nil)

		err := service.TrackPageView(ctx, weddingID, sessionID, page, req)
		require.NoError(t, err)

		analyticsRepo.AssertExpectations(t)
		weddingRepo.AssertExpectations(t)
	})

	t.Run("Error - wedding not found", func(t *testing.T) {
		// Create fresh mocks for this test
		analyticsRepo := &MockAnalyticsRepository{}
		weddingRepo := &MockWeddingRepository{}
		logger := zaptest.NewLogger(t)
		service := NewAnalyticsService(analyticsRepo, weddingRepo, logger)

		weddingRepo.On("GetByID", ctx, weddingID).Return(nil, errors.New("wedding not found"))

		err := service.TrackPageView(ctx, weddingID, sessionID, page, req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "wedding not found")

		weddingRepo.AssertExpectations(t)
	})

	t.Run("Error - unpublished wedding", func(t *testing.T) {
		// Create fresh mocks for this test
		analyticsRepo := &MockAnalyticsRepository{}
		weddingRepo := &MockWeddingRepository{}
		logger := zaptest.NewLogger(t)
		service := NewAnalyticsService(analyticsRepo, weddingRepo, logger)

		// Mock wedding exists but is not published
		wedding := &models.Wedding{
			ID:     weddingID,
			Status: string(models.WeddingStatusDraft),
		}
		weddingRepo.On("GetByID", ctx, weddingID).Return(wedding, nil)

		err := service.TrackPageView(ctx, weddingID, sessionID, page, req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot track analytics for unpublished wedding")

		weddingRepo.AssertExpectations(t)
	})

	t.Run("Error - tracking failed", func(t *testing.T) {
		// Create fresh mocks for this test
		analyticsRepo := &MockAnalyticsRepository{}
		weddingRepo := &MockWeddingRepository{}
		logger := zaptest.NewLogger(t)
		service := NewAnalyticsService(analyticsRepo, weddingRepo, logger)

		// Mock wedding exists and is published
		wedding := &models.Wedding{
			ID:     weddingID,
			Status: string(models.WeddingStatusPublished),
		}
		weddingRepo.On("GetByID", ctx, weddingID).Return(wedding, nil)

		// Mock tracking failure
		analyticsRepo.On("TrackPageView", ctx, mock.AnythingOfType("*models.PageView")).Return(errors.New("tracking failed"))

		err := service.TrackPageView(ctx, weddingID, sessionID, page, req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to track page view")

		analyticsRepo.AssertExpectations(t)
		weddingRepo.AssertExpectations(t)
	})
}

func TestAnalyticsService_TrackRSVPSubmission(t *testing.T) {
	analyticsRepo := &MockAnalyticsRepository{}
	weddingRepo := &MockWeddingRepository{}
	logger := zaptest.NewLogger(t)

	service := NewAnalyticsService(analyticsRepo, weddingRepo, logger)

	ctx := context.Background()
	weddingID := primitive.NewObjectID()
	rsvpID := primitive.NewObjectID()
	sessionID := "test-session-123"
	source := "web"
	timeToComplete := int64(120)

	// Create a mock HTTP request
	req := httptest.NewRequest("POST", "/test", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)")
	req.Header.Set("Referer", "https://facebook.com")

	t.Run("Success", func(t *testing.T) {
		// Mock wedding exists
		wedding := &models.Wedding{
			ID:     weddingID,
			Status: string(models.WeddingStatusPublished),
		}
		weddingRepo.On("GetByID", ctx, weddingID).Return(wedding, nil)

		// Mock successful RSVP event tracking
		analyticsRepo.On("TrackRSVPEvent", ctx, mock.AnythingOfType("*models.RSVPAnalytics")).Return(nil)

		// Mock successful conversion tracking
		analyticsRepo.On("TrackConversion", ctx, mock.AnythingOfType("*models.ConversionEvent")).Return(nil)

		err := service.TrackRSVPSubmission(ctx, weddingID, rsvpID, sessionID, source, timeToComplete, req)
		require.NoError(t, err)

		analyticsRepo.AssertExpectations(t)
		weddingRepo.AssertExpectations(t)
	})

	t.Run("Error - wedding not found", func(t *testing.T) {
		// Create fresh mocks for this test
		analyticsRepo := &MockAnalyticsRepository{}
		weddingRepo := &MockWeddingRepository{}
		logger := zaptest.NewLogger(t)
		service := NewAnalyticsService(analyticsRepo, weddingRepo, logger)

		weddingRepo.On("GetByID", ctx, weddingID).Return(nil, errors.New("wedding not found"))

		err := service.TrackRSVPSubmission(ctx, weddingID, rsvpID, sessionID, source, timeToComplete, req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "wedding not found")

		weddingRepo.AssertExpectations(t)
	})
}

func TestAnalyticsService_TrackRSVPAbandonment(t *testing.T) {
	analyticsRepo := &MockAnalyticsRepository{}
	weddingRepo := &MockWeddingRepository{}
	logger := zaptest.NewLogger(t)

	service := NewAnalyticsService(analyticsRepo, weddingRepo, logger)

	ctx := context.Background()
	weddingID := primitive.NewObjectID()
	sessionID := "test-session-123"
	abandonedStep := "personal_info"
	formErrors := []string{"Invalid email format", "Name is required"}

	// Create a mock HTTP request
	req := httptest.NewRequest("GET", "/test", nil)

	t.Run("Success", func(t *testing.T) {
		// Mock wedding exists
		wedding := &models.Wedding{
			ID:     weddingID,
			Status: string(models.WeddingStatusPublished),
		}
		weddingRepo.On("GetByID", ctx, weddingID).Return(wedding, nil)

		// Mock successful RSVP event tracking
		analyticsRepo.On("TrackRSVPEvent", ctx, mock.AnythingOfType("*models.RSVPAnalytics")).Return(nil)

		// Mock successful conversion tracking
		analyticsRepo.On("TrackConversion", ctx, mock.AnythingOfType("*models.ConversionEvent")).Return(nil)

		err := service.TrackRSVPAbandonment(ctx, weddingID, sessionID, abandonedStep, formErrors, req)
		require.NoError(t, err)

		analyticsRepo.AssertExpectations(t)
		weddingRepo.AssertExpectations(t)
	})

	t.Run("Error - wedding not found", func(t *testing.T) {
		weddingRepo.On("GetByID", ctx, weddingID).Return(nil, errors.New("wedding not found"))

		err := service.TrackRSVPAbandonment(ctx, weddingID, sessionID, abandonedStep, formErrors, req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "wedding not found")

		weddingRepo.AssertExpectations(t)
	})
}

func TestAnalyticsService_TrackConversion(t *testing.T) {
	analyticsRepo := &MockAnalyticsRepository{}
	weddingRepo := &MockWeddingRepository{}
	logger := zaptest.NewLogger(t)

	service := NewAnalyticsService(analyticsRepo, weddingRepo, logger)

	ctx := context.Background()
	weddingID := primitive.NewObjectID()
	sessionID := "test-session-123"
	event := "share_clicked"
	value := 1.0
	properties := map[string]interface{}{
		"platform": "facebook",
		"url":      "https://example.com/share",
	}

	t.Run("Success", func(t *testing.T) {
		// Mock wedding exists
		wedding := &models.Wedding{
			ID:     weddingID,
			Status: string(models.WeddingStatusPublished),
		}
		weddingRepo.On("GetByID", ctx, weddingID).Return(wedding, nil)

		// Mock successful conversion tracking
		analyticsRepo.On("TrackConversion", ctx, mock.AnythingOfType("*models.ConversionEvent")).Return(nil)

		err := service.TrackConversion(ctx, weddingID, sessionID, event, value, properties)
		require.NoError(t, err)

		analyticsRepo.AssertExpectations(t)
		weddingRepo.AssertExpectations(t)
	})

	t.Run("Error - wedding not found", func(t *testing.T) {
		weddingRepo.On("GetByID", ctx, weddingID).Return(nil, errors.New("wedding not found"))

		err := service.TrackConversion(ctx, weddingID, sessionID, event, value, properties)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "wedding not found")

		weddingRepo.AssertExpectations(t)
	})
}

func TestAnalyticsService_GetWeddingAnalytics(t *testing.T) {
	analyticsRepo := &MockAnalyticsRepository{}
	weddingRepo := &MockWeddingRepository{}
	logger := zaptest.NewLogger(t)

	service := NewAnalyticsService(analyticsRepo, weddingRepo, logger)

	ctx := context.Background()
	weddingID := primitive.NewObjectID()

	expectedAnalytics := &models.WeddingAnalytics{
		WeddingID:      weddingID,
		PageViews:      100,
		UniqueSessions: 50,
		RSVPCount:      25,
		ConversionRate: 25.0,
		LastUpdated:    time.Now(),
	}

	t.Run("Success", func(t *testing.T) {
		analyticsRepo.On("GetWeddingAnalytics", ctx, weddingID).Return(expectedAnalytics, nil)

		result, err := service.GetWeddingAnalytics(ctx, weddingID)
		require.NoError(t, err)
		assert.Equal(t, expectedAnalytics, result)

		analyticsRepo.AssertExpectations(t)
	})

	t.Run("Error", func(t *testing.T) {
		analyticsRepo.On("GetWeddingAnalytics", ctx, weddingID).Return(nil, assert.AnError)

		result, err := service.GetWeddingAnalytics(ctx, weddingID)
		require.Error(t, err)
		assert.Nil(t, result)

		analyticsRepo.AssertExpectations(t)
	})
}

func TestAnalyticsService_GetSystemAnalytics(t *testing.T) {
	analyticsRepo := &MockAnalyticsRepository{}
	weddingRepo := &MockWeddingRepository{}
	logger := zaptest.NewLogger(t)

	service := NewAnalyticsService(analyticsRepo, weddingRepo, logger)

	ctx := context.Background()

	expectedAnalytics := &models.SystemAnalytics{
		TotalUsers:        1000,
		TotalWeddings:     500,
		TotalRSVPs:        2000,
		ActiveWeddings:    100,
		PublishedWeddings: 250,
		NewUsersToday:     10,
		NewWeddingsToday:  5,
		NewRSVPsToday:     25,
		TotalPageViews:    50000,
		StorageUsed:       1024 * 1024 * 100, // 100MB
		LastUpdated:       time.Now(),
	}

	t.Run("Success", func(t *testing.T) {
		analyticsRepo.On("GetSystemAnalytics", ctx).Return(expectedAnalytics, nil)

		result, err := service.GetSystemAnalytics(ctx)
		require.NoError(t, err)
		assert.Equal(t, expectedAnalytics, result)

		analyticsRepo.AssertExpectations(t)
	})

	t.Run("Error", func(t *testing.T) {
		analyticsRepo.On("GetSystemAnalytics", ctx).Return(nil, assert.AnError)

		result, err := service.GetSystemAnalytics(ctx)
		require.Error(t, err)
		assert.Nil(t, result)

		analyticsRepo.AssertExpectations(t)
	})
}

func TestAnalyticsService_RefreshWeddingAnalytics(t *testing.T) {
	analyticsRepo := &MockAnalyticsRepository{}
	weddingRepo := &MockWeddingRepository{}
	logger := zaptest.NewLogger(t)

	service := NewAnalyticsService(analyticsRepo, weddingRepo, logger)

	ctx := context.Background()
	weddingID := primitive.NewObjectID()

	t.Run("Success", func(t *testing.T) {
		analyticsRepo.On("UpdateWeddingAnalytics", ctx, weddingID).Return(nil)

		err := service.RefreshWeddingAnalytics(ctx, weddingID)
		require.NoError(t, err)

		analyticsRepo.AssertExpectations(t)
	})

	t.Run("Error", func(t *testing.T) {
		analyticsRepo.On("UpdateWeddingAnalytics", ctx, weddingID).Return(assert.AnError)

		err := service.RefreshWeddingAnalytics(ctx, weddingID)
		require.Error(t, err)

		analyticsRepo.AssertExpectations(t)
	})
}

func TestAnalyticsService_RefreshSystemAnalytics(t *testing.T) {
	analyticsRepo := &MockAnalyticsRepository{}
	weddingRepo := &MockWeddingRepository{}
	logger := zaptest.NewLogger(t)

	service := NewAnalyticsService(analyticsRepo, weddingRepo, logger)

	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		analyticsRepo.On("UpdateSystemAnalytics", ctx).Return(nil)

		err := service.RefreshSystemAnalytics(ctx)
		require.NoError(t, err)

		analyticsRepo.AssertExpectations(t)
	})

	t.Run("Error", func(t *testing.T) {
		analyticsRepo.On("UpdateSystemAnalytics", ctx).Return(assert.AnError)

		err := service.RefreshSystemAnalytics(ctx)
		require.Error(t, err)

		analyticsRepo.AssertExpectations(t)
	})
}

func TestAnalyticsService_CleanupOldAnalytics(t *testing.T) {
	analyticsRepo := &MockAnalyticsRepository{}
	weddingRepo := &MockWeddingRepository{}
	logger := zaptest.NewLogger(t)

	service := NewAnalyticsService(analyticsRepo, weddingRepo, logger)

	ctx := context.Background()
	olderThan := time.Now().AddDate(0, 0, -30) // 30 days ago

	t.Run("Success", func(t *testing.T) {
		analyticsRepo.On("CleanupOldAnalytics", ctx, olderThan).Return(nil)

		err := service.CleanupOldAnalytics(ctx, olderThan)
		require.NoError(t, err)

		analyticsRepo.AssertExpectations(t)
	})

	t.Run("Error", func(t *testing.T) {
		analyticsRepo.On("CleanupOldAnalytics", ctx, olderThan).Return(assert.AnError)

		err := service.CleanupOldAnalytics(ctx, olderThan)
		require.Error(t, err)

		analyticsRepo.AssertExpectations(t)
	})
}

func TestAnalyticsService_GetAnalyticsSummary(t *testing.T) {
	analyticsRepo := &MockAnalyticsRepository{}
	weddingRepo := &MockWeddingRepository{}
	logger := zaptest.NewLogger(t)

	service := NewAnalyticsService(analyticsRepo, weddingRepo, logger)

	ctx := context.Background()
	weddingID := primitive.NewObjectID()
	period := "weekly"

	expectedSummary := &models.AnalyticsSummary{
		Period:          period,
		TotalPageViews:  1000,
		TotalSessions:   500,
		TotalRSVPs:      100,
		ConversionRate:  10.0,
		TopPages:        []models.PageStats{},
		TopSources:      []models.TrafficSourceStats{},
		DeviceBreakdown: map[string]int64{"mobile": 600, "desktop": 400},
		DailyMetrics:    []models.DailyMetrics{},
	}

	t.Run("Success", func(t *testing.T) {
		analyticsRepo.On("GetAnalyticsSummary", ctx, weddingID, period).Return(expectedSummary, nil)

		result, err := service.GetAnalyticsSummary(ctx, weddingID, period)
		require.NoError(t, err)
		assert.Equal(t, expectedSummary, result)

		analyticsRepo.AssertExpectations(t)
	})

	t.Run("Error", func(t *testing.T) {
		analyticsRepo.On("GetAnalyticsSummary", ctx, weddingID, period).Return(nil, assert.AnError)

		result, err := service.GetAnalyticsSummary(ctx, weddingID, period)
		require.Error(t, err)
		assert.Nil(t, result)

		analyticsRepo.AssertExpectations(t)
	})
}

func TestAnalyticsService_HelperMethods(t *testing.T) {
	service := &analyticsService{}

	t.Run("GenerateSessionID", func(t *testing.T) {
		sessionID1 := service.GenerateSessionID()
		sessionID2 := service.GenerateSessionID()

		assert.NotEmpty(t, sessionID1)
		assert.NotEmpty(t, sessionID2)
		assert.NotEqual(t, sessionID1, sessionID2)
		assert.Len(t, sessionID1, 24) // Hex string length
		assert.Len(t, sessionID2, 24)
	})

	t.Run("IsValidPage", func(t *testing.T) {
		assert.True(t, service.IsValidPage("invitation"))
		assert.True(t, service.IsValidPage("rsvp"))
		assert.True(t, service.IsValidPage("gallery"))
		assert.False(t, service.IsValidPage("invalid_page"))
		assert.False(t, service.IsValidPage(""))
	})

	t.Run("IsValidEvent", func(t *testing.T) {
		assert.True(t, service.IsValidEvent("rsvp_started"))
		assert.True(t, service.IsValidEvent("rsvp_completed"))
		assert.True(t, service.IsValidEvent("share_clicked"))
		assert.False(t, service.IsValidEvent("invalid_event"))
		assert.False(t, service.IsValidEvent(""))
	})

	t.Run("ValidatePeriod", func(t *testing.T) {
		assert.True(t, service.ValidatePeriod("daily"))
		assert.True(t, service.ValidatePeriod("weekly"))
		assert.True(t, service.ValidatePeriod("monthly"))
		assert.False(t, service.ValidatePeriod("invalid_period"))
		assert.False(t, service.ValidatePeriod(""))
	})

	t.Run("SanitizeReferrer", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected string
		}{
			{"https://example.com/path?param=value", "https://example.com/path"},
			{"https://example.com/path#section", "https://example.com/path"},
			{"", ""},
			{"https://example.com/" + string(make([]byte, 600, 600)), "https://example.com/" + string(make([]byte, 500, 500))},
		}

		for _, tc := range testCases {
			result := service.SanitizeReferrer(tc.input)
			assert.Equal(t, tc.expected, result)
		}
	})

	t.Run("ExtractSourceFromReferrer", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected string
		}{
			{"", "direct"},
			{"https://google.com/search?q=test", "google"},
			{"https://facebook.com/sharer", "facebook"},
			{"https://instagram.com/p/abc123", "instagram"},
			{"https://twitter.com/intent/tweet", "twitter"},
			{"https://linkedin.com/sharing/share-offsite/", "linkedin"},
			{"https://pinterest.com/pin/123", "pinterest"},
			{"https://youtube.com/watch?v=abc123", "youtube"},
			{"https://example.com", "referral"},
		}

		for _, tc := range testCases {
			result := service.ExtractSourceFromReferrer(tc.input)
			assert.Equal(t, tc.expected, result)
		}
	})

	t.Run("SanitizeCustomData", func(t *testing.T) {
		input := map[string]interface{}{
			"valid_key":    "valid_value",
			"invalid key!": "value1",
			"key@special":  "value2",
			"very_long_key_name_that_exceeds_fifty_characters_limit": "short_value",
			"normal_key": "very_long_value_that_exceeds_two_hundred_characters_limit_and_should_be_truncated_to_exactly_two_hundred_characters" + string(make([]byte, 50, 50)),
		}

		result := service.SanitizeCustomData(input)

		assert.Equal(t, "valid_value", result["valid_key"])
		assert.Equal(t, "value1", result["invalid_key_"])
		assert.Equal(t, "value2", result["key_special"])
		assert.Equal(t, "short_value", result["very_long_key_name_that_exceeds_fifty_characters_"])
		assert.Equal(t, 250, len(result["normal_key"].(string))) // 200 + 50 padding
	})

	t.Run("ParseUserAgent", func(t *testing.T) {
		testCases := []struct {
			userAgent       string
			expectedDevice  string
			expectedBrowser string
			expectedOS      string
		}{
			{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36", "desktop", "chrome", "windows"},
			{"Mozilla/5.0 (iPhone; CPU iPhone OS 14_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1.1 Mobile/15E148 Safari/604.1", "mobile", "safari", "ios"},
			{"Mozilla/5.0 (iPad; CPU OS 14_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1.1 Mobile/15E148 Safari/604.1", "tablet", "safari", "ios"},
			{"Mozilla/5.0 (X11; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0", "desktop", "firefox", "linux"},
			{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36", "desktop", "chrome", "macos"},
			{"", "unknown", "unknown", "unknown"},
		}

		for _, tc := range testCases {
			device, browser, os := service.parseUserAgent(tc.userAgent)
			assert.Equal(t, tc.expectedDevice, device, "Device mismatch for: %s", tc.userAgent)
			assert.Equal(t, tc.expectedBrowser, browser, "Browser mismatch for: %s", tc.userAgent)
			assert.Equal(t, tc.expectedOS, os, "OS mismatch for: %s", tc.userAgent)
		}
	})

	t.Run("GetClientIP", func(t *testing.T) {
		testCases := []struct {
			name       string
			headers    map[string]string
			remoteAddr string
			expectedIP string
		}{
			{
				name: "X-Forwarded-For header",
				headers: map[string]string{
					"X-Forwarded-For": "192.168.1.100, 10.0.0.1",
				},
				remoteAddr: "127.0.0.1:8080",
				expectedIP: "192.168.1.100",
			},
			{
				name: "X-Real-IP header",
				headers: map[string]string{
					"X-Real-IP": "192.168.1.200",
				},
				remoteAddr: "127.0.0.1:8080",
				expectedIP: "192.168.1.200",
			},
			{
				name:       "RemoteAddr only",
				headers:    map[string]string{},
				remoteAddr: "192.168.1.300:8080",
				expectedIP: "192.168.1.300",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				req := httptest.NewRequest("GET", "/test", nil)
				req.RemoteAddr = tc.remoteAddr
				for key, value := range tc.headers {
					req.Header.Set(key, value)
				}

				ip := service.getClientIP(req)
				assert.Equal(t, tc.expectedIP, ip)
			})
		}
	})

	t.Run("GetGeoLocation", func(t *testing.T) {
		// This is a placeholder implementation, so it should return empty values
		country, city := service.getGeoLocation("192.168.1.1")
		assert.Equal(t, "", country)
		assert.Equal(t, "", city)

		country, city = service.getGeoLocation("")
		assert.Equal(t, "", country)
		assert.Equal(t, "", city)

		country, city = service.getGeoLocation("127.0.0.1")
		assert.Equal(t, "", country)
		assert.Equal(t, "", city)
	})
}
