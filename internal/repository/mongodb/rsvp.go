package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"wedding-invitation-backend/internal/domain/models"
	"wedding-invitation-backend/internal/domain/repository"
)

type mongoRSVPRepository struct {
	collection *mongo.Collection
}

func NewMongoRSVPRepository(db *mongo.Database) repository.RSVPRepository {
	return &mongoRSVPRepository{
		collection: db.Collection("rsvps"),
	}
}

func (r *mongoRSVPRepository) Create(ctx context.Context, rsvp *models.RSVP) error {
	_, err := r.collection.InsertOne(ctx, rsvp)
	return err
}

func (r *mongoRSVPRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.RSVP, error) {
	var rsvp models.RSVP
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&rsvp)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return &rsvp, nil
}

func (r *mongoRSVPRepository) GetByEmail(ctx context.Context, weddingID primitive.ObjectID, email string) (*models.RSVP, error) {
	var rsvp models.RSVP
	err := r.collection.FindOne(ctx, bson.M{
		"wedding_id": weddingID,
		"email":      email,
	}).Decode(&rsvp)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return &rsvp, nil
}

func (r *mongoRSVPRepository) ListByWedding(ctx context.Context, weddingID primitive.ObjectID, page, pageSize int, filters repository.RSVPFilters) ([]*models.RSVP, int64, error) {
	filter := bson.M{"wedding_id": weddingID}

	// Apply filters
	if filters.Status != "" {
		filter["status"] = filters.Status
	}
	if filters.Source != "" {
		filter["source"] = filters.Source
	}
	if filters.Search != "" {
		filter["$or"] = []bson.M{
			{"first_name": bson.M{"$regex": filters.Search, "$options": "i"}},
			{"last_name": bson.M{"$regex": filters.Search, "$options": "i"}},
			{"email": bson.M{"$regex": filters.Search, "$options": "i"}},
		}
	}
	if filters.SubmittedAfter != nil {
		filter["submitted_at"] = bson.M{"$gte": filters.SubmittedAfter}
	}
	if filters.SubmittedBefore != nil {
		filter["submitted_at"] = bson.M{"$lte": filters.SubmittedBefore}
	}

	// Count total documents
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Pagination
	skip := (page - 1) * pageSize
	opts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(pageSize)).
		SetSort(bson.D{{"submitted_at", -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var rsvps []*models.RSVP
	for cursor.Next(ctx) {
		var rsvp models.RSVP
		if err := cursor.Decode(&rsvp); err != nil {
			return nil, 0, err
		}
		rsvps = append(rsvps, &rsvp)
	}

	return rsvps, total, nil
}

func (r *mongoRSVPRepository) Update(ctx context.Context, rsvp *models.RSVP) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": rsvp.ID},
		bson.M{"$set": rsvp},
	)
	return err
}

func (r *mongoRSVPRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *mongoRSVPRepository) GetStatistics(ctx context.Context, weddingID primitive.ObjectID) (*models.RSVPStatistics, error) {
	// Match stage
	matchStage := bson.D{{"$match", bson.D{{"wedding_id", weddingID}}}}

	// Group by status to get counts
	groupStage := bson.D{
		{"$group", bson.D{
			{"_id", "$status"},
			{"count", bson.D{{"$sum", 1}}},
			{"totalGuests", bson.D{{"$sum", "$attendance_count"}}},
			{"plusOnes", bson.D{{"$sum", "$plus_one_count"}}},
		}},
	}

	cursor, err := r.collection.Aggregate(ctx, mongo.Pipeline{matchStage, groupStage})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	stats := &models.RSVPStatistics{
		DietaryCounts:   make(map[string]int),
		SubmissionTrend: []models.DailyCount{},
	}

	for cursor.Next(ctx) {
		var result struct {
			ID          string `bson:"_id"`
			Count       int    `bson:"count"`
			TotalGuests int    `bson:"totalGuests"`
			PlusOnes    int    `bson:"plusOnes"`
		}
		if err := cursor.Decode(&result); err != nil {
			continue
		}

		switch result.ID {
		case "attending":
			stats.Attending = result.Count
		case "not-attending":
			stats.NotAttending = result.Count
		case "maybe":
			stats.Maybe = result.Count
		}

		stats.TotalGuests += result.TotalGuests
		stats.PlusOnesCount += result.PlusOnes
		stats.TotalResponses += result.Count
	}

	// Get dietary counts
	dietaryPipeline := mongo.Pipeline{
		matchStage,
		bson.D{{"$unwind", bson.D{{"path", "$dietary_selected"}, {"preserveNullAndEmptyArrays", false}}}},
		bson.D{
			{"$group", bson.D{
				{"_id", "$dietary_selected"},
				{"count", bson.D{{"$sum", 1}}},
			}},
		},
	}

	dietaryCursor, err := r.collection.Aggregate(ctx, dietaryPipeline)
	if err == nil {
		defer dietaryCursor.Close(ctx)
		for dietaryCursor.Next(ctx) {
			var result struct {
				ID    string `bson:"_id"`
				Count int    `bson:"count"`
			}
			if err := dietaryCursor.Decode(&result); err == nil && result.ID != "" {
				stats.DietaryCounts[result.ID] = result.Count
			}
		}
	}

	// Get submission trend for last 30 days
	stats.SubmissionTrend, _ = r.GetSubmissionTrend(ctx, weddingID, 30)

	return stats, nil
}

func (r *mongoRSVPRepository) MarkConfirmationSent(ctx context.Context, id primitive.ObjectID) error {
	now := time.Now()
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{
			"$set": bson.M{
				"confirmation_sent":    true,
				"confirmation_sent_at": now,
			},
		},
	)
	return err
}

func (r *mongoRSVPRepository) GetSubmissionTrend(ctx context.Context, weddingID primitive.ObjectID, days int) ([]models.DailyCount, error) {
	startDate := time.Now().AddDate(0, 0, -days).Truncate(24 * time.Hour)

	matchStage := bson.D{
		{"$match", bson.D{
			{"wedding_id", weddingID},
			{"submitted_at", bson.D{{"$gte", startDate}}},
		}},
	}

	groupStage := bson.D{
		{"$group", bson.D{
			{"_id", bson.D{{"$dateToString", bson.D{{"format", "%Y-%m-%d"}, {"date", "$submitted_at"}}}}},
			{"count", bson.D{{"$sum", 1}}},
		}},
	}

	sortStage := bson.D{{"$sort", bson.D{{"_id", 1}}}}

	cursor, err := r.collection.Aggregate(ctx, mongo.Pipeline{matchStage, groupStage, sortStage})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		Date  string `bson:"_id"`
		Count int    `bson:"count"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	// Fill in missing days with zero count
	trend := make([]models.DailyCount, days)
	now := time.Now().Truncate(24 * time.Hour)

	for i := 0; i < days; i++ {
		date := now.AddDate(0, 0, -days+1+i).Format("2006-01-02")
		trend[i] = models.DailyCount{Date: date, Count: 0}
	}

	// Fill in actual counts
	for _, result := range results {
		for i, day := range trend {
			if day.Date == result.Date {
				trend[i].Count = result.Count
				break
			}
		}
	}

	return trend, nil
}
