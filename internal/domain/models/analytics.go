package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PageView represents a page view event for analytics
type PageView struct {
	ID           primitive.ObjectID          `bson:"_id,omitempty" json:"id"`
	WeddingID    primitive.ObjectID          `bson:"wedding_id" json:"wedding_id"`
	SessionID    string                      `bson:"session_id" json:"session_id"`
	IPAddress    string                      `bson:"ip_address" json:"-"`
	UserAgent    string                      `bson:"user_agent" json:"-"`
	Referrer     string                      `bson:"referrer,omitempty" json:"-"`
	Page         string                      `bson:"page" json:"page"` // e.g., "invitation", "rsvp", "gallery"
	Timestamp    time.Time                   `bson:"timestamp" json:"timestamp"`
	Duration     int64                       `bson:"duration,omitempty" json:"duration"` // Time spent on page in seconds
	Device       string                      `bson:"device,omitempty" json:"device"`     // mobile, desktop, tablet
	Browser      string                      `bson:"browser,omitempty" json:"browser"`
	OS           string                      `bson:"os,omitempty" json:"os"`
	Country      string                      `bson:"country,omitempty" json:"country"`
	City         string                      `bson:"city,omitempty" json:"city"`
	Metadata     map[string]interface{}      `bson:"metadata,omitempty" json:"metadata,omitempty"`
}

// RSVPAnalytics represents analytics data for RSVP submissions
type RSVPAnalytics struct {
	ID              primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	WeddingID       primitive.ObjectID   `bson:"wedding_id" json:"wedding_id"`
	RSVPID          primitive.ObjectID   `bson:"rsvp_id" json:"rsvp_id"`
	SessionID       string               `bson:"session_id" json:"session_id"`
	TimeToComplete  int64                `bson:"time_to_complete" json:"time_to_complete"` // Seconds from page view to submission
	Source          string               `bson:"source" json:"source"`                    // web, direct_link, qr_code, manual
	Device          string               `bson:"device,omitempty" json:"device"`
	Browser         string               `bson:"browser,omitempty" json:"browser"`
	Referrer        string               `bson:"referrer,omitempty" json:"referrer"`
	Timestamp       time.Time            `bson:"timestamp" json:"timestamp"`
	AbandonedStep   string               `bson:"abandoned_step,omitempty" json:"abandoned_step"` // For incomplete submissions
	FormErrors      []string             `bson:"form_errors,omitempty" json:"form_errors"`
}

// ConversionEvent represents conversion events
type ConversionEvent struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	WeddingID  primitive.ObjectID `bson:"wedding_id" json:"wedding_id"`
	SessionID  string             `bson:"session_id" json:"session_id"`
	Event      string             `bson:"event" json:"event"` // rsvp_started, rsvp_completed, share_clicked, etc.
	Value      float64            `bson:"value,omitempty" json:"value"` // Optional value (e.g., for goal tracking)
	Currency   string             `bson:"currency,omitempty" json:"currency"`
	Timestamp  time.Time          `bson:"timestamp" json:"timestamp"`
	Properties map[string]interface{} `bson:"properties,omitempty" json:"properties,omitempty"`
}

// WeddingAnalytics represents aggregated analytics for a wedding
type WeddingAnalytics struct {
	WeddingID           primitive.ObjectID          `bson:"_id" json:"wedding_id"`
	PageViews           int64                       `bson:"page_views" json:"page_views"`
	UniqueSessions      int64                       `bson:"unique_sessions" json:"unique_sessions"`
	RSVPCount           int64                       `bson:"rsvp_count" json:"rsvp_count"`
	CompletedRSVPs      int64                       `bson:"completed_rsvps" json:"completed_rsvps"`
	ConversionRate      float64                     `bson:"conversion_rate" json:"conversion_rate"` // RSVPs / PageViews
	PopularPages        map[string]int64            `bson:"popular_pages" json:"popular_pages"`
	TrafficSources      map[string]int64            `bson:"traffic_sources" json:"traffic_sources"`
	DeviceBreakdown     map[string]int64            `bson:"device_breakdown" json:"device_breakdown"`
	ViewsByDate         map[string]int64            `bson:"views_by_date" json:"views_by_date"`
	RSVPsByDate         map[string]int64            `bson:"rsvps_by_date" json:"rsvps_by_date"`
	AverageTimeOnPage   float64                     `bson:"average_time_on_page" json:"average_time_on_page"`
	BounceRate          float64                     `bson:"bounce_rate" json:"bounce_rate"`
	LastUpdated         time.Time                   `bson:"last_updated" json:"last_updated"`
}

// SystemAnalytics represents system-wide analytics
type SystemAnalytics struct {
	TotalUsers        int64                     `bson:"total_users" json:"total_users"`
	TotalWeddings     int64                     `bson:"total_weddings" json:"total_weddings"`
	TotalRSVPs        int64                     `bson:"total_rsvps" json:"total_rsvps"`
	ActiveWeddings    int64                     `bson:"active_weddings" json:"active_weddings"`
	PublishedWeddings int64                     `bson:"published_weddings" json:"published_weddings"`
	NewUsersToday     int64                     `bson:"new_users_today" json:"new_users_today"`
	NewWeddingsToday  int64                     `bson:"new_weddings_today" json:"new_weddings_today"`
	NewRSVPsToday     int64                     `bson:"new_rsvps_today" json:"new_rsvps_today"`
	TotalPageViews    int64                     `bson:"total_page_views" json:"total_page_views"`
	StorageUsed       int64                     `bson:"storage_used" json:"storage_used"` // In bytes
	LastUpdated       time.Time                 `bson:"last_updated" json:"last_updated"`
	MetricsByDate     map[string]interface{}    `bson:"metrics_by_date" json:"metrics_by_date"`
}

// AnalyticsFilter represents filters for analytics queries
type AnalyticsFilter struct {
	WeddingID    *primitive.ObjectID `json:"wedding_id,omitempty"`
	StartDate    *time.Time          `json:"start_date,omitempty"`
	EndDate      *time.Time          `json:"end_date,omitempty"`
	Device       string              `json:"device,omitempty"`
	Source       string              `json:"source,omitempty"`
	Page         string              `json:"page,omitempty"`
	Event        string              `json:"event,omitempty"`
	Limit        int                 `json:"limit,omitempty"`
	Offset       int                 `json:"offset,omitempty"`
}

// AnalyticsSummary represents a summary report
type AnalyticsSummary struct {
	Period           string                    `json:"period"` // daily, weekly, monthly
	TotalPageViews   int64                     `json:"total_page_views"`
	TotalSessions    int64                     `json:"total_sessions"`
	TotalRSVPs       int64                     `json:"total_rsvps"`
	ConversionRate   float64                   `json:"conversion_rate"`
	TopPages         []PageStats               `json:"top_pages"`
	TopSources       []TrafficSourceStats      `json:"top_sources"`
	DeviceBreakdown  map[string]int64          `json:"device_breakdown"`
	DailyMetrics     []DailyMetrics            `json:"daily_metrics"`
}

// PageStats represents statistics for a specific page
type PageStats struct {
	Page        string  `json:"page"`
	Views       int64   `json:"views"`
	UniqueViews int64   `json:"unique_views"`
	AvgTime     float64 `json:"avg_time_on_page"`
}

// TrafficSourceStats represents statistics for traffic sources
type TrafficSourceStats struct {
	Source   string `json:"source"`
	Visitors int64  `json:"visitors"`
	Views    int64  `json:"views"`
}

// DailyMetrics represents metrics for a specific day
type DailyMetrics struct {
	Date       string  `json:"date"`
	PageViews  int64   `json:"page_views"`
	Sessions   int64   `json:"sessions"`
	RSVPs      int64   `json:"rsvps"`
	Conversions float64 `json:"conversion_rate"`
}