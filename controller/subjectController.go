package controller

import (
	"back-end/model"
	"back-end/token"
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"net/http"
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
	token := r.Header.Get("token")
	if token == "" {
		http.Error(w, "Token not provided", http.StatusBadRequest)
		return
	}
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

	// Decode the subject details from the request body
	var subject model.NewSubject
	if err := json.NewDecoder(r.Body).Decode(&subject); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

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
