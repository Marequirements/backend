package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type Task struct {
	Id          primitive.ObjectID `json:"id" bson:"_id"`
	Deadline    primitive.DateTime `json:"deadline" bson:"deadline"`
	Priority    string             `json:"priority" bson:"priority"`
	Students    []StudentStatus    `json:"students" bson:"students"`
	Title       string             `json:"title" bson:"title"`
	Description string             `json:"description" bson:"description"`
	Lesson      primitive.ObjectID `json:"lesson" bson:"lesson"`
	Status      string             `json:"status" bson:"status"`
}

type StudentStatus struct {
	StudentID primitive.ObjectID `bson:"studentid" json:"studentid"`
	Status    string             `bson:"status" json:"status"`
}

type TaskWithStudent struct {
	StudentID primitive.ObjectID `json:"studentId"`
	TaskName  string             `json:"taskName"`
}
