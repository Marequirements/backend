package controller

import (
	"back-end/util"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"strings"

	"back-end/model"
	"back-end/token"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type StudentController struct {
	db *mongo.Client
	ts *token.Storage
}
type UserController struct {
	db *mongo.Client
}

type User model.Student

func NewStudentController(db *mongo.Client, ts *token.Storage) *StudentController {
	return &StudentController{db: db, ts: ts}
}

func (sc *StudentController) HandleLogin(w http.ResponseWriter, r *http.Request) {
	log.Println("Function HandleLogin called")
	var loginRequest model.LoginRequest

	log.Println("HandleLogin: Extracting information from body")
	// Check if the expected fields are present in the JSON object
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&loginRequest); err != nil {
		log.Println("HandleLogin: Failed to get data from body")
		util.WriteErrorResponse(w, http.StatusBadRequest, "JSON parameters not provided")
		return
	}
	log.Println("HandleLogin: Got body data= ", loginRequest)

	log.Println("HandleLogin: Validating login")
	isValid, role := sc.CheckLogin(loginRequest.Username, loginRequest.Password)

	//if username or password are not correct sends error
	if !isValid {
		log.Println("HandleLogin: Incorrect login data, username= ", loginRequest.Username, " password= ", loginRequest.Password)
		util.WriteErrorResponse(w, http.StatusUnauthorized, "Incorrect username or password")

		return
	}
	log.Println("HandleLogin: Login data correct, username= ", loginRequest.Username, " password= ", loginRequest.Password)

	log.Println("HandleLogin: generating token for user= ", loginRequest.Username)
	// Generate a new userToken and write it to the response body.
	userToken := sc.ts.GenerateToken()
	sc.ts.AddToken(loginRequest.Username, userToken, role)
	log.Println("HandleLogin: Token= ", userToken, " generated and added to token storage for user= ", loginRequest.Username)

	log.Println("HandleLogin: Returning token in authHeader")
	// Write the userToken to the response header using the Bearer scheme.
	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", userToken))
	log.Println("HandleLogin: Returning role")
	w.Header().Set("Access-Control-Expose-Headers", "Authorization")

	response := struct {
		Role string `json:"role"`
	}{Role: role}

	log.Println("HandleLogin: Returned token= ", userToken, " role= ", response)

	util.WriteSuccessResponse(w, http.StatusOK, response)

}

func (sc *StudentController) CheckLogin(username, password string) (bool, string) {
	log.Println("Function CheckLogin called")

	// Get a handle to the "user" collection.
	collection := sc.db.Database("BrainBoard").Collection("user")

	// Search for a user with the specified username.
	var user model.Student
	filter := bson.M{"username": username}

	log.Println("CheckLogin: Searching for username= ", username)

	err := collection.FindOne(context.Background(), filter).Decode(&user)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Println("CheckLogin: No matching user found:", username)
		} else {
			log.Println("CheckLogin: ", err)
		}
		return false, ""
	}

	log.Println("CheckLogin: User= ", username, " found:", user)

	log.Println("CheckLogin: Comparing supplied password with stored hashed password ")
	// Compare the supplied password with the stored hashed password.
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		log.Println("CheckLogin: ", err)
		return false, ""
	}
	log.Println("The password is correct")

	// If we've made it this far, the login is valid!
	return true, user.Role
}

func (sc *StudentController) HandleLogout(w http.ResponseWriter, r *http.Request) {
	log.Println("Function HandleLogout called")

	log.Println("HandleLogout: Getting authHeader")
	// Get the userToken from the Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		log.Println("HandleLogout: Auth header is empty, header= ", authHeader)
		util.WriteErrorResponse(w, http.StatusUnauthorized, "Token or username is incorrect")
		return
	}
	log.Println("HandleLogout: Got authHeader= " + authHeader)

	log.Println("HandleLogout: Getting user token")

	userToken := strings.TrimPrefix(authHeader, "Bearer ")
	if userToken == "" {
		log.Println("HandleLogout: Token is empty: ", userToken)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		errorResponse := struct {
			Error string `json:"error"`
		}{Error: "Token or username is incorrect"}
		if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}
	log.Println("HandleLogout: Got user token= ", userToken)

	log.Println("HandleLogout: Getting username from body")
	// Get the username from the request body
	var body model.Student
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&body)
	if err != nil {
		log.Println("HandleLogout: Could not get body, body= ", body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		errorResponse := struct {
			Error string `json:"error"`
		}{Error: "JSON parameters not provided"}
		if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}
	log.Println("HandleLogout: Got username from body: ", body.Username)

	log.Println("HandleLogout: Deleting token of user= ", body.Username, " with token= ", userToken)
	// Check if the username has the given userToken value in the map
	err = sc.ts.DeleteToken(body.Username, userToken)
	if err != nil {
		log.Println("HandleLogout: Failed to delete token= ", userToken, " for user= ", body.Username)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		errorResponse := struct {
			Error string `json:"error"`
		}{Error: "Token or username is incorrect"}
		if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	log.Println("HandleLogout: Token= ", userToken, " deleted for user= ", body.Username)

	w.WriteHeader(http.StatusOK)
}

func (sc *StudentController) HandleAddStudent(w http.ResponseWriter, r *http.Request) {
	log.Println("Function HandleAddStudent called")

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		respondWithError(w, http.StatusBadRequest, "Authorization header not provided")
		return
	}
	userToken := strings.TrimPrefix(authHeader, "Bearer ")

	username, err := sc.ts.GetUsernameByToken(userToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Token is invalid")
		return
	}

	role, err := sc.GetUserRole(username)
	if err != nil || role != "teacher" {
		respondWithError(w, http.StatusForbidden, "User does not have permission for this request")
		return
	}

	var student model.NewStudent
	if err := json.NewDecoder(r.Body).Decode(&student); err != nil {
		respondWithError(w, http.StatusBadRequest, "JSON parameters not provided")
		return
	}
	// Validate the required fields here. Assuming username and classTitle as required fields.
	if student.Username == "" || student.ClassTitle == "" {
		respondWithError(w, http.StatusBadRequest, "JSON parameters not provided")
		return
	}

	classID, err := sc.GetClassIDByTitle(student.ClassTitle)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Class does not exist")
		return
	}

	student.Class = classID
	err = sc.AddStudent(student)
	if err != nil {
		if err.Error() == "Username already taken" {
			respondWithError(w, http.StatusConflict, "Username already exists")
		} else {
			respondWithError(w, http.StatusInternalServerError, "Unexpected server error")
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (sc *StudentController) AddStudent(student model.NewStudent) error {
	log.Println("Function AddStudent called")
	collection := sc.db.Database("BrainBoard").Collection("user")

	filter := bson.M{"username": student.Username}
	var existingUser model.NewStudent
	err := collection.FindOne(context.Background(), filter).Decode(&existingUser)
	if err != mongo.ErrNoDocuments {
		return fmt.Errorf("Username already taken")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("brainboard"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	student.Password = string(hashedPassword)
	student.Role = "student"

	_, err = collection.InsertOne(context.Background(), student)
	return err
}

func (sc *StudentController) GetUserRole(username string) (string, error) {
	log.Println("Function GetUserRole called")
	// Get a handle to the "user" collection.
	collection := sc.db.Database("BrainBoard").Collection("user")

	log.Println("GetUserRole: Getting role of user= ", username, " from database")
	// Search for a user with the specified username.
	var user model.NewStudent
	filter := bson.M{"username": username}
	err := collection.FindOne(context.Background(), filter).Decode(&user)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Println("GetUserRole: No matching user found:", username)
		} else {
			log.Println("GetUserRole:", err)
		}
		return "", err
	}

	log.Println("GetUserRole: The role of user ", username, " is ", user.Role)

	// Return the user's role
	return user.Role, nil
}

func (sc *StudentController) HandleDeleteStudent(w http.ResponseWriter, r *http.Request) {
	log.Println("Function HandleDeleteStudent called")

	// Get the userToken from the header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		respondWithError(w, http.StatusBadRequest, "Authorization header not provided")
		return
	}
	userToken := strings.TrimPrefix(authHeader, "Bearer ")

	// Get the username associated with the userToken
	username, err := sc.ts.GetUsernameByToken(userToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Token is invalid")
		return
	}

	// Check if the user has the teacher role
	role, err := sc.GetUserRole(username)
	if err != nil || role != "teacher" {
		respondWithError(w, http.StatusForbidden, "User does not have permission for this request")
		return
	}

	// Get the student's ID from the request body
	var body struct {
		Username string `json:"username"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondWithError(w, http.StatusBadRequest, "JSON parameters not provided")
		return
	}

	// Delete the student from the database
	err = sc.DeleteStudent(body.Username)
	if err != nil {
		if err.Error() == "no student found with the specified username" {
			respondWithError(w, http.StatusNotFound, "The Subject does not exists")
		} else {
			respondWithError(w, http.StatusInternalServerError, "Unexpected server error")
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (sc *StudentController) DeleteStudent(username string) error {
	log.Println("Function DeleteStudent called")

	// Get a handle to the "user" collection
	collection := sc.db.Database("BrainBoard").Collection("user")

	// Delete the student from the collection
	filter := bson.M{"username": username, "role": "student"}
	res, err := collection.DeleteOne(context.Background(), filter)
	if err != nil {
		return err
	}

	if res.DeletedCount == 0 {
		return fmt.Errorf("no student found with the specified username")
	}

	return nil
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	resp := map[string]string{"error": message}
	jsonResp, _ := json.Marshal(resp)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(jsonResp)
}

func (sc *StudentController) HandleEditStudent(w http.ResponseWriter, r *http.Request) {
	log.Println("Function HandleEditStudent called")

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		respondWithError(w, http.StatusBadRequest, "Authorization header not provided")
		return
	}
	userToken := strings.TrimPrefix(authHeader, "Bearer ")

	username, err := sc.ts.GetUsernameByToken(userToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Token is invalid")
		return
	}

	role, err := sc.GetUserRole(username)
	if err != nil || role != "teacher" {
		respondWithError(w, http.StatusForbidden, "User does not have permission for this request")
		return
	}

	var editRequest model.EditStudentRequest
	if err := json.NewDecoder(r.Body).Decode(&editRequest); err != nil {
		respondWithError(w, http.StatusBadRequest, "JSON parameters not provided")
		return
	}

	classID, err := sc.GetClassIDByTitle(editRequest.NewStudent.ClassTitle)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Class does not exist")
		return
	}

	editRequest.NewStudent.Class = classID
	err = sc.EditStudent(editRequest)
	if err != nil {
		if err.Error() == "No student found with the specified username" {
			respondWithError(w, http.StatusNotFound, "Username does not exist")
		} else {
			respondWithError(w, http.StatusInternalServerError, "Unexpected server error")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (sc *StudentController) EditStudent(editRequest model.EditStudentRequest) error {
	log.Println("Function EditStudent called")
	collection := sc.db.Database("BrainBoard").Collection("user")

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(editRequest.NewStudent.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"username": editRequest.NewStudent.Username,
			"password": string(hashedPassword),
			"name":     editRequest.NewStudent.Name,
			"surname":  editRequest.NewStudent.Surname,
			"class":    editRequest.NewStudent.Class,
		},
	}
	filter := bson.M{"username": editRequest.OldStudentUsername, "role": "student"}

	res, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	if res.MatchedCount == 0 {
		return fmt.Errorf("No student found with the specified username")
	}

	return nil
}

func (uc *UserController) GetUserRole(username string) (string, error) {
	log.Println("Function GetUserRole called")
	collection := uc.db.Database("BrainBoard").Collection("user")
	filter := bson.M{"username": username}
	var user model.Student

	log.Println("GetUserRole: Getting role by usename= ", username, " from database")

	err := collection.FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
		log.Println("GetUserRole: Failed to get role for user= ", username, " from database")
		return "", err
	}

	log.Println("GetUserRole: Got role for user= ", username, " role= ", user.Role)
	return user.Role, nil
}

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
func (sc *StudentController) GetClassIDByTitle(classTitle string) (primitive.ObjectID, error) {
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
func (sc *StudentController) HandleGetStudentsFromClass(w http.ResponseWriter, r *http.Request) {
	log.Println("Function HandleGetStudentsFromClass called")

	classTitle := chi.URLParam(r, "classTitle")
	if classTitle == "" {
		log.Println("HandleGetStudentsFromClass: classTitle parameter is missing")
		util.WriteErrorResponse(w, http.StatusBadRequest, "classTitle parameter is missing")
		return
	}

	classCollection := sc.db.Database("BrainBoard").Collection("class")
	var classDoc struct {
		ID primitive.ObjectID `bson:"_id"`
	}
	err := classCollection.FindOne(context.Background(), bson.M{"name": classTitle}).Decode(&classDoc)
	if err != nil {
		log.Println("HandleGetStudentsFromClass: Invalid class title")
		util.WriteErrorResponse(w, http.StatusBadRequest, "Invalid class title")
		return
	}

	userCollection := sc.db.Database("BrainBoard").Collection("user")

	filter := bson.M{"class": classDoc.ID, "role": "student"}

	cur, err := userCollection.Find(context.Background(), filter)
	if err != nil {
		log.Printf("Error getting students: %v", err)
		util.WriteErrorResponse(w, http.StatusInternalServerError, "Error getting students")
		return
	}

	var students []model.Student
	if err = cur.All(context.Background(), &students); err != nil {
		log.Printf("Error decoding students: %v", err)
		util.WriteErrorResponse(w, http.StatusInternalServerError, "Error decoding students")
		return
	}

	util.WriteSuccessResponse(w, http.StatusOK, students)
}

func (sc *StudentController) GetStudentsByClass(w http.ResponseWriter, r *http.Request) {
	// First check if the user is a teacher
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		util.WriteErrorResponse(w, http.StatusBadRequest, "Authorization header not provided")
		return
	}
	userToken := strings.TrimPrefix(authHeader, "Bearer ")

	username, err := sc.ts.GetUsernameByToken(userToken)
	if err != nil {
		util.WriteErrorResponse(w, http.StatusUnauthorized, "Token is invalid")
		return
	}

	role, err := sc.GetUserRole(username)
	if err != nil || role != "teacher" {
		util.WriteErrorResponse(w, http.StatusForbidden, "User does not have permission for this request")
		return
	}
	classTitle := chi.URLParam(r, "class")

	// Get a handle to the "user" collection.
	collection := sc.db.Database("BrainBoard").Collection("user")

	// Search for all students in the specified class.
	filter := bson.M{"classTitle": classTitle, "role": "student"}

	log.Println("GetStudentsByClass: Searching for class= ", classTitle)

	var students []model.Student
	cur, err := collection.Find(context.Background(), filter)
	if err != nil {
		log.Println("GetStudentsByClass: ", err)
		util.WriteErrorResponse(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	defer cur.Close(context.Background())

	for cur.Next(context.Background()) {
		var student model.Student
		err := cur.Decode(&student)
		if err != nil {
			log.Println("GetStudentsByClass: ", err)
			util.WriteErrorResponse(w, http.StatusInternalServerError, "Internal Server Error")
			return
		}
		students = append(students, student)
	}

	if err := cur.Err(); err != nil {
		log.Println("GetStudentsByClass: ", err)
		util.WriteErrorResponse(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	log.Println("GetStudentsByClass: Returning list of students in class= ", classTitle)
	util.WriteSuccessResponse(w, http.StatusOK, students)
}
