package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"wedding-invitation-backend/internal/domain/models"
	"wedding-invitation-backend/internal/domain/repository"
)

func createTestWedding() *models.Wedding {
	return &models.Wedding{
		Title: "Test Wedding",
		Slug:  "test-wedding",
		Couple: models.CoupleInfo{
			Partner1: struct {
				FirstName   string            `bson:"first_name" json:"first_name" validate:"required"`
				LastName    string            `bson:"last_name" json:"last_name" validate:"required"`
				FullName    string            `bson:"full_name" json:"full_name"`
				PhotoURL    string            `bson:"photo_url,omitempty" json:"photo_url,omitempty"`
				SocialLinks map[string]string `bson:"social_links,omitempty" json:"social_links,omitempty"`
			}{
				FirstName: "John",
				LastName:  "Doe",
			},
			Partner2: struct {
				FirstName   string            `bson:"first_name" json:"first_name" validate:"required"`
				LastName    string            `bson:"last_name" json:"last_name" validate:"required"`
				FullName    string            `bson:"full_name" json:"full_name"`
				PhotoURL    string            `bson:"photo_url,omitempty" json:"photo_url,omitempty"`
				SocialLinks map[string]string `bson:"social_links,omitempty" json:"social_links,omitempty"`
			}{
				FirstName: "Jane",
				LastName:  "Smith",
			},
		},
		Event: models.EventDetails{
			Title:        "Wedding Ceremony",
			Date:         time.Now().AddDate(0, 6, 0),
			VenueName:    "Test Venue",
			VenueAddress: "123 Test St",
		},
		Theme: models.ThemeSettings{
			ThemeID: "default",
		},
		RSVP: models.RSVPSettings{
			Enabled:      true,
			AllowPlusOne: true,
			MaxPlusOnes:  2,
		},
	}
}

func TestWeddingService_CreateWedding(t *testing.T) {
	ctx := context.Background()
	mockWeddingRepo := new(MockWeddingRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewWeddingService(mockWeddingRepo, mockUserRepo)

	userID := primitive.NewObjectID()
	wedding := createTestWedding()

	// Test successful creation
	mockWeddingRepo.On("ExistsBySlug", ctx, wedding.Slug).Return(false, nil)
	mockWeddingRepo.On("Create", ctx, mock.AnythingOfType("*models.Wedding")).Return(nil)
	mockUserRepo.On("AddWeddingID", ctx, userID, mock.AnythingOfType("primitive.ObjectID")).Return(nil)

	err := service.CreateWedding(ctx, wedding, userID)
	assert.NoError(t, err)
	assert.Equal(t, userID, wedding.UserID)
	assert.Equal(t, string(models.WeddingStatusDraft), wedding.Status)
	assert.Equal(t, 0, wedding.RSVPCount)
	assert.Equal(t, 0, wedding.GuestCount)
	assert.Equal(t, 0, wedding.TotalAttending)
	assert.Equal(t, 0, wedding.ViewCount)
	assert.False(t, wedding.GalleryEnabled)
	assert.False(t, wedding.IsPublic)

	mockWeddingRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestWeddingService_CreateWedding_WithAutoSlug(t *testing.T) {
	ctx := context.Background()
	mockWeddingRepo := new(MockWeddingRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewWeddingService(mockWeddingRepo, mockUserRepo)

	userID := primitive.NewObjectID()
	wedding := createTestWedding()
	wedding.Slug = "" // Empty slug to test auto-generation

	// Test slug generation
	mockWeddingRepo.On("ExistsBySlug", ctx, "test-wedding").Return(true, nil)
	mockWeddingRepo.On("ExistsBySlug", ctx, "test-wedding-1").Return(false, nil)
	mockWeddingRepo.On("Create", ctx, mock.AnythingOfType("*models.Wedding")).Return(nil)
	mockUserRepo.On("AddWeddingID", ctx, userID, mock.AnythingOfType("primitive.ObjectID")).Return(nil)

	err := service.CreateWedding(ctx, wedding, userID)
	assert.NoError(t, err)
	assert.Equal(t, "test-wedding-1", wedding.Slug)

	mockWeddingRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestWeddingService_CreateWedding_SlugExists(t *testing.T) {
	ctx := context.Background()
	mockWeddingRepo := new(MockWeddingRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewWeddingService(mockWeddingRepo, mockUserRepo)

	userID := primitive.NewObjectID()
	wedding := createTestWedding()

	// Test slug already exists
	mockWeddingRepo.On("ExistsBySlug", ctx, wedding.Slug).Return(true, nil)

	err := service.CreateWedding(ctx, wedding, userID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "slug already exists")

	mockWeddingRepo.AssertExpectations(t)
}

func TestWeddingService_CreateWedding_InvalidData(t *testing.T) {
	ctx := context.Background()
	mockWeddingRepo := new(MockWeddingRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewWeddingService(mockWeddingRepo, mockUserRepo)

	userID := primitive.NewObjectID()

	// Test missing title
	wedding := createTestWedding()
	wedding.Title = ""

	err := service.CreateWedding(ctx, wedding, userID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "title is required")

	// Test missing partner info
	wedding = createTestWedding()
	wedding.Couple.Partner1.FirstName = ""

	err = service.CreateWedding(ctx, wedding, userID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "partner1 first name and last name are required")
}

func TestWeddingService_GetWeddingByID(t *testing.T) {
	ctx := context.Background()
	mockWeddingRepo := new(MockWeddingRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewWeddingService(mockWeddingRepo, mockUserRepo)

	userID := primitive.NewObjectID()
	weddingID := primitive.NewObjectID()
	wedding := createTestWedding()
	wedding.ID = weddingID
	wedding.UserID = userID
	wedding.IsPublic = true
	wedding.Status = string(models.WeddingStatusPublished)

	// Test successful retrieval (owner)
	mockWeddingRepo.On("GetByID", ctx, weddingID).Return(wedding, nil)

	result, err := service.GetWeddingByID(ctx, weddingID, userID)
	assert.NoError(t, err)
	assert.Equal(t, wedding, result)

	mockWeddingRepo.AssertExpectations(t)

	// Reset mock for next test
	mockWeddingRepo.ExpectedCalls = nil

	// Test successful retrieval (public access)
	otherUserID := primitive.NewObjectID()
	mockWeddingRepo.On("GetByID", ctx, weddingID).Return(wedding, nil)
	mockWeddingRepo.On("IncrementViewCount", ctx, weddingID).Return(nil)

	result, err = service.GetWeddingByID(ctx, weddingID, otherUserID)
	assert.NoError(t, err)
	assert.Equal(t, wedding, result)

	mockWeddingRepo.AssertExpectations(t)
}

func TestWeddingService_GetWeddingByID_NotFound(t *testing.T) {
	ctx := context.Background()
	mockWeddingRepo := new(MockWeddingRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewWeddingService(mockWeddingRepo, mockUserRepo)

	weddingID := primitive.NewObjectID()

	// Test not found
	mockWeddingRepo.On("GetByID", ctx, weddingID).Return(nil, nil)

	result, err := service.GetWeddingByID(ctx, weddingID, primitive.NewObjectID())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "wedding not found")
	assert.Nil(t, result)

	mockWeddingRepo.AssertExpectations(t)
}

func TestWeddingService_GetWeddingByID_AccessDenied(t *testing.T) {
	ctx := context.Background()
	mockWeddingRepo := new(MockWeddingRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewWeddingService(mockWeddingRepo, mockUserRepo)

	userID := primitive.NewObjectID()
	weddingID := primitive.NewObjectID()
	wedding := createTestWedding()
	wedding.ID = weddingID
	wedding.UserID = primitive.NewObjectID()
	wedding.IsPublic = false
	wedding.Status = string(models.WeddingStatusDraft)

	// Test access denied
	mockWeddingRepo.On("GetByID", ctx, weddingID).Return(wedding, nil)

	result, err := service.GetWeddingByID(ctx, weddingID, userID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "access denied")
	assert.Nil(t, result)

	mockWeddingRepo.AssertExpectations(t)
}

func TestWeddingService_GetUserWeddings(t *testing.T) {
	ctx := context.Background()
	mockWeddingRepo := new(MockWeddingRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewWeddingService(mockWeddingRepo, mockUserRepo)

	userID := primitive.NewObjectID()
	weddings := []*models.Wedding{createTestWedding()}
	filters := repository.WeddingFilters{}

	mockWeddingRepo.On("GetByUserID", ctx, userID, 1, 20, filters).Return(weddings, int64(1), nil)

	result, total, err := service.GetUserWeddings(ctx, userID, 1, 20, filters)
	assert.NoError(t, err)
	assert.Equal(t, weddings, result)
	assert.Equal(t, int64(1), total)

	mockWeddingRepo.AssertExpectations(t)
}

func TestWeddingService_UpdateWedding(t *testing.T) {
	ctx := context.Background()
	mockWeddingRepo := new(MockWeddingRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewWeddingService(mockWeddingRepo, mockUserRepo)

	userID := primitive.NewObjectID()
	weddingID := primitive.NewObjectID()
	existingWedding := createTestWedding()
	existingWedding.ID = weddingID
	existingWedding.UserID = userID
	updatedWedding := createTestWedding()
	updatedWedding.ID = weddingID
	updatedWedding.Title = "Updated Wedding"

	// Test successful update
	mockWeddingRepo.On("GetByID", ctx, weddingID).Return(existingWedding, nil)
	mockWeddingRepo.On("ExistsBySlug", ctx, updatedWedding.Slug).Return(false, nil)
	mockWeddingRepo.On("Update", ctx, updatedWedding).Return(nil)

	err := service.UpdateWedding(ctx, updatedWedding, userID)
	assert.NoError(t, err)

	mockWeddingRepo.AssertExpectations(t)
}

func TestWeddingService_UpdateWedding_NotOwner(t *testing.T) {
	ctx := context.Background()
	mockWeddingRepo := new(MockWeddingRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewWeddingService(mockWeddingRepo, mockUserRepo)

	userID := primitive.NewObjectID()
	weddingID := primitive.NewObjectID()
	existingWedding := createTestWedding()
	existingWedding.ID = weddingID
	existingWedding.UserID = primitive.NewObjectID() // Different user
	updatedWedding := createTestWedding()
	updatedWedding.ID = weddingID

	// Test access denied
	mockWeddingRepo.On("GetByID", ctx, weddingID).Return(existingWedding, nil)

	err := service.UpdateWedding(ctx, updatedWedding, userID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "access denied")

	mockWeddingRepo.AssertExpectations(t)
}

func TestWeddingService_DeleteWedding(t *testing.T) {
	ctx := context.Background()
	mockWeddingRepo := new(MockWeddingRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewWeddingService(mockWeddingRepo, mockUserRepo)

	userID := primitive.NewObjectID()
	weddingID := primitive.NewObjectID()
	wedding := createTestWedding()
	wedding.ID = weddingID
	wedding.UserID = userID

	// Test successful deletion
	mockWeddingRepo.On("GetByID", ctx, weddingID).Return(wedding, nil)
	mockWeddingRepo.On("Delete", ctx, weddingID).Return(nil)
	mockUserRepo.On("RemoveWeddingID", ctx, userID, weddingID).Return(nil)

	err := service.DeleteWedding(ctx, weddingID, userID)
	assert.NoError(t, err)

	mockWeddingRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestWeddingService_PublishWedding(t *testing.T) {
	ctx := context.Background()
	mockWeddingRepo := new(MockWeddingRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewWeddingService(mockWeddingRepo, mockUserRepo)

	userID := primitive.NewObjectID()
	weddingID := primitive.NewObjectID()
	wedding := createTestWedding()
	wedding.ID = weddingID
	wedding.UserID = userID

	// Test successful publishing
	mockWeddingRepo.On("GetByID", ctx, weddingID).Return(wedding, nil).Twice()
	mockWeddingRepo.On("Update", ctx, mock.AnythingOfType("*models.Wedding")).Return(nil)

	err := service.PublishWedding(ctx, weddingID, userID)
	assert.NoError(t, err)

	mockWeddingRepo.AssertExpectations(t)
}

func TestWeddingService_ListPublicWeddings(t *testing.T) {
	ctx := context.Background()
	mockWeddingRepo := new(MockWeddingRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewWeddingService(mockWeddingRepo, mockUserRepo)

	weddings := []*models.Wedding{createTestWedding()}
	filters := repository.PublicWeddingFilters{}

	mockWeddingRepo.On("ListPublic", ctx, 1, 20, filters).Return(weddings, int64(1), nil)

	result, total, err := service.ListPublicWeddings(ctx, 1, 20, filters)
	assert.NoError(t, err)
	assert.Equal(t, weddings, result)
	assert.Equal(t, int64(1), total)

	mockWeddingRepo.AssertExpectations(t)
}

func TestWeddingService_ValidateWedding_InvalidTheme(t *testing.T) {
	ctx := context.Background()
	mockWeddingRepo := new(MockWeddingRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewWeddingService(mockWeddingRepo, mockUserRepo)

	userID := primitive.NewObjectID()
	wedding := createTestWedding()
	wedding.Theme.ThemeID = "" // Invalid theme

	err := service.CreateWedding(ctx, wedding, userID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "theme ID is required")
}

func TestWeddingService_ValidateWedding_InvalidRSVP(t *testing.T) {
	ctx := context.Background()
	mockWeddingRepo := new(MockWeddingRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewWeddingService(mockWeddingRepo, mockUserRepo)

	userID := primitive.NewObjectID()
	wedding := createTestWedding()
	wedding.RSVP.MaxPlusOnes = 10 // Invalid max plus ones

	err := service.CreateWedding(ctx, wedding, userID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "max plus ones must be between 0 and 5")
}
