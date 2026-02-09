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

type BillRepository struct {
	collection *mongo.Collection
}

func NewBillRepository(db *database.MongoDB) *BillRepository {
	return &BillRepository{
		collection: db.Collection(database.CollectionBills),
	}
}

func (r *BillRepository) Create(ctx context.Context, bill *models.Bill) error {
	bill.CreatedAt = time.Now()
	bill.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, bill)
	if err != nil {
		return err
	}

	bill.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *BillRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Bill, error) {
	var bill models.Bill
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&bill)
	if err != nil {
		return nil, err
	}
	return &bill, nil
}

func (r *BillRepository) FindByGroupID(ctx context.Context, groupID primitive.ObjectID) ([]models.Bill, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.collection.Find(ctx, bson.M{"group_id": groupID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var bills []models.Bill
	if err := cursor.All(ctx, &bills); err != nil {
		return nil, err
	}
	return bills, nil
}

func (r *BillRepository) FindActiveByGroupID(ctx context.Context, groupID primitive.ObjectID) ([]models.Bill, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.collection.Find(ctx, bson.M{
		"group_id": groupID,
		"status":   bson.M{"$ne": models.BillCancelled},
	}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var bills []models.Bill
	if err := cursor.All(ctx, &bills); err != nil {
		return nil, err
	}
	return bills, nil
}

func (r *BillRepository) Update(ctx context.Context, bill *models.Bill) error {
	bill.UpdatedAt = time.Now()
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": bill.ID},
		bson.M{"$set": bill},
	)
	return err
}

func (r *BillRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{
			"status":     models.BillCancelled,
			"updated_at": time.Now(),
		}},
	)
	return err
}
