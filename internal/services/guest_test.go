package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"wedding-invitation-backend/internal/domain/models"
	"wedding-invitation-backend/internal/domain/repository"
)

// MockGuestRepository for testing
type MockGuestRepository struct {
	guests      map[primitive.ObjectID]*models.Guest
	batchGuests map[string][]*models.Guest
	createError error
	getError    error
	updateError error
	deleteError error
}

func NewMockGuestRepository() *MockGuestRepository {
	return &MockGuestRepository{
		guests:      make(map[primitive.ObjectID]*models.Guest),
		batchGuests: make(map[string][]*models.Guest),
	}
}

func (m *MockGuestRepository) Create(ctx context.Context, guest *models.Guest) error {
	if m.createError != nil {
		return m.createError
	}

	id := primitive.NewObjectID()
	guest.ID = id
	guest.CreatedAt = time.Now()
	guest.UpdatedAt = time.Now()
	m.guests[id] = guest
	return nil
}

func (m *MockGuestRepository) CreateMany(ctx context.Context, guests []*models.Guest) error {
	for _, guest := range guests {
		id := primitive.NewObjectID()
		guest.ID = id
		guest.CreatedAt = time.Now()
		guest.UpdatedAt = time.Now()
		m.guests[id] = guest
	}
	return nil
}

func (m *MockGuestRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.Guest, error) {
	if m.getError != nil {
		return nil, m.getError
	}

	guest, exists := m.guests[id]
	if !exists {
		return nil, repository.ErrNotFound
	}
	return guest, nil
}

func (m *MockGuestRepository) GetByEmail(ctx context.Context, weddingID primitive.ObjectID, email string) (*models.Guest, error) {
	for _, guest := range m.guests {
		if guest.WeddingID == weddingID && guest.Email == email {
			return guest, nil
		}
	}
	return nil, repository.ErrNotFound
}

func (m *MockGuestRepository) ListByWedding(ctx context.Context, weddingID primitive.ObjectID, page, pageSize int, filters repository.GuestFilters) ([]*models.Guest, int64, error) {
	var guests []*models.Guest

	for _, guest := range m.guests {
		if guest.WeddingID == weddingID {
			// Apply filters
			if filters.Search != "" {
				search := filters.Search
				if guest.FirstName == search || guest.LastName == search || guest.Email == search {
					guests = append(guests, guest)
				}
			} else {
				guests = append(guests, guest)
			}
		}
	}

	return guests, int64(len(guests)), nil
}

func (m *MockGuestRepository) Update(ctx context.Context, guest *models.Guest) error {
	if m.updateError != nil {
		return m.updateError
	}

	if _, exists := m.guests[guest.ID]; !exists {
		return repository.ErrNotFound
	}

	guest.UpdatedAt = time.Now()
	m.guests[guest.ID] = guest
	return nil
}

func (m *MockGuestRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	if m.deleteError != nil {
		return m.deleteError
	}

	if _, exists := m.guests[id]; !exists {
		return repository.ErrNotFound
	}

	delete(m.guests, id)
	return nil
}

func (m *MockGuestRepository) ImportBatch(ctx context.Context, guests []*models.Guest, batchID string) error {
	for _, guest := range guests {
		guest.ImportBatchID = batchID
		id := primitive.NewObjectID()
		guest.ID = id
		guest.CreatedAt = time.Now()
		guest.UpdatedAt = time.Now()
		m.guests[id] = guest
	}

	m.batchGuests[batchID] = guests
	return nil
}

func (m *MockGuestRepository) GetByImportBatch(ctx context.Context, weddingID primitive.ObjectID, batchID string) ([]*models.Guest, error) {
	guests, exists := m.batchGuests[batchID]
	if !exists {
		return []*models.Guest{}, nil
	}

	var result []*models.Guest
	for _, guest := range guests {
		if guest.WeddingID == weddingID {
			result = append(result, guest)
		}
	}

	return result, nil
}

func TestGuestService_CreateGuest(t *testing.T) {
	guestRepo := NewMockGuestRepository()
	weddingRepo := &MockWeddingRepository{}
	service := NewGuestService(guestRepo, weddingRepo)

	// Create test wedding
	weddingID := primitive.NewObjectID()
	userID := primitive.NewObjectID()
	wedding := &models.Wedding{
		ID:     weddingID,
		UserID: userID,
	}

	// Setup mock expectations
	weddingRepo.On("GetByID", mock.Anything, weddingID).Return(wedding, nil)

	// Test data
	guest := &models.Guest{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
		Side:      "groom",
	}

	// Test successful creation
	err := service.CreateGuest(context.Background(), weddingID, userID, guest)
	assert.NoError(t, err)
	assert.NotEmpty(t, guest.ID)

	weddingRepo.AssertExpectations(t)
}

func TestGuestService_CreateGuest_Unauthorized(t *testing.T) {
	guestRepo := NewMockGuestRepository()
	weddingRepo := &MockWeddingRepository{}
	service := NewGuestService(guestRepo, weddingRepo)

	// Create test wedding
	weddingID := primitive.NewObjectID()
	userID := primitive.NewObjectID()
	otherUserID := primitive.NewObjectID()
	wedding := &models.Wedding{
		ID:     weddingID,
		UserID: otherUserID, // Different user
	}
	weddingRepo.weddings[weddingID] = wedding

	// Test data
	guest := &models.Guest{
		FirstName: "John",
		LastName:  "Doe",
	}

	// Test unauthorized creation
	err := service.CreateGuest(context.Background(), weddingID, userID, guest)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unauthorized")
}

func TestGuestService_CreateGuest_ValidationError(t *testing.T) {
	guestRepo := NewMockGuestRepository()
	weddingRepo := &MockWeddingRepository{}
	service := NewGuestService(guestRepo, weddingRepo)

	// Create test wedding
	weddingID := primitive.NewObjectID()
	userID := primitive.NewObjectID()
	wedding := &models.Wedding{
		ID:     weddingID,
		UserID: userID,
	}
	weddingRepo.weddings[weddingID] = wedding

	// Test data with missing required fields
	guest := &models.Guest{
		FirstName: "", // Missing first name
		LastName:  "Doe",
	}

	// Test validation error
	err := service.CreateGuest(context.Background(), weddingID, userID, guest)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "first name is required")
}

func TestGuestService_CreateGuest_DuplicateEmail(t *testing.T) {
	guestRepo := NewMockGuestRepository()
	weddingRepo := &MockWeddingRepository{}
	service := NewGuestService(guestRepo, weddingRepo)

	// Create test wedding
	weddingID := primitive.NewObjectID()
	userID := primitive.NewObjectID()
	wedding := &models.Wedding{
		ID:     weddingID,
		UserID: userID,
	}
	weddingRepo.weddings[weddingID] = wedding

	// Create first guest
	guest1 := &models.Guest{
		WeddingID: weddingID,
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
		CreatedBy: userID,
	}
	guestRepo.Create(context.Background(), guest1)

	// Test creating second guest with same email
	guest2 := &models.Guest{
		FirstName: "Jane",
		LastName:  "Doe",
		Email:     "john@example.com", // Same email
	}

	err := service.CreateGuest(context.Background(), weddingID, userID, guest2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestGuestService_GetGuestByID(t *testing.T) {
	guestRepo := NewMockGuestRepository()
	weddingRepo := &MockWeddingRepository{}
	service := NewGuestService(guestRepo, weddingRepo)

	// Create test wedding
	weddingID := primitive.NewObjectID()
	userID := primitive.NewObjectID()
	wedding := &models.Wedding{
		ID:     weddingID,
		UserID: userID,
	}
	weddingRepo.weddings[weddingID] = wedding

	// Create test guest
	guest := &models.Guest{
		WeddingID: weddingID,
		FirstName: "John",
		LastName:  "Doe",
		CreatedBy: userID,
	}
	guestRepo.Create(context.Background(), guest)

	// Test successful retrieval
	found, err := service.GetGuestByID(context.Background(), guest.ID, userID)
	assert.NoError(t, err)
	assert.Equal(t, guest.FirstName, found.FirstName)
}

func TestGuestService_UpdateGuest(t *testing.T) {
	guestRepo := NewMockGuestRepository()
	weddingRepo := &MockWeddingRepository{}
	service := NewGuestService(guestRepo, weddingRepo)

	// Create test wedding
	weddingID := primitive.NewObjectID()
	userID := primitive.NewObjectID()
	wedding := &models.Wedding{
		ID:     weddingID,
		UserID: userID,
	}
	weddingRepo.weddings[weddingID] = wedding

	// Create test guest
	guest := &models.Guest{
		WeddingID: weddingID,
		FirstName: "John",
		LastName:  "Doe",
		CreatedBy: userID,
	}
	guestRepo.Create(context.Background(), guest)

	// Update guest
	guest.FirstName = "Updated"
	err := service.UpdateGuest(context.Background(), guest.ID, userID, guest)
	assert.NoError(t, err)
}

func TestGuestService_DeleteGuest(t *testing.T) {
	guestRepo := NewMockGuestRepository()
	weddingRepo := &MockWeddingRepository{}
	service := NewGuestService(guestRepo, weddingRepo)

	// Create test wedding
	weddingID := primitive.NewObjectID()
	userID := primitive.NewObjectID()
	wedding := &models.Wedding{
		ID:     weddingID,
		UserID: userID,
	}
	weddingRepo.weddings[weddingID] = wedding

	// Create test guest
	guest := &models.Guest{
		WeddingID: weddingID,
		FirstName: "John",
		LastName:  "Doe",
		CreatedBy: userID,
	}
	guestRepo.Create(context.Background(), guest)

	// Delete guest
	err := service.DeleteGuest(context.Background(), guest.ID, userID)
	assert.NoError(t, err)

	// Verify deletion
	_, err = service.GetGuestByID(context.Background(), guest.ID, userID)
	assert.Error(t, err)
}

func TestGuestService_CreateManyGuests(t *testing.T) {
	guestRepo := NewMockGuestRepository()
	weddingRepo := &MockWeddingRepository{}
	service := NewGuestService(guestRepo, weddingRepo)

	// Create test wedding
	weddingID := primitive.NewObjectID()
	userID := primitive.NewObjectID()
	wedding := &models.Wedding{
		ID:     weddingID,
		UserID: userID,
	}
	weddingRepo.weddings[weddingID] = wedding

	// Create test guests
	guests := []*models.Guest{
		{
			FirstName: "Guest1",
			LastName:  "Test1",
		},
		{
			FirstName: "Guest2",
			LastName:  "Test2",
		},
	}

	// Test successful creation
	err := service.CreateManyGuests(context.Background(), weddingID, userID, guests)
	assert.NoError(t, err)

	// Verify all guests were created
	for _, guest := range guests {
		assert.NotEmpty(t, guest.ID)
		assert.Equal(t, weddingID, guest.WeddingID)
		assert.Equal(t, userID, guest.CreatedBy)
	}
}
