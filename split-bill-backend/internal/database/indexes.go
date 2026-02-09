package database

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// EnsureIndexes creates all necessary MongoDB indexes for optimal query performance.
// This should be called once during application startup.
func EnsureIndexes(db *MongoDB) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Println("📇 Creating MongoDB indexes...")

	// Users collection indexes
	createIndexes(ctx, db.Collection("users"), []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "firebase_uid", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("idx_users_firebase_uid"),
		},
		{
			Keys:    bson.D{{Key: "phone", Value: 1}},
			Options: options.Index().SetSparse(true).SetName("idx_users_phone"),
		},
		{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetSparse(true).SetName("idx_users_email"),
		},
	})

	// Groups collection indexes
	createIndexes(ctx, db.Collection("groups"), []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "invite_code", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("idx_groups_invite_code"),
		},
		{
			Keys:    bson.D{{Key: "members.user_id", Value: 1}},
			Options: options.Index().SetName("idx_groups_member_user_id"),
		},
		{
			Keys:    bson.D{{Key: "created_by", Value: 1}},
			Options: options.Index().SetName("idx_groups_created_by"),
		},
		{
			Keys:    bson.D{{Key: "status", Value: 1}},
			Options: options.Index().SetName("idx_groups_status"),
		},
	})

	// Bills collection indexes
	createIndexes(ctx, db.Collection("bills"), []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "group_id", Value: 1}, {Key: "created_at", Value: -1}},
			Options: options.Index().SetName("idx_bills_group_id_created_at"),
		},
		{
			Keys:    bson.D{{Key: "group_id", Value: 1}, {Key: "status", Value: 1}},
			Options: options.Index().SetName("idx_bills_group_id_status"),
		},
		{
			Keys:    bson.D{{Key: "created_by", Value: 1}},
			Options: options.Index().SetName("idx_bills_created_by"),
		},
		{
			Keys:    bson.D{{Key: "category", Value: 1}},
			Options: options.Index().SetName("idx_bills_category"),
		},
	})

	// Transactions collection indexes
	createIndexes(ctx, db.Collection("transactions"), []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "group_id", Value: 1}, {Key: "created_at", Value: -1}},
			Options: options.Index().SetName("idx_transactions_group_id_created_at"),
		},
		{
			Keys:    bson.D{{Key: "from_user", Value: 1}},
			Options: options.Index().SetName("idx_transactions_from_user"),
		},
		{
			Keys:    bson.D{{Key: "to_user", Value: 1}},
			Options: options.Index().SetName("idx_transactions_to_user"),
		},
		{
			Keys:    bson.D{{Key: "group_id", Value: 1}, {Key: "status", Value: 1}},
			Options: options.Index().SetName("idx_transactions_group_id_status"),
		},
	})

	// Activities collection indexes
	createIndexes(ctx, db.Collection("activities"), []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "group_id", Value: 1}, {Key: "created_at", Value: -1}},
			Options: options.Index().SetName("idx_activities_group_id_created_at"),
		},
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}, {Key: "created_at", Value: -1}},
			Options: options.Index().SetName("idx_activities_user_id_created_at"),
		},
	})

	// OCR results collection indexes
	createIndexes(ctx, db.Collection("ocr_results"), []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}, {Key: "status", Value: 1}},
			Options: options.Index().SetName("idx_ocr_results_user_id_status"),
		},
		{
			Keys:    bson.D{{Key: "group_id", Value: 1}, {Key: "created_at", Value: -1}},
			Options: options.Index().SetName("idx_ocr_results_group_id_created_at"),
		},
	})

	log.Println("✅ MongoDB indexes created successfully")
}

func createIndexes(ctx context.Context, collection *mongo.Collection, indexes []mongo.IndexModel) {
	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		log.Printf("⚠️  Warning: Failed to create indexes for %s: %v", collection.Name(), err)
	}
}
