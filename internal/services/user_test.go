package services

import (
	"context"
	"testing"
	"time"
	"wedding-invitation-backend/internal/domain/models"
	"wedding-invitation-backend/internal/domain/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	user := args.Get(0)
	if user == nil {
		return nil, args.Error(1)
	}
	return user.(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByVerificationToken(ctx context.Context, token string) (*models.User, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByResetToken(ctx context.Context, token string) (*models.User, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, page, pageSize int, filters repository.UserFilters) ([]*models.User, int64, error) {
	args := m.Called(ctx, page, pageSize, filters)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*models.User), args.Get(1).(int64), args.Error(2)
}

func (m *MockUserRepository) AddWeddingID(ctx context.Context, userID, weddingID primitive.ObjectID) error {
	args := m.Called(ctx, userID, weddingID)
	return args.Error(0)
}

func (m *MockUserRepository) RemoveWeddingID(ctx context.Context, userID, weddingID primitive.ObjectID) error {
	args := m.Called(ctx, userID, weddingID)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateLastLogin(ctx context.Context, userID primitive.ObjectID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepository) SetEmailVerified(ctx context.Context, userID primitive.ObjectID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func setupUserService() (*UserService, *MockUserRepository) {
	mockRepo := &MockUserRepository{}
	service := NewUserService(mockRepo)
	return service, mockRepo
}

func createTestUser() *models.User {
	userID := primitive.NewObjectID()
	return &models.User{
		ID:            userID,
		Email:         "test@example.com",
		FirstName:     "John",
		LastName:      "Doe",
		Phone:         "+1234567890",
		EmailVerified: true,
		Status:        models.UserStatusActive,
		Role:          "user",
		WeddingIDs:    []primitive.ObjectID{},
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

func TestUserService_GetUserProfile(t *testing.T) {
	service, mockRepo := setupUserService()
	ctx := context.Background()
	userID := primitive.NewObjectID()

	t.Run("success", func(t *testing.T) {
		user := createTestUser()
		user.ID = userID

		mockRepo.On("GetByID", ctx, userID).Return(user, nil)

		result, err := service.GetUserProfile(ctx, userID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, userID, result.ID)
		assert.Empty(t, result.PasswordHash) // Sensitive data should be cleared
		assert.Empty(t, result.EmailVerificationToken)
		assert.Empty(t, result.PasswordResetToken)

		mockRepo.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		// Create a new mock for this specific test to avoid conflicts
		mockRepo := &MockUserRepository{}
		service := NewUserService(mockRepo)

		mockRepo.On("GetByID", ctx, userID).Return((*models.User)(nil), nil)

		result, err := service.GetUserProfile(ctx, userID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "user not found")

		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		// Create a new mock for this specific test to avoid conflicts
		mockRepo := &MockUserRepository{}
		service := NewUserService(mockRepo)

		mockRepo.On("GetByID", ctx, userID).Return((*models.User)(nil), assert.AnError)

		result, err := service.GetUserProfile(ctx, userID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to get user")

		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_UpdateUserProfile(t *testing.T) {
	service, mockRepo := setupUserService()
	ctx := context.Background()
	userID := primitive.NewObjectID()

	t.Run("success", func(t *testing.T) {
		user := createTestUser()
		user.ID = userID

		profile := &UserProfile{
			FirstName: stringPtr("Jane"),
			LastName:  stringPtr("Smith"),
			Phone:     stringPtr("+9876543210"),
		}

		mockRepo.On("GetByID", ctx, userID).Return(user, nil)
		mockRepo.On("Update", ctx, mock.AnythingOfType("*models.User")).Return(nil)

		result, err := service.UpdateUserProfile(ctx, userID, profile)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Jane", result.FirstName)
		assert.Equal(t, "Smith", result.LastName)
		assert.Equal(t, "+9876543210", result.Phone)

		mockRepo.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		profile := &UserProfile{}
		// Create a new mock for this specific test to avoid conflicts
		mockRepo := &MockUserRepository{}
		service := NewUserService(mockRepo)

		mockRepo.On("GetByID", ctx, userID).Return((*models.User)(nil), nil)

		result, err := service.UpdateUserProfile(ctx, userID, profile)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "user not found")

		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_UpdateUserStatus(t *testing.T) {
	service, mockRepo := setupUserService()
	ctx := context.Background()
	userID := primitive.NewObjectID()

	t.Run("success", func(t *testing.T) {
		user := createTestUser()
		user.ID = userID

		mockRepo.On("GetByID", ctx, userID).Return(user, nil)
		mockRepo.On("Update", ctx, mock.AnythingOfType("*models.User")).Return(nil)

		err := service.UpdateUserStatus(ctx, userID, models.UserStatusInactive)

		assert.NoError(t, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid status", func(t *testing.T) {
		err := service.UpdateUserStatus(ctx, userID, models.UserStatus("invalid"))

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid user status")
	})

	t.Run("user not found", func(t *testing.T) {
		// Create a new mock for this specific test to avoid conflicts
		mockRepo := &MockUserRepository{}
		service := NewUserService(mockRepo)

		mockRepo.On("GetByID", ctx, userID).Return((*models.User)(nil), nil)

		err := service.UpdateUserStatus(ctx, userID, models.UserStatusInactive)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")

		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_GetUsersList(t *testing.T) {
	service, mockRepo := setupUserService()
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		users := []*models.User{createTestUser(), createTestUser()}
		filters := repository.UserFilters{}

		mockRepo.On("List", ctx, 1, 20, filters).Return(users, int64(2), nil)

		result, err := service.GetUsersList(ctx, 1, 20, filters)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Users, 2)
		assert.Equal(t, int64(2), result.Total)
		assert.Equal(t, 1, result.Page)
		assert.Equal(t, 20, result.PageSize)
		assert.Equal(t, 1, result.TotalPages)

		// Check that sensitive data is cleared
		for _, user := range result.Users {
			assert.Empty(t, user.PasswordHash)
			assert.Empty(t, user.EmailVerificationToken)
			assert.Empty(t, user.PasswordResetToken)
		}

		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid pagination defaults", func(t *testing.T) {
		users := []*models.User{}
		filters := repository.UserFilters{}

		mockRepo.On("List", ctx, 1, 20, filters).Return(users, int64(0), nil)

		result, err := service.GetUsersList(ctx, 0, 150, filters)

		assert.NoError(t, err)
		assert.Equal(t, 1, result.Page)
		assert.Equal(t, 20, result.PageSize)

		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_SearchUsers(t *testing.T) {
	service, mockRepo := setupUserService()
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		users := []*models.User{createTestUser()}
		filters := repository.UserFilters{Search: "test"}

		mockRepo.On("List", ctx, 1, 20, filters).Return(users, int64(1), nil)

		result, err := service.SearchUsers(ctx, "test", 20)

		assert.NoError(t, err)
		assert.Len(t, result, 1)

		mockRepo.AssertExpectations(t)
	})

	t.Run("limit defaults", func(t *testing.T) {
		users := []*models.User{}
		filters := repository.UserFilters{Search: "test"}
		// Create a new mock for this specific test to avoid conflicts
		mockRepo := &MockUserRepository{}
		service := NewUserService(mockRepo)

		mockRepo.On("List", ctx, 1, 20, filters).Return(users, int64(0), nil)

		result, err := service.SearchUsers(ctx, "test", 0)

		assert.NoError(t, err)
		assert.Len(t, result, 0)

		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_AddWeddingToUser(t *testing.T) {
	service, mockRepo := setupUserService()
	ctx := context.Background()
	userID := primitive.NewObjectID()
	weddingID := primitive.NewObjectID()

	t.Run("success", func(t *testing.T) {
		user := createTestUser()
		user.ID = userID

		mockRepo.On("GetByID", ctx, userID).Return(user, nil)
		mockRepo.On("AddWeddingID", ctx, userID, weddingID).Return(nil)

		err := service.AddWeddingToUser(ctx, userID, weddingID)

		assert.NoError(t, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		// Create a new mock for this specific test to avoid conflicts
		mockRepo := &MockUserRepository{}
		service := NewUserService(mockRepo)

		mockRepo.On("GetByID", ctx, userID).Return((*models.User)(nil), nil)

		err := service.AddWeddingToUser(ctx, userID, weddingID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")

		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_ValidateUser(t *testing.T) {
	service, _ := setupUserService()

	t.Run("valid user", func(t *testing.T) {
		user := createTestUser()

		err := service.ValidateUser(user)

		assert.NoError(t, err)
	})

	t.Run("nil user", func(t *testing.T) {
		err := service.ValidateUser(nil)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user cannot be nil")
	})

	t.Run("missing email", func(t *testing.T) {
		user := createTestUser()
		user.Email = ""

		err := service.ValidateUser(user)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "email is required")
	})

	t.Run("invalid email", func(t *testing.T) {
		user := createTestUser()
		user.Email = "invalid-email"

		err := service.ValidateUser(user)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid email format")
	})

	t.Run("missing first name", func(t *testing.T) {
		user := createTestUser()
		user.FirstName = ""

		err := service.ValidateUser(user)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "first name is required")
	})

	t.Run("missing last name", func(t *testing.T) {
		user := createTestUser()
		user.LastName = ""

		err := service.ValidateUser(user)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "last name is required")
	})

	t.Run("invalid phone", func(t *testing.T) {
		user := createTestUser()
		user.Phone = "invalid-phone"

		err := service.ValidateUser(user)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid phone number format")
	})
}

func TestUserService_IsEmailAvailable(t *testing.T) {
	service, mockRepo := setupUserService()
	ctx := context.Background()
	email := "test@example.com"

	t.Run("email available", func(t *testing.T) {
		// Create a new mock for this specific test to avoid conflicts
		mockRepo := &MockUserRepository{}
		service := NewUserService(mockRepo)

		mockRepo.On("GetByEmail", ctx, email).Return((*models.User)(nil), nil)

		available, err := service.IsEmailAvailable(ctx, email, nil)

		assert.NoError(t, err)
		assert.True(t, available)

		mockRepo.AssertExpectations(t)
	})

	t.Run("email taken", func(t *testing.T) {
		user := createTestUser()
		user.Email = email

		mockRepo.On("GetByEmail", ctx, email).Return(user, nil)

		available, err := service.IsEmailAvailable(ctx, email, nil)

		assert.NoError(t, err)
		assert.False(t, available)

		mockRepo.AssertExpectations(t)
	})

	t.Run("email available for same user", func(t *testing.T) {
		user := createTestUser()
		user.Email = email
		userID := user.ID // Use the same ID as the user

		// Create a new mock for this specific test to avoid conflicts
		mockRepo := &MockUserRepository{}
		service := NewUserService(mockRepo)

		mockRepo.On("GetByEmail", ctx, email).Return(user, nil)

		available, err := service.IsEmailAvailable(ctx, email, &userID)

		assert.NoError(t, err)
		assert.True(t, available)

		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_isValidUserStatus(t *testing.T) {
	service, _ := setupUserService()

	t.Run("valid statuses", func(t *testing.T) {
		validStatuses := []models.UserStatus{
			models.UserStatusActive,
			models.UserStatusInactive,
			models.UserStatusUnverified,
			models.UserStatusSuspended,
		}

		for _, status := range validStatuses {
			assert.True(t, service.isValidUserStatus(status), "Status %s should be valid", status)
		}
	})

	t.Run("invalid status", func(t *testing.T) {
		invalidStatus := models.UserStatus("invalid")
		assert.False(t, service.isValidUserStatus(invalidStatus))
	})
}

func TestUserService_isValidEmail(t *testing.T) {
	service, _ := setupUserService()

	t.Run("valid emails", func(t *testing.T) {
		validEmails := []string{
			"test@example.com",
			"user.name@domain.co.uk",
			"user+tag@example.org",
		}

		for _, email := range validEmails {
			assert.True(t, service.isValidEmail(email), "Email %s should be valid", email)
		}
	})

	t.Run("invalid emails", func(t *testing.T) {
		invalidEmails := []string{
			"",
			"invalid-email",
			"@example.com",
			"user@",
			"user@.com",
			"user@example.",
		}

		for _, email := range invalidEmails {
			assert.False(t, service.isValidEmail(email), "Email %s should be invalid", email)
		}
	})
}

func TestUserService_isValidPhone(t *testing.T) {
	service, _ := setupUserService()

	t.Run("valid phones", func(t *testing.T) {
		validPhones := []string{
			"+1234567890",
			"+441234567890",
			"+123456789",
		}

		for _, phone := range validPhones {
			assert.True(t, service.isValidPhone(phone), "Phone %s should be valid", phone)
		}
	})

	t.Run("invalid phones", func(t *testing.T) {
		invalidPhones := []string{
			"",
			"1234567890",
			"+abc123456",
			"+123456789012345",
			"+123",
		}

		for _, phone := range invalidPhones {
			assert.False(t, service.isValidPhone(phone), "Phone %s should be invalid", phone)
		}
	})
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
