package mongodb

import (
	"context"
	"testing"
	"time"

	"wedding-invitation-backend/internal/config"
	"wedding-invitation-backend/internal/domain/models"
	"wedding-invitation-backend/internal/domain/repository"
	"wedding-invitation-backend/pkg/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// WeddingRepositoryTestSuite contains all tests for the wedding repository
type WeddingRepositoryTestSuite struct {
	suite.Suite
	db      *database.MongoDB
	repo    repository.WeddingRepository
	ctx     context.Context
	cleanup func()
}

// SetupSuite runs once before all tests
func (suite *WeddingRepositoryTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	// Setup test database
	testDBConfig := &config.DatabaseConfig{
		URI:      "mongodb://localhost:27017",
		Database: "wedding_test_" + primitive.NewObjectID().Hex(),
		Timeout:  10,
	}

	db, err := database.NewMongoDB(testDBConfig)
	if err != nil {
		suite.T().Skipf("Skipping integration tests: MongoDB not available: %v", err)
		return
	}

	suite.db = db
	suite.repo = NewMongoWeddingRepository(db.Database)

	// Setup cleanup function
	suite.cleanup = func() {
		if db != nil {
			db.Close(suite.ctx)
			// Drop test database
			client, err := mongo.Connect(suite.ctx, options.Client().ApplyURI(testDBConfig.URI))
			if err == nil {
				client.Database(testDBConfig.Database).Drop(suite.ctx)
				client.Disconnect(suite.ctx)
			}
		}
	}
}

// TearDownSuite runs once after all tests
func (suite *WeddingRepositoryTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

// SetupTest runs before each test
func (suite *WeddingRepositoryTestSuite) SetupTest() {
	if suite.db == nil {
		suite.T().Skip("MongoDB not available")
		return
	}

	// Clear the weddings collection before each test
	err := suite.db.Collection("weddings").Drop(suite.ctx)
	require.NoError(suite.T(), err)
}

// createTestWedding creates a test wedding for use in tests
func (suite *WeddingRepositoryTestSuite) createTestWedding() *models.Wedding {
	userID := primitive.NewObjectID()
	now := time.Now()

	return &models.Wedding{
		ID:       primitive.NewObjectID(), // Set ID explicitly
		UserID:   userID,
		Slug:     "test-wedding-" + primitive.NewObjectID().Hex(),
		Title:    "Test Wedding",
		IsPublic: true,
		Event: models.EventDetails{
			Title:        "Wedding Ceremony",
			Date:         now.AddDate(0, 6, 0),
			Time:         "16:00",
			VenueName:    "Test Venue",
			VenueAddress: "123 Test Street",
			DressCode:    "Formal",
		},
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
				FullName:  "John Doe",
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
				FullName:  "Jane Smith",
			},
			Story: "We met in college",
		},
		Theme: models.ThemeSettings{
			ThemeID:        "default",
			PrimaryColor:   "#FF6B6B",
			SecondaryColor: "#4ECDC4",
			FontFamily:     "Playfair Display",
			CustomCSS:      "body { margin: 0; }",
		},
		RSVP: models.RSVPSettings{
			Enabled:        true,
			AllowPlusOne:   true,
			CollectEmail:   true,
			CollectPhone:   false,
			CollectDietary: true,
			DietaryOptions: []string{"Vegetarian", "Vegan", "Gluten-Free"},
			CustomQuestions: []models.CustomQuestion{
				{
					ID:       primitive.NewObjectID().Hex(),
					Type:     "text",
					Question: "Dietary restrictions",
					Required: false,
					Order:    1,
				},
			},
		},
		GalleryImages: []models.GalleryImage{
			{
				ID:           primitive.NewObjectID().Hex(),
				URL:          "https://example.com/image1.jpg",
				ThumbnailURL: "https://example.com/thumb1.jpg",
				Caption:      "Our engagement",
				Order:        1,
				UploadedAt:   now,
				FileSize:     1024000,
			},
		},
		Status:         string(models.WeddingStatusDraft),
		CreatedAt:      now,
		UpdatedAt:      now,
		RSVPCount:      0,
		GuestCount:     0,
		TotalAttending: 0,
		ViewCount:      0,
	}
}

// TestCreate tests creating a new wedding
func (suite *WeddingRepositoryTestSuite) TestCreate() {
	if suite.db == nil {
		suite.T().Skip("MongoDB not available")
	}

	wedding := suite.createTestWedding()

	err := suite.repo.Create(suite.ctx, wedding)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), wedding.ID)

	// Verify the wedding was actually created
	found, err := suite.repo.GetByID(suite.ctx, wedding.ID)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), found)
	assert.Equal(suite.T(), wedding.Title, found.Title)
	assert.Equal(suite.T(), wedding.Slug, found.Slug)
}

// TestGetByID tests retrieving a wedding by ID
func (suite *WeddingRepositoryTestSuite) TestGetByID() {
	if suite.db == nil {
		suite.T().Skip("MongoDB not available")
	}

	wedding := suite.createTestWedding()

	// Create wedding
	err := suite.repo.Create(suite.ctx, wedding)
	require.NoError(suite.T(), err)

	// Get wedding by ID
	found, err := suite.repo.GetByID(suite.ctx, wedding.ID)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), found)
	assert.Equal(suite.T(), wedding.ID, found.ID)
	assert.Equal(suite.T(), wedding.Title, found.Title)
	assert.Equal(suite.T(), wedding.UserID, found.UserID)
}

// TestGetBySlug tests retrieving a wedding by slug
func (suite *WeddingRepositoryTestSuite) TestGetBySlug() {
	if suite.db == nil {
		suite.T().Skip("MongoDB not available")
	}

	wedding := suite.createTestWedding()

	// Create wedding
	err := suite.repo.Create(suite.ctx, wedding)
	require.NoError(suite.T(), err)

	// Get wedding by slug
	found, err := suite.repo.GetBySlug(suite.ctx, wedding.Slug)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), found)
	assert.Equal(suite.T(), wedding.Slug, found.Slug)
	assert.Equal(suite.T(), wedding.Title, found.Title)
}

// TestUpdate tests updating a wedding
func (suite *WeddingRepositoryTestSuite) TestUpdate() {
	if suite.db == nil {
		suite.T().Skip("MongoDB not available")
	}

	wedding := suite.createTestWedding()

	// Create wedding
	err := suite.repo.Create(suite.ctx, wedding)
	require.NoError(suite.T(), err)

	// Update wedding
	wedding.Title = "Updated Wedding Title"
	wedding.Status = string(models.WeddingStatusPublished)
	wedding.UpdatedAt = time.Now()

	err = suite.repo.Update(suite.ctx, wedding)
	assert.NoError(suite.T(), err)

	// Verify update
	found, err := suite.repo.GetByID(suite.ctx, wedding.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Updated Wedding Title", found.Title)
	assert.Equal(suite.T(), string(models.WeddingStatusPublished), found.Status)
}

// TestDelete tests deleting a wedding
func (suite *WeddingRepositoryTestSuite) TestDelete() {
	if suite.db == nil {
		suite.T().Skip("MongoDB not available")
	}

	wedding := suite.createTestWedding()

	// Create wedding
	err := suite.repo.Create(suite.ctx, wedding)
	require.NoError(suite.T(), err)

	// Delete wedding
	err = suite.repo.Delete(suite.ctx, wedding.ID)
	assert.NoError(suite.T(), err)

	// Verify deletion
	found, err := suite.repo.GetByID(suite.ctx, wedding.ID)
	// Note: GetByID returns nil, nil for not found
	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), found)
}

// TestGetByUserID tests retrieving weddings by user ID
func (suite *WeddingRepositoryTestSuite) TestGetByUserID() {
	if suite.db == nil {
		suite.T().Skip("MongoDB not available")
	}

	userID := primitive.NewObjectID()

	// Create multiple weddings for the same user
	for i := 0; i < 3; i++ {
		wedding := suite.createTestWedding()
		wedding.UserID = userID
		wedding.Slug = "test-wedding-" + primitive.NewObjectID().Hex()
		err := suite.repo.Create(suite.ctx, wedding)
		require.NoError(suite.T(), err)
	}

	// Get weddings by user ID
	weddings, _, err := suite.repo.GetByUserID(suite.ctx, userID, 1, 10, repository.WeddingFilters{})
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), weddings, 3)

	// Verify all weddings belong to the correct user
	for _, w := range weddings {
		assert.Equal(suite.T(), userID, w.UserID)
	}
}

// TestExistsBySlug tests checking if a slug exists
func (suite *WeddingRepositoryTestSuite) TestExistsBySlug() {
	if suite.db == nil {
		suite.T().Skip("MongoDB not available")
	}

	wedding := suite.createTestWedding()

	// Create wedding
	err := suite.repo.Create(suite.ctx, wedding)
	require.NoError(suite.T(), err)

	// Test existing slug
	exists, err := suite.repo.ExistsBySlug(suite.ctx, wedding.Slug)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), exists)

	// Test non-existing slug
	exists, err = suite.repo.ExistsBySlug(suite.ctx, "non-existent-slug")
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), exists)
}

// TestWeddingRepositoryIntegration runs the test suite
func TestWeddingRepositoryIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
		return
	}

	suite.Run(t, new(WeddingRepositoryTestSuite))
}

// TestWeddingRepositoryUnit tests basic functionality without MongoDB
func TestWeddingRepositoryUnit(t *testing.T) {
	// Test that the repository constructor works when passed a nil database (should panic)
	// This is expected behavior, so we recover from it
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic when creating repository with nil database, but got none")
		}
	}()

	NewMongoWeddingRepository(nil)
}
