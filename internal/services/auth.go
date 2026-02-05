package services

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"wedding-invitation-backend/internal/domain/models"
	"wedding-invitation-backend/internal/domain/repository"
	"wedding-invitation-backend/internal/utils"
)

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrEmailAlreadyExists  = errors.New("email already exists")
	ErrInvalidPassword     = errors.New("password does not meet requirements")
	ErrAccountNotVerified  = errors.New("account not verified")
	ErrAccountDisabled     = errors.New("account is disabled")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
)

type AuthService interface {
	Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error)
	Login(ctx context.Context, req LoginRequest) (*AuthResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error)
	Logout(ctx context.Context, userID string, tokenID string) error
	ChangePassword(ctx context.Context, userID primitive.ObjectID, req ChangePasswordRequest) error
	ForgotPassword(ctx context.Context, email string) (*PasswordResetResponse, error)
	ResetPassword(ctx context.Context, req ResetPasswordRequest) error
	VerifyEmail(ctx context.Context, token string) error
	GetProfile(ctx context.Context, userID primitive.ObjectID) (*models.User, error)
}

type authService struct {
	userRepo      repository.UserRepository
	jwtManager    *utils.JWTManager
	passValidator *utils.PasswordValidator
}

type RegisterRequest struct {
	FirstName string `json:"first_name" validate:"required,min=2,max=50"`
	LastName  string `json:"last_name" validate:"required,min=2,max=50"`
	Email     string `json:"email" validate:"required,email,max=100"`
	Password  string `json:"password" validate:"required,min=8,max=72"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	User         *models.User `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresAt    time.Time    `json:"expires_at"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8,max=72"`
}

type PasswordResetRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type PasswordResetResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

type ResetPasswordRequest struct {
	Token    string `json:"token" validate:"required"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

func NewAuthService(userRepo repository.UserRepository, jwtManager *utils.JWTManager) AuthService {
	return &authService{
		userRepo:      userRepo,
		jwtManager:    jwtManager,
		passValidator: utils.NewPasswordValidator(),
	}
}

func (s *authService) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	// Validate email uniqueness
	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, ErrEmailAlreadyExists
	}

	// Validate password strength
	if err := s.passValidator.Validate(req.Password); err != nil {
		return nil, ErrInvalidPassword
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &models.User{
		ID:           primitive.NewObjectID(),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Status:       models.UserStatusUnverified,
		Role:         "user",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Generate tokens
	tokenPair, err := s.jwtManager.GenerateTokenPair(user.ID, user.Email, []string{user.Role})
	if err != nil {
		return nil, err
	}

	// Update user login info
	now := time.Now()
	user.LastLoginAt = &now
	user.UpdatedAt = now
	s.userRepo.Update(ctx, user)

	return &AuthResponse{
		User:         user,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt,
	}, nil
}

func (s *authService) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// Check password
	if !utils.CheckPassword(user.PasswordHash, req.Password) {
		return nil, ErrInvalidCredentials
	}

	// Check account status
	if user.Status == models.UserStatusSuspended {
		return nil, ErrAccountDisabled
	}

	if user.Status == models.UserStatusUnverified {
		return nil, ErrAccountNotVerified
	}

	// Generate tokens
	tokenPair, err := s.jwtManager.GenerateTokenPair(user.ID, user.Email, []string{user.Role})
	if err != nil {
		return nil, err
	}

	// Update user login info
	now := time.Now()
	user.LastLoginAt = &now
	user.UpdatedAt = now
	s.userRepo.Update(ctx, user)

	return &AuthResponse{
		User:         user,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt,
	}, nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error) {
	// Validate refresh token
	tokenPair, err := s.jwtManager.RefreshAccessToken(refreshToken)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	// Extract user info from new access token
	userID, err := s.jwtManager.ExtractUserIDFromToken(tokenPair.AccessToken)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// Check account status
	if user.Status == models.UserStatusSuspended {
		return nil, ErrAccountDisabled
	}

	// Note: Refresh token management would require additional fields in User model
	// For now, we rely on JWT validation
	s.userRepo.Update(ctx, user)

	return &AuthResponse{
		User:         user,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt,
	}, nil
}

func (s *authService) Logout(ctx context.Context, userID string, tokenID string) error {
	_, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	// Note: Token blacklisting would require additional fields in User model
	// For now, we rely on token expiration
	return nil
}

func (s *authService) ChangePassword(ctx context.Context, userID primitive.ObjectID, req ChangePasswordRequest) error {
	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	// Verify current password
	if !utils.CheckPassword(user.PasswordHash, req.CurrentPassword) {
		return ErrInvalidCredentials
	}

	// Validate new password
	if err := s.passValidator.Validate(req.NewPassword); err != nil {
		return ErrInvalidPassword
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	// Update password
	user.PasswordHash = hashedPassword
	user.UpdatedAt = time.Now()

	return s.userRepo.Update(ctx, user)
}

func (s *authService) ForgotPassword(ctx context.Context, email string) (*PasswordResetResponse, error) {
	// Get user
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// Generate reset token
	resetToken, err := utils.GenerateResetToken()
	if err != nil {
		return nil, err
	}

	// Update user with reset token
	expiresAt := time.Now().Add(1 * time.Hour)
	user.PasswordResetToken = resetToken
	user.PasswordResetExpires = &expiresAt
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return &PasswordResetResponse{
		Token:     resetToken,
		ExpiresAt: expiresAt,
	}, nil
}

func (s *authService) ResetPassword(ctx context.Context, req ResetPasswordRequest) error {
	// Get user by reset token
	user, err := s.userRepo.GetByResetToken(ctx, req.Token)
	if err != nil {
		return ErrUserNotFound
	}

	// Check if token is still valid
	if user.PasswordResetExpires == nil || time.Now().After(*user.PasswordResetExpires) {
		return errors.New("reset token has expired")
	}

	// Validate new password
	if err := s.passValidator.Validate(req.Password); err != nil {
		return ErrInvalidPassword
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return err
	}

	// Update password and clear reset token
	user.PasswordHash = hashedPassword
	user.PasswordResetToken = ""
	user.PasswordResetExpires = nil
	user.UpdatedAt = time.Now()

	return s.userRepo.Update(ctx, user)
}

func (s *authService) VerifyEmail(ctx context.Context, token string) error {
	// Get user by verification token
	user, err := s.userRepo.GetByVerificationToken(ctx, token)
	if err != nil {
		return ErrUserNotFound
	}

	// Update user status
	now := time.Now()
	user.Status = models.UserStatusActive
	user.EmailVerificationToken = ""
	user.EmailVerifiedAt = &now
	user.UpdatedAt = now

	return s.userRepo.Update(ctx, user)
}

func (s *authService) GetProfile(ctx context.Context, userID primitive.ObjectID) (*models.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}
