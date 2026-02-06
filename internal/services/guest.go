package services

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"wedding-invitation-backend/internal/domain/models"
	"wedding-invitation-backend/internal/domain/repository"
)

// GuestService handles guest-related business logic
type GuestService struct {
	guestRepo   repository.GuestRepository
	weddingRepo repository.WeddingRepository
}

// NewGuestService creates a new guest service
func NewGuestService(guestRepo repository.GuestRepository, weddingRepo repository.WeddingRepository) *GuestService {
	return &GuestService{
		guestRepo:   guestRepo,
		weddingRepo: weddingRepo,
	}
}

// CreateGuest creates a new guest
func (s *GuestService) CreateGuest(ctx context.Context, weddingID, userID primitive.ObjectID, guest *models.Guest) error {
	// Verify wedding exists and user owns it
	wedding, err := s.weddingRepo.GetByID(ctx, weddingID)
	if err != nil {
		return fmt.Errorf("wedding not found: %w", err)
	}

	if wedding.UserID != userID {
		return errors.New("unauthorized: you don't own this wedding")
	}

	// Set wedding ID
	guest.WeddingID = weddingID
	guest.CreatedBy = userID

	// Validate guest data
	if err := s.validateGuest(guest); err != nil {
		return fmt.Errorf("invalid guest data: %w", err)
	}

	// Check for duplicate email within the same wedding
	if guest.Email != "" {
		existingGuest, err := s.guestRepo.GetByEmail(ctx, weddingID, guest.Email)
		if err == nil && existingGuest != nil {
			return errors.New("guest with this email already exists for this wedding")
		}
	}

	return s.guestRepo.Create(ctx, guest)
}

// GetGuestByID retrieves a guest by ID
func (s *GuestService) GetGuestByID(ctx context.Context, guestID, userID primitive.ObjectID) (*models.Guest, error) {
	guest, err := s.guestRepo.GetByID(ctx, guestID)
	if err != nil {
		return nil, err
	}

	// Verify user owns the wedding
	if err := s.verifyWeddingOwnership(ctx, guest.WeddingID, userID); err != nil {
		return nil, err
	}

	return guest, nil
}

// ListGuests retrieves guests for a wedding with pagination and filtering
func (s *GuestService) ListGuests(ctx context.Context, weddingID, userID primitive.ObjectID, page, pageSize int, filters repository.GuestFilters) ([]*models.Guest, int64, error) {
	// Verify user owns the wedding
	if err := s.verifyWeddingOwnership(ctx, weddingID, userID); err != nil {
		return nil, 0, err
	}

	return s.guestRepo.ListByWedding(ctx, weddingID, page, pageSize, filters)
}

// UpdateGuest updates an existing guest
func (s *GuestService) UpdateGuest(ctx context.Context, guestID, userID primitive.ObjectID, guest *models.Guest) error {
	// Get existing guest
	existingGuest, err := s.guestRepo.GetByID(ctx, guestID)
	if err != nil {
		return fmt.Errorf("guest not found: %w", err)
	}

	// Verify user owns the wedding
	if err := s.verifyWeddingOwnership(ctx, existingGuest.WeddingID, userID); err != nil {
		return err
	}

	// Preserve immutable fields
	guest.ID = guestID
	guest.WeddingID = existingGuest.WeddingID
	guest.CreatedAt = existingGuest.CreatedAt
	guest.CreatedBy = existingGuest.CreatedBy

	// Validate guest data
	if err := s.validateGuest(guest); err != nil {
		return fmt.Errorf("invalid guest data: %w", err)
	}

	// Check for duplicate email (if email changed)
	if guest.Email != "" && guest.Email != existingGuest.Email {
		existingEmailGuest, err := s.guestRepo.GetByEmail(ctx, existingGuest.WeddingID, guest.Email)
		if err == nil && existingEmailGuest != nil && existingEmailGuest.ID != guestID {
			return errors.New("guest with this email already exists for this wedding")
		}
	}

	return s.guestRepo.Update(ctx, guest)
}

// DeleteGuest deletes a guest
func (s *GuestService) DeleteGuest(ctx context.Context, guestID, userID primitive.ObjectID) error {
	// Get existing guest
	guest, err := s.guestRepo.GetByID(ctx, guestID)
	if err != nil {
		return fmt.Errorf("guest not found: %w", err)
	}

	// Verify user owns the wedding
	if err := s.verifyWeddingOwnership(ctx, guest.WeddingID, userID); err != nil {
		return err
	}

	return s.guestRepo.Delete(ctx, guestID)
}

// ImportGuestsFromCSV imports guests from a CSV file
func (s *GuestService) ImportGuestsFromCSV(ctx context.Context, weddingID, userID primitive.ObjectID, csvData io.Reader) (*models.GuestImportResult, error) {
	// Verify user owns the wedding
	if err := s.verifyWeddingOwnership(ctx, weddingID, userID); err != nil {
		return nil, err
	}

	// Generate batch ID
	batchID := fmt.Sprintf("%s_%d", userID.Hex(), time.Now().Unix())

	// Parse CSV
	records, err := csv.NewReader(csvData).ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}

	if len(records) < 2 {
		return nil, errors.New("CSV file must contain at least a header row and one data row")
	}

	// Get headers
	headers := records[0]
	var guests []*models.Guest
	var errors []string
	successCount := 0

	// Process each row
	for i := 1; i < len(records); i++ {
		row := records[i]
		guest, err := s.parseGuestFromCSV(row, headers, weddingID, userID, batchID)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Row %d: %v", i+1, err))
			continue
		}

		// Validate guest
		if err := s.validateGuest(guest); err != nil {
			errors = append(errors, fmt.Sprintf("Row %d: %v", i+1, err))
			continue
		}

		guests = append(guests, guest)
		successCount++
	}

	// Import valid guests
	if len(guests) > 0 {
		if err := s.guestRepo.ImportBatch(ctx, guests, batchID); err != nil {
			return nil, fmt.Errorf("failed to import guests: %w", err)
		}
	}

	result := &models.GuestImportResult{
		SuccessCount: successCount,
		ErrorCount:   len(errors),
		Errors:       errors,
		BatchID:      batchID,
	}

	return result, nil
}

// GetImportBatch retrieves guests from a specific import batch
func (s *GuestService) GetImportBatch(ctx context.Context, weddingID, userID primitive.ObjectID, batchID string) ([]*models.Guest, error) {
	// Verify user owns the wedding
	if err := s.verifyWeddingOwnership(ctx, weddingID, userID); err != nil {
		return nil, err
	}

	return s.guestRepo.GetByImportBatch(ctx, weddingID, batchID)
}

// CreateManyGuests creates multiple guests at once
func (s *GuestService) CreateManyGuests(ctx context.Context, weddingID, userID primitive.ObjectID, guests []*models.Guest) error {
	// Verify user owns the wedding
	if err := s.verifyWeddingOwnership(ctx, weddingID, userID); err != nil {
		return err
	}

	// Set required fields for all guests
	for _, guest := range guests {
		guest.WeddingID = weddingID
		guest.CreatedBy = userID

		// Validate guest
		if err := s.validateGuest(guest); err != nil {
			return fmt.Errorf("invalid guest data: %w", err)
		}
	}

	return s.guestRepo.CreateMany(ctx, guests)
}

// verifyWeddingOwnership verifies that the user owns the wedding
func (s *GuestService) verifyWeddingOwnership(ctx context.Context, weddingID, userID primitive.ObjectID) error {
	wedding, err := s.weddingRepo.GetByID(ctx, weddingID)
	if err != nil {
		return fmt.Errorf("wedding not found: %w", err)
	}

	if wedding.UserID != userID {
		return errors.New("unauthorized: you don't own this wedding")
	}

	return nil
}

// validateGuest validates guest data
func (s *GuestService) validateGuest(guest *models.Guest) error {
	if guest.FirstName == "" {
		return errors.New("first name is required")
	}

	if guest.LastName == "" {
		return errors.New("last name is required")
	}

	if len(guest.FirstName) > 50 {
		return errors.New("first name must be 50 characters or less")
	}

	if len(guest.LastName) > 50 {
		return errors.New("last name must be 50 characters or less")
	}

	// Validate email if provided
	if guest.Email != "" {
		if len(guest.Email) > 100 {
			return errors.New("email must be 100 characters or less")
		}
		// Basic email validation
		if !strings.Contains(guest.Email, "@") || !strings.Contains(guest.Email, ".") {
			return errors.New("invalid email format")
		}
	}

	// Validate side if provided
	if guest.Side != "" && guest.Side != "bride" && guest.Side != "groom" && guest.Side != "both" {
		return errors.New("side must be one of: bride, groom, both")
	}

	// Validate invited via if provided
	if guest.InvitedVia != "" && guest.InvitedVia != "digital" && guest.InvitedVia != "manual" {
		return errors.New("invited via must be one of: digital, manual")
	}

	// Validate invitation status if provided
	if guest.InvitationStatus != "" {
		validStatuses := []string{"pending", "sent", "delivered", "failed"}
		valid := false
		for _, status := range validStatuses {
			if guest.InvitationStatus == status {
				valid = true
				break
			}
		}
		if !valid {
			return errors.New("invitation status must be one of: pending, sent, delivered, failed")
		}
	}

	// Validate RSVP status if provided
	if guest.RSVPStatus != "" {
		validStatuses := []string{"attending", "not-attending", "maybe", "pending"}
		valid := false
		for _, status := range validStatuses {
			if guest.RSVPStatus == status {
				valid = true
				break
			}
		}
		if !valid {
			return errors.New("RSVP status must be one of: attending, not-attending, maybe, pending")
		}
	}

	// Validate max plus ones
	if guest.MaxPlusOnes < 0 || guest.MaxPlusOnes > 5 {
		return errors.New("max plus ones must be between 0 and 5")
	}

	return nil
}

// parseGuestFromCSV parses a guest from a CSV row
func (s *GuestService) parseGuestFromCSV(row []string, headers []string, weddingID, userID primitive.ObjectID, batchID string) (*models.Guest, error) {
	guest := &models.Guest{
		WeddingID:     weddingID,
		CreatedBy:     userID,
		ImportBatchID: batchID,
	}

	// Create header to column index map
	headerMap := make(map[string]int)
	for i, header := range headers {
		headerMap[strings.ToLower(strings.TrimSpace(header))] = i
	}

	// Map CSV columns to guest fields
	if idx, exists := headerMap["first_name"]; exists && idx < len(row) {
		guest.FirstName = strings.TrimSpace(row[idx])
	}

	if idx, exists := headerMap["last_name"]; exists && idx < len(row) {
		guest.LastName = strings.TrimSpace(row[idx])
	}

	if idx, exists := headerMap["email"]; exists && idx < len(row) {
		guest.Email = strings.TrimSpace(row[idx])
	}

	if idx, exists := headerMap["phone"]; exists && idx < len(row) {
		guest.Phone = strings.TrimSpace(row[idx])
	}

	if idx, exists := headerMap["relationship"]; exists && idx < len(row) {
		guest.Relationship = strings.TrimSpace(row[idx])
	}

	if idx, exists := headerMap["side"]; exists && idx < len(row) {
		guest.Side = strings.TrimSpace(row[idx])
	}

	if idx, exists := headerMap["invited_via"]; exists && idx < len(row) {
		guest.InvitedVia = strings.TrimSpace(row[idx])
	} else {
		guest.InvitedVia = "manual" // Default for CSV imports
	}

	if idx, exists := headerMap["invitation_status"]; exists && idx < len(row) {
		guest.InvitationStatus = strings.TrimSpace(row[idx])
	} else {
		guest.InvitationStatus = "pending" // Default
	}

	if idx, exists := headerMap["allow_plus_one"]; exists && idx < len(row) {
		plusOne := strings.ToLower(strings.TrimSpace(row[idx]))
		guest.AllowPlusOne = plusOne == "true" || plusOne == "yes" || plusOne == "1"
	}

	if idx, exists := headerMap["max_plus_ones"]; exists && idx < len(row) {
		maxPlusOnes := strings.TrimSpace(row[idx])
		if maxPlusOnes != "" {
			fmt.Sscanf(maxPlusOnes, "%d", &guest.MaxPlusOnes)
		}
	}

	if idx, exists := headerMap["vip"]; exists && idx < len(row) {
		vip := strings.ToLower(strings.TrimSpace(row[idx]))
		guest.VIP = vip == "true" || vip == "yes" || vip == "1"
	}

	if idx, exists := headerMap["notes"]; exists && idx < len(row) {
		guest.Notes = strings.TrimSpace(row[idx])
	}

	// Validate required fields
	if guest.FirstName == "" || guest.LastName == "" {
		return nil, errors.New("first name and last name are required")
	}

	return guest, nil
}