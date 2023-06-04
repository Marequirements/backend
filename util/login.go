package util

import (
	"back-end/model"
	"back-end/token"
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"net/http"
	"strings"
)

// TeacherLogin returns username, if returns error just end the function that called this one with return
func TeacherLogin(functionName string, db *mongo.Client, ts *token.Storage, w http.ResponseWriter, r *http.Request) (string, error) {
	log.Println("Function ", functionName, " called")

	log.Println(functionName, ":getting suth header")
	authHeader, err := getAuthHeader(w, r)
	if err != nil {
		return "", err
	}
	log.Println(functionName, ":got auth header= ", authHeader)

	log.Println(functionName, ":getting token")
	userToken := getToken(authHeader)
	log.Println(functionName, ":token=", userToken)

	log.Println(functionName, ":getting username from token")
	username, err := getUsernameFromToken(userToken, ts, w)
	if err != nil {
		return "", err
	}
	log.Println(functionName, ":username= ", username, " from token= ", userToken)

	log.Println(functionName, ":getting role from username")
	role, err := getUserRoleFromUsername(username, w, db)
	if err != nil {
		return "", err
	}
	log.Println(functionName, ":role= ", role, " username= ", username)

	log.Println(functionName, ":checking if role is teacher")
	err = checkTeacherRole(role, w, username)
	if err != nil {
		log.Println(functionName, ":username= ", username, " checked role is teacher,role= ", role)
		return "", err
	}
	return username, nil
}

func StudentLogin(db *mongo.Client, ts *token.Storage, w http.ResponseWriter, r *http.Request) (string, error) {
	log.Println("Function StudentLogin called")

	log.Println("StudentLogin:getting suth header")
	authHeader, err := getAuthHeader(w, r)
	if err != nil {
		return "", err
	}
	log.Println("StudentLogin:got auth header= ", authHeader)

	log.Println("StudentLogin:getting token")
	userToken := getToken(authHeader)
	log.Println("StudentLogin:token=", userToken)

	log.Println("StudentLogin:getting username from token")
	username, err := getUsernameFromToken(userToken, ts, w)
	if err != nil {
		return "", err
	}
	log.Println("StudentLogin:username= ", username, " from token= ", userToken)

	log.Println("StudentLogin:getting role from username")
	role, err := getUserRoleFromUsername(username, w, db)
	if err != nil {
		return "", err
	}
	log.Println("StudentLogin:role= ", role, " username= ", username)

	log.Println("StudentLogin:checking if role is student")
	err = checkStudentRole(role, w, username)
	if err != nil {
		log.Println("StudentLogin:username= ", username, " checked role is teacher,role= ", role)
		return "", err
	}
	return username, nil
}

func getAuthHeader(w http.ResponseWriter, r *http.Request) (string, error) {
	log.Println("getAuthHeader: Getting auth header")
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		log.Println("getAuthHeader: Failed to get auth header= ", authHeader)
		WriteErrorResponse(w, http.StatusBadRequest, "Could not get token parameter")
		return "", fmt.Errorf("failed to get authHeader")
	}
	log.Println("getAuthHeader: Got authHeader= ", authHeader)
	return authHeader, nil
}

func getToken(authHeader string) string {
	log.Println("getToken: Getting token")
	userToken := strings.TrimPrefix(authHeader, "Bearer ")
	log.Println("getToken: Token= ", userToken)
	return userToken
}

func getUsernameFromToken(userToken string, ts *token.Storage, w http.ResponseWriter) (string, error) {
	log.Println("getUsernameFromToken: Getting username from token")
	//username, err := tc.ts.GetUsernameByToken(userToken)
	username, err := ts.GetUsernameByToken(userToken)
	if err != nil {
		log.Println("getUsernameFromToken: Failed to get username from token= ", userToken)
		log.Println("getUsernameFromToken: Returning status code 401")
		WriteErrorResponse(w, http.StatusUnauthorized, "Token is invalid")
		return "", err
	}
	log.Println("getUsernameFromToken: username= ", username, " token= ", userToken)
	return username, nil
}

func getUserRoleFromUsername(username string, w http.ResponseWriter, db *mongo.Client) (string, error) {
	log.Println("getUserRoleFromUsername: Getting role from username= ", username)
	role, err := getUserRole(username, db)
	if err != nil {
		log.Println("getUserRoleFromUsername: Failed to get role of user= ", username)
		log.Println("getUserRoleFromUsername: Returning status code 500")
		WriteErrorResponse(w, http.StatusInternalServerError, "failed to get role of user")
		return "", err
	}
	return role, nil
}

func checkTeacherRole(role string, w http.ResponseWriter, username string) error {
	if role != "teacher" {
		log.Println("checkTeacherRole: User= ", username, " does not have teacher role, role= ", role)
		log.Println("checkTeacherRole: Returning status code 403")
		WriteErrorResponse(w, http.StatusForbidden, "User does not have permission for this request")
		return fmt.Errorf("user does not have teacher role")
	}
	log.Println("checkTeacherRole: user role= ", role, " is teacher")
	return nil
}

func checkStudentRole(role string, w http.ResponseWriter, username string) error {
	if role != "student" {
		log.Println("checkStudentRole: User= ", username, " does not have student role, role= ", role)
		log.Println("checkStudentRole: Returning status code 403")
		WriteErrorResponse(w, http.StatusForbidden, "User does not have permission for this request")
		return fmt.Errorf("user does not have student role")
	}
	log.Println("checkStudentRole: user role= ", role, " is student")
	return nil
}

func getUserRole(username string, db *mongo.Client) (string, error) {
	log.Println("Function GetUserRole called")
	// Get a handle to the "user" collection.
	collection := db.Database("BrainBoard").Collection("user")

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
