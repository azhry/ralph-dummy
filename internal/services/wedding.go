package services

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strings"
	"time"
	"wedding-invitation-backend/internal/domain/models"
	"wedding-invitation-backend/internal/domain/repository"
	"wedding-invitation-backend/internal/utils"
)

// WeddingService provides business logic for wedding management
type WeddingService struct {
	weddingRepo repository.WeddingRepository
	userRepo    repository.UserRepository
}

// NewWeddingService creates a new wedding service
func NewWeddingService(weddingRepo repository.WeddingRepository, userRepo repository.UserRepository) *WeddingService {
	return &WeddingService{
		weddingRepo: weddingRepo,
		userRepo:    userRepo,
	}
}

// CreateWedding creates a new wedding
func (s *WeddingService) CreateWedding(ctx context.Context, wedding *models.Wedding, userID primitive.ObjectID) error {
	// Validate wedding data
	if err := s.validateWedding(wedding, true); err != nil {
		return err
	}

	// Set user ID
	wedding.UserID = userID

	// Generate unique slug if not provided
	if wedding.Slug == "" {
		slug, err := s.generateUniqueSlug(ctx, wedding.Title)
		if err != nil {
			return fmt.Errorf("failed to generate slug: %w", err)
		}
		wedding.Slug = slug
	} else {
		// Check if slug is available
		exists, err := s.weddingRepo.ExistsBySlug(ctx, wedding.Slug)
		if err != nil {
			return fmt.Errorf("failed to check slug availability: %w", err)
		}
		if exists {
			return errors.New("slug already exists")
		}
	}

	// Set default values
	wedding.Status = string(models.WeddingStatusDraft)
	wedding.RSVPCount = 0
	wedding.GuestCount = 0
	wedding.TotalAttending = 0
	wedding.ViewCount = 0
	wedding.GalleryEnabled = false
	wedding.IsPublic = false

	// Set default theme if not provided
	if wedding.Theme.ThemeID == "" {
		wedding.Theme.ThemeID = "default"
	}

	// Validate theme settings
	if err := s.validateThemeSettings(&wedding.Theme); err != nil {
		return err
	}

	// Validate RSVP settings
	if err := s.validateRSVPSettings(&wedding.RSVP); err != nil {
		return err
	}

	// Create wedding
	if err := s.weddingRepo.Create(ctx, wedding); err != nil {
		return fmt.Errorf("failed to create wedding: %w", err)
	}

	// Add wedding ID to user's weddings list
	if err := s.userRepo.AddWeddingID(ctx, userID, wedding.ID); err != nil {
		// Log error but don't fail the operation
		// In production, you might want to handle this more gracefully
	}

	return nil
}

// GetWeddingByID retrieves a wedding by ID
func (s *WeddingService) GetWeddingByID(ctx context.Context, id primitive.ObjectID, requestingUserID primitive.ObjectID) (*models.Wedding, error) {
	wedding, err := s.weddingRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get wedding: %w", err)
	}

	if wedding == nil {
		return nil, errors.New("wedding not found")
	}

	// Check access permissions
	if !s.canAccessWedding(wedding, requestingUserID) {
		return nil, errors.New("access denied")
	}

	// Increment view count if not the owner
	if wedding.UserID != requestingUserID {
		if err := s.weddingRepo.IncrementViewCount(ctx, id); err != nil {
			// Log error but don't fail the request
		}
	}

	return wedding, nil
}

// GetWeddingBySlug retrieves a wedding by slug
func (s *WeddingService) GetWeddingBySlug(ctx context.Context, slug string, requestingUserID primitive.ObjectID) (*models.Wedding, error) {
	wedding, err := s.weddingRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to get wedding: %w", err)
	}

	if wedding == nil {
		return nil, errors.New("wedding not found")
	}

	// Check access permissions
	if !s.canAccessWedding(wedding, requestingUserID) {
		return nil, errors.New("access denied")
	}

	// Increment view count if not the owner
	if wedding.UserID != requestingUserID {
		if err := s.weddingRepo.IncrementViewCount(ctx, wedding.ID); err != nil {
			// Log error but don't fail the request
		}
	}

	return wedding, nil
}

// GetUserWeddings retrieves all weddings for a user with pagination
func (s *WeddingService) GetUserWeddings(ctx context.Context, userID primitive.ObjectID, page, pageSize int, filters repository.WeddingFilters) ([]*models.Wedding, int64, error) {
	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	weddings, total, err := s.weddingRepo.GetByUserID(ctx, userID, page, pageSize, filters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get user weddings: %w", err)
	}

	return weddings, total, nil
}

// UpdateWedding updates an existing wedding
func (s *WeddingService) UpdateWedding(ctx context.Context, wedding *models.Wedding, requestingUserID primitive.ObjectID) error {
	// Get existing wedding
	existingWedding, err := s.weddingRepo.GetByID(ctx, wedding.ID)
	if err != nil {
		return fmt.Errorf("failed to get existing wedding: %w", err)
	}

	if existingWedding == nil {
		return errors.New("wedding not found")
	}

	// Check ownership
	if existingWedding.UserID != requestingUserID {
		return errors.New("access denied")
	}

	// Validate wedding data
	if err := s.validateWedding(wedding, false); err != nil {
		return err
	}

	// Check if slug changed and is available
	if wedding.Slug != existingWedding.Slug {
		exists, err := s.weddingRepo.ExistsBySlug(ctx, wedding.Slug)
		if err != nil {
			return fmt.Errorf("failed to check slug availability: %w", err)
		}
		if exists {
			return errors.New("slug already exists")
		}
	}

	// Validate theme settings
	if err := s.validateThemeSettings(&wedding.Theme); err != nil {
		return err
	}

	// Validate RSVP settings
	if err := s.validateRSVPSettings(&wedding.RSVP); err != nil {
		return err
	}

	// Preserve certain fields that shouldn't be changed via update
	wedding.UserID = existingWedding.UserID
	wedding.CreatedAt = existingWedding.CreatedAt
	wedding.ViewCount = existingWedding.ViewCount
	wedding.RSVPCount = existingWedding.RSVPCount
	wedding.GuestCount = existingWedding.GuestCount
	wedding.TotalAttending = existingWedding.TotalAttending

	// Handle status changes
	if wedding.Status != existingWedding.Status {
		if err := s.handleStatusChange(ctx, wedding, existingWedding); err != nil {
			return err
		}
	}

	// Update wedding
	if err := s.weddingRepo.Update(ctx, wedding); err != nil {
		return fmt.Errorf("failed to update wedding: %w", err)
	}

	return nil
}

// DeleteWedding deletes a wedding
func (s *WeddingService) DeleteWedding(ctx context.Context, weddingID primitive.ObjectID, requestingUserID primitive.ObjectID) error {
	// Get wedding to check ownership
	wedding, err := s.weddingRepo.GetByID(ctx, weddingID)
	if err != nil {
		return fmt.Errorf("failed to get wedding: %w", err)
	}

	if wedding == nil {
		return errors.New("wedding not found")
	}

	// Check ownership
	if wedding.UserID != requestingUserID {
		return errors.New("access denied")
	}

	// Delete wedding
	if err := s.weddingRepo.Delete(ctx, weddingID); err != nil {
		return fmt.Errorf("failed to delete wedding: %w", err)
	}

	// Remove wedding ID from user's weddings list
	if err := s.userRepo.RemoveWeddingID(ctx, requestingUserID, weddingID); err != nil {
		// Log error but don't fail the operation
	}

	return nil
}

// PublishWedding publishes a wedding
func (s *WeddingService) PublishWedding(ctx context.Context, weddingID primitive.ObjectID, requestingUserID primitive.ObjectID) error {
	wedding, err := s.weddingRepo.GetByID(ctx, weddingID)
	if err != nil {
		return fmt.Errorf("failed to get wedding: %w", err)
	}

	if wedding == nil {
		return errors.New("wedding not found")
	}

	// Check ownership
	if wedding.UserID != requestingUserID {
		return errors.New("access denied")
	}

	// Validate wedding is ready for publishing
	if err := s.validateWeddingForPublishing(wedding); err != nil {
		return err
	}

	// Update status and publish date
	now := time.Now()
	wedding.Status = string(models.WeddingStatusPublished)
	wedding.PublishedAt = &now
	wedding.UpdatedAt = now

	// Update wedding
	if err := s.weddingRepo.Update(ctx, wedding); err != nil {
		return fmt.Errorf("failed to publish wedding: %w", err)
	}

	return nil
}

// ListPublicWeddings retrieves public weddings with pagination
func (s *WeddingService) ListPublicWeddings(ctx context.Context, page, pageSize int, filters repository.PublicWeddingFilters) ([]*models.Wedding, int64, error) {
	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	weddings, total, err := s.weddingRepo.ListPublic(ctx, page, pageSize, filters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get public weddings: %w", err)
	}

	return weddings, total, nil
}

// Helper functions

func (s *WeddingService) validateWedding(wedding *models.Wedding, isNew bool) error {
	// Validate basic required fields
	if wedding.Title == "" {
		return errors.New("title is required")
	}

	if wedding.Slug != "" {
		if err := utils.ValidateSlug(wedding.Slug); err != nil {
			return fmt.Errorf("invalid slug: %w", err)
		}
	}

	// Validate couple information
	if wedding.Couple.Partner1.FirstName == "" || wedding.Couple.Partner1.LastName == "" {
		return errors.New("partner1 first name and last name are required")
	}

	if wedding.Couple.Partner2.FirstName == "" || wedding.Couple.Partner2.LastName == "" {
		return errors.New("partner2 first name and last name are required")
	}

	// Validate event details
	if wedding.Event.Title == "" {
		return errors.New("event title is required")
	}

	if wedding.Event.VenueName == "" {
		return errors.New("venue name is required")
	}

	if wedding.Event.VenueAddress == "" {
		return errors.New("venue address is required")
	}

	if wedding.Event.Date.IsZero() {
		return errors.New("event date is required")
	}

	// Validate status
	validStatuses := []string{
		string(models.WeddingStatusDraft),
		string(models.WeddingStatusPublished),
		string(models.WeddingStatusExpired),
		string(models.WeddingStatusArchived),
	}

	if wedding.Status != "" && !utils.Contains(validStatuses, wedding.Status) {
		return errors.New("invalid wedding status")
	}

	return nil
}

func (s *WeddingService) validateThemeSettings(theme *models.ThemeSettings) error {
	if theme.ThemeID == "" {
		return errors.New("theme ID is required")
	}

	// Validate hex colors if provided
	if theme.PrimaryColor != "" {
		if err := utils.ValidateHexColor(theme.PrimaryColor); err != nil {
			return fmt.Errorf("invalid primary color: %w", err)
		}
	}

	if theme.SecondaryColor != "" {
		if err := utils.ValidateHexColor(theme.SecondaryColor); err != nil {
			return fmt.Errorf("invalid secondary color: %w", err)
		}
	}

	if theme.BackgroundColor != "" {
		if err := utils.ValidateHexColor(theme.BackgroundColor); err != nil {
			return fmt.Errorf("invalid background color: %w", err)
		}
	}

	return nil
}

func (s *WeddingService) validateRSVPSettings(rsvp *models.RSVPSettings) error {
	if rsvp.Enabled {
		// Validate max plus ones
		if rsvp.MaxPlusOnes < 0 || rsvp.MaxPlusOnes > 5 {
			return errors.New("max plus ones must be between 0 and 5")
		}

		// Validate custom questions
		for i, q := range rsvp.CustomQuestions {
			if q.Question == "" {
				return fmt.Errorf("custom question %d: question is required", i+1)
			}

			validTypes := []string{"text", "textarea", "select", "checkbox", "radio"}
			if !utils.Contains(validTypes, q.Type) {
				return fmt.Errorf("custom question %d: invalid question type", i+1)
			}

			// For select/radio types, options are required
			if (q.Type == "select" || q.Type == "radio") && len(q.Options) == 0 {
				return fmt.Errorf("custom question %d: options are required for %s type", i+1, q.Type)
			}
		}
	}

	return nil
}

func (s *WeddingService) validateWeddingForPublishing(wedding *models.Wedding) error {
	// Check if required fields are present for publishing
	if wedding.Couple.Partner1.FirstName == "" || wedding.Couple.Partner1.LastName == "" {
		return errors.New("partner1 information is required for publishing")
	}

	if wedding.Couple.Partner2.FirstName == "" || wedding.Couple.Partner2.LastName == "" {
		return errors.New("partner2 information is required for publishing")
	}

	if wedding.Event.Title == "" || wedding.Event.VenueName == "" || wedding.Event.VenueAddress == "" {
		return errors.New("complete event information is required for publishing")
	}

	if wedding.Event.Date.IsZero() {
		return errors.New("event date is required for publishing")
	}

	return nil
}

func (s *WeddingService) generateUniqueSlug(ctx context.Context, title string) (string, error) {
	// Generate base slug from title
	baseSlug := strings.ToLower(strings.ReplaceAll(title, " ", "-"))
	baseSlug = utils.SanitizeSlug(baseSlug)

	// If base slug is available, use it
	exists, err := s.weddingRepo.ExistsBySlug(ctx, baseSlug)
	if err != nil {
		return "", err
	}
	if !exists {
		return baseSlug, nil
	}

	// Try with random suffix
	for i := 1; i <= 100; i++ {
		candidateSlug := fmt.Sprintf("%s-%d", baseSlug, i)
		exists, err := s.weddingRepo.ExistsBySlug(ctx, candidateSlug)
		if err != nil {
			return "", err
		}
		if !exists {
			return candidateSlug, nil
		}
	}

	return "", errors.New("failed to generate unique slug")
}

func (s *WeddingService) canAccessWedding(wedding *models.Wedding, requestingUserID primitive.ObjectID) bool {
	// Owner can always access
	if wedding.UserID == requestingUserID {
		return true
	}

	// Public weddings can be accessed by anyone
	if wedding.IsPublic && wedding.Status == string(models.WeddingStatusPublished) {
		return true
	}

	// Published weddings can be accessed (this would include password check in future)
	if wedding.Status == string(models.WeddingStatusPublished) {
		return true
	}

	return false
}

func (s *WeddingService) handleStatusChange(ctx context.Context, newWedding *models.Wedding, oldWedding *models.Wedding) error {
	// Handle transition to published
	if newWedding.Status == string(models.WeddingStatusPublished) && oldWedding.Status != string(models.WeddingStatusPublished) {
		now := time.Now()
		newWedding.PublishedAt = &now

		// Validate wedding is ready for publishing
		if err := s.validateWeddingForPublishing(newWedding); err != nil {
			return err
		}
	}

	return nil
}

// GetWeddingBySlugForPublic retrieves a wedding by slug for public access
func (s *WeddingService) GetWeddingBySlugForPublic(ctx context.Context, slug string) (*models.Wedding, error) {
	wedding, err := s.weddingRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to get wedding: %w", err)
	}

	if wedding == nil {
		return nil, errors.New("wedding not found")
	}

	// Check if wedding is published
	if wedding.Status != string(models.WeddingStatusPublished) {
		return nil, errors.New("wedding not published")
	}

	// Increment view count for public access
	if err := s.weddingRepo.IncrementViewCount(ctx, wedding.ID); err != nil {
		// Log error but don't fail the request
	}

	return wedding, nil
}
