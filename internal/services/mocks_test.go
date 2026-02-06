package services

import (
	"context"
	"wedding-invitation-backend/internal/domain/models"
	"wedding-invitation-backend/internal/domain/repository"

	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MockWeddingRepository is a mock implementation of WeddingRepository
type MockWeddingRepository struct {
	mock.Mock
}

func (m *MockWeddingRepository) Create(ctx context.Context, wedding *models.Wedding) error {
	args := m.Called(ctx, wedding)
	return args.Error(0)
}

func (m *MockWeddingRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.Wedding, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Wedding), args.Error(1)
}

func (m *MockWeddingRepository) GetBySlug(ctx context.Context, slug string) (*models.Wedding, error) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Wedding), args.Error(1)
}

func (m *MockWeddingRepository) GetByUserID(ctx context.Context, userID primitive.ObjectID, page, pageSize int, filters repository.WeddingFilters) ([]*models.Wedding, int64, error) {
	args := m.Called(ctx, userID, page, pageSize, filters)
	return args.Get(0).([]*models.Wedding), args.Get(1).(int64), args.Error(2)
}

func (m *MockWeddingRepository) Update(ctx context.Context, wedding *models.Wedding) error {
	args := m.Called(ctx, wedding)
	return args.Error(0)
}

func (m *MockWeddingRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockWeddingRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	args := m.Called(ctx, slug)
	return args.Bool(0), args.Error(1)
}

func (m *MockWeddingRepository) ListPublic(ctx context.Context, page, pageSize int, filters repository.PublicWeddingFilters) ([]*models.Wedding, int64, error) {
	args := m.Called(ctx, page, pageSize, filters)
	return args.Get(0).([]*models.Wedding), args.Get(1).(int64), args.Error(2)
}

func (m *MockWeddingRepository) IncrementViewCount(ctx context.Context, id primitive.ObjectID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockWeddingRepository) UpdateRSVPCount(ctx context.Context, weddingID primitive.ObjectID) error {
	args := m.Called(ctx, weddingID)
	return args.Error(0)
}

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
	user := args.Get(0)
	if user == nil {
		return nil, args.Error(1)
	}
	return user.(*models.User), args.Error(1)
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
	user := args.Get(0)
	if user == nil {
		return nil, args.Error(1)
	}
	return user.(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByResetToken(ctx context.Context, token string) (*models.User, error) {
	args := m.Called(ctx, token)
	user := args.Get(0)
	if user == nil {
		return nil, args.Error(1)
	}
	return user.(*models.User), args.Error(1)
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

// Helper functions for tests
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}
