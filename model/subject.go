package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type Subject struct {
	Id      primitive.ObjectID `json:"id" bson:"_id"`
	Class   primitive.ObjectID `json:"class" bson:"class"`
	Teacher primitive.ObjectID `json:"teacher" bson:"teacher"`
	Title   string             `json:"title" bson:"title"`
}

type NewSubject struct {
	ClassTitle  string             `json:"classTitle" bson:"-"`
	Class       primitive.ObjectID `json:"class" bson:"class"`
	Title       string             `json:"title" bson:"title"`
	Teacher     primitive.ObjectID `json:"teacher" bson:"teacher"`
	TeacherName string             `json:"teacherName" bson:"-"`
}

type FormSubjects struct {
	Title string `json:"title" bson:"title"`
	Class string `json:"class" bson:"class"`
}
