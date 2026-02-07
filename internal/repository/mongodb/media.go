package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"wedding-invitation-backend/internal/domain/models"
	"wedding-invitation-backend/internal/domain/repository"
)

type mediaRepository struct {
	collection *mongo.Collection
}

// NewMediaRepository creates a new MongoDB media repository
func NewMediaRepository(db *mongo.Database) repository.MediaRepository {
	return &mediaRepository{
		collection: db.Collection("media"),
	}
}

// Create creates a new media record
func (r *mediaRepository) Create(ctx context.Context, media *models.Media) error {
	// Generate ID if not set
	if media.ID.IsZero() {
		media.ID = primitive.NewObjectID()
	}

	media.BeforeCreate()

	result, err := r.collection.InsertOne(ctx, media)
	if err != nil {
		return fmt.Errorf("failed to insert media: %w", err)
	}

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		media.ID = oid
	}

	return nil
}

// GetByID retrieves a media record by ID
func (r *mediaRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.Media, error) {
	var media models.Media
	err := r.collection.FindOne(ctx, bson.M{"_id": id, "deletedAt": bson.M{"$exists": false}}).Decode(&media)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("media not found")
		}
		return nil, fmt.Errorf("failed to get media: %w", err)
	}
	return &media, nil
}

// GetByStorageKey retrieves a media record by storage key
func (r *mediaRepository) GetByStorageKey(ctx context.Context, key string) (*models.Media, error) {
	var media models.Media
	err := r.collection.FindOne(ctx, bson.M{"storageKey": key, "deletedAt": bson.M{"$exists": false}}).Decode(&media)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("media not found")
		}
		return nil, fmt.Errorf("failed to get media: %w", err)
	}
	return &media, nil
}

// List retrieves media records with filtering and pagination
func (r *mediaRepository) List(ctx context.Context, filter repository.MediaFilter, opts repository.ListOptions) ([]*models.Media, int64, error) {
	query := bson.M{"deletedAt": bson.M{"$exists": false}}

	if filter.MimeType != "" {
		query["mimeType"] = filter.MimeType
	}
	if filter.CreatedBy != nil {
		query["createdBy"] = *filter.CreatedBy
	}
	if filter.CreatedAfter != nil {
		query["createdAt"] = bson.M{"$gte": *filter.CreatedAfter}
	}
	if filter.CreatedBefore != nil {
		if _, ok := query["createdAt"]; ok {
			query["createdAt"] = bson.M{
				"$gte": query["createdAt"].(bson.M)["$gte"],
				"$lte": *filter.CreatedBefore,
			}
		} else {
			query["createdAt"] = bson.M{"$lte": *filter.CreatedBefore}
		}
	}
	if filter.HasThumbnails {
		query["thumbnails"] = bson.M{"$exists": true, "$ne": bson.M{}}
	}

	// Get total count
	total, err := r.collection.CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count media: %w", err)
	}

	// Set default sort
	if len(opts.Sort) == 0 {
		opts.Sort = bson.D{{Key: "createdAt", Value: -1}}
	}

	findOpts := options.Find().
		SetLimit(opts.Limit).
		SetSkip(opts.Offset).
		SetSort(opts.Sort)

	cursor, err := r.collection.Find(ctx, query, findOpts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list media: %w", err)
	}
	defer cursor.Close(ctx)

	var media []*models.Media
	if err := cursor.All(ctx, &media); err != nil {
		return nil, 0, fmt.Errorf("failed to decode media: %w", err)
	}

	return media, total, nil
}

// Update updates a media record
func (r *mediaRepository) Update(ctx context.Context, media *models.Media) error {
	media.BeforeUpdate()

	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": media.ID},
		bson.M{"$set": media},
	)
	if err != nil {
		return fmt.Errorf("failed to update media: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("media not found")
	}

	return nil
}

// Delete permanently deletes a media record
func (r *mediaRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("failed to delete media: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("media not found")
	}

	return nil
}

// SoftDelete marks a media record as deleted
func (r *mediaRepository) SoftDelete(ctx context.Context, id primitive.ObjectID) error {
	now := time.Now()
	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"deletedAt": now, "updatedAt": now}},
	)
	if err != nil {
		return fmt.Errorf("failed to soft delete media: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("media not found")
	}

	return nil
}

// GetOrphaned retrieves media that were soft deleted before the specified time
func (r *mediaRepository) GetOrphaned(ctx context.Context, before time.Time) ([]*models.Media, error) {
	query := bson.M{
		"deletedAt": bson.M{"$lt": before},
	}

	cursor, err := r.collection.Find(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to find orphaned media: %w", err)
	}
	defer cursor.Close(ctx)

	var media []*models.Media
	if err := cursor.All(ctx, &media); err != nil {
		return nil, fmt.Errorf("failed to decode orphaned media: %w", err)
	}

	return media, nil
}

// GetByCreatedBy retrieves media created by a specific user
func (r *mediaRepository) GetByCreatedBy(ctx context.Context, userID primitive.ObjectID, opts repository.ListOptions) ([]*models.Media, int64, error) {
	filter := repository.MediaFilter{
		CreatedBy: &userID,
	}
	return r.List(ctx, filter, opts)
}
