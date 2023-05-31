package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Task struct {
	Id          primitive.ObjectID `json:"id" bson:"_id"`
	Deadline    primitive.DateTime `json:"deadline" bson:"deadline"`
	Students    []StudentStatus    `json:"students" bson:"students"`
	Title       string             `json:"title" bson:"title"`
	Description string             `json:"description" bson:"description"`
	Subject     primitive.ObjectID `json:"subject" bson:"subject"`
	Status      string             `json:"status" bson:"status"`
}

type StudentStatus struct {
	StudentID primitive.ObjectID `bson:"studentid" json:"studentid"`
	Status    string             `bson:"status" json:"status"`
}

type TaskWithStudent struct {
	StudentID   primitive.ObjectID `json:"studentId"`
	TaskName    string             `json:"taskName"`
	Subject     string             `json:"subject"`
	Description string             `json:"description"`
	Name        string             `json:"name"`
	Surname     string             `json:"surname"`
	Deadline    primitive.DateTime `json:"deadline"`
}

type TaskAggregationResult struct {
	ID          primitive.ObjectID `bson:"_id"`
	Title       string             `bson:"title"`
	Description string             `bson:"description"`
	Deadline    primitive.DateTime `bson:"deadline"`
	Subject     struct {
		Title string `bson:"title"`
	} `bson:"subjectDetails"`
}

type ClassTasks struct {
	ID          primitive.ObjectID `bson:"_id"`
	Title       string             `bson:"title"`
	Description string             `bson:"description"`
	Deadline    primitive.DateTime `bson:"deadline"`
	Subject     string             `bson:"subject"`
}

type NewTask struct {
	Title       string             `bson:"title"`
	Description string             `bson:"description"`
	Deadline    primitive.DateTime `bson:"deadline"`
	Class       primitive.ObjectID `bson:"class"`
	Subject     primitive.ObjectID `bson:"subject"`
	Students    []StudentStatus    `bson:"students"`
}

type StudentDashboardTask struct {
	Title       string             `bson:"title" json:"title"`
	Description string             `bson:"description" json:"description"`
	Deadline    primitive.DateTime `bson:"deadline" json:"deadline"`
	Students    []StudentStatus    `bson:"students" json:"-"`
}
