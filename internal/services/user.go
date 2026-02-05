package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"wedding-invitation-backend/internal/domain/models"
	"wedding-invitation-backend/internal/domain/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserService provides business logic for user management
type UserService struct {
	userRepo repository.UserRepository
}

// NewUserService creates a new user service
func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// UserProfile represents user profile data for updates
type UserProfile struct {
	FirstName *string `json:"first_name" validate:"omitempty,min=1,max=50"`
	LastName  *string `json:"last_name" validate:"omitempty,min=1,max=50"`
	Phone     *string `json:"phone" validate:"omitempty,e164"`
}

// UserListResponse represents the response for user list
type UserListResponse struct {
	Users      []*models.User `json:"users"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}

// isValidUserStatus checks if the user status is valid
func (s *UserService) isValidUserStatus(status models.UserStatus) bool {
	validStatuses := []models.UserStatus{
		models.UserStatusActive,
		models.UserStatusInactive,
		models.UserStatusUnverified,
		models.UserStatusSuspended,
	}
	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}
	return false
}

// GetUserProfile retrieves a user's profile by ID
func (s *UserService) GetUserProfile(ctx context.Context, userID primitive.ObjectID) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// Clear sensitive data
	user.PasswordHash = ""
	user.EmailVerificationToken = ""
	user.PasswordResetToken = ""

	return user, nil
}

// UpdateUserProfile updates a user's profile information
func (s *UserService) UpdateUserProfile(ctx context.Context, userID primitive.ObjectID, profile *UserProfile) (*models.User, error) {
	// Get existing user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// Update fields if provided
	if profile.FirstName != nil {
		user.FirstName = *profile.FirstName
	}
	if profile.LastName != nil {
		user.LastName = *profile.LastName
	}
	if profile.Phone != nil {
		user.Phone = *profile.Phone
	}

	user.UpdatedAt = time.Now()

	// Save changes
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Clear sensitive data for response
	user.PasswordHash = ""
	user.EmailVerificationToken = ""
	user.PasswordResetToken = ""

	return user, nil
}

// UpdateUserStatus updates a user's status (admin only)
func (s *UserService) UpdateUserStatus(ctx context.Context, userID primitive.ObjectID, status models.UserStatus) error {
	// Validate status
	if !s.isValidUserStatus(status) {
		return errors.New("invalid user status")
	}

	// Get existing user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return errors.New("user not found")
	}

	// Update status
	user.Status = status
	user.UpdatedAt = time.Now()

	// Save changes
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user status: %w", err)
	}

	return nil
}

// DeleteUser deletes a user (soft delete by setting status to inactive)
func (s *UserService) DeleteUser(ctx context.Context, userID primitive.ObjectID) error {
	return s.UpdateUserStatus(ctx, userID, models.UserStatusInactive)
}

// GetUsersList retrieves a paginated list of users (admin only)
func (s *UserService) GetUsersList(ctx context.Context, page, pageSize int, filters repository.UserFilters) (*UserListResponse, error) {
	// Validate pagination
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Get users and total count
	users, total, err := s.userRepo.List(ctx, page, pageSize, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get users list: %w", err)
	}

	// Clear sensitive data
	for _, user := range users {
		user.PasswordHash = ""
		user.EmailVerificationToken = ""
		user.PasswordResetToken = ""
	}

	// Calculate total pages
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	return &UserListResponse{
		Users:      users,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// SearchUsers searches for users by email or name (admin only)
func (s *UserService) SearchUsers(ctx context.Context, query string, limit int) ([]*models.User, error) {
	if limit < 1 || limit > 50 {
		limit = 20
	}

	// Use the List method with search filter
	filters := repository.UserFilters{
		Search: query,
	}

	users, _, err := s.userRepo.List(ctx, 1, limit, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}

	// Clear sensitive data
	for _, user := range users {
		user.PasswordHash = ""
		user.EmailVerificationToken = ""
		user.PasswordResetToken = ""
	}

	return users, nil
}

// GetUserStats retrieves user statistics (admin only)
func (s *UserService) GetUserStats(ctx context.Context) (map[string]int64, error) {
	stats := make(map[string]int64)

	// Get counts by status using List method
	statuses := []models.UserStatus{
		models.UserStatusActive,
		models.UserStatusInactive,
		models.UserStatusUnverified,
		models.UserStatusSuspended,
	}

	for _, status := range statuses {
		filters := repository.UserFilters{
			Status: string(status),
		}
		_, count, err := s.userRepo.List(ctx, 1, 1, filters)
		if err != nil {
			return nil, fmt.Errorf("failed to get user stats for status %s: %w", status, err)
		}
		stats[string(status)] = count
	}

	return stats, nil
}

// AddWeddingToUser adds a wedding ID to a user's wedding list
func (s *UserService) AddWeddingToUser(ctx context.Context, userID, weddingID primitive.ObjectID) error {
	// Verify user exists
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return errors.New("user not found")
	}

	// Add wedding ID
	if err := s.userRepo.AddWeddingID(ctx, userID, weddingID); err != nil {
		return fmt.Errorf("failed to add wedding to user: %w", err)
	}

	return nil
}

// RemoveWeddingFromUser removes a wedding ID from a user's wedding list
func (s *UserService) RemoveWeddingFromUser(ctx context.Context, userID, weddingID primitive.ObjectID) error {
	// Verify user exists
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return errors.New("user not found")
	}

	// Remove wedding ID
	if err := s.userRepo.RemoveWeddingID(ctx, userID, weddingID); err != nil {
		return fmt.Errorf("failed to remove wedding from user: %w", err)
	}

	return nil
}

// GetUserWeddings retrieves all wedding IDs for a user
func (s *UserService) GetUserWeddings(ctx context.Context, userID primitive.ObjectID) ([]primitive.ObjectID, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	return user.WeddingIDs, nil
}

// ValidateUser validates user data and returns any errors
func (s *UserService) ValidateUser(user *models.User) error {
	if user == nil {
		return errors.New("user cannot be nil")
	}

	// Validate email
	if user.Email == "" {
		return errors.New("email is required")
	}
	if !s.isValidEmail(user.Email) {
		return errors.New("invalid email format")
	}

	// Validate name
	if user.FirstName == "" {
		return errors.New("first name is required")
	}
	if user.LastName == "" {
		return errors.New("last name is required")
	}

	// Validate status
	if !s.isValidUserStatus(user.Status) {
		return errors.New("invalid user status")
	}

	// Validate phone if provided
	if user.Phone != "" && !s.isValidPhone(user.Phone) {
		return errors.New("invalid phone number format")
	}

	return nil
}

// IsEmailAvailable checks if an email is available for registration
func (s *UserService) IsEmailAvailable(ctx context.Context, email string, excludeUserID *primitive.ObjectID) (bool, error) {
	existingUser, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return false, fmt.Errorf("failed to check email availability: %w", err)
	}

	// Email is available if no user exists or it's the same user (for updates)
	if existingUser == nil {
		return true, nil
	}

	if excludeUserID != nil && existingUser.ID == *excludeUserID {
		return true, nil
	}

	return false, nil
}

// isValidEmail validates email format
func (s *UserService) isValidEmail(email string) bool {
	email = strings.TrimSpace(email)
	if email == "" {
		return false
	}

	// Basic email validation
	at := strings.LastIndex(email, "@")
	if at <= 0 || at == len(email)-1 {
		return false
	}

	dot := strings.LastIndex(email, ".")
	if dot <= at+1 || dot == len(email)-1 {
		return false
	}

	return true
}

// isValidPhone validates phone number format (basic E.164 check)
func (s *UserService) isValidPhone(phone string) bool {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return false
	}

	// Basic E.164 validation: starts with + and contains only digits
	if !strings.HasPrefix(phone, "+") {
		return false
	}

	for _, char := range phone[1:] {
		if char < '0' || char > '9' {
			return false
		}
	}

	return len(phone) >= 8 && len(phone) <= 15
}
