package services

import (
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"wedding-invitation-backend/internal/domain/models"
	"wedding-invitation-backend/internal/domain/repository"
)

// PublicWeddingService defines methods needed for public wedding operations
type PublicWeddingService interface {
	GetWeddingBySlugForPublic(ctx context.Context, slug string) (*models.Wedding, error)
}

// PublicRSVPService defines methods needed for public RSVP operations
type PublicRSVPService interface {
	SubmitRSVP(ctx context.Context, weddingID primitive.ObjectID, req SubmitRSVPRequest) (*models.RSVP, error)
}

// RSVPServiceInterface defines the full interface for RSVP service
type RSVPServiceInterface interface {
	SubmitRSVP(ctx context.Context, weddingID primitive.ObjectID, req SubmitRSVPRequest) (*models.RSVP, error)
	GetRSVPByID(ctx context.Context, id primitive.ObjectID) (*models.RSVP, error)
	UpdateRSVP(ctx context.Context, id primitive.ObjectID, req UpdateRSVPRequest) (*models.RSVP, error)
	DeleteRSVP(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) error
	ListRSVPs(ctx context.Context, weddingID primitive.ObjectID, userID primitive.ObjectID, page, pageSize int, filters repository.RSVPFilters) ([]*models.RSVP, int64, error)
	GetRSVPStatistics(ctx context.Context, weddingID primitive.ObjectID, userID primitive.ObjectID) (*models.RSVPStatistics, error)
	ExportRSVPs(ctx context.Context, weddingID primitive.ObjectID, userID primitive.ObjectID) ([]*models.RSVP, error)
}

// WeddingServiceInterface defines the full interface for Wedding service
type WeddingServiceInterface interface {
	CreateWedding(ctx context.Context, wedding *models.Wedding, userID primitive.ObjectID) error
	GetWeddingByID(ctx context.Context, id primitive.ObjectID, requestingUserID primitive.ObjectID) (*models.Wedding, error)
	GetWeddingBySlug(ctx context.Context, slug string, requestingUserID primitive.ObjectID) (*models.Wedding, error)
	GetUserWeddings(ctx context.Context, userID primitive.ObjectID, page, pageSize int, filters repository.WeddingFilters) ([]*models.Wedding, int64, error)
	UpdateWedding(ctx context.Context, wedding *models.Wedding, requestingUserID primitive.ObjectID) error
	DeleteWedding(ctx context.Context, weddingID primitive.ObjectID, requestingUserID primitive.ObjectID) error
	PublishWedding(ctx context.Context, weddingID primitive.ObjectID, requestingUserID primitive.ObjectID) error
	ListPublicWeddings(ctx context.Context, page, pageSize int, filters repository.PublicWeddingFilters) ([]*models.Wedding, int64, error)
}

// WeddingFilters represents filters for wedding listings
type WeddingFilters struct {
	Status   string
	IsPublic *bool
	Search   string
	DateFrom *primitive.DateTime
	DateTo   *primitive.DateTime
}

// WeddingStatistics represents wedding statistics
type WeddingStatistics struct {
	TotalViews        int64   `json:"total_views"`
	TotalRSVPs        int64   `json:"total_rsvps"`
	TotalAttending    int64   `json:"total_attending"`
	TotalNotAttending int64   `json:"total_not_attending"`
	TotalMaybe        int64   `json:"total_maybe"`
	RSVPRate          float64 `json:"rsvp_rate"`
}

// RSVPStatistics represents RSVP statistics
type RSVPStatistics struct {
	TotalRSVPs          int64    `json:"total_rsvps"`
	Attending           int64    `json:"attending"`
	NotAttending        int64    `json:"not_attending"`
	Maybe               int64    `json:"maybe"`
	TotalGuests         int64    `json:"total_guests"`
	AttendingGuests     int64    `json:"attending_guests"`
	DietaryRestrictions []string `json:"dietary_restrictions"`
	PlusOnes            int64    `json:"plus_ones"`
}
