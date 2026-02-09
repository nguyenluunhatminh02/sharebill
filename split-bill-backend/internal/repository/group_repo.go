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

type GroupRepository struct {
	collection *mongo.Collection
}

func NewGroupRepository(db *database.MongoDB) *GroupRepository {
	return &GroupRepository{
		collection: db.Collection(database.CollectionGroups),
	}
}

func (r *GroupRepository) Create(ctx context.Context, group *models.Group) error {
	group.CreatedAt = time.Now()
	group.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, group)
	if err != nil {
		return err
	}

	group.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *GroupRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Group, error) {
	var group models.Group
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&group)
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *GroupRepository) FindByInviteCode(ctx context.Context, code string) (*models.Group, error) {
	var group models.Group
	err := r.collection.FindOne(ctx, bson.M{"invite_code": code, "is_active": true}).Decode(&group)
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *GroupRepository) FindByMemberUserID(ctx context.Context, userID primitive.ObjectID) ([]models.Group, error) {
	opts := options.Find().SetSort(bson.D{{Key: "updated_at", Value: -1}})
	cursor, err := r.collection.Find(ctx, bson.M{
		"members.user_id": userID,
		"is_active":       true,
	}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var groups []models.Group
	if err := cursor.All(ctx, &groups); err != nil {
		return nil, err
	}
	return groups, nil
}

func (r *GroupRepository) Update(ctx context.Context, group *models.Group) error {
	group.UpdatedAt = time.Now()
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": group.ID},
		bson.M{"$set": group},
	)
	return err
}

func (r *GroupRepository) AddMember(ctx context.Context, groupID primitive.ObjectID, member models.GroupMember) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": groupID},
		bson.M{
			"$push": bson.M{"members": member},
			"$set":  bson.M{"updated_at": time.Now()},
		},
	)
	return err
}

func (r *GroupRepository) RemoveMember(ctx context.Context, groupID, userID primitive.ObjectID) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": groupID},
		bson.M{
			"$pull": bson.M{"members": bson.M{"user_id": userID}},
			"$set":  bson.M{"updated_at": time.Now()},
		},
	)
	return err
}

func (r *GroupRepository) IsMember(ctx context.Context, groupID, userID primitive.ObjectID) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{
		"_id":             groupID,
		"members.user_id": userID,
	})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *GroupRepository) Delete(ctx context.Context, groupID primitive.ObjectID) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": groupID},
		bson.M{"$set": bson.M{"is_active": false, "updated_at": time.Now()}},
	)
	return err
}
