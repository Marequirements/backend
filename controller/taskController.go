package controller

import (
	"back-end/model"
	"back-end/token"
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
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

func (tc *TaskController) HandleTeacherTasks(w http.ResponseWriter, r *http.Request) {
	log.Println("Function HandleTeacherTasks called")

	log.Println("HandleTeacherTasks: Getting auth header")
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		log.Println("HandleTeacherTasks: Failed to get auth headr= ", authHeader)
		respondWithError(w, http.StatusBadRequest, "Authorization header not provided")
		return
	}
	log.Println("HandleTeacherTasks: Got authHeader= ", authHeader)
	log.Println("HandleTeacherTasks: Getting token")
	userToken := strings.TrimPrefix(authHeader, "Bearer ")
	log.Println("HandleTeacherTasks: Token= ", userToken)

	log.Println("HandleTeacherTasks: Getting username from token")
	username, err := tc.ts.GetUsernameByToken(userToken)
	if err != nil {
		log.Println("HandleTeacherTasks: Failed to get username from token= ", userToken)
		log.Println("HandleTeacherTasks: Returning status code 401")
		respondWithError(w, http.StatusUnauthorized, "Token is invalid")
		return
	}

	log.Println("HandleTeacherTasks: GEtting role from username= ", username)
	role, err := tc.GetUserRole(username)
	if err != nil {
		log.Println("HandleTeacherTasks: Failed to get role of user= ", username)
		log.Println("HandleTeacherTasks: Returning status code 500")
		respondWithError(w, http.StatusInternalServerError, "failed to get role of user")
		return
	}
	if role != "teacher" {
		log.Println("HandleTeacherTasks: User= ", username, " does not have teacher role, role= ", role)
		log.Println("HandleTeacherTasks: Returning status code 403")
		respondWithError(w, http.StatusForbidden, "User does not have permission for this request")
		return
	}

	log.Println("HandleTeacherTasks: Getting path variable")
	classTitle := chi.URLParam(r, "classTitle")

	log.Println("HandleTeacherTasks: Path cariable classTitle= ", classTitle)
	log.Println("HandleTeacherTasks: Getting classid for classTitle")
	classId, err := tc.GetClassIdByClassTitle(classTitle)
	if err != nil {
		log.Println("HandleTeacherTasks: Failed to get classid for class= ", classTitle)
		log.Println("HandleTeacherTasks: Returned status code 500")
		respondWithError(w, http.StatusInternalServerError, "Failed to get classid from class")
		return
	}

	log.Println("HandleTeacherTasks: classtitle= ", classTitle, " classId= ", classId)

	log.Println("HandleTeacherTasks: Getting userid from user= ", username)
	userId, err := tc.GetIdByUsername(username)
	if err != nil {
		log.Println("HandleTeacherTasks: Failed to get userID from user= ", username)
		log.Println("HandleTeacherTasks: Returning status code 500")
		respondWithError(w, http.StatusInternalServerError, "Failed to get userid from user")
		return
	}
	log.Println("HandleTeacherTasks: username= ", username, " userID= ", userId)

	log.Println("HandleTeacherTasks: Getting getting tasks for class= ", classTitle, " user= ", username)
	tasks, err := tc.GetClassTask(*classId, *userId)
	if err != nil {
		log.Println("HandleTeacherTasks: Failed to get tasks from classid= ", classId, " userID= ", userId)
		respondWithError(w, http.StatusInternalServerError, "Failed to get tasks from classID and userID")
		return
	}
	log.Println("HandleTeacherTasks: Returning tasks= ,", tasks)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(tasks); err != nil {
		log.Println("GetTasks: Failed to write tasks to body, tasks= ", tasks)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println("HandleTeacherTasks: Tasks sent succesfully")
}

func (tc *TaskController) GetClassTask(class primitive.ObjectID, teacher primitive.ObjectID) ([]model.ClassTasks, error) {
	log.Println("Function GetClassTask called")

	//get connection
	collection := tc.db.Database("BrainBoard").Collection("task")
	// Lookup the subject based on the teacher and class IDs
	lookupStage := bson.D{
		{"$lookup", bson.D{
			{"from", "subject"},
			{"localField", "subject"},
			{"foreignField", "_id"},
			{"as", "subject"},
		}},
	}

	unwindStage := bson.D{
		{"$unwind", "$subject"},
	}

	// Match the tasks that have the specified teacher and class
	matchStage := bson.D{
		{"$match", bson.D{
			{"subject.teacher", teacher},
			{"subject.class", class},
		}},
	}

	// Pipeline for aggregation
	pipeline := mongo.Pipeline{lookupStage, unwindStage, matchStage}

	// Perform the aggregation
	cursor, err := collection.Aggregate(context.Background(), pipeline)
	if err != nil {
		log.Fatal(err)
	}

	var tasks []model.ClassTasks

	// Iterate over the cursor and print the filtered tasks
	defer cursor.Close(context.Background())
	//for cursor.Next(context.Background()) {
	//	var task bson.M
	//	if err := cursor.Decode(&task); err != nil {
	//		log.Fatal(err)
	//	}
	//	fmt.Println(task)
	//}
	if err := cursor.All(context.Background(), &tasks); err != nil {
		log.Println("GetClassTasks: Failed saving tasks Error: ", err)
		return nil, err
	}

	if err := cursor.Err(); err != nil {
		log.Fatal(err)
	}

	log.Println("GetClassTasks: Returned tasks= ", tasks)
	return tasks, nil

}

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

func (tc *TaskController) GetUserRole(username string) (string, error) {
	log.Println("Function GetUserRole called")
	// Get a handle to the "user" collection.
	collection := tc.db.Database("BrainBoard").Collection("user")

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

func (tc *TaskController) GetIdByUsername(username string) (*primitive.ObjectID, error) {
	log.Println("Function GetIdByUsername called")
	collection := tc.db.Database("BrainBoard").Collection("user")

	filter := bson.M{"username": username}
	var id struct {
		ID primitive.ObjectID `bson:"_id"`
	}

	log.Println("GetIdByUsername: Searching for user= ", username, " in database")
	err := collection.FindOne(context.Background(), filter).Decode(&id)
	if err != nil {
		log.Println("GetIdByUsername: Failed to find user= ", username)
		return nil, err
	}

	log.Println("GetIdByUsername: Found user= ", username, " in database and id= ", id)
	log.Println("GetIdByUsername: Returned values id= ", id, " error= nil")
	return &id.ID, nil
}

func (tc *TaskController) GetClassIdByClassTitle(classTitle string) (*primitive.ObjectID, error) {
	log.Println("Function GetClassIdByClassTitle called")
	collection := tc.db.Database("BrainBoard").Collection("class")

	filter := bson.M{"name": classTitle}
	//var id *primitive.ObjectID

	var id struct {
		ID primitive.ObjectID `bson:"_id"`
	}

	log.Println("GetClassIdByClassTitle: Searching for class= ", classTitle, " in database")
	err := collection.FindOne(context.Background(), filter).Decode(&id)
	if err != nil {
		log.Println("GetClassIdByClassTitle: Failed to find class= ", classTitle, " Error: ", err)
		return nil, err
	}

	log.Println("GetClassIdByClassTitle: Found class= ", classTitle, " in database and id= ", id)
	log.Println("GetClassIdByClassTitle: Returned values id= ", id, " error= nil")
	return &id.ID, nil
}
