package repository

import (
	"context"
	"time"

	"go-fiber/app/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoAchievementRepository struct {
	collection *mongo.Collection
}

func NewMongoAchievementRepository(coll *mongo.Collection) *MongoAchievementRepository {
	return &MongoAchievementRepository{collection: coll}
}

func (r *MongoAchievementRepository) CreateAchievement(ctx context.Context, a *model.Achievement) (primitive.ObjectID, error) {
	now := time.Now()
	a.CreatedAt = now
	a.UpdatedAt = now

	res, err := r.collection.InsertOne(ctx, a)
	if err != nil {
		return primitive.NilObjectID, err
	}
	id := res.InsertedID.(primitive.ObjectID)
	return id, nil
}

func (r *MongoAchievementRepository) UpdateAchievement(ctx context.Context, id primitive.ObjectID, update bson.M) error {
	update["updatedAt"] = time.Now()
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": update})
	return err
}

func (r *MongoAchievementRepository) DeleteAchievement(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *MongoAchievementRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*model.Achievement, error) {
	var a model.Achievement
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&a)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *MongoAchievementRepository) AddAttachment(
    ctx context.Context,
    id primitive.ObjectID,
    att model.Attachment,
) error {
    _, err := r.collection.UpdateOne(
        ctx,
        bson.M{"_id": id},
        bson.M{
            "$push": bson.M{"attachments": att},
        },
    )
    return err
}
