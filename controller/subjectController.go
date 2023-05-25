package controller

import (
	"back-end/model"
	"back-end/token"
	"back-end/util"
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"net/http"
)

type SubjectController struct {
	db *mongo.Client
	ts *token.Storage
	uc *StudentController
}

func NewSubjectController(db *mongo.Client, ts *token.Storage, uc *StudentController) *SubjectController {
	return &SubjectController{db: db, ts: ts, uc: uc}
}

func (sc *SubjectController) HandleGetFormSubjects(w http.ResponseWriter, r *http.Request) {
	_, err := util.TeacherLogin("HandleGetFormSubjects", sc.db, sc.ts, w, r)
	if err != nil {
		return
	}

	response := sc.GetAllSubjects()
	util.WriteSuccessResponse(w, 200, response)
}

func (sc *SubjectController) HandleNewSubject(w http.ResponseWriter, r *http.Request) {
	username, err := util.TeacherLogin("HandleNewSubject", sc.db, sc.ts, w, r)
	if err != nil {
		return
	}

	log.Println("HandleNewSubject:  Getting objectid of teacher/user= ", username)

	// Get the ObjectId of the teacher using their username
	teacherID, err := sc.GetTeacherIDByUsername(username)
	if err != nil {
		log.Println("HandleNewSubject: Could not get teacherid of teacher= ", username)

		errMsg := "Teacher does not exist"
		util.WriteErrorResponse(w, 404, errMsg)
		return
	}
	log.Println("HandleNewSubject: Success got teacherid= ", teacherID, " of teacher= ", username)

	log.Println("HandleNewSubject: Decoding body")

	// Decode the subject details from the request body
	var subject model.NewSubject
	if err := json.NewDecoder(r.Body).Decode(&subject); err != nil {
		log.Println("HandleNewSubject: Failed decoding the body ", r.Body)

		errMsg := "JSON parameters not provided"
		util.WriteErrorResponse(w, 400, errMsg)
		return
	}

	log.Println("HandleNewSubject: Got body data ", r.Body)

	// Set the Teacher field of the subject with the obtained ObjectId
	subject.Teacher = teacherID

	log.Println("HandleNewSubject: Getting objectid of class", subject.ClassTitle)

	// Get the ObjectId of the class using its title
	classID, err := sc.GetClassIDByTitle(subject.ClassTitle)
	if err != nil {
		log.Println("HandleNewSubject: Could not get classid of class= ", subject.ClassTitle)

		errMsg := "Class does not exists"
		util.WriteErrorResponse(w, 404, errMsg)
		return
	}
	log.Println("HandleNewSubject: Success, got classid= ", classID, " of class", subject.ClassTitle)

	// Set the Class field of the subject with the obtained ObjectId
	subject.Class = classID

	log.Println("HandleNewSubject: Adding subject to database")

	// Add the subject to the database
	err = sc.AddSubject(subject)
	if err != nil {
		log.Println("HandleNewSubject: Failed to add subject", subject.Title)

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("HandleNewSubject: Successfully added subject", subject.Title)

	w.WriteHeader(http.StatusCreated)
}

func (sc *SubjectController) HandleDeleteSubject(w http.ResponseWriter, r *http.Request) {
	_, err := util.TeacherLogin("HandleDeleteSubject", sc.db, sc.ts, w, r)
	if err != nil {
		return
	}

	log.Println("HandleDeleteSubject: Getting data from body ")

	// Decode the subject details from the request body
	var subject model.NewSubject
	if err := json.NewDecoder(r.Body).Decode(&subject); err != nil {
		log.Println("HandleDeleteSubject: JSON parameters not provided ", r.Body)

		errMsg := "JSON parameters not provided"
		util.WriteErrorResponse(w, 400, errMsg)
		return
	}
	log.Println("HandleDeleteSubject: Data decoded successfully= ", r.Body)

	log.Println("HandleDeleteSubject: Getting objectid of class ", subject.ClassTitle)
	// Get the ObjectId of the class using its title
	classID, err := sc.GetClassIDByTitle(subject.ClassTitle)
	if err != nil {
		log.Println("HandleDeleteSubject: Could not get objectid of class ", subject.ClassTitle)
		errMsg := "Class does not exists"
		util.WriteErrorResponse(w, 404, errMsg)
		return
	}
	log.Println("HandleDeleteSubject: Got object id= ", classID, " of class", subject.ClassTitle)

	// Set the Class field of the subject with the obtained ObjectId
	subject.Class = classID

	log.Println("HandleDeleteSubject: Deleting the subject ")
	// Delete the subject from the database
	err = sc.DeleteSubject(subject)
	if err != nil {
		log.Println("HandleDeleteSubject: Failed to delete the subject ", subject.ClassTitle)

		errMsg := "Subject does not exist"
		util.WriteErrorResponse(w, 404, errMsg)
		return
	}

	log.Println("HandleDeleteSubject: Subject= ", subject.ClassTitle, " deleted successfully")

	w.WriteHeader(204)
}

func (sc *SubjectController) AddSubject(subject model.NewSubject) error {
	log.Println("Function AddSubject was called")

	collection := sc.db.Database("BrainBoard").Collection("subject")
	filter := bson.M{"title": subject.Title}

	var existingSubject model.Subject

	log.Println("AddSubject: Adding new subject= ", subject.Title, " for class= ", subject.ClassTitle)

	err := collection.FindOne(context.Background(), filter).Decode(&existingSubject)
	if err != mongo.ErrNoDocuments {
		log.Println("AddSubject: The subject ", subject.Title, " sor class ", subject.ClassTitle, " is already in database")
		return err
	}

	// Insert the new subject into the collection
	_, err = collection.InsertOne(context.Background(), subject)

	if err != nil {
		log.Println("AddSubject: Failed inserting new subject= ", subject.Title, " for class= ", subject.ClassTitle, " to database")
		return err

	}

	log.Println("AddSubject: new subject= ", subject.Title, " for class= ", subject.ClassTitle, " successfully added")
	return err
}

func (sc *SubjectController) DeleteSubject(subject model.NewSubject) error {
	log.Println("Function DeleteSubject was called")

	collection := sc.db.Database("BrainBoard").Collection("subject")
	filter := bson.M{"title": subject.Title, "class": subject.Class}

	var existingSubject model.Subject

	log.Println("DeleteSubject: Searching for subject= ", subject.Title, " in class= ", subject.ClassTitle)
	err := collection.FindOneAndDelete(context.Background(), filter).Decode(&existingSubject)
	if err != nil {
		log.Println("DeleteSubject: Failed could not find subject= ", subject.Title, " in class= ", subject.ClassTitle)
		return err
	}

	log.Println("DeleteSubject: Subject= ", subject.Title, " in class= ", subject.ClassTitle, " was deleted")
	return nil
}

func (sc *SubjectController) GetUserRole(username string) (string, error) {
	log.Println("Function GetUserRole was called")

	// Get a handle to the "user" collection.
	collection := sc.db.Database("BrainBoard").Collection("user")

	// Search for a user with the specified username.
	var user model.NewStudent
	filter := bson.M{"username": username}

	log.Println("GetUserRole: Searching for role with user= ", username)

	err := collection.FindOne(context.Background(), filter).Decode(&user)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Println("GetUserRole: No matching user found:", username)
		} else {
			log.Println(err)
		}
		return "", err
	}

	log.Println("GetUserRole: Found user= ", username, " with role= ", user.Role)

	// Return the user's role
	return user.Role, nil
}

func (sc *SubjectController) GetClassIDByTitle(classTitle string) (primitive.ObjectID, error) {
	log.Println("Function GetClassIDByTitle was called")

	collection := sc.db.Database("BrainBoard").Collection("class")
	filter := bson.M{"name": classTitle}
	var class model.Class

	log.Println("GetClassIDByTitle: searching ID for class= ", classTitle)

	err := collection.FindOne(context.Background(), filter).Decode(&class)
	if err != nil {
		log.Println("GetClassIDByTitle: Fail class= ", classTitle, " not found")
		return primitive.NilObjectID, err
	}
	log.Println("GetClassIDByTitle: Found ID= ", class.Id, " for class= ", classTitle)
	return class.Id, nil
}

func (sc *SubjectController) GetTeacherIDByUsername(username string) (primitive.ObjectID, error) {
	log.Println("Function GetTeacherIDByUsername was called")

	// Get a handle to the "user" collection.
	collection := sc.db.Database("BrainBoard").Collection("user")

	// Search for a user with the specified username.
	var user model.Teacher
	filter := bson.M{"username": username, "role": "teacher"}

	log.Println("GetTeacherIDByUsername: searching ID for user/teacher= ", username)

	err := collection.FindOne(context.Background(), filter).Decode(&user)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Println("GetTeacherIDByUsername: No matching user found:", username)
		} else {
			log.Println(err)
		}
		return primitive.NilObjectID, err
	}

	log.Println("GetTeacherIDByUsername: Found User ID= ", user.Id, " for user/teacher= ", username)
	// Return the ObjectId of the user
	return user.Id, nil
}

func (sc *SubjectController) GetAllSubjects() []model.FormSubjects {
	log.Println("Function GetAllSubject called")
	subjectsCollection := sc.db.Database("BrainBoard").Collection("subject")
	classesCollection := sc.db.Database("BrainBoard").Collection("class")

	cur, _ := subjectsCollection.Find(context.Background(), bson.M{})
	defer cur.Close(context.Background())

	var results []model.FormSubjects

	for cur.Next(context.Background()) {
		var subject model.Subject
		cur.Decode(&subject)

		var class Class
		log.Println("GetAllSubject: Searching for class with id= ", subject.Class)
		classesCollection.FindOne(context.Background(), bson.M{"_id": subject.Class}).Decode(&class)

		results = append(results, model.FormSubjects{
			Class: class.Name,
			Title: subject.Title,
		})
	}

	for _, result := range results {
		fmt.Printf("Name: %s, Title: %s\n", result.Class, result.Title)
	}

	return results
}
