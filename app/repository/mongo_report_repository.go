package repository

import (
	"context"
	"errors"

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

type TypeCount struct {
	Type  string `bson:"_id" json:"type"`
	Count int    `bson:"count" json:"count"`
}

type KeyCount struct {
	Key   string `bson:"_id" json:"key"`
	Count int    `bson:"count" json:"count"`
}

type StudentPoints struct {
	StudentID string `bson:"_id" json:"studentId"`
	Points    int    `bson:"points" json:"points"`
}

func (r *MongoReportRepository) GetTotalByType(ctx context.Context, studentIDs []string) (map[string]int, error) {
	match := bson.D{}
	if len(studentIDs) > 0 {
		match = bson.D{{Key: "$match", Value: bson.D{{Key: "studentId", Value: bson.D{{Key: "$in", Value: studentIDs}}}}}}
	}

	pipeline := mongo.Pipeline{}
	if len(match) > 0 {
		pipeline = append(pipeline, match)
	}
	pipeline = append(pipeline,
		bson.D{{Key: "$group", Value: bson.D{{Key: "_id", Value: "$achievementType"}, {Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}}}}},
	)

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	out := map[string]int{}
	for cursor.Next(ctx) {
		var tc TypeCount
		if err := cursor.Decode(&tc); err != nil {
			return nil, err
		}
		out[tc.Type] = tc.Count
	}
	return out, nil
}

// GetCompetitionLevelDistribution groups by details.competitionLevel (if exists)
func (r *MongoReportRepository) GetCompetitionLevelDistribution(ctx context.Context, studentIDs []string) (map[string]int, error) {
	match := bson.D{}
	if len(studentIDs) > 0 {
		match = bson.D{{Key: "$match", Value: bson.D{{Key: "studentId", Value: bson.D{{Key: "$in", Value: studentIDs}}}}}}
	}

	pipeline := mongo.Pipeline{}
	if len(match) > 0 {
		pipeline = append(pipeline, match)
	}
	// group by details.competitionLevel
	pipeline = append(pipeline,
		bson.D{{Key: "$group", Value: bson.D{{Key: "_id", Value: "$details.competitionLevel"}, {Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}}}}},
	)

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	out := map[string]int{}
	for cursor.Next(ctx) {
		var kc KeyCount
		if err := cursor.Decode(&kc); err != nil {
			return nil, err
		}
		key := kc.Key
		if key == "" {
			key = "unknown"
		}
		out[key] = kc.Count
	}
	return out, nil
}

// GetTopStudentsByPoints returns top N students ordered by total points (from achievements.points)
func (r *MongoReportRepository) GetTopStudentsByPoints(ctx context.Context, limit int64, studentIDs []string) ([]StudentPoints, error) {
	match := bson.D{}
	if len(studentIDs) > 0 {
		match = bson.D{{Key: "$match", Value: bson.D{{Key: "studentId", Value: bson.D{{Key: "$in", Value: studentIDs}}}}}}
	}

	pipeline := mongo.Pipeline{}
	if len(match) > 0 {
		pipeline = append(pipeline, match)
	}
	pipeline = append(pipeline,
		bson.D{{Key: "$group", Value: bson.D{{Key: "_id", Value: "$studentId"}, {Key: "points", Value: bson.D{{Key: "$sum", Value: "$points"}}}}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "points", Value: -1}}}},
		bson.D{{Key: "$limit", Value: limit}},
	)

	opts := options.Aggregate()
	cursor, err := r.collection.Aggregate(ctx, pipeline, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	out := []StudentPoints{}
	for cursor.Next(ctx) {
		var sp StudentPoints
		if err := cursor.Decode(&sp); err != nil {
			return nil, err
		}
		out = append(out, sp)
	}
	return out, nil
}

// GetMonthlyCounts returns map["YYYY-MM"] = count for N months back (including current)
func (r *MongoReportRepository) GetMonthlyCounts(ctx context.Context, months int, studentIDs []string) (map[string]int, error) {
	if months <= 0 {
		return nil, errors.New("months must be > 0")
	}

	match := bson.D{}
	if len(studentIDs) > 0 {
		match = bson.D{{Key: "$match", Value: bson.D{{Key: "studentId", Value: bson.D{{Key: "$in", Value: studentIDs}}}}}}
	}

	// project year-month from createdAt
	pipeline := mongo.Pipeline{}
	if len(match) > 0 {
		pipeline = append(pipeline, match)
	}
	pipeline = append(pipeline,
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "ym", Value: bson.D{{Key: "$dateToString", Value: bson.D{{Key: "format", Value: "%Y-%m"}, {Key: "date", Value: "$createdAt"}}}}},
		}}},
		bson.D{{Key: "$group", Value: bson.D{{Key: "_id", Value: "$ym"}, {Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}}}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "_id", Value: 1}}}},
	)

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	out := map[string]int{}
	for cursor.Next(ctx) {
		var kc KeyCount
		if err := cursor.Decode(&kc); err != nil {
			return nil, err
		}
		out[kc.Key] = kc.Count
	}
	return out, nil
}
