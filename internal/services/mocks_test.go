package services

import (
	"context"
	"time"
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

// MockAnalyticsRepository is a mock implementation of AnalyticsRepository
type MockAnalyticsRepository struct {
	mock.Mock
}

func (m *MockAnalyticsRepository) TrackPageView(ctx context.Context, pageView *models.PageView) error {
	args := m.Called(ctx, pageView)
	return args.Error(0)
}

func (m *MockAnalyticsRepository) GetPageViews(ctx context.Context, weddingID primitive.ObjectID, filter *models.AnalyticsFilter) ([]*models.PageView, int64, error) {
	args := m.Called(ctx, weddingID, filter)
	return args.Get(0).([]*models.PageView), args.Get(1).(int64), args.Error(2)
}

func (m *MockAnalyticsRepository) TrackRSVPEvent(ctx context.Context, event *models.RSVPAnalytics) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockAnalyticsRepository) GetRSVPAnalytics(ctx context.Context, weddingID primitive.ObjectID, filter *models.AnalyticsFilter) ([]*models.RSVPAnalytics, int64, error) {
	args := m.Called(ctx, weddingID, filter)
	return args.Get(0).([]*models.RSVPAnalytics), args.Get(1).(int64), args.Error(2)
}

func (m *MockAnalyticsRepository) TrackConversion(ctx context.Context, event *models.ConversionEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockAnalyticsRepository) GetConversions(ctx context.Context, weddingID primitive.ObjectID, filter *models.AnalyticsFilter) ([]*models.ConversionEvent, int64, error) {
	args := m.Called(ctx, weddingID, filter)
	return args.Get(0).([]*models.ConversionEvent), args.Get(1).(int64), args.Error(2)
}

func (m *MockAnalyticsRepository) GetWeddingAnalytics(ctx context.Context, weddingID primitive.ObjectID) (*models.WeddingAnalytics, error) {
	args := m.Called(ctx, weddingID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.WeddingAnalytics), args.Error(1)
}

func (m *MockAnalyticsRepository) UpdateWeddingAnalytics(ctx context.Context, weddingID primitive.ObjectID) error {
	args := m.Called(ctx, weddingID)
	return args.Error(0)
}

func (m *MockAnalyticsRepository) RefreshWeddingAnalytics(ctx context.Context, weddingID primitive.ObjectID) error {
	args := m.Called(ctx, weddingID)
	return args.Error(0)
}

func (m *MockAnalyticsRepository) GetSystemAnalytics(ctx context.Context) (*models.SystemAnalytics, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SystemAnalytics), args.Error(1)
}

func (m *MockAnalyticsRepository) UpdateSystemAnalytics(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockAnalyticsRepository) RefreshSystemAnalytics(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockAnalyticsRepository) GetAnalyticsSummary(ctx context.Context, weddingID primitive.ObjectID, period string) (*models.AnalyticsSummary, error) {
	args := m.Called(ctx, weddingID, period)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AnalyticsSummary), args.Error(1)
}

func (m *MockAnalyticsRepository) GetPopularPages(ctx context.Context, weddingID primitive.ObjectID, limit int) ([]models.PageStats, error) {
	args := m.Called(ctx, weddingID, limit)
	return args.Get(0).([]models.PageStats), args.Error(1)
}

func (m *MockAnalyticsRepository) GetTrafficSources(ctx context.Context, weddingID primitive.ObjectID, limit int) ([]models.TrafficSourceStats, error) {
	args := m.Called(ctx, weddingID, limit)
	return args.Get(0).([]models.TrafficSourceStats), args.Error(1)
}

func (m *MockAnalyticsRepository) GetDailyMetrics(ctx context.Context, weddingID primitive.ObjectID, startDate, endDate time.Time) ([]models.DailyMetrics, error) {
	args := m.Called(ctx, weddingID, startDate, endDate)
	return args.Get(0).([]models.DailyMetrics), args.Error(1)
}

func (m *MockAnalyticsRepository) CleanupOldAnalytics(ctx context.Context, olderThan time.Time) error {
	args := m.Called(ctx, olderThan)
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
