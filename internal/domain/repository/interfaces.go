package repository

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
	"wedding-invitation-backend/internal/domain/models"
)

var (
	ErrNotFound = errors.New("document not found")
)

// UserRepository defines database operations for users
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByVerificationToken(ctx context.Context, token string) (*models.User, error)
	GetByResetToken(ctx context.Context, token string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	List(ctx context.Context, page, pageSize int, filters UserFilters) ([]*models.User, int64, error)
	AddWeddingID(ctx context.Context, userID, weddingID primitive.ObjectID) error
	RemoveWeddingID(ctx context.Context, userID, weddingID primitive.ObjectID) error
	UpdateLastLogin(ctx context.Context, userID primitive.ObjectID) error
	SetEmailVerified(ctx context.Context, userID primitive.ObjectID) error
}

// WeddingRepository defines database operations for weddings
type WeddingRepository interface {
	Create(ctx context.Context, wedding *models.Wedding) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*models.Wedding, error)
	GetBySlug(ctx context.Context, slug string) (*models.Wedding, error)
	GetByUserID(ctx context.Context, userID primitive.ObjectID, page, pageSize int, filters WeddingFilters) ([]*models.Wedding, int64, error)
	Update(ctx context.Context, wedding *models.Wedding) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	ExistsBySlug(ctx context.Context, slug string) (bool, error)
	ListPublic(ctx context.Context, page, pageSize int, filters PublicWeddingFilters) ([]*models.Wedding, int64, error)
	IncrementViewCount(ctx context.Context, id primitive.ObjectID) error
	UpdateRSVPCount(ctx context.Context, weddingID primitive.ObjectID) error
}

// RSVPRepository defines database operations for RSVPs
type RSVPRepository interface {
	Create(ctx context.Context, rsvp *models.RSVP) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*models.RSVP, error)
	GetByEmail(ctx context.Context, weddingID primitive.ObjectID, email string) (*models.RSVP, error)
	ListByWedding(ctx context.Context, weddingID primitive.ObjectID, page, pageSize int, filters RSVPFilters) ([]*models.RSVP, int64, error)
	Update(ctx context.Context, rsvp *models.RSVP) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	GetStatistics(ctx context.Context, weddingID primitive.ObjectID) (*models.RSVPStatistics, error)
	MarkConfirmationSent(ctx context.Context, id primitive.ObjectID) error
	GetSubmissionTrend(ctx context.Context, weddingID primitive.ObjectID, days int) ([]models.DailyCount, error)
}

// GuestRepository defines database operations for guests (for Phase 3)
type GuestRepository interface {
	Create(ctx context.Context, guest *models.Guest) error
	CreateMany(ctx context.Context, guests []*models.Guest) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*models.Guest, error)
	GetByEmail(ctx context.Context, weddingID primitive.ObjectID, email string) (*models.Guest, error)
	ListByWedding(ctx context.Context, weddingID primitive.ObjectID, page, pageSize int, filters GuestFilters) ([]*models.Guest, int64, error)
	Update(ctx context.Context, guest *models.Guest) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	ImportBatch(ctx context.Context, guests []*models.Guest, batchID string) error
	GetByImportBatch(ctx context.Context, weddingID primitive.ObjectID, batchID string) ([]*models.Guest, error)
}

// MediaRepository defines database operations for media files (for Phase 2)
type MediaRepository interface {
	Create(ctx context.Context, media *models.Media) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*models.Media, error)
	GetByStorageKey(ctx context.Context, key string) (*models.Media, error)
	List(ctx context.Context, filter MediaFilter, opts ListOptions) ([]*models.Media, int64, error)
	Update(ctx context.Context, media *models.Media) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	SoftDelete(ctx context.Context, id primitive.ObjectID) error
	GetOrphaned(ctx context.Context, before time.Time) ([]*models.Media, error)
	GetByCreatedBy(ctx context.Context, userID primitive.ObjectID, opts ListOptions) ([]*models.Media, int64, error)
}

// AnalyticsRepository defines database operations for analytics (for Phase 4)
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

// Filter types for repository queries

type UserFilters struct {
	Status        string     `json:"status"`
	Search        string     `json:"search"`
	CreatedAfter  *time.Time `json:"created_after"`
	CreatedBefore *time.Time `json:"created_before"`
}

type WeddingFilters struct {
	Status        string     `json:"status"`
	Search        string     `json:"search"`
	CreatedAfter  *time.Time `json:"created_after"`
	CreatedBefore *time.Time `json:"created_before"`
	EventDate     *time.Time `json:"event_date"`
}

type PublicWeddingFilters struct {
	Search    string     `json:"search"`
	EventDate *time.Time `json:"event_date"`
}

type RSVPFilters struct {
	Status          string     `json:"status"`
	Search          string     `json:"search"`
	SubmittedAfter  *time.Time `json:"submitted_after"`
	SubmittedBefore *time.Time `json:"submitted_before"`
	Source          string     `json:"source"`
}

type GuestFilters struct {
	RSVPStatus       string `json:"rsvp_status"`
	Side             string `json:"side"`
	Relationship     string `json:"relationship"`
	Search           string `json:"search"`
	VIP              *bool  `json:"vip"`
	InvitationStatus string `json:"invitation_status"`
	InvitedVia       string `json:"invited_via"`
	AllowPlusOne     *bool  `json:"allow_plus_one"`
}

type GuestStatistics struct {
	TotalGuests      int64 `json:"total_guests"`
	InvitedDigital   int64 `json:"invited_digital"`
	InvitedManual    int64 `json:"invited_manual"`
	RSVPAttending    int64 `json:"rsvp_attending"`
	RSVPNotAttending int64 `json:"rsvp_not_attending"`
	RSVPPending      int64 `json:"rsvp_pending"`
	PlusOnesAllowed  int64 `json:"plus_ones_allowed"`
	VIPGuests        int64 `json:"vip_guests"`
}

type MediaFilter struct {
	MimeType      string              `json:"mimeType"`
	CreatedBy     *primitive.ObjectID `json:"createdBy"`
	CreatedAfter  *time.Time          `json:"createdAfter"`
	CreatedBefore *time.Time          `json:"createdBefore"`
	HasThumbnails bool                `json:"hasThumbnails"`
}

type ListOptions struct {
	Limit  int64  `json:"limit"`
	Offset int64  `json:"offset"`
	Sort   bson.D `json:"sort"`
}

type DateRange struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}
