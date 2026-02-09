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

type TransactionRepository struct {
	collection *mongo.Collection
}

func NewTransactionRepository(db *database.MongoDB) *TransactionRepository {
	return &TransactionRepository{
		collection: db.Collection(database.CollectionTransactions),
	}
}

func (r *TransactionRepository) Create(ctx context.Context, tx *models.Transaction) error {
	tx.CreatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, tx)
	if err != nil {
		return err
	}

	tx.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *TransactionRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Transaction, error) {
	var tx models.Transaction
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&tx)
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

func (r *TransactionRepository) FindByGroupID(ctx context.Context, groupID primitive.ObjectID) ([]models.Transaction, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.collection.Find(ctx, bson.M{"group_id": groupID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var transactions []models.Transaction
	if err := cursor.All(ctx, &transactions); err != nil {
		return nil, err
	}
	return transactions, nil
}

func (r *TransactionRepository) FindByUser(ctx context.Context, userID primitive.ObjectID) ([]models.Transaction, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.collection.Find(ctx, bson.M{
		"$or": []bson.M{
			{"from_user": userID},
			{"to_user": userID},
		},
	}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var transactions []models.Transaction
	if err := cursor.All(ctx, &transactions); err != nil {
		return nil, err
	}
	return transactions, nil
}

func (r *TransactionRepository) FindConfirmedByGroupID(ctx context.Context, groupID primitive.ObjectID) ([]models.Transaction, error) {
	cursor, err := r.collection.Find(ctx, bson.M{
		"group_id": groupID,
		"status":   models.TransactionConfirmed,
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var transactions []models.Transaction
	if err := cursor.All(ctx, &transactions); err != nil {
		return nil, err
	}
	return transactions, nil
}

func (r *TransactionRepository) Confirm(ctx context.Context, id primitive.ObjectID) error {
	now := time.Now()
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{
			"status":       models.TransactionConfirmed,
			"confirmed_at": now,
		}},
	)
	return err
}

func (r *TransactionRepository) Reject(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"status": models.TransactionRejected}},
	)
	return err
}
