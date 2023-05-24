package util

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

const connectionString = "mongodb+srv://mareklescinsky:EUFZTs6jcdkqqEUk@brainboard.lrpc8h3.mongodb.net/test"

const (
	dbName            = "BrainBoard"
	userCollection    = "user"
	taskCollection    = "task"
	subjectCollection = "subject"
)

func getObjectsFromCollection(collectionName string) {
	// Set up the MongoDB client
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(connectionString))
	if err != nil {
		log.Fatal(err)
	}

	// Ping the MongoDB server to check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}

	// Access the desired database and collection
	db := client.Database(dbName)
	collection := db.Collection(collectionName)

	// Retrieve objects from the collection
	cursor, err := collection.Find(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}

	// Iterate over the cursor and process each document
	for cursor.Next(context.TODO()) {
		var result bson.M
		err := cursor.Decode(&result)
		if err != nil {
			log.Fatal(err)
		}

		// Process the document as needed
		fmt.Println(result)
	}

	// Check if any error occurred during iteration
	if err := cursor.Err(); err != nil {
		log.Fatal(err)
	}

	// Close the cursor
	cursor.Close(context.TODO())

	// Disconnect the client
	err = client.Disconnect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Objects retrieved successfully.")
}

// GetUsers retrieves all users from the database
func GetUsers() {
	getObjectsFromCollection(userCollection)

}

// GetTasks retrieves all tasks from the database
func GetTasks() {
	getObjectsFromCollection(taskCollection)
}

// GetSubjects retrieves all subjects from the database
func GetSubjects() {
	getObjectsFromCollection(subjectCollection)
}
