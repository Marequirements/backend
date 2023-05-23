package controller

import (
	_ "back-end/model"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Class struct {
	Id      string `json:"id" bson:"_id"`
	Name    string `json:"name" bson:"name"`
	Teacher string `json:"teacher" bson:"teacher"`
}

func GetAllClasses() ([]Class, error) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb+srv://mareklescinsky:EUFZTs6jcdkqqEUk@brainboard.lrpc8h3.mongodb.net/test"))
	if err != nil {
		return nil, err
	}
	defer client.Disconnect(context.Background())

	collection := client.Database("BrainBoard").Collection("class")
	cursor, err := collection.Find(context.Background(), bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var classes []Class
	for cursor.Next(context.Background()) {
		var class Class
		err := cursor.Decode(&class)
		if err != nil {
			return nil, err
		}
		classes = append(classes, class)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return classes, nil
}
