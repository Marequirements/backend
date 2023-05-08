package controller

import (
	"back-end/model"
	"back-end/token"
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"net/http"
	"strings"
)

type SubjectController struct {
	db *mongo.Client
	ts *token.Storage
	sc *StudentController
}

func NewSubjectController(db *mongo.Client, ts *token.Storage, sc *StudentController) *SubjectController {
	return &SubjectController{db: db, ts: ts, sc: sc}
}

func (sc *SubjectController) HandleNewSubject(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Token not provided", http.StatusBadRequest)
		return
	}
	splitHeader := strings.Split(authHeader, "Bearer ")
	if len(splitHeader) != 2 {
		http.Error(w, "Invalid authorization header", http.StatusBadRequest)
		return
	}
	token := splitHeader[1]
	// Get the username associated with the token
	username, err := sc.ts.GetUsernameByToken(token)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Check if the user has the teacher role
	role, err := sc.GetUserRole(username)
	if err != nil || role != "teacher" {
		http.Error(w, "Unauthorized access", http.StatusForbidden)
		return
	}

	// Get the ObjectId of the teacher using their username
	teacherID, err := sc.GetTeacherIDByUsername(username)
	if err != nil {
		http.Error(w, "Invalid teacher username", http.StatusBadRequest)
		return
	}

	// Decode the subject details from the request body
	var subject model.NewSubject
	if err := json.NewDecoder(r.Body).Decode(&subject); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set the Teacher field of the subject with the obtained ObjectId
	subject.Teacher = teacherID

	// Get the ObjectId of the class using its title
	classID, err := sc.GetClassIDByTitle(subject.ClassTitle)
	if err != nil {
		http.Error(w, "Invalid class title", http.StatusBadRequest)
		return
	}

	// Set the Class field of the subject with the obtained ObjectId
	subject.Class = classID

	// Add the subject to the database
	err = sc.AddSubject(subject)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (sc *SubjectController) AddSubject(subject model.NewSubject) error {
	collection := sc.db.Database("BrainBoard").Collection("subject")
	filter := bson.M{"title": subject.Title}

	var existingSubject model.Subject
	err := collection.FindOne(context.Background(), filter).Decode(&existingSubject)
	if err != mongo.ErrNoDocuments {
		return err
	}
	// Insert the new subject into the collection
	_, err = collection.InsertOne(context.Background(), subject)
	return err
}

func (sc *SubjectController) GetUserRole(username string) (string, error) {
	// Get a handle to the "user" collection.
	collection := sc.db.Database("BrainBoard").Collection("user")

	// Search for a user with the specified username.
	var user model.NewStudent
	filter := bson.M{"username": username}
	err := collection.FindOne(context.Background(), filter).Decode(&user)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Println("No matching user found:", username)
		} else {
			log.Println(err)
		}
		return "", err
	}
	// Return the user's role
	return user.Role, nil
}

func (sc *SubjectController) GetClassIDByTitle(classTitle string) (primitive.ObjectID, error) {
	collection := sc.db.Database("BrainBoard").Collection("class")
	filter := bson.M{"name": classTitle}
	var class model.Class
	err := collection.FindOne(context.Background(), filter).Decode(&class)
	if err != nil {
		return primitive.NilObjectID, err
	}
	return class.Id, nil
}

func (sc *SubjectController) GetTeacherIDByUsername(username string) (primitive.ObjectID, error) {
	// Get a handle to the "user" collection.
	collection := sc.db.Database("BrainBoard").Collection("user")

	// Search for a user with the specified username.
	var user model.Teacher
	filter := bson.M{"username": username, "role": "teacher"}
	err := collection.FindOne(context.Background(), filter).Decode(&user)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Println("No matching user found:", username)
		} else {
			log.Println(err)
		}
		return primitive.NilObjectID, err
	}
	// Return the ObjectId of the user
	return user.Id, nil
}

//func GetAllSubjects() ([]Subject, error) {
//	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb+srv://mareklescinsky:EUFZTs6jcdkqqEUk@brainboard.lrpc8h3.mongodb.net/test"))
//	if err != nil {
//		return nil, err
//	}
//	defer client.Disconnect(context.Background())
//
//	collection := client.Database("BrainBoard").Collection("subject")
//	cursor, err := collection.Find(context.Background(), bson.D{})
//	if err != nil {
//		return nil, err
//	}
//	defer cursor.Close(context.Background())
//
//	var subjects []Subject
//	for cursor.Next(context.Background()) {
//		var subject Subject
//		err := cursor.Decode(&subject)
//		if err != nil {
//			return nil, err
//		}
//		subjects = append(subjects, subject)
//	}
//	if err := cursor.Err(); err != nil {
//		return nil, err
//	}
//	return subjects, nil
//}
