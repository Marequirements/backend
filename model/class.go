package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type Class struct {
	Id   primitive.ObjectID `json:"id" bson:"_id"`
	Name string             `json:"name" bson:"name"`
}
