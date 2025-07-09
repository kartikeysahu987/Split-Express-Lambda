package database

import (
	"context"
	"time"
	// "fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"

	// "go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client

func init() {
	// Load .env file (only works locally; on Lambda, env vars should be in console)
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ .env file not found, assuming AWS environment variables are set")
	}

	uri := os.Getenv("MONGODB_URI")
	// if uri == "" {
	// 	log.Fatal("❌ MONGODB_URI not set")
	// }

	// Connect with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("❌ MongoDB connection failed: %v", err)
	}

	// Ping to ensure connection
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("❌ MongoDB ping failed: %v", err)
	}

	Client = client
	log.Println("✅ MongoDB connected successfully")
}

// var Client *mongo.Client = connection_database()

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
