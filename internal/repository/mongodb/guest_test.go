package mongodb

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"wedding-invitation-backend/internal/config"
	"wedding-invitation-backend/internal/domain/models"
	"wedding-invitation-backend/internal/domain/repository"
	"wedding-invitation-backend/pkg/database"
)

func TestGuestRepository_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test repository
	repo, cleanup := setupTestGuestRepository(t)
	defer cleanup()

	// Test data
	weddingID := primitive.NewObjectID()
	guest := &models.Guest{
		WeddingID:        weddingID,
		FirstName:        "John",
		LastName:         "Doe",
		Email:            "john.doe@example.com",
		Phone:            "+1234567890",
		Relationship:     "Friend",
		Side:             "groom",
		InvitedVia:       "digital",
		InvitationStatus: "pending",
		AllowPlusOne:     true,
		MaxPlusOnes:      1,
		VIP:              false,
		Notes:            "Test guest",
		CreatedBy:        primitive.NewObjectID(),
	}

	// Test Create
	err := repo.Create(context.Background(), guest)
	assert.NoError(t, err)
	assert.NotEmpty(t, guest.ID)
	assert.NotZero(t, guest.CreatedAt)
	assert.NotZero(t, guest.UpdatedAt)
}

func TestGuestRepository_GetByID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	_, cleanup := setupTestGuestRepository(t)
	defer cleanup()

	repo, _ := setupTestGuestRepository(t)

	// Create test guest
	guest := &models.Guest{
		WeddingID: primitive.NewObjectID(),
		FirstName: "Jane",
		LastName:  "Smith",
		Email:     "jane.smith@example.com",
		CreatedBy: primitive.NewObjectID(),
	}

	err := repo.Create(context.Background(), guest)
	require.NoError(t, err)

	// Test GetByID
	found, err := repo.GetByID(context.Background(), guest.ID)
	assert.NoError(t, err)
	assert.Equal(t, guest.FirstName, found.FirstName)
	assert.Equal(t, guest.LastName, found.LastName)
	assert.Equal(t, guest.Email, found.Email)
}

func TestGuestRepository_ListByWedding(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	_, cleanup := setupTestGuestRepository(t)
	defer cleanup()

	repo, _ := setupTestGuestRepository(t)

	// Create test guests
	weddingID := primitive.NewObjectID()
	guests := []*models.Guest{
		{
			WeddingID: weddingID,
			FirstName: "Alice",
			LastName:  "Johnson",
			CreatedBy: primitive.NewObjectID(),
		},
		{
			WeddingID: weddingID,
			FirstName: "Bob",
			LastName:  "Wilson",
			CreatedBy: primitive.NewObjectID(),
		},
	}

	for _, guest := range guests {
		err := repo.Create(context.Background(), guest)
		require.NoError(t, err)
	}

	// Test ListByWedding
	found, total, err := repo.ListByWedding(context.Background(), weddingID, 1, 10, repository.GuestFilters{})
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, found, 2)
}

func TestGuestRepository_Update(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	_, cleanup := setupTestGuestRepository(t)
	defer cleanup()

	repo, _ := setupTestGuestRepository(t)

	// Create test guest
	guest := &models.Guest{
		WeddingID: primitive.NewObjectID(),
		FirstName: "Original",
		LastName:  "Name",
		Email:     "original@example.com",
		CreatedBy: primitive.NewObjectID(),
	}

	err := repo.Create(context.Background(), guest)
	require.NoError(t, err)

	// Update guest
	originalUpdatedAt := guest.UpdatedAt
	time.Sleep(time.Millisecond) // Ensure different timestamp

	guest.FirstName = "Updated"
	guest.Email = "updated@example.com"

	err = repo.Update(context.Background(), guest)
	assert.NoError(t, err)
	assert.True(t, guest.UpdatedAt.After(originalUpdatedAt))

	// Verify update
	found, err := repo.GetByID(context.Background(), guest.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated", found.FirstName)
	assert.Equal(t, "updated@example.com", found.Email)
}

func TestGuestRepository_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	_, cleanup := setupTestGuestRepository(t)
	defer cleanup()

	repo, _ := setupTestGuestRepository(t)

	// Create test guest
	guest := &models.Guest{
		WeddingID: primitive.NewObjectID(),
		FirstName: "ToDelete",
		LastName:  "Guest",
		CreatedBy: primitive.NewObjectID(),
	}

	err := repo.Create(context.Background(), guest)
	require.NoError(t, err)

	// Test Delete
	err = repo.Delete(context.Background(), guest.ID)
	assert.NoError(t, err)

	// Verify deletion
	_, err = repo.GetByID(context.Background(), guest.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "guest not found")
}

func TestGuestRepository_CreateMany(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	_, cleanup := setupTestGuestRepository(t)
	defer cleanup()

	repo, _ := setupTestGuestRepository(t)

	// Create test guests
	weddingID := primitive.NewObjectID()
	guests := []*models.Guest{
		{
			WeddingID: weddingID,
			FirstName: "Guest1",
			LastName:  "Test1",
			CreatedBy: primitive.NewObjectID(),
		},
		{
			WeddingID: weddingID,
			FirstName: "Guest2",
			LastName:  "Test2",
			CreatedBy: primitive.NewObjectID(),
		},
	}

	// Test CreateMany
	err := repo.CreateMany(context.Background(), guests)
	assert.NoError(t, err)

	// Verify creation
	for _, guest := range guests {
		assert.NotEmpty(t, guest.ID)
		assert.NotZero(t, guest.CreatedAt)
	}
}

func TestGuestRepository_ImportBatch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	_, cleanup := setupTestGuestRepository(t)
	defer cleanup()

	repo, _ := setupTestGuestRepository(t)

	// Create test guests with batch ID
	weddingID := primitive.NewObjectID()
	batchID := "test_batch_123"
	guests := []*models.Guest{
		{
			WeddingID: weddingID,
			FirstName: "Import1",
			LastName:  "Test1",
			CreatedBy: primitive.NewObjectID(),
		},
		{
			WeddingID: weddingID,
			FirstName: "Import2",
			LastName:  "Test2",
			CreatedBy: primitive.NewObjectID(),
		},
	}

	// Test ImportBatch
	err := repo.ImportBatch(context.Background(), guests, batchID)
	assert.NoError(t, err)

	// Verify import
	found, err := repo.GetByImportBatch(context.Background(), weddingID, batchID)
	assert.NoError(t, err)
	assert.Len(t, found, 2)

	for _, guest := range found {
		assert.Equal(t, batchID, guest.ImportBatchID)
	}
}

func TestGuestRepository_EnsureIndexes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Note: EnsureIndexes is not part of the GuestRepository interface
	// This test should be removed or implemented differently
	t.Skip("EnsureIndexes not implemented in GuestRepository interface")
}

func setupTestGuestRepository(t *testing.T) (repository.GuestRepository, func()) {
	// Setup test database
	testDBConfig := &config.DatabaseConfig{
		URI:      "mongodb://admin:password123@localhost:27017/wedding_invitations?authSource=admin",
		Database: "wedding_test_" + primitive.NewObjectID().Hex(),
		Timeout:  10,
	}

	db, err := database.NewMongoDB(testDBConfig)
	if err != nil {
		t.Skipf("Skipping integration tests: MongoDB not available: %v", err)
		return nil, func() {}
	}

	repo := NewGuestRepository(db.Database)

	cleanup := func() {
		db.Close(context.Background())
		client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(testDBConfig.URI))
		if err == nil {
			client.Database(testDBConfig.Database).Drop(context.Background())
			client.Disconnect(context.Background())
		}
	}

	return repo, cleanup
}
