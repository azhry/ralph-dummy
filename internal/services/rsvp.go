package services

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"wedding-invitation-backend/internal/domain/models"
	"wedding-invitation-backend/internal/domain/repository"
)

var (
	ErrRSVPNotFound        = errors.New("rsvp not found")
	ErrRSVPClosed          = errors.New("rsvp is closed for this wedding")
	ErrInvalidRSVPStatus   = errors.New("invalid rsvp status")
	ErrDuplicateRSVP       = errors.New("rsvp already submitted for this email")
	ErrWeddingNotFound     = errors.New("wedding not found")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrTooManyPlusOnes     = errors.New("too many plus ones")
	ErrRSVPCannotModify    = errors.New("rsvp cannot be modified after 24 hours")
)

// RSVPService provides business logic for RSVP management
type RSVPService struct {
	rsvpRepo    repository.RSVPRepository
	weddingRepo repository.WeddingRepository
}

// NewRSVPService creates a new RSVP service
func NewRSVPService(rsvpRepo repository.RSVPRepository, weddingRepo repository.WeddingRepository) *RSVPService {
	return &RSVPService{
		rsvpRepo:    rsvpRepo,
		weddingRepo: weddingRepo,
	}
}

// SubmitRSVPRequest represents a new RSVP submission
type SubmitRSVPRequest struct {
	FirstName           string                    `json:"first_name" validate:"required,max=50"`
	LastName            string                    `json:"last_name" validate:"required,max=50"`
	Email               string                    `json:"email,omitempty" validate:"omitempty,email,max=100"`
	Phone               string                    `json:"phone,omitempty"`
	Status              string                    `json:"status" validate:"required,oneof=attending not-attending maybe"`
	AttendanceCount     int                       `json:"attendance_count" validate:"required,min=1"`
	PlusOnes            []models.PlusOneInfo      `json:"plus_ones,omitempty"`
	DietaryRestrictions string                    `json:"dietary_restrictions,omitempty"`
	DietarySelected     []string                  `json:"dietary_selected,omitempty"`
	AdditionalNotes     string                    `json:"additional_notes,omitempty" validate:"omitempty,max=500"`
	CustomAnswers       []models.CustomAnswer     `json:"custom_answers,omitempty"`
	Source              string                    `json:"source" validate:"oneof=web direct_link qr_code manual"`
	IPAddress           string                    `json:"ip_address,omitempty"`
	UserAgent           string                    `json:"user_agent,omitempty"`
}

// UpdateRSVPRequest represents an RSVP update
type UpdateRSVPRequest struct {
	Status              *string               `json:"status,omitempty" validate:"omitempty,oneof=attending not-attending maybe"`
	AttendanceCount     *int                  `json:"attendance_count,omitempty" validate:"omitempty,min=1"`
	PlusOnes            *[]models.PlusOneInfo `json:"plus_ones,omitempty"`
	DietaryRestrictions *string               `json:"dietary_restrictions,omitempty"`
	DietarySelected     *[]string             `json:"dietary_selected,omitempty"`
	AdditionalNotes     *string               `json:"additional_notes,omitempty" validate:"omitempty,max=500"`
	CustomAnswers       *[]models.CustomAnswer `json:"custom_answers,omitempty"`
}

// SubmitRSVP handles new RSVP submission
func (s *RSVPService) SubmitRSVP(ctx context.Context, weddingID primitive.ObjectID, req SubmitRSVPRequest) (*models.RSVP, error) {
	// Get wedding to validate RSVP is open
	wedding, err := s.weddingRepo.GetByID(ctx, weddingID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrWeddingNotFound
		}
		return nil, fmt.Errorf("failed to get wedding: %w", err)
	}

	// Check if RSVP is open
	if !s.isRSVPOpen(wedding) {
		return nil, ErrRSVPClosed
	}

	// Validate request
	if err := s.validateSubmitRequest(req, wedding); err != nil {
		return nil, err
	}

	// Check for duplicate RSVP by email
	if req.Email != "" {
		existing, _ := s.rsvpRepo.GetByEmail(ctx, weddingID, req.Email)
		if existing != nil {
			return nil, ErrDuplicateRSVP
		}
	}

	// Create RSVP
	rsvp := &models.RSVP{
		ID:                  primitive.NewObjectID(),
		WeddingID:           weddingID,
		FirstName:           req.FirstName,
		LastName:            req.LastName,
		Email:               req.Email,
		Phone:               req.Phone,
		Status:              req.Status,
		AttendanceCount:     req.AttendanceCount,
		PlusOnes:            req.PlusOnes,
		PlusOneCount:        len(req.PlusOnes),
		DietaryRestrictions: req.DietaryRestrictions,
		DietarySelected:     req.DietarySelected,
		AdditionalNotes:     req.AdditionalNotes,
		CustomAnswers:       req.CustomAnswers,
		SubmittedAt:         time.Now(),
		IPAddress:           req.IPAddress,
		UserAgent:           req.UserAgent,
		Source:              req.Source,
		ConfirmationSent:    false,
	}

	if err := s.rsvpRepo.Create(ctx, rsvp); err != nil {
		return nil, fmt.Errorf("failed to create RSVP: %w", err)
	}

	// Update wedding RSVP count
	if err := s.weddingRepo.UpdateRSVPCount(ctx, weddingID); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Failed to update RSVP count: %v\n", err)
	}

	return rsvp, nil
}

// GetRSVPByID retrieves an RSVP by ID
func (s *RSVPService) GetRSVPByID(ctx context.Context, id primitive.ObjectID) (*models.RSVP, error) {
	rsvp, err := s.rsvpRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrRSVPNotFound
		}
		return nil, fmt.Errorf("failed to get RSVP: %w", err)
	}
	return rsvp, nil
}

// UpdateRSVP updates an existing RSVP
func (s *RSVPService) UpdateRSVP(ctx context.Context, id primitive.ObjectID, req UpdateRSVPRequest) (*models.RSVP, error) {
	// Get existing RSVP
	rsvp, err := s.rsvpRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrRSVPNotFound
		}
		return nil, fmt.Errorf("failed to get RSVP: %w", err)
	}

	// Check if RSVP can be modified
	if !rsvp.CanBeModified() {
		return nil, ErrRSVPCannotModify
	}

	// Update fields if provided
	if req.Status != nil {
		rsvp.Status = *req.Status
	}
	if req.AttendanceCount != nil {
		rsvp.AttendanceCount = *req.AttendanceCount
	}
	if req.PlusOnes != nil {
		rsvp.PlusOnes = *req.PlusOnes
		rsvp.PlusOneCount = len(*req.PlusOnes)
	}
	if req.DietaryRestrictions != nil {
		rsvp.DietaryRestrictions = *req.DietaryRestrictions
	}
	if req.DietarySelected != nil {
		rsvp.DietarySelected = *req.DietarySelected
	}
	if req.AdditionalNotes != nil {
		rsvp.AdditionalNotes = *req.AdditionalNotes
	}
	if req.CustomAnswers != nil {
		rsvp.CustomAnswers = *req.CustomAnswers
	}

	// Update timestamp
	now := time.Now()
	rsvp.UpdatedAt = &now

	// Validate updated RSVP
	wedding, err := s.weddingRepo.GetByID(ctx, rsvp.WeddingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wedding for validation: %w", err)
	}

	if err := s.validateRSVP(rsvp, wedding); err != nil {
		return nil, err
	}

	// Save updates
	if err := s.rsvpRepo.Update(ctx, rsvp); err != nil {
		return nil, fmt.Errorf("failed to update RSVP: %w", err)
	}

	// Update wedding RSVP count
	if err := s.weddingRepo.UpdateRSVPCount(ctx, rsvp.WeddingID); err != nil {
		fmt.Printf("Failed to update RSVP count: %v\n", err)
	}

	return rsvp, nil
}

// DeleteRSVP deletes an RSVP
func (s *RSVPService) DeleteRSVP(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) error {
	// Get RSVP to verify ownership
	rsvp, err := s.rsvpRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrRSVPNotFound
		}
		return fmt.Errorf("failed to get RSVP: %w", err)
	}

	// Get wedding to verify ownership
	wedding, err := s.weddingRepo.GetByID(ctx, rsvp.WeddingID)
	if err != nil {
		return fmt.Errorf("failed to get wedding: %w", err)
	}

	// Check ownership
	if wedding.UserID != userID {
		return ErrUnauthorized
	}

	// Delete RSVP
	if err := s.rsvpRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete RSVP: %w", err)
	}

	// Update wedding RSVP count
	if err := s.weddingRepo.UpdateRSVPCount(ctx, rsvp.WeddingID); err != nil {
		fmt.Printf("Failed to update RSVP count: %v\n", err)
	}

	return nil
}

// ListRSVPs retrieves RSVPs for a wedding
func (s *RSVPService) ListRSVPs(ctx context.Context, weddingID primitive.ObjectID, userID primitive.ObjectID, page, pageSize int, filters repository.RSVPFilters) ([]*models.RSVP, int64, error) {
	// Verify wedding ownership
	wedding, err := s.weddingRepo.GetByID(ctx, weddingID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, 0, ErrWeddingNotFound
		}
		return nil, 0, fmt.Errorf("failed to get wedding: %w", err)
	}

	if wedding.UserID != userID {
		return nil, 0, ErrUnauthorized
	}

	rsvps, total, err := s.rsvpRepo.ListByWedding(ctx, weddingID, page, pageSize, filters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list RSVPs: %w", err)
	}

	return rsvps, total, nil
}

// GetRSVPStatistics retrieves RSVP statistics for a wedding
func (s *RSVPService) GetRSVPStatistics(ctx context.Context, weddingID primitive.ObjectID, userID primitive.ObjectID) (*models.RSVPStatistics, error) {
	// Verify wedding ownership
	wedding, err := s.weddingRepo.GetByID(ctx, weddingID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrWeddingNotFound
		}
		return nil, fmt.Errorf("failed to get wedding: %w", err)
	}

	if wedding.UserID != userID {
		return nil, ErrUnauthorized
	}

	stats, err := s.rsvpRepo.GetStatistics(ctx, weddingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get RSVP statistics: %w", err)
	}

	return stats, nil
}

// ExportRSVPs exports RSVPs data (for CSV export)
func (s *RSVPService) ExportRSVPs(ctx context.Context, weddingID primitive.ObjectID, userID primitive.ObjectID) ([]*models.RSVP, error) {
	// Get all RSVPs for the wedding
	rsvps, _, err := s.ListRSVPs(ctx, weddingID, userID, 1, 10000, repository.RSVPFilters{})
	if err != nil {
		return nil, fmt.Errorf("failed to export RSVPs: %w", err)
	}

	return rsvps, nil
}

// Helper methods

func (s *RSVPService) isRSVPOpen(wedding *models.Wedding) bool {
	// If wedding is not published, RSVP is not open
	if wedding.Status != "published" {
		return false
	}

	// If RSVP settings exist, check them
	if wedding.RSVP != nil {
		now := time.Now()
		
		// Check if RSVP period is set and valid
		if wedding.RSVP.OpenDate != nil && now.Before(*wedding.RSVP.OpenDate) {
			return false
		}
		
		if wedding.RSVP.CloseDate != nil && now.After(*wedding.RSVP.CloseDate) {
			return false
		}
		
		// Check if RSVP is explicitly disabled
		if wedding.RSVP.Enabled != nil && !*wedding.RSVP.Enabled {
			return false
		}
	}

	return true
}

func (s *RSVPService) validateSubmitRequest(req SubmitRSVPRequest, wedding *models.Wedding) error {
	// Validate status
	validStatuses := []string{"attending", "not-attending", "maybe"}
	if !contains(validStatuses, req.Status) {
		return ErrInvalidRSVPStatus
	}

	// Validate plus ones
	if wedding.RSVP != nil && len(req.PlusOnes) > wedding.RSVP.MaxPlusOnes {
		return ErrTooManyPlusOnes
	}

	// Validate source
	validSources := []string{"web", "direct_link", "qr_code", "manual"}
	if req.Source == "" {
		req.Source = "web" // Default source
	} else if !contains(validSources, req.Source) {
		req.Source = "web" // Fallback to web if invalid
	}

	return nil
}

func (s *RSVPService) validateRSVP(rsvp *models.RSVP, wedding *models.Wedding) error {
	// Validate status
	validStatuses := []string{"attending", "not-attending", "maybe"}
	if !contains(validStatuses, rsvp.Status) {
		return ErrInvalidRSVPStatus
	}

	// Validate plus ones
	if wedding.RSVP != nil && rsvp.PlusOneCount > wedding.RSVP.MaxPlusOnes {
		return ErrTooManyPlusOnes
	}

	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}