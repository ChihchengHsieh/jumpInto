package database

import (
	"context"
	"fmt"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DB - MongoDB database
var DB *mongo.Database

// InitDB - Initialise the database for MongoDB
func InitDB() {
	// Setting autoload in the main funciton to get the environment variables
	clientOptions := options.Client().ApplyURI(os.Getenv("MGDB_APIKEY"))
	client, err := mongo.Connect(context.TODO(), clientOptions)

	if err != nil {
		log.Fatal(err)
	}

	// Checking
	err = client.Ping(context.TODO(), nil)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB successfully.")

	DB = client.Database("jumpInto")
}
