package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

func NewStudentController(db *mongo.Client, ts *token.Storage) *StudentController {
	return &StudentController{db: db, ts: ts}
}

//func (sc *StudentController) HandleLogin(w http.ResponseWriter, r *http.Request) {
//	var loginRequest model.LoginRequest
//
//	if err := json.NewDecoder(r.Body).Decode(&loginRequest); err != nil {
//		http.Error(w, err.Error(), http.StatusBadRequest)
//		return
//	}
//	isValid, role := sc.CheckLogin(loginRequest.Username, loginRequest.Password)
//	log.Println("validated login")
//	if !isValid {
//		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
//		return
//	}
//	log.Println("login is valid")
//	// Generate a new token and write it to the response body.
//	token := sc.ts.GenerateToken()
//	sc.ts.AddToken(loginRequest.Username, token)
//	response := struct {
//		Token string `json:"token"`
//		Role  string `json:"role"`
//	}{Token: token, Role: role}
//	w.Header().Set("Content-Type", "application/json")
//	if err := json.NewEncoder(w).Encode(response); err != nil {
//		http.Error(w, err.Error(), http.StatusInternalServerError)
//		return
//	}
//	w.WriteHeader(http.StatusOK)
//}

func (sc *StudentController) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var loginRequest model.LoginRequest

	// Check if the expected fields are present in the JSON object
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&loginRequest); err != nil {
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

	isValid, role := sc.CheckLogin(loginRequest.Username, loginRequest.Password)
	log.Println("validated login")

	//if username or password are not correct sends error
	if !isValid {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		errorResponse := struct {
			Error string `json:"error"`
		}{Error: "Incorrect username or password"}
		if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	// Generate a new userToken and write it to the response body.
	userToken := sc.ts.GenerateToken()
	sc.ts.AddToken(loginRequest.Username, userToken, role)

	// Write the userToken to the response header using the Bearer scheme.
	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", userToken))
	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Role string `json:"role"`
	}{Role: role}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (sc *StudentController) CheckLogin(username, password string) (bool, string) {

	// Get a handle to the "user" collection.
	collection := sc.db.Database("BrainBoard").Collection("user")

	// Search for a user with the specified username.
	var user model.User
	filter := bson.M{"username": username}
	err := collection.FindOne(context.Background(), filter).Decode(&user)
	fmt.Println("Filter:", filter)
	fmt.Println("User found:", user)
	fmt.Printf("Filter: %#v\n", filter)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Println("No matching user found:", username)
		} else {
			log.Println(err)
		}
		return false, ""
	}

	// Compare the supplied password with the stored hashed password.
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		log.Println(err)
		return false, ""
	}

	// If we've made it this far, the login is valid!
	return true, user.Role
}

func (sc *StudentController) HandleLogout(w http.ResponseWriter, r *http.Request) {
	log.Println("Called logout request")

	// Get the userToken from the Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
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
	log.Println("Bearer token=" + authHeader)
	userToken := strings.TrimPrefix(authHeader, "Bearer ")
	if userToken == "" {
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

	log.Println("Got userToken:", userToken)

	// Get the username from the request body
	var body model.User
	log.Println("got body")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&body)
	if err != nil {
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
	log.Println("Got body username: " + body.Username)

	// Check if the username has the given userToken value in the map
	err = sc.ts.DeleteToken(body.Username, userToken)
	if err != nil {
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

	w.WriteHeader(http.StatusOK)
	//if sc.ts.CheckToken(body.Username, userToken) {
	//	log.Println("userToken is in map")
	//	// If the userToken matches, remove the entry from the map
	//	sc.ts.DeleteToken(body.Username, userToken)
	//	log.Println("Token deleted")
	//	w.WriteHeader(http.StatusOK)
	//	log.Println("Token deleted for user:", body.Username)
	//	return
	//}
}
func (sc *StudentController) HandleAddStudent(w http.ResponseWriter, r *http.Request) {
	// Get the userToken from the header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header not provided", http.StatusBadRequest)
		return
	}
	userToken := strings.TrimPrefix(authHeader, "Bearer ")

	// Get the username associated with the userToken
	username, err := sc.ts.GetUsernameByToken(userToken)
	if err != nil {
		http.Error(w, "Invalid userToken", http.StatusUnauthorized)
		return
	}

	// Check if the user has the teacher role
	role, err := sc.GetUserRole(username)
	if err != nil || role != "teacher" {
		http.Error(w, "Unauthorized access", http.StatusForbidden)
		return
	}

	// Decode the student details from the request body
	var student model.NewStudent
	if err := json.NewDecoder(r.Body).Decode(&student); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Add the student to the database
	err = sc.AddStudent(student)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
func (sc *StudentController) AddStudent(student model.NewStudent) error {
	// Get a handle to the "user" collection
	collection := sc.db.Database("BrainBoard").Collection("user")

	// Check if the username is already taken
	filter := bson.M{"username": student.Username}
	var existingUser model.NewStudent
	err := collection.FindOne(context.Background(), filter).Decode(&existingUser)
	if err != mongo.ErrNoDocuments {
		return fmt.Errorf("Username already taken")
	}

	// Hash the student's password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(student.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Set the hashed password and role
	student.Password = string(hashedPassword)
	student.Role = "student"

	// Insert the new student into the collection
	_, err = collection.InsertOne(context.Background(), student)
	return err
}
func (sc *StudentController) GetUserRole(username string) (string, error) {
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
func (sc *StudentController) HandleDeleteStudent(w http.ResponseWriter, r *http.Request) {
	// Get the userToken from the header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header not provided", http.StatusBadRequest)
		return
	}
	userToken := strings.TrimPrefix(authHeader, "Bearer ")

	// Get the username associated with the userToken
	username, err := sc.ts.GetUsernameByToken(userToken)
	if err != nil {
		http.Error(w, "Invalid userToken", http.StatusUnauthorized)
		return
	}

	// Check if the user has the teacher role
	role, err := sc.GetUserRole(username)
	if err != nil || role != "teacher" {
		http.Error(w, "Unauthorized access", http.StatusForbidden)
		return
	}

	// Get the student's ID from the request body
	var body struct {
		StudentID string `json:"studentId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Delete the student from the database
	err = sc.DeleteStudent(body.StudentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
func (sc *StudentController) DeleteStudent(studentID string) error {
	// Get a handle to the "user" collection
	collection := sc.db.Database("BrainBoard").Collection("user")

	// Convert the studentID string to an ObjectID
	objID, err := primitive.ObjectIDFromHex(studentID)
	if err != nil {
		return fmt.Errorf("Invalid student ID")
	}

	// Delete the student from the collection
	filter := bson.M{"_id": objID, "role": "student"}
	res, err := collection.DeleteOne(context.Background(), filter)
	if err != nil {
		return err
	}

	if res.DeletedCount == 0 {
		return fmt.Errorf("no student found with the specified ID")
	}

	return nil
}

func (sc *StudentController) HandleEditStudent(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header not provided", http.StatusBadRequest)
		return
	}
	userToken := strings.TrimPrefix(authHeader, "Bearer ")

	username, err := sc.ts.GetUsernameByToken(userToken)
	if err != nil {
		http.Error(w, "Invalid userToken", http.StatusUnauthorized)
		return
	}

	role, err := sc.GetUserRole(username)
	if err != nil || role != "teacher" {
		http.Error(w, "Unauthorized access", http.StatusForbidden)
		return
	}

	var editRequest model.EditStudentRequest
	if err := json.NewDecoder(r.Body).Decode(&editRequest); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = sc.EditStudent(editRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (sc *StudentController) EditStudent(editRequest model.EditStudentRequest) error {
	collection := sc.db.Database("BrainBoard").Collection("user")

	objID, err := primitive.ObjectIDFromHex(editRequest.StudentID)
	if err != nil {
		return fmt.Errorf("Invalid student ID")
	}

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

	filter := bson.M{"_id": objID, "role": "student"}
	res, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	if res.MatchedCount == 0 {
		return fmt.Errorf("No student found with the specified ID")
	}

	return nil
}
