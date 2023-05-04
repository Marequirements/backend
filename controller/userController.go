package controller

import (
	"back-end/model"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type User model.User

func GetAllUsers() ([]User, error) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb+srv://<username>:<password>@<cluster-address>/test?w=majority"))
	if err != nil {
		return nil, err
	}
	defer client.Disconnect(context.Background())

	collection := client.Database("BrainBoard").Collection("user")
	cursor, err := collection.Find(context.Background(), bson.D{})
	if err != nil {
		return nil, err

	}
	defer cursor.Close(context.Background())

	var users []User
	for cursor.Next(context.Background()) {
		var user User
		err := cursor.Decode(&user)
		if err != nil {
			return nil, err
		}
		users = append(users, user)

	}
	if err := cursor.Err(); err != nil {
		return nil, err

	}
	return users, nil

}
