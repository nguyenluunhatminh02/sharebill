package repository

import (
	"context"
	"time"

	"github.com/splitbill/backend/internal/database"
	"github.com/splitbill/backend/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ActivityRepository struct {
	collection *mongo.Collection
}

func NewActivityRepository(db *database.MongoDB) *ActivityRepository {
	return &ActivityRepository{
		collection: db.Collection(database.CollectionActivities),
	}
}

func (r *ActivityRepository) Create(ctx context.Context, activity *models.Activity) error {
	activity.CreatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, activity)
	if err != nil {
		return err
	}

	activity.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *ActivityRepository) FindByGroupID(ctx context.Context, groupID primitive.ObjectID, limit int64) ([]models.Activity, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(limit)

	cursor, err := r.collection.Find(ctx, bson.M{"group_id": groupID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var activities []models.Activity
	if err := cursor.All(ctx, &activities); err != nil {
		return nil, err
	}
	return activities, nil
}

func (r *ActivityRepository) FindByUserGroups(ctx context.Context, groupIDs []primitive.ObjectID, limit int64) ([]models.Activity, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(limit)

	cursor, err := r.collection.Find(ctx, bson.M{
		"group_id": bson.M{"$in": groupIDs},
	}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var activities []models.Activity
	if err := cursor.All(ctx, &activities); err != nil {
		return nil, err
	}
	return activities, nil
}

func (r *ActivityRepository) FindByUserID(ctx context.Context, userID primitive.ObjectID, limit int64) ([]models.Activity, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(limit)

	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var activities []models.Activity
	if err := cursor.All(ctx, &activities); err != nil {
		return nil, err
	}
	return activities, nil
}
