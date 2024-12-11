package Mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

var mongoDBInstance *mongo.Client = nil

func ConnectToMongoDB() {
	uri := "mongodb+srv://admin:admin@cluster0.td7uetr.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0"

	// Create a new MongoDB instance with the provided URI
	newMongoInstance, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		log.Printf("(ConnectToMongoDB) There was an error creating the mongoDB instance: %v", err)
		return
	}

	log.Printf("(ConnectToMongoDB) Successfully Connected to MongoDB")
	mongoDBInstance = newMongoInstance
}

func GetMongoDB() *mongo.Client {
	if mongoDBInstance == nil {
		ConnectToMongoDB()
	}

	return mongoDBInstance
}

func GetCollection(collection string) *mongo.Collection {
	return GetMongoDB().Database("Pametni-Paketnik-baza").Collection(collection) //Change this
}
