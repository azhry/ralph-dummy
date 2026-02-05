package mongodb

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
	"wedding-invitation-backend/internal/domain/models"
	"wedding-invitation-backend/internal/domain/repository"
)

// MongoWeddingRepository implements repository.WeddingRepository for MongoDB
type MongoWeddingRepository struct {
	collection *mongo.Collection
}

// NewMongoWeddingRepository creates a new MongoDB wedding repository
func NewMongoWeddingRepository(db *mongo.Database) repository.WeddingRepository {
	return &MongoWeddingRepository{
		collection: db.Collection("weddings"),
	}
}

// Create inserts a new wedding into the database
func (r *MongoWeddingRepository) Create(ctx context.Context, wedding *models.Wedding) error {
	wedding.CreatedAt = time.Now()
	wedding.UpdatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, wedding)
	return err
}

// GetByID retrieves a wedding by ID
func (r *MongoWeddingRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.Wedding, error) {
	var wedding models.Wedding
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&wedding)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &wedding, nil
}

// GetBySlug retrieves a wedding by slug
func (r *MongoWeddingRepository) GetBySlug(ctx context.Context, slug string) (*models.Wedding, error) {
	var wedding models.Wedding
	err := r.collection.FindOne(ctx, bson.M{"slug": slug}).Decode(&wedding)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &wedding, nil
}

// GetByUserID retrieves weddings by user ID with pagination
func (r *MongoWeddingRepository) GetByUserID(ctx context.Context, userID primitive.ObjectID, page, pageSize int, filters repository.WeddingFilters) ([]*models.Wedding, int64, error) {
	// Build filter
	filter := bson.M{"user_id": userID}

	if filters.Status != "" {
		filter["status"] = filters.Status
	}

	if filters.Search != "" {
		filter["$or"] = []bson.M{
			{"title": bson.M{"$regex": filters.Search, "$options": "i"}},
			{"slug": bson.M{"$regex": filters.Search, "$options": "i"}},
			{"couple.partner1.first_name": bson.M{"$regex": filters.Search, "$options": "i"}},
			{"couple.partner1.last_name": bson.M{"$regex": filters.Search, "$options": "i"}},
			{"couple.partner2.first_name": bson.M{"$regex": filters.Search, "$options": "i"}},
			{"couple.partner2.last_name": bson.M{"$regex": filters.Search, "$options": "i"}},
		}
	}

	if filters.CreatedAfter != nil {
		filter["created_at"] = bson.M{"$gte": filters.CreatedAfter}
	}

	if filters.CreatedBefore != nil {
		if existing, ok := filter["created_at"].(bson.M); ok {
			existing["$lte"] = filters.CreatedBefore
			filter["created_at"] = existing
		} else {
			filter["created_at"] = bson.M{"$lte": filters.CreatedBefore}
		}
	}

	if filters.EventDate != nil {
		filter["event.date"] = bson.M{"$gte": filters.EventDate}
	}

	// Count total documents
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Calculate skip
	skip := (page - 1) * pageSize
	if skip < 0 {
		skip = 0
	}

	skip64 := int64(skip)
	limit64 := int64(pageSize)

	// Find weddings
	cursor, err := r.collection.Find(ctx, filter,
		&options.FindOptions{
			Skip:  &skip64,
			Limit: &limit64,
			Sort:  bson.D{{Key: "created_at", Value: -1}},
		},
	)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var weddings []*models.Wedding
	for cursor.Next(ctx) {
		var wedding models.Wedding
		if err := cursor.Decode(&wedding); err != nil {
			return nil, 0, err
		}
		weddings = append(weddings, &wedding)
	}

	if err := cursor.Err(); err != nil {
		return nil, 0, err
	}

	return weddings, total, nil
}

// Update updates a wedding in the database
func (r *MongoWeddingRepository) Update(ctx context.Context, wedding *models.Wedding) error {
	wedding.UpdatedAt = time.Now()
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": wedding.ID},
		bson.M{"$set": wedding},
	)
	return err
}

// Delete removes a wedding from the database
func (r *MongoWeddingRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// ExistsBySlug checks if a wedding with the given slug exists
func (r *MongoWeddingRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"slug": slug})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ListPublic retrieves public weddings with pagination
func (r *MongoWeddingRepository) ListPublic(ctx context.Context, page, pageSize int, filters repository.PublicWeddingFilters) ([]*models.Wedding, int64, error) {
	// Build filter for public weddings
	filter := bson.M{
		"is_public": true,
		"status":    string(models.WeddingStatusPublished),
	}

	if filters.Search != "" {
		filter["$or"] = []bson.M{
			{"title": bson.M{"$regex": filters.Search, "$options": "i"}},
			{"slug": bson.M{"$regex": filters.Search, "$options": "i"}},
			{"couple.partner1.first_name": bson.M{"$regex": filters.Search, "$options": "i"}},
			{"couple.partner1.last_name": bson.M{"$regex": filters.Search, "$options": "i"}},
			{"couple.partner2.first_name": bson.M{"$regex": filters.Search, "$options": "i"}},
			{"couple.partner2.last_name": bson.M{"$regex": filters.Search, "$options": "i"}},
		}
	}

	if filters.EventDate != nil {
		filter["event.date"] = bson.M{"$gte": filters.EventDate}
	}

	// Count total documents
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Calculate skip
	skip := (page - 1) * pageSize
	if skip < 0 {
		skip = 0
	}

	skip64 := int64(skip)
	limit64 := int64(pageSize)

	// Find weddings
	cursor, err := r.collection.Find(ctx, filter,
		&options.FindOptions{
			Skip:  &skip64,
			Limit: &limit64,
			Sort:  bson.D{{Key: "event.date", Value: 1}},
		},
	)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var weddings []*models.Wedding
	for cursor.Next(ctx) {
		var wedding models.Wedding
		if err := cursor.Decode(&wedding); err != nil {
			return nil, 0, err
		}
		weddings = append(weddings, &wedding)
	}

	if err := cursor.Err(); err != nil {
		return nil, 0, err
	}

	return weddings, total, nil
}

// IncrementViewCount increments the view count for a wedding
func (r *MongoWeddingRepository) IncrementViewCount(ctx context.Context, id primitive.ObjectID) error {
	now := time.Now()
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{
			"$inc": bson.M{"view_count": 1},
			"$set": bson.M{"last_viewed_at": &now},
		},
	)
	return err
}

// UpdateRSVPCount updates the RSVP statistics for a wedding
func (r *MongoWeddingRepository) UpdateRSVPCount(ctx context.Context, weddingID primitive.ObjectID) error {
	// This would typically involve an aggregation pipeline to count RSVPs
	// For now, we'll update the basic RSVP count
	// In a real implementation, you'd want to aggregate RSVP data and update all counts
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": weddingID},
		bson.M{"$set": bson.M{"updated_at": time.Now()}},
	)
	return err
}
