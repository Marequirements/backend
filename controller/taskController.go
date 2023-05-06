package controller

import (
	"back-end/model"
	"back-end/token"
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"strings"
)

type TaskController struct {
	db *mongo.Client
	ts *token.Storage
	uc *StudentController
}

func NewTaskController(db *mongo.Client, ts *token.Storage, uc *StudentController) *TaskController {
	return &TaskController{db: db, ts: ts, uc: uc}
}

/*
	func GetAllTasks() ([]TaskController, error) {
		client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb+srv://mareklescinsky:EUFZTs6jcdkqqEUk@brainboard.lrpc8h3.mongodb.net/test"))
		if err != nil {
			return nil, err
		}
		defer client.Disconnect(context.Background())

		collection := client.Database("BrainBoard").Collection("task")
		cursor, err := collection.Find(context.Background(), bson.D{})
		if err != nil {
			return nil, err
		}
		defer cursor.Close(context.Background())

		var tasks []TaskController
		for cursor.Next(context.Background()) {
			var task TaskController
			err := cursor.Decode(&task)
			if err != nil {
				return nil, err
			}
			tasks = append(tasks, task)
		}
		if err := cursor.Err(); err != nil {
			return nil, err
		}
		return tasks, nil
	}

	func AddTask(task TaskController) error {
		client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb+srv://mareklescinsky:EUFZTs6jcdkqqEUk@brainboard"+
			".lrpc8h3.mongodb.net/test"))
		if err != nil {
			return err
		}
		defer client.Disconnect(context.Background())

		collection := client.Database("BrainBoard").Collection("task")
		_, err = collection.InsertOne(context.Background(), task)
		if err != nil {
			return err
		}
		return nil
	}

	func DeleteTask(id string) error {
		client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb+srv://mareklescinsky:EUFZTs6jcdkqqEUk@brainboard"+
			".lrpc8h3.mongodb.net/test"))
		if err != nil {
			return err
		}
		defer client.Disconnect(context.Background())

		collection := client.Database("BrainBoard").Collection("task")
		_, err = collection.DeleteOne(context.Background(), bson.D{{"_id", id}})
		if err != nil {
			return err
		}
		return nil
	}

	func EditTask(task TaskController) error {
		client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb+srv://mareklescinsky:EUFZTs6jcdkqqEUk@brainboard"+
			".lrpc8h3.mongodb.net/test"))
		if err != nil {
			return err
		}
		defer client.Disconnect(context.Background())

		collection := client.Database("BrainBoard").Collection("task")
		_, err = collection.UpdateOne(context.Background(), bson.D{{"_id", task.Id}}, bson.D{{"$set", bson.D{{"deadline", task.Deadline}, {"priority", task.Priority}, {"students", task.Students}, {"title", task.Title}, {"description", task.Description}, {"lesson", task.Lesson}}}})
		if err != nil {
			return err
		}
		return nil
	}
*/
func (tc *TaskController) GetTasksWithStatus3() ([]model.Task, error) {
	// Get a handle to the "tasks" collection.
	collection := tc.db.Database("BrainBoard").Collection("task")

	// Define the filter for tasks with students having status "3".
	filter := bson.M{"students": bson.M{"$elemMatch": bson.M{"status": "3"}}}

	// Execute the query.
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	// Decode the results.
	var tasks []model.Task
	if err := cursor.All(context.Background(), &tasks); err != nil {
		return nil, err
	}
	fmt.Println("Tasks with status 3:", tasks) // Debug print
	return tasks, nil
}
func (tc *TaskController) GetTasks(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header not provided", http.StatusBadRequest)
		return
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")

	// Get the username associated with the token
	username, err := tc.uc.ts.GetUsernameByToken(token)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Check if the user has the teacher role
	role, err := tc.uc.GetUserRole(username)
	if err != nil || role != "teacher" {
		http.Error(w, "Unauthorized access", http.StatusForbidden)
		return
	}

	// Get tasks with status 3
	tasks, err := tc.getFilteredTasks()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(tasks); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (tc *TaskController) getFilteredTasks() ([]model.TaskWithStudent, error) {
	// Get tasks with status 3
	tasks, err := tc.GetTasksWithStatus3()
	if err != nil {
		return nil, err
	}

	// Extract required information
	filteredTasks := make([]model.TaskWithStudent, 0)
	for _, task := range tasks {
		for _, student := range task.Students {
			if student.Status == "3" {
				filteredTask := model.TaskWithStudent{
					StudentID: student.StudentID,
					TaskName:  task.Title,
				}
				filteredTasks = append(filteredTasks, filteredTask)
			}
		}
	}
	fmt.Println("Filtered tasks:", filteredTasks) // Debug print

	return filteredTasks, nil
}
