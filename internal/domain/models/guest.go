package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// Guest model for Phase 3
type Guest struct {
	ID               primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	WeddingID        primitive.ObjectID  `bson:"wedding_id" json:"wedding_id"`
	FirstName        string              `bson:"first_name" json:"first_name" validate:"required,max=50"`
	LastName         string              `bson:"last_name" json:"last_name" validate:"required,max=50"`
	Email            string              `bson:"email,omitempty" json:"email,omitempty" validate:"omitempty,email,max=100"`
	Phone            string              `bson:"phone,omitempty" json:"phone,omitempty"`
	Address          *Address            `bson:"address,omitempty" json:"address,omitempty"`
	Relationship     string              `bson:"relationship,omitempty" json:"relationship,omitempty"`
	Side             string              `bson:"side,omitempty" validate:"oneof=bride groom both"`
	InvitedVia       string              `bson:"invited_via" json:"invited_via" validate:"oneof=digital manual"`
	InvitationStatus string              `bson:"invitation_status" json:"invitation_status" validate:"oneof=pending sent delivered failed"`
	AllowPlusOne     bool                `bson:"allow_plus_one" json:"allow_plus_one"`
	MaxPlusOnes      int                 `bson:"max_plus_ones" json:"max_plus_ones" validate:"min=0,max=5"`
	RSVPStatus       string              `bson:"rsvp_status,omitempty" json:"rsvp_status,omitempty" validate:"omitempty,oneof=attending not-attending maybe pending"`
	RSVPID           *primitive.ObjectID `bson:"rsvp_id,omitempty" json:"rsvp_id,omitempty"`
	DietaryNotes     string              `bson:"dietary_notes,omitempty" json:"dietary_notes,omitempty"`
	VIP              bool                `bson:"vip,omitempty" json:"vip,omitempty"`
	Notes            string              `bson:"notes,omitempty" json:"notes,omitempty"`
	ImportBatchID    string              `bson:"import_batch_id,omitempty" json:"import_batch_id,omitempty"`
	CreatedAt        time.Time           `bson:"created_at" json:"created_at"`
	UpdatedAt        time.Time           `bson:"updated_at" json:"updated_at"`
	CreatedBy        primitive.ObjectID  `bson:"created_by" json:"created_by"`
}

type Address struct {
	Street  string `bson:"street,omitempty" json:"street,omitempty"`
	City    string `bson:"city,omitempty" json:"city,omitempty"`
	State   string `bson:"state,omitempty" json:"state,omitempty"`
	ZIP     string `bson:"zip,omitempty" json:"zip,omitempty"`
	Country string `bson:"country,omitempty" json:"country,omitempty"`
}

type GuestImportResult struct {
	SuccessCount int      `json:"success_count"`
	ErrorCount   int      `json:"error_count"`
	Errors       []string `json:"errors"`
	BatchID      string   `json:"batch_id"`
}

// Analytics models for Phase 4
type PageViewEvent struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	WeddingID primitive.ObjectID `bson:"wedding_id" json:"wedding_id"`
	Page      string             `bson:"page" json:"page"` // wedding_home, rsvp_form, gallery, etc.
	IPAddress string             `bson:"ip_address,omitempty" json:"-"`
	UserAgent string             `bson:"user_agent,omitempty" json:"-"`
	Referrer  string             `bson:"referrer,omitempty" json:"referrer,omitempty"`
	SessionID string             `bson:"session_id,omitempty" json:"session_id,omitempty"`
	Timestamp time.Time          `bson:"timestamp" json:"timestamp"`
	Duration  int                `bson:"duration,omitempty" json:"duration,omitempty"` // Time on page in seconds
}

type InteractionEvent struct {
	ID        primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	WeddingID primitive.ObjectID     `bson:"wedding_id" json:"wedding_id"`
	EventType string                 `bson:"event_type" json:"event_type"`               // button_click, form_submit, gallery_view, etc.
	Element   string                 `bson:"element,omitempty" json:"element,omitempty"` // Element identifier
	Page      string                 `bson:"page" json:"page"`
	IPAddress string                 `bson:"ip_address,omitempty" json:"-"`
	UserAgent string                 `bson:"user_agent,omitempty" json:"-"`
	SessionID string                 `bson:"session_id,omitempty" json:"session_id,omitempty"`
	Timestamp time.Time              `bson:"timestamp" json:"timestamp"`
	Metadata  map[string]interface{} `bson:"metadata,omitempty" json:"metadata,omitempty"`
}

type WeddingAnalytics struct {
	WeddingID       primitive.ObjectID `json:"wedding_id"`
	PageViews       int64              `json:"page_views"`
	UniqueSessions  int64              `json:"unique_sessions"`
	RSVPCount       int64              `json:"rsvp_count"`
	ConversionRate  float64            `json:"conversion_rate"` // RSVPs / PageViews
	PopularPages    map[string]int64   `json:"popular_pages"`
	TrafficSources  map[string]int64   `json:"traffic_sources"`
	DeviceBreakdown map[string]int64   `json:"device_breakdown"`
}

type SystemAnalytics struct {
	TotalUsers     int64 `json:"total_users"`
	TotalWeddings  int64 `json:"total_weddings"`
	TotalRSVPs     int64 `json:"total_rsvps"`
	ActiveWeddings int64 `json:"active_weddings"`
	NewUsersToday  int64 `json:"new_users_today"`
	NewRSVPsToday  int64 `json:"new_rsvps_today"`
}
