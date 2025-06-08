package database

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"

	// "go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func connection_database() *mongo.Client {

	if err := godotenv.Load(); err != nil {
		log.Println("Error loading .env file:", err)
		// return
	}

	uri := os.Getenv("MONGODB_URI")

	if uri == "" {
		log.Fatal("MONGODB_URI environment variable is not set")
		// return
	}

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}

	// Ping the database to verify connection
	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal("Failed to ping MongoDB:", err)
	}

	// result ,err :=client.ListDatabaseNames(context.Background(),bson.D{})

	// if err!=nil{
	// 	log.Fatal(err)
	// }
	// for _,db:= range result{
	// 	fmt.Println(db)
	// }

	fmt.Println("Successfully connected to MongoDB")

	return client

}

var Client *mongo.Client = connection_database()

func OpenCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	if err := godotenv.Load(); err != nil {
		log.Println("Error loading .env file:", err)
		// return
	}
	cluster_name := os.Getenv("MONGOCLUSTER")
	if cluster_name == "" {
		cluster_name = "cluster0"
	}
	//
	var collection *mongo.Collection = client.Database(cluster_name).Collection(collectionName)

	return collection
}
