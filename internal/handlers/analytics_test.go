package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"wedding-invitation-backend/internal/domain/models"
	serviceMocks "wedding-invitation-backend/test/mocks/services"
)

func TestAnalyticsHandler_TrackPageView(t *testing.T) {
	gin.SetMode(gin.TestMode)

	analyticsService := &serviceMocks.MockAnalyticsService{}
	weddingService := &serviceMocks.MockWeddingService{}
	handler := NewAnalyticsHandler(analyticsService, weddingService)

	t.Run("Success", func(t *testing.T) {
		weddingID := primitive.NewObjectID()
		reqBody := TrackPageViewRequest{
			WeddingID: weddingID.Hex(),
			SessionID: "test-session-123",
			Page:      "invitation",
		}

		analyticsService.On("IsValidPage", "invitation").Return(true)
		analyticsService.On("TrackPageView", mock.Anything, weddingID, "test-session-123", "invitation", mock.AnythingOfType("*http.Request")).Return(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/analytics/track/page-view", bytes.NewBufferString(toJSON(reqBody)))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.TrackPageView(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		analyticsService.AssertExpectations(t)
	})

	t.Run("Invalid request body", func(t *testing.T) {
		reqBody := `{"invalid": "json"`

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/analytics/track/page-view", bytes.NewBufferString(reqBody))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.TrackPageView(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Invalid wedding ID", func(t *testing.T) {
		reqBody := TrackPageViewRequest{
			WeddingID: "invalid-id",
			SessionID: "test-session-123",
			Page:      "invitation",
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/analytics/track/page-view", bytes.NewBufferString(toJSON(reqBody)))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.TrackPageView(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Invalid page", func(t *testing.T) {
		weddingID := primitive.NewObjectID()
		reqBody := TrackPageViewRequest{
			WeddingID: weddingID.Hex(),
			SessionID: "test-session-123",
			Page:      "invalid_page",
		}

		analyticsService.On("IsValidPage", "invalid_page").Return(false)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/analytics/track/page-view", bytes.NewBufferString(toJSON(reqBody)))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.TrackPageView(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		analyticsService.AssertExpectations(t)
	})

	t.Run("Wedding not found", func(t *testing.T) {
		weddingID := primitive.NewObjectID()
		reqBody := TrackPageViewRequest{
			WeddingID: weddingID.Hex(),
			SessionID: "test-session-123",
			Page:      "invitation",
		}

		analyticsService.On("IsValidPage", "invitation").Return(true)
		analyticsService.On("TrackPageView", mock.Anything, weddingID, "test-session-123", "invitation", mock.AnythingOfType("*http.Request")).Return(assert.AnError)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/analytics/track/page-view", bytes.NewBufferString(toJSON(reqBody)))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.TrackPageView(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		analyticsService.AssertExpectations(t)
	})
}

func TestAnalyticsHandler_TrackRSVPSubmission(t *testing.T) {
	gin.SetMode(gin.TestMode)

	analyticsService := &serviceMocks.MockAnalyticsService{}
	weddingService := &serviceMocks.MockWeddingService{}
	handler := NewAnalyticsHandler(analyticsService, weddingService)

	t.Run("Success", func(t *testing.T) {
		weddingID := primitive.NewObjectID()
		rsvpID := primitive.NewObjectID()
		reqBody := TrackRSVPSubmissionRequest{
			WeddingID:      weddingID.Hex(),
			RSVPID:         rsvpID.Hex(),
			SessionID:      "test-session-123",
			Source:         "web",
			TimeToComplete: 120,
		}

		analyticsService.On("TrackRSVPSubmission", mock.Anything, weddingID, rsvpID, "test-session-123", "web", int64(120), mock.AnythingOfType("*http.Request")).Return(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/analytics/track/rsvp-submission", bytes.NewBufferString(toJSON(reqBody)))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.TrackRSVPSubmission(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		analyticsService.AssertExpectations(t)
	})

	t.Run("Invalid source", func(t *testing.T) {
		weddingID := primitive.NewObjectID()
		rsvpID := primitive.NewObjectID()
		reqBody := TrackRSVPSubmissionRequest{
			WeddingID:      weddingID.Hex(),
			RSVPID:         rsvpID.Hex(),
			SessionID:      "test-session-123",
			Source:         "invalid_source",
			TimeToComplete: 120,
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/analytics/track/rsvp-submission", bytes.NewBufferString(toJSON(reqBody)))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.TrackRSVPSubmission(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Invalid wedding ID", func(t *testing.T) {
		rsvpID := primitive.NewObjectID()
		reqBody := TrackRSVPSubmissionRequest{
			WeddingID:      "invalid-id",
			RSVPID:         rsvpID.Hex(),
			SessionID:      "test-session-123",
			Source:         "web",
			TimeToComplete: 120,
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/analytics/track/rsvp-submission", bytes.NewBufferString(toJSON(reqBody)))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.TrackRSVPSubmission(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestAnalyticsHandler_TrackRSVPAbandonment(t *testing.T) {
	gin.SetMode(gin.TestMode)

	analyticsService := &serviceMocks.MockAnalyticsService{}
	weddingService := &serviceMocks.MockWeddingService{}
	handler := NewAnalyticsHandler(analyticsService, weddingService)

	t.Run("Success", func(t *testing.T) {
		weddingID := primitive.NewObjectID()
		reqBody := TrackRSVPAbandonmentRequest{
			WeddingID:     weddingID.Hex(),
			SessionID:     "test-session-123",
			AbandonedStep: "personal_info",
			FormErrors:    []string{"Invalid email", "Name required"},
		}

		analyticsService.On("TrackRSVPAbandonment", mock.Anything, weddingID, "test-session-123", "personal_info", []string{"Invalid email", "Name required"}, mock.AnythingOfType("*http.Request")).Return(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/analytics/track/rsvp-abandonment", bytes.NewBufferString(toJSON(reqBody)))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.TrackRSVPAbandonment(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		analyticsService.AssertExpectations(t)
	})

	t.Run("Invalid abandoned step", func(t *testing.T) {
		weddingID := primitive.NewObjectID()
		reqBody := TrackRSVPAbandonmentRequest{
			WeddingID:     weddingID.Hex(),
			SessionID:     "test-session-123",
			AbandonedStep: "invalid_step",
			FormErrors:    []string{"Error"},
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/analytics/track/rsvp-abandonment", bytes.NewBufferString(toJSON(reqBody)))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.TrackRSVPAbandonment(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestAnalyticsHandler_TrackConversion(t *testing.T) {
	gin.SetMode(gin.TestMode)

	analyticsService := &serviceMocks.MockAnalyticsService{}
	weddingService := &serviceMocks.MockWeddingService{}
	handler := NewAnalyticsHandler(analyticsService, weddingService)

	t.Run("Success", func(t *testing.T) {
		weddingID := primitive.NewObjectID()
		properties := map[string]interface{}{"platform": "facebook"}
		reqBody := TrackConversionRequest{
			WeddingID:  weddingID.Hex(),
			SessionID:  "test-session-123",
			Event:      "share_clicked",
			Value:      1.0,
			Properties: properties,
		}

		analyticsService.On("IsValidEvent", "share_clicked").Return(true)
		analyticsService.On("SanitizeCustomData", properties).Return(properties)
		analyticsService.On("TrackConversion", mock.Anything, weddingID, "test-session-123", "share_clicked", 1.0, properties).Return(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/analytics/track/conversion", bytes.NewBufferString(toJSON(reqBody)))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.TrackConversion(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		analyticsService.AssertExpectations(t)
	})

	t.Run("Invalid event", func(t *testing.T) {
		weddingID := primitive.NewObjectID()
		reqBody := TrackConversionRequest{
			WeddingID: weddingID.Hex(),
			SessionID: "test-session-123",
			Event:     "invalid_event",
			Value:     1.0,
		}

		analyticsService.On("IsValidEvent", "invalid_event").Return(false)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/analytics/track/conversion", bytes.NewBufferString(toJSON(reqBody)))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.TrackConversion(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		analyticsService.AssertExpectations(t)
	})
}

func TestAnalyticsHandler_GetWeddingAnalytics(t *testing.T) {
	gin.SetMode(gin.TestMode)

	analyticsService := &serviceMocks.MockAnalyticsService{}
	weddingService := &serviceMocks.MockWeddingService{}
	handler := NewAnalyticsHandler(analyticsService, weddingService)

	weddingID := primitive.NewObjectID()
	userID := primitive.NewObjectID()
	expectedAnalytics := &models.WeddingAnalytics{
		WeddingID: weddingID,
		PageViews: 100,
		RSVPCount: 25,
	}

	t.Run("Success", func(t *testing.T) {
		wedding := &models.Wedding{
			ID:     weddingID,
			UserID: userID,
		}

		weddingService.On("GetWeddingByID", mock.Anything, weddingID).Return(wedding, nil)
		analyticsService.On("GetWeddingAnalytics", mock.Anything, weddingID).Return(expectedAnalytics, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/weddings/"+weddingID.Hex()+"/analytics", nil)
		c.Set("user_id", userID.Hex())
		c.Params = gin.Params{gin.Param{Key: "id", Value: weddingID.Hex()}}

		handler.GetWeddingAnalytics(c)

		assert.Equal(t, http.StatusOK, w.Code)
		weddingService.AssertExpectations(t)
		analyticsService.AssertExpectations(t)
	})

	t.Run("Not authenticated", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/weddings/"+weddingID.Hex()+"/analytics", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: weddingID.Hex()}}

		handler.GetWeddingAnalytics(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Access denied", func(t *testing.T) {
		otherUserID := primitive.NewObjectID()
		wedding := &models.Wedding{
			ID:     weddingID,
			UserID: otherUserID, // Different user
		}

		weddingService.On("GetWeddingByID", mock.Anything, weddingID).Return(wedding, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/weddings/"+weddingID.Hex()+"/analytics", nil)
		c.Set("user_id", userID.Hex())
		c.Params = gin.Params{gin.Param{Key: "id", Value: weddingID.Hex()}}

		handler.GetWeddingAnalytics(c)

		assert.Equal(t, http.StatusForbidden, w.Code)
		weddingService.AssertExpectations(t)
	})

	t.Run("Wedding not found", func(t *testing.T) {
		weddingService.On("GetWeddingByID", mock.Anything, weddingID).Return(nil, assert.AnError)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/weddings/"+weddingID.Hex()+"/analytics", nil)
		c.Set("user_id", userID.Hex())
		c.Params = gin.Params{gin.Param{Key: "id", Value: weddingID.Hex()}}

		handler.GetWeddingAnalytics(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		weddingService.AssertExpectations(t)
	})
}

func TestAnalyticsHandler_GetAnalyticsSummary(t *testing.T) {
	gin.SetMode(gin.TestMode)

	analyticsService := &serviceMocks.MockAnalyticsService{}
	weddingService := &serviceMocks.MockWeddingService{}
	handler := NewAnalyticsHandler(analyticsService, weddingService)

	weddingID := primitive.NewObjectID()
	userID := primitive.NewObjectID()
	expectedSummary := &models.AnalyticsSummary{
		Period:         "daily",
		TotalPageViews: 1000,
		TotalRSVPs:     100,
		ConversionRate: 10.0,
	}

	t.Run("Success", func(t *testing.T) {
		wedding := &models.Wedding{
			ID:     weddingID,
			UserID: userID,
		}

		weddingService.On("GetWeddingByID", mock.Anything, weddingID).Return(wedding, nil)
		analyticsService.On("ValidatePeriod", "daily").Return(true)
		analyticsService.On("GetAnalyticsSummary", mock.Anything, weddingID, "daily").Return(expectedSummary, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/weddings/"+weddingID.Hex()+"/analytics/summary?period=daily", nil)
		c.Set("user_id", userID.Hex())
		c.Params = gin.Params{gin.Param{Key: "id", Value: weddingID.Hex()}}

		handler.GetAnalyticsSummary(c)

		assert.Equal(t, http.StatusOK, w.Code)
		weddingService.AssertExpectations(t)
		analyticsService.AssertExpectations(t)
	})

	t.Run("Invalid period", func(t *testing.T) {
		wedding := &models.Wedding{
			ID:     weddingID,
			UserID: userID,
		}

		weddingService.On("GetWeddingByID", mock.Anything, weddingID).Return(wedding, nil)
		analyticsService.On("ValidatePeriod", "invalid").Return(false)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/weddings/"+weddingID.Hex()+"/analytics/summary?period=invalid", nil)
		c.Set("user_id", userID.Hex())
		c.Params = gin.Params{gin.Param{Key: "id", Value: weddingID.Hex()}}

		handler.GetAnalyticsSummary(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		weddingService.AssertExpectations(t)
		analyticsService.AssertExpectations(t)
	})
}

func TestAnalyticsHandler_GetPageViews(t *testing.T) {
	gin.SetMode(gin.TestMode)

	analyticsService := &serviceMocks.MockAnalyticsService{}
	weddingService := &serviceMocks.MockWeddingService{}
	handler := NewAnalyticsHandler(analyticsService, weddingService)

	weddingID := primitive.NewObjectID()
	userID := primitive.NewObjectID()
	expectedPageViews := []*models.PageView{
		{
			WeddingID: weddingID,
			Page:      "invitation",
			Timestamp: time.Now(),
		},
	}

	t.Run("Success", func(t *testing.T) {
		wedding := &models.Wedding{
			ID:     weddingID,
			UserID: userID,
		}

		weddingService.On("GetWeddingByID", mock.Anything, weddingID).Return(wedding, nil)
		analyticsService.On("GetPageViews", mock.Anything, weddingID, mock.AnythingOfType("*models.AnalyticsFilter")).Return(expectedPageViews, int64(1), nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/weddings/"+weddingID.Hex()+"/analytics/page-views?limit=10&offset=0", nil)
		c.Set("user_id", userID.Hex())
		c.Params = gin.Params{gin.Param{Key: "id", Value: weddingID.Hex()}}

		handler.GetPageViews(c)

		assert.Equal(t, http.StatusOK, w.Code)
		weddingService.AssertExpectations(t)
		analyticsService.AssertExpectations(t)
	})

	t.Run("Invalid date format", func(t *testing.T) {
		wedding := &models.Wedding{
			ID:     weddingID,
			UserID: userID,
		}

		weddingService.On("GetWeddingByID", mock.Anything, weddingID).Return(wedding, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/weddings/"+weddingID.Hex()+"/analytics/page-views?start_date=invalid-date", nil)
		c.Set("user_id", userID.Hex())
		c.Params = gin.Params{gin.Param{Key: "id", Value: weddingID.Hex()}}

		handler.GetPageViews(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		weddingService.AssertExpectations(t)
	})

	t.Run("Invalid limit", func(t *testing.T) {
		wedding := &models.Wedding{
			ID:     weddingID,
			UserID: userID,
		}

		weddingService.On("GetWeddingByID", mock.Anything, weddingID).Return(wedding, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/weddings/"+weddingID.Hex()+"/analytics/page-views?limit=invalid", nil)
		c.Set("user_id", userID.Hex())
		c.Params = gin.Params{gin.Param{Key: "id", Value: weddingID.Hex()}}

		handler.GetPageViews(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		weddingService.AssertExpectations(t)
	})
}

func TestAnalyticsHandler_GetPopularPages(t *testing.T) {
	gin.SetMode(gin.TestMode)

	analyticsService := &serviceMocks.MockAnalyticsService{}
	weddingService := &serviceMocks.MockWeddingService{}
	handler := NewAnalyticsHandler(analyticsService, weddingService)

	weddingID := primitive.NewObjectID()
	userID := primitive.NewObjectID()
	expectedPages := []models.PageStats{
		{Page: "invitation", Views: 100, UniqueViews: 80},
		{Page: "rsvp", Views: 50, UniqueViews: 40},
	}

	t.Run("Success", func(t *testing.T) {
		wedding := &models.Wedding{
			ID:     weddingID,
			UserID: userID,
		}

		weddingService.On("GetWeddingByID", mock.Anything, weddingID).Return(wedding, nil)
		analyticsService.On("GetPopularPages", mock.Anything, weddingID, 10).Return(expectedPages, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/weddings/"+weddingID.Hex()+"/analytics/popular-pages?limit=10", nil)
		c.Set("user_id", userID.Hex())
		c.Params = gin.Params{gin.Param{Key: "id", Value: weddingID.Hex()}}

		handler.GetPopularPages(c)

		assert.Equal(t, http.StatusOK, w.Code)
		weddingService.AssertExpectations(t)
		analyticsService.AssertExpectations(t)
	})

	t.Run("Invalid limit", func(t *testing.T) {
		wedding := &models.Wedding{
			ID:     weddingID,
			UserID: userID,
		}

		weddingService.On("GetWeddingByID", mock.Anything, weddingID).Return(wedding, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/weddings/"+weddingID.Hex()+"/analytics/popular-pages?limit=150", nil)
		c.Set("user_id", userID.Hex())
		c.Params = gin.Params{gin.Param{Key: "id", Value: weddingID.Hex()}}

		handler.GetPopularPages(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		weddingService.AssertExpectations(t)
	})
}

func TestAnalyticsHandler_GetSystemAnalytics(t *testing.T) {
	gin.SetMode(gin.TestMode)

	analyticsService := &serviceMocks.MockAnalyticsService{}
	weddingService := &serviceMocks.MockWeddingService{}
	handler := NewAnalyticsHandler(analyticsService, weddingService)

	expectedAnalytics := &models.SystemAnalytics{
		TotalUsers:    1000,
		TotalWeddings: 500,
		TotalRSVPs:    2000,
	}

	t.Run("Success - admin user", func(t *testing.T) {
		analyticsService.On("GetSystemAnalytics", mock.Anything).Return(expectedAnalytics, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/admin/analytics/system", nil)
		c.Set("is_admin", true)

		handler.GetSystemAnalytics(c)

		assert.Equal(t, http.StatusOK, w.Code)
		analyticsService.AssertExpectations(t)
	})

	t.Run("Access denied - non-admin user", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/admin/analytics/system", nil)
		c.Set("is_admin", false)

		handler.GetSystemAnalytics(c)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Access denied - no admin flag", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/admin/analytics/system", nil)

		handler.GetSystemAnalytics(c)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestAnalyticsHandler_RefreshAnalytics(t *testing.T) {
	gin.SetMode(gin.TestMode)

	analyticsService := &serviceMocks.MockAnalyticsService{}
	weddingService := &serviceMocks.MockWeddingService{}
	handler := NewAnalyticsHandler(analyticsService, weddingService)

	weddingID := primitive.NewObjectID()
	userID := primitive.NewObjectID()

	t.Run("Success", func(t *testing.T) {
		wedding := &models.Wedding{
			ID:     weddingID,
			UserID: userID,
		}

		weddingService.On("GetWeddingByID", mock.Anything, weddingID).Return(wedding, nil)
		analyticsService.On("RefreshWeddingAnalytics", mock.Anything, weddingID).Return(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/weddings/"+weddingID.Hex()+"/analytics/refresh", nil)
		c.Set("user_id", userID.Hex())
		c.Params = gin.Params{gin.Param{Key: "id", Value: weddingID.Hex()}}

		handler.RefreshAnalytics(c)

		assert.Equal(t, http.StatusOK, w.Code)
		weddingService.AssertExpectations(t)
		analyticsService.AssertExpectations(t)
	})

	t.Run("Access denied", func(t *testing.T) {
		otherUserID := primitive.NewObjectID()
		wedding := &models.Wedding{
			ID:     weddingID,
			UserID: otherUserID,
		}

		weddingService.On("GetWeddingByID", mock.Anything, weddingID).Return(wedding, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/weddings/"+weddingID.Hex()+"/analytics/refresh", nil)
		c.Set("user_id", userID.Hex())
		c.Params = gin.Params{gin.Param{Key: "id", Value: weddingID.Hex()}}

		handler.RefreshAnalytics(c)

		assert.Equal(t, http.StatusForbidden, w.Code)
		weddingService.AssertExpectations(t)
	})
}

func TestAnalyticsHandler_RefreshSystemAnalytics(t *testing.T) {
	gin.SetMode(gin.TestMode)

	analyticsService := &serviceMocks.MockAnalyticsService{}
	weddingService := &serviceMocks.MockWeddingService{}
	handler := NewAnalyticsHandler(analyticsService, weddingService)

	t.Run("Success - admin user", func(t *testing.T) {
		analyticsService.On("RefreshSystemAnalytics", mock.Anything).Return(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/admin/analytics/refresh", nil)
		c.Set("is_admin", true)

		handler.RefreshSystemAnalytics(c)

		assert.Equal(t, http.StatusOK, w.Code)
		analyticsService.AssertExpectations(t)
	})

	t.Run("Access denied - non-admin user", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/admin/analytics/refresh", nil)
		c.Set("is_admin", false)

		handler.RefreshSystemAnalytics(c)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestAnalyticsHandler_ValidateAnalyticsFilter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	analyticsService := &serviceMocks.MockAnalyticsService{}
	weddingService := &serviceMocks.MockWeddingService{}
	handler := NewAnalyticsHandler(analyticsService, weddingService)

	t.Run("Valid filter", func(t *testing.T) {
		now := time.Now()
		filter := &AnalyticsFilterRequest{
			StartDate: &now,
			EndDate:   &now.Add(time.Hour),
			Limit:     10,
			Offset:    0,
		}

		err := handler.validateAnalyticsFilter(filter)
		assert.NoError(t, err)
	})

	t.Run("Invalid limit", func(t *testing.T) {
		filter := &AnalyticsFilterRequest{
			Limit: 150, // Exceeds 100
		}

		err := handler.validateAnalyticsFilter(filter)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "limit must be between 0 and 100")
	})

	t.Run("Invalid offset", func(t *testing.T) {
		filter := &AnalyticsFilterRequest{
			Limit:  10,
			Offset: -1,
		}

		err := handler.validateAnalyticsFilter(filter)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "offset must be non-negative")
	})

	t.Run("Start date after end date", func(t *testing.T) {
		now := time.Now()
		filter := &AnalyticsFilterRequest{
			StartDate: &now,
			EndDate:   &now.Add(-time.Hour),
		}

		err := handler.validateAnalyticsFilter(filter)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "start date must be before end date")
	})
}

// Helper function to convert struct to JSON string
func toJSON(v interface{}) string {
	bytes, _ := json.Marshal(v)
	return string(bytes)
}