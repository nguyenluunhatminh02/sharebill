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

type OCRRepository struct {
	collection *mongo.Collection
}

func NewOCRRepository(db *database.MongoDB) *OCRRepository {
	return &OCRRepository{
		collection: db.Collection(database.CollectionOCRResults),
	}
}

func (r *OCRRepository) Create(ctx context.Context, result *models.OCRResult) error {
	result.CreatedAt = time.Now()
	res, err := r.collection.InsertOne(ctx, result)
	if err != nil {
		return err
	}
	result.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *OCRRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.OCRResult, error) {
	var result models.OCRResult
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *OCRRepository) Update(ctx context.Context, result *models.OCRResult) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": result.ID},
		bson.M{"$set": result},
	)
	return err
}

func (r *OCRRepository) UpdateStatus(ctx context.Context, id primitive.ObjectID, status models.OCRStatus) error {
	update := bson.M{
		"$set": bson.M{
			"status": status,
		},
	}
	if status == models.OCRStatusConfirmed {
		now := time.Now()
		update["$set"].(bson.M)["confirmed_at"] = &now
	}
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	return err
}

func (r *OCRRepository) SetBillID(ctx context.Context, ocrID, billID primitive.ObjectID) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": ocrID},
		bson.M{"$set": bson.M{"bill_id": billID}},
	)
	return err
}

func (r *OCRRepository) FindByGroupID(ctx context.Context, groupID primitive.ObjectID, limit int64) ([]models.OCRResult, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(limit)

	cursor, err := r.collection.Find(ctx, bson.M{"group_id": groupID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []models.OCRResult
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func (r *OCRRepository) FindPendingByUser(ctx context.Context, userID primitive.ObjectID) ([]models.OCRResult, error) {
	filter := bson.M{
		"uploaded_by": userID,
		"status":      bson.M{"$in": []models.OCRStatus{models.OCRStatusCompleted, models.OCRStatusProcessing}},
	}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []models.OCRResult
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}
