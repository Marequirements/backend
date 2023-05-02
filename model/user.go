package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	Id       primitive.ObjectID `json:"id" bson:"_id"`
	Username string             `json:"username" bson:"username"`
	Password string             `json:"password" bson:"password"`
	Role     string             `json:"role" bson:"role"`
	Name     string             `json:"name" bson:"name"`
	Surname  string             `json:"surname" bson:"surname"`
	Class    primitive.ObjectID `json:"class" bson:"class"`
}
