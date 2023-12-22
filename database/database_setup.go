package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DBSet creates a new MongoDB client and connects to the database
func DBSet() *mongo.Client {
	// create a new MongoDB client
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost27017"))
	if err != nil {
		log.Fatal(err)
	}

	// create a context with a timeout of 10 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// attempt to connect to the MongoDB database
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// attempt to ping the MongoDB server
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Println("failed to connect on mongoDB. Try later")
		return nil
	}

	// log a success message
	fmt.Println("Successfully connected to mongoDB")
	return client
}

var Client *mongo.Client = DBSet()

func UserData(client *mongo.Client, collectionName string) *mongo.Collection {
	var collection *mongo.Collection = client.Database("EcommerceDB").Collection(collectionName)
	return collection
}

func ProductData(client *mongo.Client, collectionName string) *mongo.Collection {
	var productCollection *mongo.Collection = client.Database("EcommerceDB").Collection(collectionName)
	return productCollection
}
