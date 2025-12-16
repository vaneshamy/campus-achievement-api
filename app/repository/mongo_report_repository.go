package repository

import (
	"context"
	"errors"

	"go-fiber/app/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoReportRepository handles aggregation queries on achievements collection
type MongoReportRepository struct {
	collection *mongo.Collection
}

func NewMongoReportRepository(coll *mongo.Collection) *MongoReportRepository {
	return &MongoReportRepository{collection: coll}
}

func (r *MongoReportRepository) GetTotalByType(
	ctx context.Context,
	studentIDs []string,
) (map[string]int, error) {

	pipeline := mongo.Pipeline{}

	if len(studentIDs) > 0 {
		pipeline = append(pipeline,
			bson.D{{Key: "$match", Value: bson.D{
				{Key: "studentId", Value: bson.D{{Key: "$in", Value: studentIDs}}},
			}}},
		)
	}

	pipeline = append(pipeline,
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$achievementType"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
	)

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	result := map[string]int{}
	for cursor.Next(ctx) {
		var row model.TypeCount
		if err := cursor.Decode(&row); err != nil {
			return nil, err
		}
		result[row.Type] = row.Count
	}

	return result, nil
}

func (r *MongoReportRepository) GetCompetitionLevelDistribution(
	ctx context.Context,
	studentIDs []string,
) (map[string]int, error) {

	pipeline := mongo.Pipeline{}

	if len(studentIDs) > 0 {
		pipeline = append(pipeline,
			bson.D{{Key: "$match", Value: bson.D{
				{Key: "studentId", Value: bson.D{{Key: "$in", Value: studentIDs}}},
			}}},
		)
	}

	pipeline = append(pipeline,
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$details.competitionLevel"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
	)

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	result := map[string]int{}
	for cursor.Next(ctx) {
		var row model.KeyCount
		if err := cursor.Decode(&row); err != nil {
			return nil, err
		}

		key := row.Key
		if key == "" {
			key = "unknown"
		}
		result[key] = row.Count
	}

	return result, nil
}

func (r *MongoReportRepository) GetTopStudentsByPoints(
	ctx context.Context,
	limit int64,
	studentIDs []string,
) ([]model.StudentPoints, error) {

	pipeline := mongo.Pipeline{}

	if len(studentIDs) > 0 {
		pipeline = append(pipeline,
			bson.D{{Key: "$match", Value: bson.D{
				{Key: "studentId", Value: bson.D{{Key: "$in", Value: studentIDs}}},
			}}},
		)
	}

	pipeline = append(pipeline,
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$studentId"},
			{Key: "points", Value: bson.D{{Key: "$sum", Value: "$points"}}},
		}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "points", Value: -1}}}},
		bson.D{{Key: "$limit", Value: limit}},
	)

	cursor, err := r.collection.Aggregate(ctx, pipeline, options.Aggregate())
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var result []model.StudentPoints
	for cursor.Next(ctx) {
		var row model.StudentPoints
		if err := cursor.Decode(&row); err != nil {
			return nil, err
		}
		result = append(result, row)
	}

	return result, nil
}

func (r *MongoReportRepository) GetMonthlyCounts(
	ctx context.Context,
	months int,
	studentIDs []string,
) (map[string]int, error) {

	if months <= 0 {
		return nil, errors.New("months must be > 0")
	}

	pipeline := mongo.Pipeline{}

	if len(studentIDs) > 0 {
		pipeline = append(pipeline,
			bson.D{{Key: "$match", Value: bson.D{
				{Key: "studentId", Value: bson.D{{Key: "$in", Value: studentIDs}}},
			}}},
		)
	}

	pipeline = append(pipeline,
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "ym", Value: bson.D{{Key: "$dateToString", Value: bson.D{
				{Key: "format", Value: "%Y-%m"},
				{Key: "date", Value: "$createdAt"},
			}}}},
		}}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$ym"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "_id", Value: 1}}}},
	)

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	result := map[string]int{}
	for cursor.Next(ctx) {
		var row model.KeyCount
		if err := cursor.Decode(&row); err != nil {
			return nil, err
		}
		result[row.Key] = row.Count
	}

	return result, nil
}
