package mongodb

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"wedding-invitation-backend/internal/domain/models"
	"wedding-invitation-backend/internal/domain/repository"
)

func setupMediaTestDB(t *testing.T) *mongo.Collection {
	// Use in-memory MongoDB for testing
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(t, err)

	db := client.Database("test_media")
	collection := db.Collection("media")

	// Clean up before tests
	collection.Drop(ctx)

	// Clean up after tests
	t.Cleanup(func() {
		collection.Drop(ctx)
		client.Disconnect(ctx)
	})

	return collection
}

func createTestMedia(t *testing.T) *models.Media {
	userID := primitive.NewObjectID()
	return &models.Media{
		Filename:    "test-photo.jpg",
		OriginalURL: "http://example.com/test-photo.jpg",
		Thumbnails: map[string]string{
			"small":  "http://example.com/test-photo-small.jpg",
			"medium": "http://example.com/test-photo-medium.jpg",
		},
		Size:      1024000,
		MimeType:  "image/jpeg",
		Width:     1920,
		Height:    1080,
		Format:    "jpeg",
		StorageKey: "uploads/2024/01/01/abc123/original.jpg",
		CreatedBy: userID,
	}
}

func TestMediaRepository_Create(t *testing.T) {
	collection := setupMediaTestDB(t)
	repo := NewMediaRepository(collection.Database())

	media := createTestMedia(t)
	ctx := context.Background()

	err := repo.Create(ctx, media)
	assert.NoError(t, err)
	assert.NotEmpty(t, media.ID)

	// Verify the media was created
	retrieved, err := repo.GetByID(ctx, media.ID)
	assert.NoError(t, err)
	assert.Equal(t, media.Filename, retrieved.Filename)
	assert.Equal(t, media.OriginalURL, retrieved.OriginalURL)
	assert.Equal(t, media.Size, retrieved.Size)
}

func TestMediaRepository_GetByID(t *testing.T) {
	collection := setupMediaTestDB(t)
	repo := NewMediaRepository(collection.Database())
	ctx := context.Background()

	// Test non-existent media
	nonExistentID := primitive.NewObjectID()
	_, err := repo.GetByID(ctx, nonExistentID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "media not found")

	// Test existing media
	media := createTestMedia(t)
	err = repo.Create(ctx, media)
	require.NoError(t, err)

	retrieved, err := repo.GetByID(ctx, media.ID)
	assert.NoError(t, err)
	assert.Equal(t, media.ID, retrieved.ID)
	assert.Equal(t, media.Filename, retrieved.Filename)
}

func TestMediaRepository_GetByStorageKey(t *testing.T) {
	collection := setupMediaTestDB(t)
	repo := NewMediaRepository(collection.Database())
	ctx := context.Background()

	// Test non-existent storage key
	_, err := repo.GetByStorageKey(ctx, "non-existent-key")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "media not found")

	// Test existing storage key
	media := createTestMedia(t)
	err = repo.Create(ctx, media)
	require.NoError(t, err)

	retrieved, err := repo.GetByStorageKey(ctx, media.StorageKey)
	assert.NoError(t, err)
	assert.Equal(t, media.ID, retrieved.ID)
	assert.Equal(t, media.StorageKey, retrieved.StorageKey)
}

func TestMediaRepository_List(t *testing.T) {
	collection := setupMediaTestDB(t)
	repo := NewMediaRepository(collection.Database())
	ctx := context.Background()

	// Create test media with different properties
	userID1 := primitive.NewObjectID()
	userID2 := primitive.NewObjectID()

	media1 := createTestMedia(t)
	media1.CreatedBy = userID1
	media1.MimeType = "image/jpeg"

	media2 := createTestMedia(t)
	media2.Filename = "test-photo.png"
	media2.MimeType = "image/png"
	media2.CreatedBy = userID2

	media3 := createTestMedia(t)
	media3.Filename = "test-photo.webp"
	media3.MimeType = "image/webp"
	media3.CreatedBy = userID1

	// Create media
	err := repo.Create(ctx, media1)
	require.NoError(t, err)
	err = repo.Create(ctx, media2)
	require.NoError(t, err)
	err = repo.Create(ctx, media3)
	require.NoError(t, err)

	// Test list all
	mediaList, total, err := repo.List(ctx, repository.MediaFilter{}, repository.ListOptions{
		Limit:  10,
		Offset: 0,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, mediaList, 3)

	// Test filter by MIME type
	filter := repository.MediaFilter{MimeType: "image/jpeg"}
	mediaList, total, err = repo.List(ctx, filter, repository.ListOptions{})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, mediaList, 1)
	assert.Equal(t, "image/jpeg", mediaList[0].MimeType)

	// Test filter by created by
	filter = repository.MediaFilter{CreatedBy: &userID1}
	mediaList, total, err = repo.List(ctx, filter, repository.ListOptions{})
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, mediaList, 2)

	// Test filter by created after
	past := time.Now().Add(-1 * time.Hour)
	filter = repository.MediaFilter{CreatedAfter: &past}
	mediaList, total, err = repo.List(ctx, filter, repository.ListOptions{})
	assert.NoError(t, err)
	assert.Equal(t, int64(3), total)

	// Test filter by has thumbnails
	filter = repository.MediaFilter{HasThumbnails: true}
	mediaList, total, err = repo.List(ctx, filter, repository.ListOptions{})
	assert.NoError(t, err)
	assert.Equal(t, int64(3), total) // All test media have thumbnails

	// Test pagination
	mediaList, total, err = repo.List(ctx, repository.MediaFilter{}, repository.ListOptions{
		Limit:  2,
		Offset: 0,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, mediaList, 2)
}

func TestMediaRepository_Update(t *testing.T) {
	collection := setupMediaTestDB(t)
	repo := NewMediaRepository(collection.Database())
	ctx := context.Background()

	media := createTestMedia(t)
	err := repo.Create(ctx, media)
	require.NoError(t, err)

	// Update media
	media.Filename = "updated-photo.jpg"
	media.Width = 2000
	media.Height = 1500

	err = repo.Update(ctx, media)
	assert.NoError(t, err)

	// Verify update
	retrieved, err := repo.GetByID(ctx, media.ID)
	assert.NoError(t, err)
	assert.Equal(t, "updated-photo.jpg", retrieved.Filename)
	assert.Equal(t, 2000, retrieved.Width)
	assert.Equal(t, 1500, retrieved.Height)
}

func TestMediaRepository_Delete(t *testing.T) {
	collection := setupMediaTestDB(t)
	repo := NewMediaRepository(collection.Database())
	ctx := context.Background()

	media := createTestMedia(t)
	err := repo.Create(ctx, media)
	require.NoError(t, err)

	// Delete media
	err = repo.Delete(ctx, media.ID)
	assert.NoError(t, err)

	// Verify deletion
	_, err = repo.GetByID(ctx, media.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "media not found")
}

func TestMediaRepository_SoftDelete(t *testing.T) {
	collection := setupMediaTestDB(t)
	repo := NewMediaRepository(collection.Database())
	ctx := context.Background()

	media := createTestMedia(t)
	err := repo.Create(ctx, media)
	require.NoError(t, err)

	// Soft delete media
	err = repo.SoftDelete(ctx, media.ID)
	assert.NoError(t, err)

	// Verify soft delete (should not be found in normal queries)
	_, err = repo.GetByID(ctx, media.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "media not found")

	// Verify record still exists in database
	var deletedMedia models.Media
	err = collection.FindOne(ctx, bson.M{"_id": media.ID}).Decode(&deletedMedia)
	assert.NoError(t, err)
	assert.NotNil(t, deletedMedia.DeletedAt)
}

func TestMediaRepository_GetOrphaned(t *testing.T) {
	collection := setupMediaTestDB(t)
	repo := NewMediaRepository(collection.Database())
	ctx := context.Background()

	media1 := createTestMedia(t)
	media2 := createTestMedia(t)

	err := repo.Create(ctx, media1)
	require.NoError(t, err)
	err = repo.Create(ctx, media2)
	require.NoError(t, err)

	// Soft delete media1 a while ago
	oldTime := time.Now().Add(-2 * time.Hour)
	collection.UpdateOne(ctx, bson.M{"_id": media1.ID}, bson.M{"$set": bson.M{"deletedAt": oldTime}})

	// Soft delete media2 recently
	err = repo.SoftDelete(ctx, media2.ID)
	require.NoError(t, err)

	// Get orphaned media (deleted before 1 hour ago)
	cutoff := time.Now().Add(-1 * time.Hour)
	orphaned, err := repo.GetOrphaned(ctx, cutoff)
	assert.NoError(t, err)
	assert.Len(t, orphaned, 1)
	assert.Equal(t, media1.ID, orphaned[0].ID)
}

func TestMediaRepository_GetByCreatedBy(t *testing.T) {
	collection := setupMediaTestDB(t)
	repo := NewMediaRepository(collection.Database())
	ctx := context.Background()

	userID1 := primitive.NewObjectID()
	userID2 := primitive.NewObjectID()

	media1 := createTestMedia(t)
	media1.CreatedBy = userID1

	media2 := createTestMedia(t)
	media2.CreatedBy = userID2

	err := repo.Create(ctx, media1)
	require.NoError(t, err)
	err = repo.Create(ctx, media2)
	require.NoError(t, err)

	// Get media for user 1
	mediaList, total, err := repo.GetByCreatedBy(ctx, userID1, repository.ListOptions{})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, mediaList, 1)
	assert.Equal(t, userID1, mediaList[0].CreatedBy)

	// Get media for user 2
	mediaList, total, err = repo.GetByCreatedBy(ctx, userID2, repository.ListOptions{})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, mediaList, 1)
	assert.Equal(t, userID2, mediaList[0].CreatedBy)
}