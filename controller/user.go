package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"back-end/model"
	"back-end/token"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type StudentController struct {
	db *mongo.Client
	ts *token.TokenStorage
}

func NewStudentController(db *mongo.Client, ts *token.TokenStorage) *StudentController {
	return &StudentController{db: db, ts: ts}
}

func (uc *StudentController) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var loginRequest model.LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&loginRequest); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	isValid, role := uc.CheckLogin(loginRequest.Username, loginRequest.Password)
	log.Println("validated login")
	if !isValid {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}
	log.Println("login is valid")
	// Generate a new token and write it to the response body.
	token := uc.ts.GenerateToken()
	uc.ts.AddToken(loginRequest.Username, token)
	response := struct {
		Token string `json:"token"`
		Role  string `json:"role"`
	}{Token: token, Role: role}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func (uc *StudentController) CheckLogin(username, password string) (bool, string) {

	// Get a handle to the "user" collection.
	collection := uc.db.Database("BrainBoard").Collection("user")

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

func (uc *StudentController) HandleLogout(w http.ResponseWriter, r *http.Request) {
	log.Println("Called logout request")
	// Get the token from the header
	token := r.Header.Get("token")
	log.Println("got header token: " + token)
	if token == "" {
		log.Println("Token not provided")
		http.Error(w, "Token not provided", http.StatusBadRequest)
		return
	}

	// Get the username from the request body
	var body model.User
	log.Println("got body")
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		log.Println("Invalid request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	log.Println("Got body username: " + body.Username)
	// Check if the username has the given token value in the map
	if uc.ts.CheckToken(body.Username, token) {
		log.Println("token is in map")
		// If the token matches, remove the entry from the map
		uc.ts.DeleteToken(body.Username, token)
		log.Println("Token deleted")
		w.WriteHeader(http.StatusOK)
		log.Println("Token deleted for user:", body.Username)
		return
	}

	// If the token doesn't match, return an error response
	log.Println("Invalid token for user:", body.Username)
	http.Error(w, "Invalid token for the given username", http.StatusUnauthorized)
}
