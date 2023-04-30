package main

import (
	"gopkg.in/mgo.v2/bson"
)

type User struct {
	Id       bson.ObjectId   `json:"id" bson:"_id"`
	Username string          `json:"username" bson:"username"`
	Password string          `json:"password" bson:"password"`
	Role     string          `json:"role" bson:"role"`
	Name     string          `json:"name" bson:"name"`
	Surname  string          `json:"surname" bson:"surname"`
	Class    []bson.ObjectId `json:"class" bson:"class"`
}
