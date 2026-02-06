package mongodb

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"wedding-invitation-backend/internal/domain/models"
)

func TestAnalyticsRepository_TrackPageView(t *testing.T) {
	repo, cleanup := setupTestAnalyticsRepository(t)
	defer cleanup()

	ctx := context.Background()
	weddingID := primitive.NewObjectID()
	sessionID := "test-session-123"

	pageView := &models.PageView{
		WeddingID: weddingID,
		SessionID: sessionID,
		IPAddress: "127.0.0.1",
		UserAgent: "Mozilla/5.0 (Test Browser)",
		Referrer:  "https://google.com",
		Page:      "invitation",
		Timestamp: time.Now(),
		Device:    "desktop",
		Browser:   "chrome",
		OS:        "windows",
		Country:   "US",
		City:      "New York",
	}

	err := repo.TrackPageView(ctx, pageView)
	require.NoError(t, err)
	assert.NotEmpty(t, pageView.ID)

	// Verify the page view was saved
	pageViews, total, err := repo.GetPageViews(ctx, weddingID, &models.AnalyticsFilter{
		Limit: 10,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, pageViews, 1)
	assert.Equal(t, weddingID, pageViews[0].WeddingID)
	assert.Equal(t, sessionID, pageViews[0].SessionID)
	assert.Equal(t, "invitation", pageViews[0].Page)
}

func TestAnalyticsRepository_GetPageViews(t *testing.T) {
	repo, cleanup := setupTestAnalyticsRepository(t)
	defer cleanup()

	ctx := context.Background()
	weddingID := primitive.NewObjectID()

	// Create test data
	now := time.Now()
	pageViews := []*models.PageView{
		{
			WeddingID: weddingID,
			SessionID: "session1",
			Page:      "invitation",
			Timestamp: now.Add(-2 * time.Hour),
			Device:    "desktop",
		},
		{
			WeddingID: weddingID,
			SessionID: "session2",
			Page:      "rsvp",
			Timestamp: now.Add(-1 * time.Hour),
			Device:    "mobile",
		},
		{
			WeddingID: weddingID,
			SessionID: "session3",
			Page:      "invitation",
			Timestamp: now,
			Device:    "tablet",
		},
	}

	// Insert page views
	for _, pv := range pageViews {
		err := repo.TrackPageView(ctx, pv)
		require.NoError(t, err)
	}

	t.Run("Get all page views", func(t *testing.T) {
		results, total, err := repo.GetPageViews(ctx, weddingID, &models.AnalyticsFilter{
			Limit: 10,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, results, 3)
	})

	t.Run("Filter by device", func(t *testing.T) {
		results, total, err := repo.GetPageViews(ctx, weddingID, &models.AnalyticsFilter{
			Device: "desktop",
			Limit:  10,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, results, 1)
		assert.Equal(t, "desktop", results[0].Device)
	})

	t.Run("Filter by page", func(t *testing.T) {
		results, total, err := repo.GetPageViews(ctx, weddingID, &models.AnalyticsFilter{
			Page:  "invitation",
			Limit: 10,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, results, 2)
		for _, result := range results {
			assert.Equal(t, "invitation", result.Page)
		}
	})

	t.Run("Filter by date range", func(t *testing.T) {
		startDate := now.Add(-90 * time.Minute)
		endDate := now.Add(-30 * time.Minute)
		results, total, err := repo.GetPageViews(ctx, weddingID, &models.AnalyticsFilter{
			StartDate: &startDate,
			EndDate:   &endDate,
			Limit:     10,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, results, 1)
		assert.Equal(t, "rsvp", results[0].Page)
	})

	t.Run("Pagination", func(t *testing.T) {
		// Get first page
		results1, total1, err := repo.GetPageViews(ctx, weddingID, &models.AnalyticsFilter{
			Limit:  2,
			Offset: 0,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(3), total1)
		assert.Len(t, results1, 2)

		// Get second page
		results2, total2, err := repo.GetPageViews(ctx, weddingID, &models.AnalyticsFilter{
			Limit:  2,
			Offset: 2,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(3), total2)
		assert.Len(t, results2, 1)
	})
}

func TestAnalyticsRepository_TrackRSVPEvent(t *testing.T) {
	repo, cleanup := setupTestAnalyticsRepository(t)
	defer cleanup()

	ctx := context.Background()
	weddingID := primitive.NewObjectID()
	rsvpID := primitive.NewObjectID()
	sessionID := "test-session-123"

	event := &models.RSVPAnalytics{
		WeddingID:      weddingID,
		RSVPID:         rsvpID,
		SessionID:      sessionID,
		TimeToComplete: 120, // 2 minutes
		Source:         "web",
		Device:         "desktop",
		Browser:        "chrome",
		Referrer:       "https://google.com",
		Timestamp:      time.Now(),
	}

	err := repo.TrackRSVPEvent(ctx, event)
	require.NoError(t, err)
	assert.NotEmpty(t, event.ID)

	// Verify the event was saved
	events, total, err := repo.GetRSVPAnalytics(ctx, weddingID, &models.AnalyticsFilter{
		Limit: 10,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, events, 1)
	assert.Equal(t, weddingID, events[0].WeddingID)
	assert.Equal(t, rsvpID, events[0].RSVPID)
	assert.Equal(t, sessionID, events[0].SessionID)
	assert.Equal(t, "web", events[0].Source)
	assert.Equal(t, int64(120), events[0].TimeToComplete)
}

func TestAnalyticsRepository_TrackConversion(t *testing.T) {
	repo, cleanup := setupTestAnalyticsRepository(t)
	defer cleanup()

	ctx := context.Background()
	weddingID := primitive.NewObjectID()
	sessionID := "test-session-123"

	conversion := &models.ConversionEvent{
		WeddingID:  weddingID,
		SessionID:  sessionID,
		Event:      "rsvp_completed",
		Value:      1.0,
		Currency:   "USD",
		Timestamp:  time.Now(),
		Properties: map[string]interface{}{
			"source": "web",
			"page":   "rsvp",
		},
	}

	err := repo.TrackConversion(ctx, conversion)
	require.NoError(t, err)
	assert.NotEmpty(t, conversion.ID)

	// Verify the conversion was saved
	conversions, total, err := repo.GetConversions(ctx, weddingID, &models.AnalyticsFilter{
		Limit: 10,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, conversions, 1)
	assert.Equal(t, weddingID, conversions[0].WeddingID)
	assert.Equal(t, sessionID, conversions[0].SessionID)
	assert.Equal(t, "rsvp_completed", conversions[0].Event)
	assert.Equal(t, 1.0, conversions[0].Value)
}

func TestAnalyticsRepository_WeddingAnalytics(t *testing.T) {
	repo, cleanup := setupTestAnalyticsRepository(t)
	defer cleanup()

	ctx := context.Background()
	weddingID := primitive.NewObjectID()

	t.Run("Get analytics for new wedding", func(t *testing.T) {
		analytics, err := repo.GetWeddingAnalytics(ctx, weddingID)
		require.NoError(t, err)
		assert.Equal(t, weddingID, analytics.WeddingID)
		assert.Equal(t, int64(0), analytics.PageViews)
		assert.Equal(t, int64(0), analytics.UniqueSessions)
		assert.Equal(t, int64(0), analytics.RSVPCount)
		assert.Equal(t, float64(0), analytics.ConversionRate)
	})

	t.Run("Update analytics with data", func(t *testing.T) {
		// Create some test data
		for i := 0; i < 3; i++ {
			pageView := &models.PageView{
				WeddingID: weddingID,
				SessionID: primitive.NewObjectID().Hex(),
				Page:      "invitation",
				Timestamp: time.Now(),
				Device:    "desktop",
			}
			err := repo.TrackPageView(ctx, pageView)
			require.NoError(t, err)
		}

		// Add RSVP analytics
		rsvpEvent := &models.RSVPAnalytics{
			WeddingID: weddingID,
			RSVPID:    primitive.NewObjectID(),
			SessionID: primitive.NewObjectID().Hex(),
			Source:    "web",
			Timestamp: time.Now(),
		}
		err := repo.TrackRSVPEvent(ctx, rsvpEvent)
		require.NoError(t, err)

		// Update analytics
		err = repo.UpdateWeddingAnalytics(ctx, weddingID)
		require.NoError(t, err)

		// Verify updated analytics
		analytics, err := repo.GetWeddingAnalytics(ctx, weddingID)
		require.NoError(t, err)
		assert.Equal(t, weddingID, analytics.WeddingID)
		assert.Greater(t, analytics.PageViews, int64(0))
		assert.Greater(t, analytics.UniqueSessions, int64(0))
		assert.Greater(t, analytics.RSVPCount, int64(0))
	})
}

func TestAnalyticsRepository_SystemAnalytics(t *testing.T) {
	repo, cleanup := setupTestAnalyticsRepository(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("Get system analytics initially", func(t *testing.T) {
		analytics, err := repo.GetSystemAnalytics(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(0), analytics.TotalUsers)
		assert.Equal(t, int64(0), analytics.TotalWeddings)
		assert.Equal(t, int64(0), analytics.TotalRSVPs)
		assert.Equal(t, int64(0), analytics.NewUsersToday)
	})

	t.Run("Update system analytics", func(t *testing.T) {
		err := repo.UpdateSystemAnalytics(ctx)
		require.NoError(t, err)

		analytics, err := repo.GetSystemAnalytics(ctx)
		require.NoError(t, err)
		// Should have some data now, even if it's just 0 for empty collections
		assert.NotNil(t, analytics.LastUpdated)
	})
}

func TestAnalyticsRepository_GetPopularPages(t *testing.T) {
	repo, cleanup := setupTestAnalyticsRepository(t)
	defer cleanup()

	ctx := context.Background()
	weddingID := primitive.NewObjectID()

	// Create test page views with different pages
	pages := []string{"invitation", "rsvp", "gallery", "invitation", "rsvp", "invitation"}
	for i, page := range pages {
		pageView := &models.PageView{
			WeddingID: weddingID,
			SessionID: primitive.NewObjectID().Hex(),
			Page:      page,
			Timestamp: time.Now().Add(time.Duration(i) * time.Minute),
			Device:    "desktop",
			Duration:  int64(30 + i*10), // Varying durations
		}
		err := repo.TrackPageView(ctx, pageView)
		require.NoError(t, err)
	}

	popularPages, err := repo.GetPopularPages(ctx, weddingID, 5)
	require.NoError(t, err)
	assert.Len(t, popularPages, 3) // Should have 3 unique pages

	// Verify ordering (invitation should be first with 3 views)
	assert.Equal(t, "invitation", popularPages[0].Page)
	assert.Equal(t, int64(3), popularPages[0].Views)

	// Verify other pages
	rsvpPage := findPageStats(popularPages, "rsvp")
	require.NotNil(t, rsvpPage)
	assert.Equal(t, int64(2), rsvpPage.Views)

	galleryPage := findPageStats(popularPages, "gallery")
	require.NotNil(t, galleryPage)
	assert.Equal(t, int64(1), galleryPage.Views)
}

func TestAnalyticsRepository_GetTrafficSources(t *testing.T) {
	repo, cleanup := setupTestAnalyticsRepository(t)
	defer cleanup()

	ctx := context.Background()
	weddingID := primitive.NewObjectID()

	// Create test page views with different referrers
	referrers := []string{
		"https://google.com",
		"https://facebook.com",
		"https://google.com",
		"https://instagram.com",
	}

	for i, referrer := range referrers {
		pageView := &models.PageView{
			WeddingID: weddingID,
			SessionID: primitive.NewObjectID().Hex(),
			Page:      "invitation",
			Referrer:  referrer,
			Timestamp: time.Now().Add(time.Duration(i) * time.Minute),
		}
		err := repo.TrackPageView(ctx, pageView)
		require.NoError(t, err)
	}

	trafficSources, err := repo.GetTrafficSources(ctx, weddingID, 10)
	require.NoError(t, err)
	assert.Len(t, trafficSources, 3) // Should have 3 unique referrers

	// Verify Google is first with 2 visitors
	assert.Equal(t, "https://google.com", trafficSources[0].Source)
	assert.Equal(t, int64(2), trafficSources[0].Views)

	// Verify other sources
	facebookSource := findTrafficSourceStats(trafficSources, "https://facebook.com")
	require.NotNil(t, facebookSource)
	assert.Equal(t, int64(1), facebookSource.Views)

	instaSource := findTrafficSourceStats(trafficSources, "https://instagram.com")
	require.NotNil(t, instaSource)
	assert.Equal(t, int64(1), instaSource.Views)
}

func TestAnalyticsRepository_GetDailyMetrics(t *testing.T) {
	repo, cleanup := setupTestAnalyticsRepository(t)
	defer cleanup()

	ctx := context.Background()
	weddingID := primitive.NewObjectID()

	now := time.Now()
	startDate := now.AddDate(0, 0, -7) // 7 days ago
	endDate := now

	// Create test data spanning multiple days
	for i := 0; i < 7; i++ {
		date := startDate.AddDate(0, 0, i)
		
		// Create 2-3 page views per day
		for j := 0; j < 2+i%2; j++ {
			pageView := &models.PageView{
				WeddingID: weddingID,
				SessionID: primitive.NewObjectID().Hex(),
				Page:      "invitation",
				Timestamp: date,
				Device:    "desktop",
			}
			err := repo.TrackPageView(ctx, pageView)
			require.NoError(t, err)
		}

		// Create 0-1 RSVP events per day
		if i%3 == 0 {
			rsvpEvent := &models.RSVPAnalytics{
				WeddingID: weddingID,
				RSVPID:    primitive.NewObjectID(),
				SessionID: primitive.NewObjectID().Hex(),
				Source:    "web",
				Timestamp: date,
			}
			err := repo.TrackRSVPEvent(ctx, rsvpEvent)
			require.NoError(t, err)
		}
	}

	dailyMetrics, err := repo.GetDailyMetrics(ctx, weddingID, startDate, endDate)
	require.NoError(t, err)
	assert.Greater(t, len(dailyMetrics), 0)

	// Verify we have data for each day
	for _, metric := range dailyMetrics {
		assert.NotEmpty(t, metric.Date)
		assert.GreaterOrEqual(t, metric.PageViews, int64(0))
		assert.GreaterOrEqual(t, metric.Sessions, int64(0))
		assert.GreaterOrEqual(t, metric.RSVPs, int64(0))
		assert.GreaterOrEqual(t, metric.Conversions, float64(0))
	}
}

func TestAnalyticsRepository_CleanupOldAnalytics(t *testing.T) {
	repo, cleanup := setupTestAnalyticsRepository(t)
	defer cleanup()

	ctx := context.Background()
	weddingID := primitive.NewObjectID()

	now := time.Now()
	oldDate := now.AddDate(0, 0, -90) // 90 days ago

	// Create recent analytics
	recentPageView := &models.PageView{
		WeddingID: weddingID,
		SessionID: "recent-session",
		Page:      "invitation",
		Timestamp: now,
		Device:    "desktop",
	}
	err := repo.TrackPageView(ctx, recentPageView)
	require.NoError(t, err)

	// Create old analytics
	oldPageView := &models.PageView{
		WeddingID: weddingID,
		SessionID: "old-session",
		Page:      "invitation",
		Timestamp: oldDate,
		Device:    "desktop",
	}
	err = repo.TrackPageView(ctx, oldPageView)
	require.NoError(t, err)

	// Verify we have 2 page views initially
	pageViews, total, err := repo.GetPageViews(ctx, weddingID, &models.AnalyticsFilter{Limit: 100})
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)

	// Cleanup old analytics
	cutoffDate := now.AddDate(0, 0, -30) // 30 days ago
	err = repo.CleanupOldAnalytics(ctx, cutoffDate)
	require.NoError(t, err)

	// Verify only recent page views remain
	pageViews, total, err = repo.GetPageViews(ctx, weddingID, &models.AnalyticsFilter{Limit: 100})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, pageViews, 1)
	assert.Equal(t, "recent-session", pageViews[0].SessionID)
}

// Helper functions

func setupTestAnalyticsRepository(t *testing.T) (AnalyticsRepository, func()) {
	db, cleanup := setupTestDB(t)
	repo := NewAnalyticsRepository(db.Database)
	return repo, cleanup
}

func findPageStats(stats []models.PageStats, page string) *models.PageStats {
	for _, stat := range stats {
		if stat.Page == page {
			return &stat
		}
	}
	return nil
}

func findTrafficSourceStats(sources []models.TrafficSourceStats, source string) *models.TrafficSourceStats {
	for _, stat := range sources {
		if stat.Source == source {
			return &stat
		}
	}
	return nil
}