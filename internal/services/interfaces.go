package services

import (
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"wedding-invitation-backend/internal/domain/models"
)

// PublicWeddingService defines methods needed for public wedding operations
type PublicWeddingService interface {
	GetWeddingBySlugForPublic(ctx context.Context, slug string) (*models.Wedding, error)
}

// PublicRSVPService defines methods needed for public RSVP operations
type PublicRSVPService interface {
	SubmitRSVP(ctx context.Context, weddingID primitive.ObjectID, req SubmitRSVPRequest) (*models.RSVP, error)
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
