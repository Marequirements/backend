package controller

import (
	"back-end/util"
	"context"
	"encoding/json"
	"fmt"
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

	log.Println("HandleAddStudent: Getting authHeader")
	// Get the userToken from the header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		log.Println("HandleAddStudent: AuthHeader is empty, header= ", authHeader)
		http.Error(w, "Authorization header not provided", http.StatusBadRequest)
		return
	}
	log.Println("HandleAddStudent: Got authHeader, header= ", authHeader)

	log.Println("HandleAddStudent: Getting user token")
	userToken := strings.TrimPrefix(authHeader, "Bearer ")
	log.Println("HandleAddStudent: user token= ", userToken)

	log.Println("HandleAddStudent: Getting username form token= ", userToken)
	// Get the username associated with the userToken
	username, err := sc.ts.GetUsernameByToken(userToken)
	if err != nil {
		log.Println("HandleAddStudent: Could not get username from token= ", userToken)
		http.Error(w, "Invalid userToken", http.StatusUnauthorized)
		return
	}
	log.Println("HandleAddStudent: Got username= ", username, " form token= ", userToken)

	log.Println("HandleAddStudent: Checking role of user= ", username)
	// Check if the user has the teacher role
	role, err := sc.GetUserRole(username)
	if err != nil {
		log.Println("HandleAddStudent: Could not get role from user= ", username)
		http.Error(w, "Failed getting username", http.StatusInternalServerError)
		return
	}
	if role != "teacher" {
		log.Println("HandleAddStudent: User= ", username, " does not have teacher role, role= ", role)
		http.Error(w, "Unauthorized access", http.StatusForbidden)
		return
	}
	log.Println("HandleAddStudent: User= ", username, " has role= ", role)

	log.Println("HandleAddStudent: Getting data from body")
	// Decode the student details from the request body
	var student model.NewStudent
	if err := json.NewDecoder(r.Body).Decode(&student); err != nil {
		log.Println("HandleAddStudent: Failed to get data from body= ", r.Body)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Println("HandleAddStudent: Successfully got data from body= ", r.Body)

	log.Println("HandleAddStudent: Getting objectid of class= ", student.ClassTitle)
	classID, err := sc.GetClassIDByTitle(student.ClassTitle)
	if err != nil {
		log.Println("HandleAddStudent: Could not get objectid of class= ", student.ClassTitle)
		http.Error(w, "Invalid class title", http.StatusBadRequest)
		return
	}
	log.Println("HandleAddStudent: Objectid found, objectid =", classID, " calss= ", student.ClassTitle)

	student.Class = classID

	log.Println("HandleAddStudent: Adding student= ", student.Username, " to database")
	// Add the student to the database
	err = sc.AddStudent(student)
	if err != nil {
		log.Println("HandleAddStudent: Failed to add student ", student.Username, " to database")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("HandleAddStudent: Student ", student, " successfully added")

	util.WriteSuccessResponse(w, http.StatusCreated, nil)
}
func (sc *StudentController) AddStudent(student model.NewStudent) error {
	log.Println("Function AddStudent called")
	// Get a handle to the "user" collection
	collection := sc.db.Database("BrainBoard").Collection("user")

	log.Println("AddStudent: Checking if username= ", student.Username, " is already in database")
	// Check if the username is already taken
	filter := bson.M{"username": student.Username}
	var existingUser model.NewStudent
	err := collection.FindOne(context.Background(), filter).Decode(&existingUser)
	if err != mongo.ErrNoDocuments {
		log.Println("AddStudent: Username= ", student.Username, " is already in database")
		return fmt.Errorf("Username already taken")
	}
	log.Println("AddStudent: Username=", student.Username, " not yet in database")
	/*var existingStudent model.Student
	err = collection.FindOne(context.Background(), filter).Decode(&existingStudent)
	if err != mongo.ErrNoDocuments {
		return err
	}
	// Insert the new subject into the collection
	_, err = collection.InsertOne(context.Background(), existingStudent)
	return err*/

	log.Println("AddStudent: Hashing password")
	// Hash the default password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("brainboard"), bcrypt.DefaultCost)
	if err != nil {
		log.Println("AddStudent: Failed to hash password ")
		return err
	}
	log.Println("AddStudent: Password hashed=", hashedPassword)

	// Set the hashed password and role
	student.Password = string(hashedPassword)
	student.Role = "student"

	log.Println("AddStudent: Inserting student to database")
	// Insert the new student into the collection
	_, err = collection.InsertOne(context.Background(), student)
	if err == nil {
		log.Println("AddStudent: Student ", student, " added to database")
	}
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

	log.Println("HandleDeleteStudent: Getting authHeader")
	// Get the userToken from the header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		log.Println("HandleDeleteStudent: authHeader is empty= ", authHeader)
		http.Error(w, "Authorization header not provided", http.StatusBadRequest)
		return
	}
	log.Println("HandleDeleteStudent: Got authHeader= ", authHeader)
	userToken := strings.TrimPrefix(authHeader, "Bearer ")

	log.Println("HandleDeleteStudent: User token is ", userToken)

	log.Println("HandleDeleteStudent: Getting username from from user token= ", userToken)
	// Get the username associated with the userToken
	username, err := sc.ts.GetUsernameByToken(userToken)
	if err != nil {
		log.Println("HandleDeleteStudent: Failed to get username from token= ", userToken)
		http.Error(w, "Invalid userToken", http.StatusUnauthorized)
		return
	}

	log.Println("HandleDeleteStudent: Got username= ", username, " for token ", userToken)

	log.Println("HandleDeleteStudent: Checking the role of user= ", username)
	// Check if the user has the teacher role
	role, err := sc.GetUserRole(username)
	if err != nil || role != "teacher" {
		log.Println("HandleDeleteStudent: Failed to get role of user= ", username)
		http.Error(w, "Unauthorized access", http.StatusForbidden)
		return
	}
	log.Println("HandleDeleteStudent: role= ", role, " username= ", username)

	log.Println("HandleDeleteStudent: Getting student username from body")
	// Get the student's ID from the request body
	var body struct {
		Username string `json:"username"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.Println("HandleDeleteStudent: Failed to get username ")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Println("HandleDeleteStudent: username from body is ", body.Username)
	log.Println("HandleDeleteStudent: Deleting user= ", body.Username)
	// Delete the student from the database
	err = sc.DeleteStudent(body.Username)
	if err != nil {
		log.Println("HandleDeleteStudent: Failed to delete user= ", body.Username)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println("HandleDeleteStudent: user= ", body.Username, " was deleted")
	w.WriteHeader(http.StatusOK)
}
func (sc *StudentController) DeleteStudent(username string) error {
	log.Println("Function DeleteStudent called")
	// Get a handle to the "user" collection
	collection := sc.db.Database("BrainBoard").Collection("user")

	log.Println("DeleteStudent: Deleting student= ", username, " from database")
	// Delete the student from the collection
	filter := bson.M{"username": username, "role": "student"}
	res, err := collection.DeleteOne(context.Background(), filter)
	if err != nil {
		log.Println("DeleteStudent: Failed to delete user= ", username)
		return err
	}

	if res.DeletedCount == 0 {
		log.Println("DeleteStudent: Student with username= ", username, " was not found in database")
		return fmt.Errorf("no student found with the specified username")
	}

	log.Println("DeleteStudent: user= ", username, " was deleted from database")
	return nil
}

func (sc *StudentController) HandleEditStudent(w http.ResponseWriter, r *http.Request) {
	log.Println("Function HandleEditStudent called")

	log.Println("HandleEditStudent: Getting authHeader ")
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		log.Println("HandleEditStudent: authHeader is empty= ", authHeader)
		http.Error(w, "Authorization header not provided", http.StatusBadRequest)
		return
	}
	log.Println("HandleEditStudent: Got authHeader= ", authHeader)
	userToken := strings.TrimPrefix(authHeader, "Bearer ")

	log.Println("HandleEditStudent: user token= ", userToken)

	log.Println("HandleEditStudent: Getting username from token, user token= ", userToken)
	username, err := sc.ts.GetUsernameByToken(userToken)
	if err != nil {
		log.Println("HandleEditStudent: Failed to get username from token= ", userToken)
		http.Error(w, "Invalid userToken", http.StatusUnauthorized)
		return
	}

	log.Println("HandleEditStudent: username = ", username, "token= ", userToken)

	log.Println("HandleEditStudent: Getting role of user= ", username)
	role, err := sc.GetUserRole(username)
	if err != nil {
		log.Println("HandleEditStudent: Failed to get role if user= ", username)
		http.Error(w, "Unauthorized access", http.StatusForbidden)
		return
	}
	if role != "teacher" {
		log.Println("HandleEditStudent: user= ", username, " does not have role of teacher, role= ", role)
		http.Error(w, "Unauthorized access", http.StatusForbidden)
		return
	}

	log.Println("HandleEditStudent: user= ", username, " role= ", role)

	log.Println("HandleEditStudent: Getting data from body")
	var editRequest model.EditStudentRequest
	if err := json.NewDecoder(r.Body).Decode(&editRequest); err != nil {
		log.Println("HandleEditStudent: Failed to get information from body= ", r.Body)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Println("HandleEditStudent: Got data from body= ", r.Body)

	log.Println("HandleEditStudent: Getting class objectid by title= ", editRequest.NewStudent.ClassTitle)
	// Get the class ID by the class title
	classID, err := sc.GetClassIDByTitle(editRequest.NewStudent.ClassTitle)
	if err != nil {
		log.Println("HandleEditStudent: Failed to get objectid of class= ", editRequest.NewStudent.ClassTitle)
		http.Error(w, "Invalid class title", http.StatusBadRequest)
		return
	}

	log.Println("HandleEditStudent: ObjectId of class= ", editRequest.NewStudent.ClassTitle, " : ", classID)

	// Set the class field of the NewStudent struct
	editRequest.NewStudent.Class = classID
	log.Println("HandleEditStudent: Editing student= ", editRequest)
	err = sc.EditStudent(editRequest)
	if err != nil {
		log.Println("HandleEditStudent: Failed to edit student= ", editRequest)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("HandleEditStudent: Successfully edited student= ", editRequest)

	util.WriteSuccessResponse(w, http.StatusNoContent, nil)
}

func (sc *StudentController) EditStudent(editRequest model.EditStudentRequest) error {
	log.Println("Function EditStudent called")
	collection := sc.db.Database("BrainBoard").Collection("user")

	log.Println("EditStudent: Old username= ", editRequest.OldStudentUsername)
	log.Println("EditStudent: New username= ", editRequest.NewStudent.Username)

	log.Println("EditStudent: Hashing the new password")
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(editRequest.NewStudent.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("EditStudent: Failed to hash password= ", editRequest.NewStudent.Password)
		return err
	}
	log.Println("EditStudent: new password hashed")

	update := bson.M{
		"$set": bson.M{
			"username": editRequest.NewStudent.Username,
			"password": string(hashedPassword),
			"name":     editRequest.NewStudent.Name,
			"surname":  editRequest.NewStudent.Surname,
			"class":    editRequest.NewStudent.Class,
		},
	}
	log.Println("EditStudent: Updating user= ", editRequest.OldStudentUsername, "to new user= ", editRequest.NewStudent)
	filter := bson.M{"username": editRequest.OldStudentUsername, "role": "student"} // Use OldStudentUsername here

	res, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Println("EditStudent: Failed to udate user= ", editRequest.OldStudentUsername)
		return err
	}

	if res.MatchedCount == 0 {
		log.Println("EditStudent: No student found with the specified username= ", editRequest.OldStudentUsername)
		return fmt.Errorf("No student found with the specified username")
	}
	log.Println("EditStudent: Successfully udate user= ", editRequest.OldStudentUsername, " to new data= ", editRequest.NewStudent)
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
