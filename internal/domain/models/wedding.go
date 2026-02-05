package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// EventDetails represents wedding ceremony and reception info
type EventDetails struct {
	Title          string    `bson:"title" json:"title" validate:"required,max=100"`
	Date           time.Time `bson:"date" json:"date" validate:"required"`
	Time           string    `bson:"time,omitempty" json:"time,omitempty"`
	VenueName      string    `bson:"venue_name" json:"venue_name" validate:"required,max=200"`
	VenueAddress   string    `bson:"venue_address" json:"venue_address" validate:"required,max=500"`
	VenueMapURL    string    `bson:"venue_map_url,omitempty" json:"venue_map_url,omitempty" validate:"omitempty,url"`
	DressCode      string    `bson:"dress_code,omitempty" json:"dress_code,omitempty"`
	AdditionalInfo string    `bson:"additional_info,omitempty" json:"additional_info,omitempty"`
}

// CoupleInfo represents bride and groom details
type CoupleInfo struct {
	Partner1 struct {
		FirstName   string            `bson:"first_name" json:"first_name" validate:"required"`
		LastName    string            `bson:"last_name" json:"last_name" validate:"required"`
		FullName    string            `bson:"full_name" json:"full_name"`
		PhotoURL    string            `bson:"photo_url,omitempty" json:"photo_url,omitempty"`
		SocialLinks map[string]string `bson:"social_links,omitempty" json:"social_links,omitempty"`
	} `bson:"partner1" json:"partner1"`

	Partner2 struct {
		FirstName   string            `bson:"first_name" json:"first_name" validate:"required"`
		LastName    string            `bson:"last_name" json:"last_name" validate:"required"`
		FullName    string            `bson:"full_name" json:"full_name"`
		PhotoURL    string            `bson:"photo_url,omitempty" json:"photo_url,omitempty"`
		SocialLinks map[string]string `bson:"social_links,omitempty" json:"social_links,omitempty"`
	} `bson:"partner2" json:"partner2"`

	Story      string `bson:"story,omitempty" json:"story,omitempty" validate:"omitempty,max=2000"`
	Engagement struct {
		Date     *time.Time `bson:"date,omitempty" json:"date,omitempty"`
		Story    string     `bson:"story,omitempty" json:"story,omitempty"`
		PhotoURL string     `bson:"photo_url,omitempty" json:"photo_url,omitempty"`
	} `bson:"engagement,omitempty" json:"engagement,omitempty"`
}

// ThemeSettings represents visual customization
type ThemeSettings struct {
	ThemeID         string                 `bson:"theme_id" json:"theme_id" validate:"required"`
	PrimaryColor    string                 `bson:"primary_color" json:"primary_color" validate:"omitempty,hexcolor"`
	SecondaryColor  string                 `bson:"secondary_color" json:"secondary_color" validate:"omitempty,hexcolor"`
	BackgroundColor string                 `bson:"background_color" json:"background_color" validate:"omitempty,hexcolor"`
	FontFamily      string                 `bson:"font_family" json:"font_family"`
	CustomCSS       string                 `bson:"custom_css,omitempty" json:"-"` // Never expose in API
	CustomSettings  map[string]interface{} `bson:"custom_settings,omitempty" json:"custom_settings,omitempty"`
}

// CustomQuestion for RSVP forms
type CustomQuestion struct {
	ID       string   `bson:"id" json:"id"`
	Question string   `bson:"question" json:"question" validate:"required,max=200"`
	Type     string   `bson:"type" json:"type" validate:"oneof=text textarea select checkbox radio"`
	Required bool     `bson:"required" json:"required"`
	Options  []string `bson:"options,omitempty" json:"options,omitempty"`
	Order    int      `bson:"order" json:"order"`
}

// RSVPSettings configures RSVP form behavior
type RSVPSettings struct {
	Enabled           bool             `bson:"enabled" json:"enabled"`
	Deadline          *time.Time       `bson:"deadline,omitempty" json:"deadline,omitempty"`
	AllowPlusOne      bool             `bson:"allow_plus_one" json:"allow_plus_one"`
	MaxPlusOnes       int              `bson:"max_plus_ones" json:"max_plus_ones" validate:"min=0,max=5"`
	CollectEmail      bool             `bson:"collect_email" json:"collect_email"`
	CollectPhone      bool             `bson:"collect_phone" json:"collect_phone"`
	CollectDietary    bool             `bson:"collect_dietary" json:"collect_dietary"`
	DietaryOptions    []string         `bson:"dietary_options,omitempty" json:"dietary_options,omitempty"`
	CustomQuestions   []CustomQuestion `bson:"custom_questions,omitempty" json:"custom_questions,omitempty"`
	ConfirmationEmail bool             `bson:"confirmation_email" json:"confirmation_email"`
	EmailTemplate     string           `bson:"email_template,omitempty" json:"email_template,omitempty"`
}

// GalleryImage represents a photo in gallery
type GalleryImage struct {
	ID           string    `bson:"id" json:"id"`
	URL          string    `bson:"url" json:"url"`
	ThumbnailURL string    `bson:"thumbnail_url" json:"thumbnail_url"`
	Caption      string    `bson:"caption,omitempty" json:"caption,omitempty"`
	Order        int       `bson:"order" json:"order"`
	UploadedAt   time.Time `bson:"uploaded_at" json:"uploaded_at"`
	FileSize     int64     `bson:"file_size" json:"file_size"`
}

// Wedding is the main collection document
type Wedding struct {
	ID     primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID primitive.ObjectID `bson:"user_id" json:"user_id"` // Reference to owner

	// URL and Access
	Slug         string `bson:"slug" json:"slug" validate:"required,min=3,max=50,slug"`
	PasswordHash string `bson:"password_hash,omitempty" json:"-"` // For private weddings
	IsPublic     bool   `bson:"is_public" json:"is_public"`

	// Content
	Title  string       `bson:"title" json:"title" validate:"required,max=100"`
	Couple CoupleInfo   `bson:"couple" json:"couple"`
	Event  EventDetails `bson:"event" json:"event"`

	// Media
	CoverImageURL  string         `bson:"cover_image_url,omitempty" json:"cover_image_url,omitempty"`
	GalleryImages  []GalleryImage `bson:"gallery_images,omitempty" json:"gallery_images,omitempty"`
	GalleryEnabled bool           `bson:"gallery_enabled" json:"gallery_enabled"`

	// Settings
	Theme ThemeSettings `bson:"theme" json:"theme"`
	RSVP  RSVPSettings  `bson:"rsvp" json:"rsvp"`

	// Social/Sharing
	ShareMessage string `bson:"share_message,omitempty" json:"share_message,omitempty" validate:"omitempty,max=280"`

	// Status
	Status      string     `bson:"status" json:"status" validate:"oneof=draft published expired archived"`
	PublishedAt *time.Time `bson:"published_at,omitempty" json:"published_at,omitempty"`
	ExpiresAt   *time.Time `bson:"expires_at,omitempty" json:"expires_at,omitempty"`

	// Counts (denormalized for performance)
	RSVPCount      int `bson:"rsvp_count" json:"rsvp_count"`
	GuestCount     int `bson:"guest_count" json:"guest_count"`
	TotalAttending int `bson:"total_attending" json:"total_attending"`

	// Metadata
	CreatedAt    time.Time  `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `bson:"updated_at" json:"updated_at"`
	LastViewedAt *time.Time `bson:"last_viewed_at,omitempty" json:"last_viewed_at,omitempty"`
	ViewCount    int64      `bson:"view_count" json:"view_count"`
}

// WeddingStatus represents possible wedding statuses
type WeddingStatus string

const (
	WeddingStatusDraft     WeddingStatus = "draft"
	WeddingStatusPublished WeddingStatus = "published"
	WeddingStatusExpired   WeddingStatus = "expired"
	WeddingStatusArchived  WeddingStatus = "archived"
)

// Helper methods
func (w *Wedding) IsRSVPOpen() bool {
	if !w.RSVP.Enabled {
		return false
	}

	if w.RSVP.Deadline != nil {
		return time.Now().Before(*w.RSVP.Deadline)
	}

	return true
}

func (w *Wedding) IsAccessible() bool {
	if w.IsPublic {
		return w.Status == string(WeddingStatusPublished)
	}

	// For private weddings, allow access if published or has password protection
	return w.Status == string(WeddingStatusPublished)
}

func (w *Wedding) IsExpired() bool {
	if w.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*w.ExpiresAt)
}
