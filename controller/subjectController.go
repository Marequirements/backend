package controller

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Subject struct {
	Id      primitive.ObjectID `json:"id" bson:"_id"`
	Class   primitive.ObjectID `json:"class" bson:"class"`
	Teacher primitive.ObjectID `json:"teacher" bson:"teacher"`
	Title   string             `json:"title" bson:"title"`
}

func GetAllSubjects() ([]Subject, error) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb+srv://mareklescinsky:EUFZTs6jcdkqqEUk@brainboard.lrpc8h3.mongodb.net/test"))
	if err != nil {
		return nil, err
	}
	defer client.Disconnect(context.Background())

	collection := client.Database("BrainBoard").Collection("subject")
	cursor, err := collection.Find(context.Background(), bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var subjects []Subject
	for cursor.Next(context.Background()) {
		var subject Subject
		err := cursor.Decode(&subject)
		if err != nil {
			return nil, err
		}
		subjects = append(subjects, subject)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return subjects, nil
}
