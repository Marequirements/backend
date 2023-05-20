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
func (tc *TaskController) GetTasksWithStatus3() ([]model.TaskWithStudent, error) {
	log.Println("Function GetTaskWithStatus3 called")
	// Get a handle to the "tasks" collection.
	collection := tc.db.Database("BrainBoard").Collection("task")

	// Define the filter for tasks with students having status "3".
	filter := bson.M{"students": bson.M{"$elemMatch": bson.M{"status": "3"}}}

	log.Println("GetTaskWithStatus3: Searching for tasks with status 3 in task collection")

	// Execute the query.
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		log.Println("GetTaskWithStatus3: Error with finding tasks with status 3")
		return nil, err
	}
	defer cursor.Close(context.Background())

	log.Println("GetTaskWithStatus3: Saving found tasks")

	// Decode the results.
	var tasks []model.Task
	if err := cursor.All(context.Background(), &tasks); err != nil {
		log.Println("GetTaskWithStatus3: Failed saving tasks")
		return nil, err
	}

	// Create a map to store the subject titles
	subjectTitleMap := make(map[primitive.ObjectID]string)

	// Change collection to the "subject" collection.
	collection = tc.db.Database("BrainBoard").Collection("subject")

	// Retrieve subjects for each task
	for _, task := range tasks {
		var subject model.Subject
		err = collection.FindOne(context.Background(), bson.M{"_id": task.Subject}).Decode(&subject)
		if err != nil {
			log.Println("Failed to fetch subject for task")
			return nil, err
		}
		subjectTitleMap[task.Subject] = subject.Title
	}

	log.Println("GetTaskWithStatus3:Found tasks with status 3", tasks)

	// Now we can construct our response
	var response []model.TaskWithStudent

	// Get a handle to the "students" collection.
	studentCollection := tc.db.Database("BrainBoard").Collection("user")

	for _, task := range tasks {
		for _, studentStatus := range task.Students {
			if studentStatus.Status == "3" {
				// Fetch student details
				var student model.Student
				err := studentCollection.FindOne(context.Background(), bson.M{"_id": studentStatus.StudentID}).Decode(&student)
				if err != nil {
					log.Println("Failed to fetch student for task")
					return nil, err
				}

				response = append(response, model.TaskWithStudent{
					StudentID:   studentStatus.StudentID,
					TaskName:    task.Title,
					Subject:     subjectTitleMap[task.Subject],
					Description: task.Description,
					Name:        student.Name,
					Surname:     student.Surname,
					Deadline:    task.Deadline,
				})
			}
		}
	}

	return response, nil
}

func (tc *TaskController) GetTasks(w http.ResponseWriter, r *http.Request) {
	log.Println("Function GetTasks called")

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		log.Println("GetTasks: Failed to get authHeader= ", authHeader)

		http.Error(w, "Authorization header not provided", http.StatusBadRequest)
		return
	}
	log.Println("GetTasks: Got authHeader= ", authHeader)

	log.Println("GetTasks: Getting token")

	authToken := strings.TrimPrefix(authHeader, "Bearer ")

	log.Println("GetTasks: Got token= ", authToken)

	log.Println("GetTasks: Getting username from token ", authToken)
	// Get the username associated with the authToken
	username, err := tc.uc.ts.GetUsernameByToken(authToken)
	if err != nil {
		log.Println("GetTasks: Failed could not get the username from token ", authToken)
		http.Error(w, "Invalid authToken", http.StatusUnauthorized)
		return
	}
	log.Println("GetTasks: Got username= ", username, " from token= ", authToken)

	log.Println("GetTasks: Checking the role of user= ", username)

	// Check if the user has the teacher role
	role, err := tc.uc.GetUserRole(username)
	if err != nil {
		log.Println("GetTasks: Failed to get user role for user= ", username)
		http.Error(w, "Unauthorized access", http.StatusInternalServerError)
		return
	}
	if role != "teacher" {
		log.Println("GetTasks: user= ", username, " does not have teacher role, role= ", role)
		http.Error(w, "Unauthorized access", http.StatusForbidden)
		return
	}
	log.Println("GetTasks: user= ", username, " does have teacher role, role= ", role)

	log.Println("GetTasks: Getting tasks with status 3")
	// Get tasks with status 3
	tasks, err := tc.GetTasksWithStatus3()
	if err != nil {
		log.Println("GetTasks: Failed to get tasks")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println("GetTasks: got tasks= ", tasks)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(tasks); err != nil {
		log.Println("GetTasks: Failed to write tasks to body, tasks= ", tasks)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println("GetTasks: Tasks sent successfully")
}
