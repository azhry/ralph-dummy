package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"wedding-invitation-backend/internal/domain/models"
	"wedding-invitation-backend/internal/domain/repository"
)

// GuestRepository implements repository.GuestRepository interface
type GuestRepository struct {
	collection *mongo.Collection
}

// NewGuestRepository creates a new guest repository
func NewGuestRepository(db *mongo.Database) repository.GuestRepository {
	return &GuestRepository{
		collection: db.Collection("guests"),
	}
}

// Create creates a new guest
func (r *GuestRepository) Create(ctx context.Context, guest *models.Guest) error {
	now := time.Now()
	guest.CreatedAt = now
	guest.UpdatedAt = now

	_, err := r.collection.InsertOne(ctx, guest)
	if err != nil {
		return fmt.Errorf("failed to create guest: %w", err)
	}

	return nil
}

// GetByID retrieves a guest by ID
func (r *GuestRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.Guest, error) {
	var guest models.Guest
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&guest)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get guest: %w", err)
	}
	return &guest, nil
}

// GetByEmail retrieves a guest by email within a wedding
func (r *GuestRepository) GetByEmail(ctx context.Context, weddingID primitive.ObjectID, email string) (*models.Guest, error) {
	var guest models.Guest
	err := r.collection.FindOne(ctx, bson.M{
		"wedding_id": weddingID,
		"email":      email,
	}).Decode(&guest)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get guest by email: %w", err)
	}
	return &guest, nil
}

// CreateMany creates multiple guests in a single operation
func (r *GuestRepository) CreateMany(ctx context.Context, guests []*models.Guest) error {
	if len(guests) == 0 {
		return nil
	}

	now := time.Now()
	var docs []interface{}

	for _, guest := range guests {
		guest.CreatedAt = now
		guest.UpdatedAt = now
		docs = append(docs, guest)
	}

	_, err := r.collection.InsertMany(ctx, docs)
	if err != nil {
		return fmt.Errorf("failed to create many guests: %w", err)
	}

	return nil
}

// ListByWedding retrieves guests by wedding ID with pagination
func (r *GuestRepository) ListByWedding(ctx context.Context, weddingID primitive.ObjectID, page, pageSize int, filters repository.GuestFilters) ([]*models.Guest, int64, error) {
	// Calculate offset from page
	offset := 0
	if page > 0 {
		offset = (page - 1) * pageSize
	}

	// Build base filter
	baseFilter := bson.M{"wedding_id": weddingID}
	filter := r.buildFilters(baseFilter, filters)

	// Count total documents matching filter
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count guests: %w", err)
	}

	// Find guests with pagination
	opts := options.Find()
	opts.SetSort(bson.M{"last_name": 1, "first_name": 1})
	if pageSize > 0 {
		opts.SetLimit(int64(pageSize))
	}
	if offset > 0 {
		opts.SetSkip(int64(offset))
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get guests: %w", err)
	}
	defer cursor.Close(ctx)

	var guests []*models.Guest
	for cursor.Next(ctx) {
		var guest models.Guest
		if err := cursor.Decode(&guest); err != nil {
			return nil, 0, fmt.Errorf("failed to decode guest: %w", err)
		}
		guests = append(guests, &guest)
	}

	return guests, total, nil
}

// Update updates an existing guest
func (r *GuestRepository) Update(ctx context.Context, guest *models.Guest) error {
	guest.UpdatedAt = time.Now()

	update := bson.M{"$set": guest}
	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": guest.ID}, update)
	if err != nil {
		return fmt.Errorf("failed to update guest: %w", err)
	}

	if result.MatchedCount == 0 {
		return errors.New("guest not found")
	}

	return nil
}

// Delete deletes a guest
func (r *GuestRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("failed to delete guest: %w", err)
	}

	if result.DeletedCount == 0 {
		return errors.New("guest not found")
	}

	return nil
}

// ImportBatch imports multiple guests with a batch ID
func (r *GuestRepository) ImportBatch(ctx context.Context, guests []*models.Guest, batchID string) error {
	if len(guests) == 0 {
		return nil
	}

	now := time.Now()
	var docs []interface{}

	for _, guest := range guests {
		guest.CreatedAt = now
		guest.UpdatedAt = now
		guest.ImportBatchID = batchID
		docs = append(docs, guest)
	}

	_, err := r.collection.InsertMany(ctx, docs)
	if err != nil {
		return fmt.Errorf("failed to import guest batch: %w", err)
	}

	return nil
}

// GetByImportBatch retrieves guests by wedding ID and batch ID
func (r *GuestRepository) GetByImportBatch(ctx context.Context, weddingID primitive.ObjectID, batchID string) ([]*models.Guest, error) {
	filter := bson.M{
		"wedding_id":      weddingID,
		"import_batch_id": batchID,
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get guests by batch: %w", err)
	}
	defer cursor.Close(ctx)

	var guests []*models.Guest
	for cursor.Next(ctx) {
		var guest models.Guest
		if err := cursor.Decode(&guest); err != nil {
			return nil, fmt.Errorf("failed to decode guest: %w", err)
		}
		guests = append(guests, &guest)
	}

	return guests, nil
}

// buildFilters constructs the MongoDB filter based on the provided filters
func (r *GuestRepository) buildFilters(baseFilter bson.M, filters repository.GuestFilters) bson.M {
	if filters.Search != "" {
		searchRegex := primitive.Regex{Pattern: filters.Search, Options: "i"}
		baseFilter["$or"] = []bson.M{
			{"first_name": bson.M{"$regex": searchRegex}},
			{"last_name": bson.M{"$regex": searchRegex}},
			{"email": bson.M{"$regex": searchRegex}},
			{"relationship": bson.M{"$regex": searchRegex}},
		}
	}

	if filters.Side != "" {
		baseFilter["side"] = filters.Side
	}

	if filters.RSVPStatus != "" {
		baseFilter["rsvp_status"] = filters.RSVPStatus
	}

	if filters.Relationship != "" {
		baseFilter["relationship"] = filters.Relationship
	}

	if filters.VIP != nil {
		baseFilter["vip"] = *filters.VIP
	}

	return baseFilter
}

// EnsureIndexes creates necessary indexes for the guests collection
func (r *GuestRepository) EnsureIndexes(ctx context.Context) error {
	indexModels := []mongo.IndexModel{
		{
			Keys:    bson.M{"wedding_id": 1},
			Options: options.Index().SetName("wedding_id_index"),
		},
		{
			Keys:    bson.M{"wedding_id": 1, "email": 1},
			Options: options.Index().SetName("wedding_email_index").SetUnique(true),
		},
		{
			Keys:    bson.M{"wedding_id": 1, "side": 1},
			Options: options.Index().SetName("wedding_side_index"),
		},
		{
			Keys:    bson.M{"wedding_id": 1, "rsvp_status": 1},
			Options: options.Index().SetName("wedding_rsvp_status_index"),
		},
		{
			Keys:    bson.M{"wedding_id": 1, "invitation_status": 1},
			Options: options.Index().SetName("wedding_invitation_status_index"),
		},
		{
			Keys:    bson.M{"import_batch_id": 1},
			Options: options.Index().SetName("import_batch_id_index"),
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexModels)
	if err != nil {
		return fmt.Errorf("failed to create guest indexes: %w", err)
	}

	return nil
}
