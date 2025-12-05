package config

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func NewMongoClient() (*mongo.Client, error) {
	uri := GetEnv("MONGO_URI", "mongodb://localhost:27017")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}
	log.Println("Connected to MongoDB:", uri)
	return client, nil
}

func GetMongoDatabase(client *mongo.Client) *mongo.Database {
	dbName := GetEnv("MONGO_DB", "prestasi")
	return client.Database(dbName)
}
