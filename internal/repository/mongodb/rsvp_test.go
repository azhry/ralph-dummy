package mongodb

import (
	"context"
	"fmt"
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
)

func TestMongoRSVPRepository_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Setup test database
	client, db := setupTestDB(t, "test_rsvps")
	defer client.Disconnect(context.Background())

	repo := NewMongoRSVPRepository(db)

	// Test data
	weddingID := primitive.NewObjectID()
	rsvp := &models.RSVP{
		ID:               primitive.NewObjectID(),
		WeddingID:        weddingID,
		FirstName:        "John",
		LastName:         "Doe",
		Email:            "john.doe@example.com",
		Status:           "attending",
		AttendanceCount:  2,
		PlusOnes:         []models.PlusOneInfo{{FirstName: "Jane", LastName: "Doe"}},
		PlusOneCount:     1,
		SubmittedAt:      time.Now(),
		Source:           "web",
		ConfirmationSent: false,
	}

	// Test Create
	err := repo.Create(context.Background(), rsvp)
	require.NoError(t, err)

	// Verify creation
	found, err := repo.GetByID(context.Background(), rsvp.ID)
	require.NoError(t, err)
	assert.Equal(t, rsvp.FirstName, found.FirstName)
	assert.Equal(t, rsvp.LastName, found.LastName)
	assert.Equal(t, rsvp.Email, found.Email)
	assert.Equal(t, rsvp.Status, found.Status)
}

func TestMongoRSVPRepository_GetByID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	client, db := setupTestDB(t, "test_rsvps_getbyid")
	defer client.Disconnect(context.Background())

	repo := NewMongoRSVPRepository(db)

	// Test non-existent RSVP
	nonExistentID := primitive.NewObjectID()
	_, err := repo.GetByID(context.Background(), nonExistentID)
	assert.Error(t, err)
	assert.Equal(t, repository.ErrNotFound, err)

	// Test existing RSVP
	rsvp := &models.RSVP{
		ID:              primitive.NewObjectID(),
		WeddingID:       primitive.NewObjectID(),
		FirstName:       "Jane",
		LastName:        "Smith",
		Status:          "not-attending",
		AttendanceCount: 1,
		SubmittedAt:     time.Now(),
		Source:          "web",
	}

	err = repo.Create(context.Background(), rsvp)
	require.NoError(t, err)

	found, err := repo.GetByID(context.Background(), rsvp.ID)
	require.NoError(t, err)
	assert.Equal(t, rsvp.FirstName, found.FirstName)
	assert.Equal(t, rsvp.LastName, found.LastName)
}

func TestMongoRSVPRepository_GetByEmail(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	client, db := setupTestDB(t, "test_rsvps_getbyemail")
	defer client.Disconnect(context.Background())

	repo := NewMongoRSVPRepository(db)

	weddingID := primitive.NewObjectID()
	email := "test@example.com"

	// Test non-existent RSVP
	_, err := repo.GetByEmail(context.Background(), weddingID, email)
	assert.Error(t, err)
	assert.Equal(t, repository.ErrNotFound, err)

	// Test existing RSVP
	rsvp := &models.RSVP{
		ID:          primitive.NewObjectID(),
		WeddingID:   weddingID,
		FirstName:   "Test",
		LastName:    "User",
		Email:       email,
		Status:      "attending",
		SubmittedAt: time.Now(),
		Source:      "web",
	}

	err = repo.Create(context.Background(), rsvp)
	require.NoError(t, err)

	found, err := repo.GetByEmail(context.Background(), weddingID, email)
	require.NoError(t, err)
	assert.Equal(t, rsvp.FirstName, found.FirstName)
	assert.Equal(t, rsvp.Email, found.Email)
}

func TestMongoRSVPRepository_ListByWedding(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	client, db := setupTestDB(t, "test_rsvps_list")
	defer client.Disconnect(context.Background())

	repo := NewMongoRSVPRepository(db)

	weddingID := primitive.NewObjectID()

	// Create test RSVPs
	rsvps := make([]*models.RSVP, 5)
	for i := 0; i < 5; i++ {
		rsvps[i] = &models.RSVP{
			ID:              primitive.NewObjectID(),
			WeddingID:       weddingID,
			FirstName:       fmt.Sprintf("User%d", i),
			LastName:        "Test",
			Email:           fmt.Sprintf("user%d@test.com", i),
			Status:          "attending",
			AttendanceCount: 1,
			SubmittedAt:     time.Now().Add(time.Duration(i) * time.Hour),
			Source:          "web",
		}
		err := repo.Create(context.Background(), rsvps[i])
		require.NoError(t, err)
	}

	// Test ListByWedding
	results, total, err := repo.ListByWedding(context.Background(), weddingID, 1, 10, repository.RSVPFilters{})
	require.NoError(t, err)
	assert.Equal(t, 5, len(results))
	assert.Equal(t, int64(5), total)

	// Test pagination
	results, total, err = repo.ListByWedding(context.Background(), weddingID, 1, 2, repository.RSVPFilters{})
	require.NoError(t, err)
	assert.Equal(t, 2, len(results))
	assert.Equal(t, int64(5), total)

	// Test filters by status
	results, total, err = repo.ListByWedding(context.Background(), weddingID, 1, 10, repository.RSVPFilters{
		Status: "attending",
	})
	require.NoError(t, err)
	assert.Equal(t, 5, len(results))
	assert.Equal(t, int64(5), total)

	// Test filters by search
	results, total, err = repo.ListByWedding(context.Background(), weddingID, 1, 10, repository.RSVPFilters{
		Search: "User0",
	})
	require.NoError(t, err)
	assert.Equal(t, 1, len(results))
	assert.Equal(t, int64(1), total)
}

func TestMongoRSVPRepository_Update(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	client, db := setupTestDB(t, "test_rsvps_update")
	defer client.Disconnect(context.Background())

	repo := NewMongoRSVPRepository(db)

	// Create RSVP
	rsvp := &models.RSVP{
		ID:              primitive.NewObjectID(),
		WeddingID:       primitive.NewObjectID(),
		FirstName:       "Original",
		LastName:        "Name",
		Status:          "attending",
		AttendanceCount: 1,
		SubmittedAt:     time.Now(),
		Source:          "web",
	}

	err := repo.Create(context.Background(), rsvp)
	require.NoError(t, err)

	// Update RSVP
	rsvp.FirstName = "Updated"
	rsvp.Status = "not-attending"
	rsvp.AttendanceCount = 2

	err = repo.Update(context.Background(), rsvp)
	require.NoError(t, err)

	// Verify update
	found, err := repo.GetByID(context.Background(), rsvp.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated", found.FirstName)
	assert.Equal(t, "not-attending", found.Status)
	assert.Equal(t, 2, found.AttendanceCount)
}

func TestMongoRSVPRepository_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	client, db := setupTestDB(t, "test_rsvps_delete")
	defer client.Disconnect(context.Background())

	repo := NewMongoRSVPRepository(db)

	// Create RSVP
	rsvp := &models.RSVP{
		ID:              primitive.NewObjectID(),
		WeddingID:       primitive.NewObjectID(),
		FirstName:       "ToDelete",
		LastName:        "User",
		Status:          "attending",
		AttendanceCount: 1,
		SubmittedAt:     time.Now(),
		Source:          "web",
	}

	err := repo.Create(context.Background(), rsvp)
	require.NoError(t, err)

	// Verify exists
	_, err = repo.GetByID(context.Background(), rsvp.ID)
	require.NoError(t, err)

	// Delete RSVP
	err = repo.Delete(context.Background(), rsvp.ID)
	require.NoError(t, err)

	// Verify deleted
	_, err = repo.GetByID(context.Background(), rsvp.ID)
	assert.Error(t, err)
	assert.Equal(t, repository.ErrNotFound, err)
}

func TestMongoRSVPRepository_GetStatistics(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	client, db := setupTestDB(t, "test_rsvps_stats")
	defer client.Disconnect(context.Background())

	repo := NewMongoRSVPRepository(db)

	weddingID := primitive.NewObjectID()

	// Create test RSVPs with different statuses
	statuses := []string{"attending", "not-attending", "maybe", "attending"}
	for i, status := range statuses {
		rsvp := &models.RSVP{
			ID:              primitive.NewObjectID(),
			WeddingID:       weddingID,
			FirstName:       fmt.Sprintf("User%d", i),
			LastName:        "Test",
			Status:          status,
			AttendanceCount: 1,
			PlusOneCount:    i % 2, // Alternate 0 and 1
			DietarySelected: []string{"vegetarian", "vegan", "gluten-free", "vegetarian"}[i : i+1],
			SubmittedAt:     time.Now(),
			Source:          "web",
		}
		err := repo.Create(context.Background(), rsvp)
		require.NoError(t, err)
	}

	// Get statistics
	stats, err := repo.GetStatistics(context.Background(), weddingID)
	require.NoError(t, err)

	// Verify counts
	assert.Equal(t, 4, stats.TotalResponses)
	assert.Equal(t, 2, stats.Attending)
	assert.Equal(t, 1, stats.NotAttending)
	assert.Equal(t, 1, stats.Maybe)
	assert.True(t, stats.TotalGuests > 0)
	assert.True(t, stats.PlusOnesCount > 0)

	// Verify dietary counts
	assert.Contains(t, stats.DietaryCounts, "vegetarian")
	assert.Contains(t, stats.DietaryCounts, "vegan")
	assert.Contains(t, stats.DietaryCounts, "gluten-free")
}

func TestMongoRSVPRepository_MarkConfirmationSent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	client, db := setupTestDB(t, "test_rsvps_confirmation")
	defer client.Disconnect(context.Background())

	repo := NewMongoRSVPRepository(db)

	// Create RSVP
	rsvp := &models.RSVP{
		ID:               primitive.NewObjectID(),
		WeddingID:        primitive.NewObjectID(),
		FirstName:        "Test",
		LastName:         "User",
		Status:           "attending",
		AttendanceCount:  1,
		SubmittedAt:      time.Now(),
		Source:           "web",
		ConfirmationSent: false,
	}

	err := repo.Create(context.Background(), rsvp)
	require.NoError(t, err)

	// Mark confirmation sent
	err = repo.MarkConfirmationSent(context.Background(), rsvp.ID)
	require.NoError(t, err)

	// Verify updated
	found, err := repo.GetByID(context.Background(), rsvp.ID)
	require.NoError(t, err)
	assert.True(t, found.ConfirmationSent)
	assert.NotNil(t, found.ConfirmationSentAt)
}

func TestMongoRSVPRepository_GetSubmissionTrend(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	client, db := setupTestDB(t, "test_rsvps_trend")
	defer client.Disconnect(context.Background())

	repo := NewMongoRSVPRepository(db)

	weddingID := primitive.NewObjectID()

	// Create test RSVPs over the last few days
	for i := 0; i < 5; i++ {
		rsvp := &models.RSVP{
			ID:              primitive.NewObjectID(),
			WeddingID:       weddingID,
			FirstName:       fmt.Sprintf("User%d", i),
			LastName:        "Test",
			Status:          "attending",
			AttendanceCount: 1,
			SubmittedAt:     time.Now().AddDate(0, 0, -i), // i days ago
			Source:          "web",
		}
		err := repo.Create(context.Background(), rsvp)
		require.NoError(t, err)
	}

	// Get submission trend
	trend, err := repo.GetSubmissionTrend(context.Background(), weddingID, 7)
	require.NoError(t, err)
	assert.Equal(t, 7, len(trend)) // 7 days requested

	// Should have some counts in the trend
	totalCount := 0
	for _, day := range trend {
		totalCount += day.Count
	}
	assert.True(t, totalCount > 0)
}

// Temporary setupTestDB function to make tests compile
func setupTestDB(t *testing.T, dbName string) (*mongo.Client, *mongo.Database) {
	testDBConfig := &config.DatabaseConfig{
		URI:      "mongodb://admin:password123@localhost:27017/wedding_invitations?authSource=admin",
		Database: "wedding_test_" + primitive.NewObjectID().Hex(),
		Timeout:  10,
	}

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(testDBConfig.URI))
	if err != nil {
		t.Skipf("Skipping integration tests: Cannot connect to MongoDB: %v", err)
		return nil, nil
	}

	return client, client.Database(testDBConfig.Database)
}
