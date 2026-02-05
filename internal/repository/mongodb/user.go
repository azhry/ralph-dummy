package repository

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

// MongoUserRepository implements repository.UserRepository for MongoDB
type MongoUserRepository struct {
	collection *mongo.Collection
}

// NewMongoUserRepository creates a new MongoDB user repository
func NewMongoUserRepository(db *mongo.Database) repository.UserRepository {
	return &MongoUserRepository{
		collection: db.Collection("users"),
	}
}

// Create inserts a new user into the database
func (r *MongoUserRepository) Create(ctx context.Context, user *models.User) error {
	_, err := r.collection.InsertOne(ctx, user)
	return err
}

// GetByID retrieves a user by ID
func (r *MongoUserRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *MongoUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

// GetByVerificationToken retrieves a user by verification token
func (r *MongoUserRepository) GetByVerificationToken(ctx context.Context, token string) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"email_verification_token": token}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

// GetByResetToken retrieves a user by password reset token
func (r *MongoUserRepository) GetByResetToken(ctx context.Context, token string) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"password_reset_token": token}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

// Update updates a user in the database
func (r *MongoUserRepository) Update(ctx context.Context, user *models.User) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": user.ID},
		bson.M{"$set": user},
	)
	return err
}

// Delete removes a user from the database
func (r *MongoUserRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// List retrieves a paginated list of users with optional filters
func (r *MongoUserRepository) List(ctx context.Context, page, pageSize int, filters repository.UserFilters) ([]*models.User, int64, error) {
	// Build filter
	filter := bson.M{}

	if filters.Status != "" {
		filter["status"] = filters.Status
	}

	if filters.Search != "" {
		filter["$or"] = []bson.M{
			{"first_name": bson.M{"$regex": filters.Search, "$options": "i"}},
			{"last_name": bson.M{"$regex": filters.Search, "$options": "i"}},
			{"email": bson.M{"$regex": filters.Search, "$options": "i"}},
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

	// Find users
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

	var users []*models.User
	for cursor.Next(ctx) {
		var user models.User
		if err := cursor.Decode(&user); err != nil {
			return nil, 0, err
		}
		users = append(users, &user)
	}

	if err := cursor.Err(); err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// AddWeddingID adds a wedding ID to the user's wedding_ids array
func (r *MongoUserRepository) AddWeddingID(ctx context.Context, userID, weddingID primitive.ObjectID) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": userID},
		bson.M{"$addToSet": bson.M{"wedding_ids": weddingID}},
	)
	return err
}

// RemoveWeddingID removes a wedding ID from the user's wedding_ids array
func (r *MongoUserRepository) RemoveWeddingID(ctx context.Context, userID, weddingID primitive.ObjectID) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": userID},
		bson.M{"$pull": bson.M{"wedding_ids": weddingID}},
	)
	return err
}

// UpdateLastLogin updates the user's last login time
func (r *MongoUserRepository) UpdateLastLogin(ctx context.Context, userID primitive.ObjectID) error {
	now := primitive.NewDateTimeFromTime(time.Now())
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": userID},
		bson.M{"$set": bson.M{"last_login_at": now}},
	)
	return err
}

// SetEmailVerified marks a user's email as verified
func (r *MongoUserRepository) SetEmailVerified(ctx context.Context, userID primitive.ObjectID) error {
	now := primitive.NewDateTimeFromTime(time.Now())
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": userID},
		bson.M{
			"$set": bson.M{
				"email_verified":           true,
				"email_verified_at":        now,
				"email_verification_token": "",
			},
		},
	)
	return err
}

// UpdatePassword updates a user's password
func (r *MongoUserRepository) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": userID},
		bson.M{
			"$set": bson.M{
				"password_hash": passwordHash,
				"updated_at":    primitive.NewDateTimeFromTime(time.Now()),
			},
		},
	)
	return err
}

// FindByEmail finds users by email with partial matching
func (r *MongoUserRepository) FindByEmail(ctx context.Context, email string) ([]*models.User, error) {
	filter := bson.M{"email": bson.M{"$regex": email, "$options": "i"}}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []*models.User
	for cursor.Next(ctx) {
		var user models.User
		if err := cursor.Decode(&user); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	return users, cursor.Err()
}

// CountByStatus returns the count of users by status
func (r *MongoUserRepository) CountByStatus(ctx context.Context, status string) (int64, error) {
	filter := bson.M{"status": status}
	return r.collection.CountDocuments(ctx, filter)
}

// FindInactiveUsers finds users who haven't logged in for a certain period
func (r *MongoUserRepository) FindInactiveUsers(ctx context.Context, since time.Time) ([]*models.User, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"last_login_at": bson.M{"$lt": primitive.NewDateTimeFromTime(since)}},
			{"last_login_at": bson.M{"$exists": false}},
		},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []*models.User
	for cursor.Next(ctx) {
		var user models.User
		if err := cursor.Decode(&user); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	return users, cursor.Err()
}
