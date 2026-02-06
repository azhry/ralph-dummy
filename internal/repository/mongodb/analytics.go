package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"wedding-invitation-backend/internal/domain/models"
)

// AnalyticsRepository represents the analytics repository interface
type AnalyticsRepository interface {
	// Page Views
	TrackPageView(ctx context.Context, pageView *models.PageView) error
	GetPageViews(ctx context.Context, weddingID primitive.ObjectID, filter *models.AnalyticsFilter) ([]*models.PageView, int64, error)

	// RSVP Analytics
	TrackRSVPEvent(ctx context.Context, event *models.RSVPAnalytics) error
	GetRSVPAnalytics(ctx context.Context, weddingID primitive.ObjectID, filter *models.AnalyticsFilter) ([]*models.RSVPAnalytics, int64, error)

	// Conversion Events
	TrackConversion(ctx context.Context, event *models.ConversionEvent) error
	GetConversions(ctx context.Context, weddingID primitive.ObjectID, filter *models.AnalyticsFilter) ([]*models.ConversionEvent, int64, error)

	// Aggregated Analytics
	GetWeddingAnalytics(ctx context.Context, weddingID primitive.ObjectID) (*models.WeddingAnalytics, error)
	UpdateWeddingAnalytics(ctx context.Context, weddingID primitive.ObjectID) error

	// System Analytics
	GetSystemAnalytics(ctx context.Context) (*models.SystemAnalytics, error)
	UpdateSystemAnalytics(ctx context.Context) error

	// Reports
	GetAnalyticsSummary(ctx context.Context, weddingID primitive.ObjectID, period string) (*models.AnalyticsSummary, error)
	GetPopularPages(ctx context.Context, weddingID primitive.ObjectID, limit int) ([]models.PageStats, error)
	GetTrafficSources(ctx context.Context, weddingID primitive.ObjectID, limit int) ([]models.TrafficSourceStats, error)
	GetDailyMetrics(ctx context.Context, weddingID primitive.ObjectID, startDate, endDate time.Time) ([]models.DailyMetrics, error)

	// Cleanup
	CleanupOldAnalytics(ctx context.Context, olderThan time.Time) error
}

type analyticsRepository struct {
	db               *mongo.Database
	pageViews        *mongo.Collection
	rsvpEvents       *mongo.Collection
	conversions      *mongo.Collection
	weddingAnalytics *mongo.Collection
	systemAnalytics  *mongo.Collection
}

// NewAnalyticsRepository creates a new analytics repository
func NewAnalyticsRepository(db *mongo.Database) AnalyticsRepository {
	return &analyticsRepository{
		db:               db,
		pageViews:        db.Collection("page_views"),
		rsvpEvents:       db.Collection("rsvp_analytics"),
		conversions:      db.Collection("conversion_events"),
		weddingAnalytics: db.Collection("wedding_analytics"),
		systemAnalytics:  db.Collection("system_analytics"),
	}
}

// TrackPageView records a page view event
func (r *analyticsRepository) TrackPageView(ctx context.Context, pageView *models.PageView) error {
	if pageView.ID.IsZero() {
		pageView.ID = primitive.NewObjectID()
	}
	if pageView.Timestamp.IsZero() {
		pageView.Timestamp = time.Now()
	}

	_, err := r.pageViews.InsertOne(ctx, pageView)
	if err != nil {
		return fmt.Errorf("failed to track page view: %w", err)
	}

	// Update wedding analytics asynchronously
	go func() {
		r.UpdateWeddingAnalytics(context.Background(), pageView.WeddingID)
	}()

	return nil
}

// GetPageViews retrieves page views with filtering
func (r *analyticsRepository) GetPageViews(ctx context.Context, weddingID primitive.ObjectID, filter *models.AnalyticsFilter) ([]*models.PageView, int64, error) {
	query := bson.M{"wedding_id": weddingID}

	// Apply filters
	if filter != nil {
		if filter.StartDate != nil {
			query["timestamp"] = bson.M{"$gte": *filter.StartDate}
		}
		if filter.EndDate != nil {
			if query["timestamp"] == nil {
				query["timestamp"] = bson.M{}
			}
			query["timestamp"].(bson.M)["$lte"] = *filter.EndDate
		}
		if filter.Device != "" {
			query["device"] = filter.Device
		}
		if filter.Page != "" {
			query["page"] = filter.Page
		}
	}

	// Get total count
	total, err := r.pageViews.CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count page views: %w", err)
	}

	// Apply pagination
	opts := options.Find()
	if filter != nil {
		if filter.Limit > 0 {
			opts.SetLimit(int64(filter.Limit))
		}
		if filter.Offset > 0 {
			opts.SetSkip(int64(filter.Offset))
		}
	}
	opts.SetSort(bson.M{"timestamp": -1})

	cursor, err := r.pageViews.Find(ctx, query, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find page views: %w", err)
	}
	defer cursor.Close(ctx)

	var pageViews []*models.PageView
	if err = cursor.All(ctx, &pageViews); err != nil {
		return nil, 0, fmt.Errorf("failed to decode page views: %w", err)
	}

	return pageViews, total, nil
}

// TrackRSVPEvent records an RSVP analytics event
func (r *analyticsRepository) TrackRSVPEvent(ctx context.Context, event *models.RSVPAnalytics) error {
	if event.ID.IsZero() {
		event.ID = primitive.NewObjectID()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	_, err := r.rsvpEvents.InsertOne(ctx, event)
	if err != nil {
		return fmt.Errorf("failed to track RSVP event: %w", err)
	}

	// Update wedding analytics asynchronously
	go func() {
		r.UpdateWeddingAnalytics(context.Background(), event.WeddingID)
	}()

	return nil
}

// GetRSVPAnalytics retrieves RSVP analytics with filtering
func (r *analyticsRepository) GetRSVPAnalytics(ctx context.Context, weddingID primitive.ObjectID, filter *models.AnalyticsFilter) ([]*models.RSVPAnalytics, int64, error) {
	query := bson.M{"wedding_id": weddingID}

	// Apply filters
	if filter != nil {
		if filter.StartDate != nil {
			query["timestamp"] = bson.M{"$gte": *filter.StartDate}
		}
		if filter.EndDate != nil {
			if query["timestamp"] == nil {
				query["timestamp"] = bson.M{}
			}
			query["timestamp"].(bson.M)["$lte"] = *filter.EndDate
		}
		if filter.Device != "" {
			query["device"] = filter.Device
		}
		if filter.Source != "" {
			query["source"] = filter.Source
		}
	}

	// Get total count
	total, err := r.rsvpEvents.CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count RSVP analytics: %w", err)
	}

	// Apply pagination
	opts := options.Find()
	if filter != nil {
		if filter.Limit > 0 {
			opts.SetLimit(int64(filter.Limit))
		}
		if filter.Offset > 0 {
			opts.SetSkip(int64(filter.Offset))
		}
	}
	opts.SetSort(bson.M{"timestamp": -1})

	cursor, err := r.rsvpEvents.Find(ctx, query, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find RSVP analytics: %w", err)
	}
	defer cursor.Close(ctx)

	var events []*models.RSVPAnalytics
	if err = cursor.All(ctx, &events); err != nil {
		return nil, 0, fmt.Errorf("failed to decode RSVP analytics: %w", err)
	}

	return events, total, nil
}

// TrackConversion records a conversion event
func (r *analyticsRepository) TrackConversion(ctx context.Context, event *models.ConversionEvent) error {
	if event.ID.IsZero() {
		event.ID = primitive.NewObjectID()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	_, err := r.conversions.InsertOne(ctx, event)
	if err != nil {
		return fmt.Errorf("failed to track conversion: %w", err)
	}

	// Update wedding analytics asynchronously
	go func() {
		r.UpdateWeddingAnalytics(context.Background(), event.WeddingID)
	}()

	return nil
}

// GetConversions retrieves conversion events with filtering
func (r *analyticsRepository) GetConversions(ctx context.Context, weddingID primitive.ObjectID, filter *models.AnalyticsFilter) ([]*models.ConversionEvent, int64, error) {
	query := bson.M{"wedding_id": weddingID}

	// Apply filters
	if filter != nil {
		if filter.StartDate != nil {
			query["timestamp"] = bson.M{"$gte": *filter.StartDate}
		}
		if filter.EndDate != nil {
			if query["timestamp"] == nil {
				query["timestamp"] = bson.M{}
			}
			query["timestamp"].(bson.M)["$lte"] = *filter.EndDate
		}
		if filter.Event != "" {
			query["event"] = filter.Event
		}
	}

	// Get total count
	total, err := r.conversions.CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count conversions: %w", err)
	}

	// Apply pagination
	opts := options.Find()
	if filter != nil {
		if filter.Limit > 0 {
			opts.SetLimit(int64(filter.Limit))
		}
		if filter.Offset > 0 {
			opts.SetSkip(int64(filter.Offset))
		}
	}
	opts.SetSort(bson.M{"timestamp": -1})

	cursor, err := r.conversions.Find(ctx, query, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find conversions: %w", err)
	}
	defer cursor.Close(ctx)

	var conversions []*models.ConversionEvent
	if err = cursor.All(ctx, &conversions); err != nil {
		return nil, 0, fmt.Errorf("failed to decode conversions: %w", err)
	}

	return conversions, total, nil
}

// GetWeddingAnalytics retrieves aggregated analytics for a wedding
func (r *analyticsRepository) GetWeddingAnalytics(ctx context.Context, weddingID primitive.ObjectID) (*models.WeddingAnalytics, error) {
	var analytics models.WeddingAnalytics
	err := r.weddingAnalytics.FindOne(ctx, bson.M{"_id": weddingID}).Decode(&analytics)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// Return empty analytics if not found
			return &models.WeddingAnalytics{
				WeddingID:       weddingID,
				PageViews:       0,
				UniqueSessions:  0,
				RSVPCount:       0,
				CompletedRSVPs:  0,
				ConversionRate:  0,
				PopularPages:    make(map[string]int64),
				TrafficSources:  make(map[string]int64),
				DeviceBreakdown: make(map[string]int64),
				ViewsByDate:     make(map[string]int64),
				RSVPsByDate:     make(map[string]int64),
				LastUpdated:     time.Now(),
			}, nil
		}
		return nil, fmt.Errorf("failed to get wedding analytics: %w", err)
	}

	return &analytics, nil
}

// UpdateWeddingAnalytics recalculates and updates wedding analytics
func (r *analyticsRepository) UpdateWeddingAnalytics(ctx context.Context, weddingID primitive.ObjectID) error {
	// Get basic metrics
	pageViews, err := r.pageViews.CountDocuments(ctx, bson.M{"wedding_id": weddingID})
	if err != nil {
		return fmt.Errorf("failed to count page views: %w", err)
	}

	// Get unique sessions
	pipeline := []bson.M{
		{"$match": bson.M{"wedding_id": weddingID}},
		{"$group": bson.M{"_id": "$session_id"}},
		{"$count": "unique_sessions"},
	}
	cursor, err := r.pageViews.Aggregate(ctx, pipeline)
	if err != nil {
		return fmt.Errorf("failed to aggregate sessions: %w", err)
	}
	defer cursor.Close(ctx)

	var uniqueSessions int64 = 0
	if cursor.Next(ctx) {
		var result struct {
			UniqueSessions int64 `bson:"unique_sessions"`
		}
		if err := cursor.Decode(&result); err == nil {
			uniqueSessions = result.UniqueSessions
		}
	}

	// Get RSVP count
	rsvpCount, err := r.rsvpEvents.CountDocuments(ctx, bson.M{"wedding_id": weddingID})
	if err != nil {
		return fmt.Errorf("failed to count RSVP analytics: %w", err)
	}

	// Calculate popular pages
	popularPagesPipeline := []bson.M{
		{"$match": bson.M{"wedding_id": weddingID}},
		{"$group": bson.M{"_id": "$page", "count": bson.M{"$sum": 1}}},
		{"$sort": bson.M{"count": -1}},
		{"$limit": 10},
	}
	popularPagesCursor, err := r.pageViews.Aggregate(ctx, popularPagesPipeline)
	if err != nil {
		return fmt.Errorf("failed to aggregate popular pages: %w", err)
	}
	defer popularPagesCursor.Close(ctx)

	popularPages := make(map[string]int64)
	for popularPagesCursor.Next(ctx) {
		var result struct {
			Page  string `bson:"_id"`
			Count int64  `bson:"count"`
		}
		if err := popularPagesCursor.Decode(&result); err == nil {
			popularPages[result.Page] = result.Count
		}
	}

	// Calculate device breakdown
	devicePipeline := []bson.M{
		{"$match": bson.M{"wedding_id": weddingID, "device": bson.M{"$ne": ""}}},
		{"$group": bson.M{"_id": "$device", "count": bson.M{"$sum": 1}}},
	}
	deviceCursor, err := r.pageViews.Aggregate(ctx, devicePipeline)
	if err != nil {
		return fmt.Errorf("failed to aggregate device breakdown: %w", err)
	}
	defer deviceCursor.Close(ctx)

	deviceBreakdown := make(map[string]int64)
	for deviceCursor.Next(ctx) {
		var result struct {
			Device string `bson:"_id"`
			Count  int64  `bson:"count"`
		}
		if err := deviceCursor.Decode(&result); err == nil {
			deviceBreakdown[result.Device] = result.Count
		}
	}

	// Calculate conversion rate
	var conversionRate float64 = 0
	if pageViews > 0 {
		conversionRate = float64(rsvpCount) / float64(pageViews) * 100
	}

	// Create update
	analytics := &models.WeddingAnalytics{
		WeddingID:         weddingID,
		PageViews:         pageViews,
		UniqueSessions:    uniqueSessions,
		RSVPCount:         rsvpCount,
		CompletedRSVPs:    rsvpCount, // For now, all RSVPs are considered completed
		ConversionRate:    conversionRate,
		PopularPages:      popularPages,
		TrafficSources:    make(map[string]int64), // TODO: implement traffic source tracking
		DeviceBreakdown:   deviceBreakdown,
		ViewsByDate:       make(map[string]int64), // TODO: implement date-based tracking
		RSVPsByDate:       make(map[string]int64), // TODO: implement date-based RSVP tracking
		AverageTimeOnPage: 0,                      // TODO: implement average time on page
		BounceRate:        0,                      // TODO: implement bounce rate calculation
		LastUpdated:       time.Now(),
	}

	// Upsert analytics
	filter := bson.M{"_id": weddingID}
	update := bson.M{"$set": analytics}
	options := options.Update().SetUpsert(true)

	_, err = r.weddingAnalytics.UpdateOne(ctx, filter, update, options)
	if err != nil {
		return fmt.Errorf("failed to update wedding analytics: %w", err)
	}

	return nil
}

// GetSystemAnalytics retrieves system-wide analytics
func (r *analyticsRepository) GetSystemAnalytics(ctx context.Context) (*models.SystemAnalytics, error) {
	var analytics models.SystemAnalytics
	err := r.systemAnalytics.FindOne(ctx, bson.M{}).Decode(&analytics)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// Return empty analytics if not found
			return &models.SystemAnalytics{
				TotalUsers:        0,
				TotalWeddings:     0,
				TotalRSVPs:        0,
				ActiveWeddings:    0,
				PublishedWeddings: 0,
				NewUsersToday:     0,
				NewWeddingsToday:  0,
				NewRSVPsToday:     0,
				TotalPageViews:    0,
				StorageUsed:       0,
				LastUpdated:       time.Now(),
				MetricsByDate:     make(map[string]interface{}),
			}, nil
		}
		return nil, fmt.Errorf("failed to get system analytics: %w", err)
	}

	return &analytics, nil
}

// UpdateSystemAnalytics recalculates and updates system analytics
func (r *analyticsRepository) UpdateSystemAnalytics(ctx context.Context) error {
	// Get total users
	totalUsers, err := r.db.Collection("users").CountDocuments(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("failed to count users: %w", err)
	}

	// Get total weddings
	totalWeddings, err := r.db.Collection("weddings").CountDocuments(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("failed to count weddings: %w", err)
	}

	// Get published weddings
	publishedWeddings, err := r.db.Collection("weddings").CountDocuments(ctx, bson.M{"status": "published"})
	if err != nil {
		return fmt.Errorf("failed to count published weddings: %w", err)
	}

	// Get total RSVPs
	totalRSVPs, err := r.db.Collection("rsvps").CountDocuments(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("failed to count RSVPs: %w", err)
	}

	// Get today's date
	today := time.Now().UTC()
	startOfDay := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.Add(24 * time.Hour)

	// Get new users today
	newUsersToday, err := r.db.Collection("users").CountDocuments(ctx, bson.M{
		"created_at": bson.M{"$gte": startOfDay, "$lt": endOfDay},
	})
	if err != nil {
		return fmt.Errorf("failed to count new users today: %w", err)
	}

	// Get new weddings today
	newWeddingsToday, err := r.db.Collection("weddings").CountDocuments(ctx, bson.M{
		"created_at": bson.M{"$gte": startOfDay, "$lt": endOfDay},
	})
	if err != nil {
		return fmt.Errorf("failed to count new weddings today: %w", err)
	}

	// Get new RSVPs today
	newRSVPsToday, err := r.db.Collection("rsvps").CountDocuments(ctx, bson.M{
		"submitted_at": bson.M{"$gte": startOfDay, "$lt": endOfDay},
	})
	if err != nil {
		return fmt.Errorf("failed to count new RSVPs today: %w", err)
	}

	// Calculate active weddings (weddings with RSVPs in last 30 days)
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	activeWeddings, err := r.db.Collection("rsvps").Distinct(ctx, "wedding_id", bson.M{
		"submitted_at": bson.M{"$gte": thirtyDaysAgo},
	})
	if err != nil {
		return fmt.Errorf("failed to get active weddings: %w", err)
	}

	// Get total page views
	totalPageViews, err := r.pageViews.CountDocuments(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("failed to count page views: %w", err)
	}

	analytics := &models.SystemAnalytics{
		TotalUsers:        totalUsers,
		TotalWeddings:     totalWeddings,
		TotalRSVPs:        totalRSVPs,
		ActiveWeddings:    int64(len(activeWeddings)),
		PublishedWeddings: publishedWeddings,
		NewUsersToday:     newUsersToday,
		NewWeddingsToday:  newWeddingsToday,
		NewRSVPsToday:     newRSVPsToday,
		TotalPageViews:    totalPageViews,
		StorageUsed:       0, // TODO: implement storage calculation
		LastUpdated:       time.Now(),
		MetricsByDate:     make(map[string]interface{}), // TODO: implement date-based tracking
	}

	// Upsert analytics
	filter := bson.M{}
	update := bson.M{"$set": analytics}
	options := options.Update().SetUpsert(true)

	_, err = r.systemAnalytics.UpdateOne(ctx, filter, update, options)
	if err != nil {
		return fmt.Errorf("failed to update system analytics: %w", err)
	}

	return nil
}

// GetAnalyticsSummary generates a summary report for a wedding
func (r *analyticsRepository) GetAnalyticsSummary(ctx context.Context, weddingID primitive.ObjectID, period string) (*models.AnalyticsSummary, error) {
	// Calculate date range based on period
	var startDate time.Time
	now := time.Now()

	switch period {
	case "daily":
		startDate = now.AddDate(0, 0, -7) // Last 7 days
	case "weekly":
		startDate = now.AddDate(0, 0, -28) // Last 4 weeks
	case "monthly":
		startDate = now.AddDate(0, -3, 0) // Last 3 months
	default:
		startDate = now.AddDate(0, 0, -30) // Default to last 30 days
	}

	filter := &models.AnalyticsFilter{
		WeddingID: &weddingID,
		StartDate: &startDate,
		EndDate:   &now,
	}

	// Get page views and RSVPs
	pageViews, _, err := r.GetPageViews(ctx, weddingID, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get page views: %w", err)
	}

	rsvpAnalytics, _, err := r.GetRSVPAnalytics(ctx, weddingID, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get RSVP analytics: %w", err)
	}

	// Calculate totals
	totalPageViews := int64(len(pageViews))
	totalRSVPs := int64(len(rsvpAnalytics))

	// Get unique sessions (approximate)
	sessions := make(map[string]bool)
	for _, pv := range pageViews {
		sessions[pv.SessionID] = true
	}
	totalSessions := int64(len(sessions))

	// Calculate conversion rate
	var conversionRate float64 = 0
	if totalPageViews > 0 {
		conversionRate = float64(totalRSVPs) / float64(totalPageViews) * 100
	}

	// Get popular pages
	popularPages, err := r.GetPopularPages(ctx, weddingID, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to get popular pages: %w", err)
	}

	// Get traffic sources
	trafficSources, err := r.GetTrafficSources(ctx, weddingID, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to get traffic sources: %w", err)
	}

	// Get device breakdown
	analytics, err := r.GetWeddingAnalytics(ctx, weddingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wedding analytics: %w", err)
	}

	// Get daily metrics
	dailyMetrics, err := r.GetDailyMetrics(ctx, weddingID, startDate, now)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily metrics: %w", err)
	}

	return &models.AnalyticsSummary{
		Period:          period,
		TotalPageViews:  totalPageViews,
		TotalSessions:   totalSessions,
		TotalRSVPs:      totalRSVPs,
		ConversionRate:  conversionRate,
		TopPages:        popularPages,
		TopSources:      trafficSources,
		DeviceBreakdown: analytics.DeviceBreakdown,
		DailyMetrics:    dailyMetrics,
	}, nil
}

// GetPopularPages returns the most popular pages for a wedding
func (r *analyticsRepository) GetPopularPages(ctx context.Context, weddingID primitive.ObjectID, limit int) ([]models.PageStats, error) {
	pipeline := []bson.M{
		{"$match": bson.M{"wedding_id": weddingID}},
		{"$group": bson.M{
			"_id":          "$page",
			"views":        bson.M{"$sum": 1},
			"unique_views": bson.M{"$addToSet": "$session_id"},
		}},
		{"$project": bson.M{
			"page":         "$_id",
			"views":        1,
			"unique_views": bson.M{"$size": "$unique_views"},
			"avg_time":     bson.M{"$avg": "$duration"},
		}},
		{"$sort": bson.M{"views": -1}},
		{"$limit": int64(limit)},
	}

	cursor, err := r.pageViews.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate popular pages: %w", err)
	}
	defer cursor.Close(ctx)

	var pages []models.PageStats
	for cursor.Next(ctx) {
		var result struct {
			Page        string  `bson:"page"`
			Views       int64   `bson:"views"`
			UniqueViews int64   `bson:"unique_views"`
			AvgTime     float64 `bson:"avg_time"`
		}
		if err := cursor.Decode(&result); err != nil {
			continue
		}

		pages = append(pages, models.PageStats{
			Page:        result.Page,
			Views:       result.Views,
			UniqueViews: result.UniqueViews,
			AvgTime:     result.AvgTime,
		})
	}

	return pages, nil
}

// GetTrafficSources returns traffic sources for a wedding
func (r *analyticsRepository) GetTrafficSources(ctx context.Context, weddingID primitive.ObjectID, limit int) ([]models.TrafficSourceStats, error) {
	pipeline := []bson.M{
		{"$match": bson.M{
			"wedding_id": weddingID,
			"referrer":   bson.M{"$ne": ""},
		}},
		{"$group": bson.M{
			"_id":      "$referrer",
			"visitors": bson.M{"$addToSet": "$session_id"},
			"views":    bson.M{"$sum": 1},
		}},
		{"$project": bson.M{
			"source":   "$_id",
			"visitors": bson.M{"$size": "$visitors"},
			"views":    1,
		}},
		{"$sort": bson.M{"visitors": -1}},
		{"$limit": int64(limit)},
	}

	cursor, err := r.pageViews.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate traffic sources: %w", err)
	}
	defer cursor.Close(ctx)

	var sources []models.TrafficSourceStats
	for cursor.Next(ctx) {
		var result struct {
			Source   string `bson:"source"`
			Visitors int64  `bson:"visitors"`
			Views    int64  `bson:"views"`
		}
		if err := cursor.Decode(&result); err != nil {
			continue
		}

		sources = append(sources, models.TrafficSourceStats{
			Source:   result.Source,
			Visitors: result.Visitors,
			Views:    result.Views,
		})
	}

	return sources, nil
}

// GetDailyMetrics returns daily metrics for a date range
func (r *analyticsRepository) GetDailyMetrics(ctx context.Context, weddingID primitive.ObjectID, startDate, endDate time.Time) ([]models.DailyMetrics, error) {
	// Get daily page views
	pageViewsPipeline := []bson.M{
		{"$match": bson.M{
			"wedding_id": weddingID,
			"timestamp":  bson.M{"$gte": startDate, "$lte": endDate},
		}},
		{"$group": bson.M{
			"_id": bson.M{
				"year":  bson.M{"$year": "$timestamp"},
				"month": bson.M{"$month": "$timestamp"},
				"day":   bson.M{"$dayOfMonth": "$timestamp"},
			},
			"page_views": bson.M{"$sum": 1},
			"sessions":   bson.M{"$addToSet": "$session_id"},
		}},
		{"$project": bson.M{
			"date":       bson.M{"$dateToString": bson.M{"format": "%Y-%m-%d", "date": "$_id"}},
			"page_views": 1,
			"sessions":   bson.M{"$size": "$sessions"},
		}},
		{"$sort": bson.M{"date": 1}},
	}

	pageViewsCursor, err := r.pageViews.Aggregate(ctx, pageViewsPipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate daily page views: %w", err)
	}
	defer pageViewsCursor.Close(ctx)

	// Create map for page views by date
	pageViewsByDate := make(map[string]models.DailyMetrics)
	for pageViewsCursor.Next(ctx) {
		var result struct {
			Date      string `bson:"date"`
			PageViews int64  `bson:"page_views"`
			Sessions  int64  `bson:"sessions"`
		}
		if err := pageViewsCursor.Decode(&result); err != nil {
			continue
		}

		pageViewsByDate[result.Date] = models.DailyMetrics{
			Date:      result.Date,
			PageViews: result.PageViews,
			Sessions:  result.Sessions,
		}
	}

	// Get daily RSVPs
	rsvpPipeline := []bson.M{
		{"$match": bson.M{
			"wedding_id": weddingID,
			"timestamp":  bson.M{"$gte": startDate, "$lte": endDate},
		}},
		{"$group": bson.M{
			"_id": bson.M{
				"year":  bson.M{"$year": "$timestamp"},
				"month": bson.M{"$month": "$timestamp"},
				"day":   bson.M{"$dayOfMonth": "$timestamp"},
			},
			"rsvps": bson.M{"$sum": 1},
		}},
		{"$project": bson.M{
			"date":  bson.M{"$dateToString": bson.M{"format": "%Y-%m-%d", "date": "$_id"}},
			"rsvps": 1,
		}},
	}

	rsvpCursor, err := r.rsvpEvents.Aggregate(ctx, rsvpPipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate daily RSVPs: %w", err)
	}
	defer rsvpCursor.Close(ctx)

	// Combine data and calculate conversion rates
	var metrics []models.DailyMetrics
	for rsvpCursor.Next(ctx) {
		var result struct {
			Date  string `bson:"date"`
			RSVPS int64  `bson:"rsvps"`
		}
		if err := rsvpCursor.Decode(&result); err != nil {
			continue
		}

		dailyMetric := pageViewsByDate[result.Date]
		dailyMetric.Date = result.Date
		dailyMetric.RSVPs = result.RSVPS

		if dailyMetric.PageViews > 0 {
			dailyMetric.Conversions = float64(result.RSVPS) / float64(dailyMetric.PageViews) * 100
		}

		metrics = append(metrics, dailyMetric)
		delete(pageViewsByDate, result.Date)
	}

	// Add remaining dates without RSVPs
	for _, dailyMetric := range pageViewsByDate {
		metrics = append(metrics, dailyMetric)
	}

	// Sort by date
	for i := 0; i < len(metrics)-1; i++ {
		for j := i + 1; j < len(metrics); j++ {
			if metrics[i].Date > metrics[j].Date {
				metrics[i], metrics[j] = metrics[j], metrics[i]
			}
		}
	}

	return metrics, nil
}

// CleanupOldAnalytics removes analytics data older than the specified date
func (r *analyticsRepository) CleanupOldAnalytics(ctx context.Context, olderThan time.Time) error {
	// Cleanup old page views
	_, err := r.pageViews.DeleteMany(ctx, bson.M{"timestamp": bson.M{"$lt": olderThan}})
	if err != nil {
		return fmt.Errorf("failed to cleanup old page views: %w", err)
	}

	// Cleanup old RSVP analytics
	_, err = r.rsvpEvents.DeleteMany(ctx, bson.M{"timestamp": bson.M{"$lt": olderThan}})
	if err != nil {
		return fmt.Errorf("failed to cleanup old RSVP analytics: %w", err)
	}

	// Cleanup old conversion events
	_, err = r.conversions.DeleteMany(ctx, bson.M{"timestamp": bson.M{"$lt": olderThan}})
	if err != nil {
		return fmt.Errorf("failed to cleanup old conversions: %w", err)
	}

	return nil
}
