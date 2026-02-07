package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"wedding-invitation-backend/internal/domain/models"
	"wedding-invitation-backend/internal/domain/repository"
)

// MockRSVPRepository for testing
type MockRSVPRepository struct {
	rsvps       map[primitive.ObjectID]*models.RSVP
	createError error
	getError    error
}

func NewMockRSVPRepository() *MockRSVPRepository {
	return &MockRSVPRepository{
		rsvps: make(map[primitive.ObjectID]*models.RSVP),
	}
}

func (m *MockRSVPRepository) Create(ctx context.Context, rsvp *models.RSVP) error {
	if m.createError != nil {
		return m.createError
	}
	m.rsvps[rsvp.ID] = rsvp
	return nil
}

func (m *MockRSVPRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.RSVP, error) {
	if m.getError != nil {
		return nil, m.getError
	}
	if rsvp, exists := m.rsvps[id]; exists {
		return rsvp, nil
	}
	return nil, repository.ErrNotFound
}

func (m *MockRSVPRepository) GetByEmail(ctx context.Context, weddingID primitive.ObjectID, email string) (*models.RSVP, error) {
	for _, rsvp := range m.rsvps {
		if rsvp.WeddingID == weddingID && rsvp.Email == email {
			return rsvp, nil
		}
	}
	return nil, repository.ErrNotFound
}

func (m *MockRSVPRepository) ListByWedding(ctx context.Context, weddingID primitive.ObjectID, page, pageSize int, filters repository.RSVPFilters) ([]*models.RSVP, int64, error) {
	var results []*models.RSVP
	for _, rsvp := range m.rsvps {
		if rsvp.WeddingID == weddingID {
			results = append(results, rsvp)
		}
	}
	return results, int64(len(results)), nil
}

func (m *MockRSVPRepository) Update(ctx context.Context, rsvp *models.RSVP) error {
	m.rsvps[rsvp.ID] = rsvp
	return nil
}

func (m *MockRSVPRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	delete(m.rsvps, id)
	return nil
}

func (m *MockRSVPRepository) GetStatistics(ctx context.Context, weddingID primitive.ObjectID) (*models.RSVPStatistics, error) {
	stats := &models.RSVPStatistics{
		TotalResponses:  len(m.rsvps),
		DietaryCounts:   make(map[string]int),
		SubmissionTrend: []models.DailyCount{},
	}
	return stats, nil
}

func (m *MockRSVPRepository) MarkConfirmationSent(ctx context.Context, id primitive.ObjectID) error {
	if rsvp, exists := m.rsvps[id]; exists {
		now := time.Now()
		rsvp.ConfirmationSent = true
		rsvp.ConfirmationSentAt = &now
	}
	return nil
}

func (m *MockRSVPRepository) GetSubmissionTrend(ctx context.Context, weddingID primitive.ObjectID, days int) ([]models.DailyCount, error) {
	return []models.DailyCount{}, nil
}

func TestRSVPService_SubmitRSVP(t *testing.T) {
	// Setup
	rsvpRepo := NewMockRSVPRepository()
	weddingRepo := &MockWeddingRepository{}
	service := NewRSVPService(rsvpRepo, weddingRepo)

	weddingID := primitive.NewObjectID()
	userID := primitive.NewObjectID()

	// Create a published wedding
	wedding := &models.Wedding{
		ID:     weddingID,
		UserID: userID,
		Status: "published",
		RSVP: models.RSVPSettings{
			Enabled:     true,
			MaxPlusOnes: 2,
		},
	}
	weddingRepo.On("GetByID", mock.Anything, weddingID).Return(wedding, nil)
	weddingRepo.On("UpdateRSVPCount", mock.Anything, weddingID).Return(nil)

	// Create existing RSVP
	existingRSVP := &models.RSVP{
		ID:        primitive.NewObjectID(),
		WeddingID: weddingID,
		FirstName: "Existing",
		LastName:  "User",
		Email:     "duplicate@example.com",
		Status:    "attending",
	}
	rsvpRepo.rsvps[existingRSVP.ID] = existingRSVP

	// Try to submit with same email
	req := SubmitRSVPRequest{
		FirstName:       "John",
		LastName:        "Doe",
		Email:           "duplicate@example.com", // Same email
		Status:          "attending",
		AttendanceCount: 1,
	}

	_, err := service.SubmitRSVP(context.Background(), weddingID, req)
	assert.Error(t, err)
	assert.Equal(t, ErrDuplicateRSVP, err)
}

func TestRSVPService_SubmitRSVP_TooManyPlusOnes(t *testing.T) {
	rsvpRepo := NewMockRSVPRepository()
	weddingRepo := &MockWeddingRepository{}
	service := NewRSVPService(rsvpRepo, weddingRepo)

	weddingID := primitive.NewObjectID()
	userID := primitive.NewObjectID()

	// Create wedding with max plus ones = 1
	wedding := &models.Wedding{
		ID:     weddingID,
		UserID: userID,
		Status: "published",
		RSVP: models.RSVPSettings{
			Enabled:     true,
			MaxPlusOnes: 1,
		},
	}
	weddingRepo.On("GetByID", mock.Anything, weddingID).Return(wedding, nil)

	// Try to submit with 2 plus ones
	req := SubmitRSVPRequest{
		FirstName:       "John",
		LastName:        "Doe",
		Status:          "attending",
		AttendanceCount: 1,
		PlusOnes: []models.PlusOneInfo{
			{FirstName: "Jane", LastName: "Doe"},
			{FirstName: "Bob", LastName: "Smith"}, // Too many
		},
	}

	_, err := service.SubmitRSVP(context.Background(), weddingID, req)
	assert.Error(t, err)
	assert.Equal(t, ErrTooManyPlusOnes, err)
}

func TestRSVPService_UpdateRSVP(t *testing.T) {
	rsvpRepo := NewMockRSVPRepository()
	weddingRepo := &MockWeddingRepository{}
	service := NewRSVPService(rsvpRepo, weddingRepo)

	weddingID := primitive.NewObjectID()
	userID := primitive.NewObjectID()

	// Create wedding
	wedding := &models.Wedding{
		ID:     weddingID,
		UserID: userID,
		Status: "published",
		RSVP: models.RSVPSettings{
			Enabled:     true,
			MaxPlusOnes: 2,
		},
	}
	weddingRepo.On("GetByID", mock.Anything, weddingID).Return(wedding, nil)
	weddingRepo.On("UpdateRSVPCount", mock.Anything, weddingID).Return(nil)

	// Create existing RSVP
	rsvp := &models.RSVP{
		ID:              primitive.NewObjectID(),
		WeddingID:       weddingID,
		FirstName:       "John",
		LastName:        "Doe",
		Status:          "attending",
		AttendanceCount: 1,
		PlusOneCount:    0,
		SubmittedAt:     time.Now().Add(-1 * time.Hour), // 1 hour ago
	}
	rsvpRepo.rsvps[rsvp.ID] = rsvp

	// Update RSVP
	newStatus := "not-attending"
	req := UpdateRSVPRequest{
		Status:          &newStatus,
		AttendanceCount: intPtr(2),
	}

	updatedRSVP, err := service.UpdateRSVP(context.Background(), rsvp.ID, req)
	require.NoError(t, err)
	assert.Equal(t, "not-attending", updatedRSVP.Status)
	assert.Equal(t, 2, updatedRSVP.AttendanceCount)
	assert.NotNil(t, updatedRSVP.UpdatedAt)
}

func TestRSVPService_UpdateRSVP_NotFound(t *testing.T) {
	rsvpRepo := NewMockRSVPRepository()
	weddingRepo := &MockWeddingRepository{}
	service := NewRSVPService(rsvpRepo, weddingRepo)

	nonExistentID := primitive.NewObjectID()
	req := UpdateRSVPRequest{
		Status: stringPtr("not-attending"),
	}

	_, err := service.UpdateRSVP(context.Background(), nonExistentID, req)
	assert.Error(t, err)
	assert.Equal(t, ErrRSVPNotFound, err)
}

func TestRSVPService_UpdateRSVP_CannotModify(t *testing.T) {
	rsvpRepo := NewMockRSVPRepository()
	weddingRepo := &MockWeddingRepository{}
	service := NewRSVPService(rsvpRepo, weddingRepo)

	weddingID := primitive.NewObjectID()
	userID := primitive.NewObjectID()

	// Create wedding
	wedding := &models.Wedding{
		ID:     weddingID,
		UserID: userID,
		Status: "published",
	}
	weddingRepo.On("GetByID", mock.Anything, weddingID).Return(wedding, nil)

	// Create old RSVP (more than 24 hours ago)
	rsvp := &models.RSVP{
		ID:              primitive.NewObjectID(),
		WeddingID:       weddingID,
		FirstName:       "John",
		LastName:        "Doe",
		Status:          "attending",
		AttendanceCount: 1,
		SubmittedAt:     time.Now().Add(-25 * time.Hour), // 25 hours ago
	}
	rsvpRepo.rsvps[rsvp.ID] = rsvp

	req := UpdateRSVPRequest{
		Status: stringPtr("not-attending"),
	}

	_, err := service.UpdateRSVP(context.Background(), rsvp.ID, req)
	assert.Error(t, err)
	assert.Equal(t, ErrRSVPCannotModify, err)
}

func TestRSVPService_DeleteRSVP(t *testing.T) {
	rsvpRepo := NewMockRSVPRepository()
	weddingRepo := &MockWeddingRepository{}
	service := NewRSVPService(rsvpRepo, weddingRepo)

	weddingID := primitive.NewObjectID()
	userID := primitive.NewObjectID()

	// Create wedding
	wedding := &models.Wedding{
		ID:     weddingID,
		UserID: userID,
		Status: "published",
	}
	weddingRepo.On("GetByID", mock.Anything, weddingID).Return(wedding, nil)

	// Create RSVP
	rsvp := &models.RSVP{
		ID:        primitive.NewObjectID(),
		WeddingID: weddingID,
		FirstName: "John",
		LastName:  "Doe",
		Status:    "attending",
	}
	rsvpRepo.rsvps[rsvp.ID] = rsvp

	// Delete RSVP
	err := service.DeleteRSVP(context.Background(), rsvp.ID, userID)
	require.NoError(t, err)

	// Verify deleted
	_, err = service.GetRSVPByID(context.Background(), rsvp.ID)
	assert.Error(t, err)
	assert.Equal(t, ErrRSVPNotFound, err)
}

func TestRSVPService_DeleteRSVP_Unauthorized(t *testing.T) {
	rsvpRepo := NewMockRSVPRepository()
	weddingRepo := &MockWeddingRepository{}
	service := NewRSVPService(rsvpRepo, weddingRepo)

	weddingID := primitive.NewObjectID()
	userID := primitive.NewObjectID()
	otherUserID := primitive.NewObjectID()

	// Create wedding owned by different user
	wedding := &models.Wedding{
		ID:     weddingID,
		UserID: otherUserID, // Different owner
		Status: "published",
	}
	weddingRepo.On("GetByID", mock.Anything, weddingID).Return(wedding, nil)

	// Create RSVP
	rsvp := &models.RSVP{
		ID:        primitive.NewObjectID(),
		WeddingID: weddingID,
		FirstName: "John",
		LastName:  "Doe",
		Status:    "attending",
	}
	rsvpRepo.rsvps[rsvp.ID] = rsvp

	// Try to delete with wrong user
	err := service.DeleteRSVP(context.Background(), rsvp.ID, userID)
	assert.Error(t, err)
	assert.Equal(t, ErrUnauthorized, err)
}

func TestRSVPService_ListRSVPs(t *testing.T) {
	rsvpRepo := NewMockRSVPRepository()
	weddingRepo := &MockWeddingRepository{}
	service := NewRSVPService(rsvpRepo, weddingRepo)

	weddingID := primitive.NewObjectID()
	userID := primitive.NewObjectID()

	// Create wedding
	wedding := &models.Wedding{
		ID:     weddingID,
		UserID: userID,
		Status: "published",
	}
	weddingRepo.On("GetByID", mock.Anything, weddingID).Return(wedding, nil)

	// Create RSVPs
	for i := 0; i < 3; i++ {
		rsvp := &models.RSVP{
			ID:        primitive.NewObjectID(),
			WeddingID: weddingID,
			FirstName: fmt.Sprintf("User%d", i),
			LastName:  "Test",
			Status:    "attending",
		}
		rsvpRepo.rsvps[rsvp.ID] = rsvp
	}

	rsvps, total, err := service.ListRSVPs(context.Background(), weddingID, userID, 1, 10, repository.RSVPFilters{})
	require.NoError(t, err)
	assert.Equal(t, 3, len(rsvps))
	assert.Equal(t, int64(3), total)
}

func TestRSVPService_GetRSVPStatistics(t *testing.T) {
	rsvpRepo := NewMockRSVPRepository()
	weddingRepo := &MockWeddingRepository{}
	service := NewRSVPService(rsvpRepo, weddingRepo)

	weddingID := primitive.NewObjectID()
	userID := primitive.NewObjectID()

	// Create wedding
	wedding := &models.Wedding{
		ID:     weddingID,
		UserID: userID,
		Status: "published",
	}
	weddingRepo.On("GetByID", mock.Anything, weddingID).Return(wedding, nil)

	stats, err := service.GetRSVPStatistics(context.Background(), weddingID, userID)
	require.NoError(t, err)
	assert.NotNil(t, stats)
}

func TestRSVPService_ExportRSVPs(t *testing.T) {
	rsvpRepo := NewMockRSVPRepository()
	weddingRepo := &MockWeddingRepository{}
	service := NewRSVPService(rsvpRepo, weddingRepo)

	weddingID := primitive.NewObjectID()
	userID := primitive.NewObjectID()

	// Create wedding
	wedding := &models.Wedding{
		ID:     weddingID,
		UserID: userID,
		Status: "published",
	}
	weddingRepo.On("GetByID", mock.Anything, weddingID).Return(wedding, nil)

	// Create RSVP
	rsvp := &models.RSVP{
		ID:        primitive.NewObjectID(),
		WeddingID: weddingID,
		FirstName: "John",
		LastName:  "Doe",
		Status:    "attending",
	}
	rsvpRepo.rsvps[rsvp.ID] = rsvp

	rsvps, err := service.ExportRSVPs(context.Background(), weddingID, userID)
	require.NoError(t, err)
	assert.Equal(t, 1, len(rsvps))
	assert.Equal(t, "John", rsvps[0].FirstName)
}
