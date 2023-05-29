package controller

import (
	"back-end/model"
	"back-end/token"
	"back-end/util"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"net/http"
	"strings"
	"time"
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
	username, err := util.TeacherLogin("HandleTeacherTasks", tc.db, tc.ts, w, r)
	if err != nil {
		return
	}

	log.Println("HandleTeacherTasks: Getting path variable")
	classTitle := chi.URLParam(r, "classTitle")

	log.Println("HandleTeacherTasks: Path variable classTitle= ", classTitle)
	log.Println("HandleTeacherTasks: Getting classid from classTitle")
	classId, err := tc.GetClassIdByClassTitle(classTitle)
	if err != nil {
		log.Println("HandleTeacherTasks: Failed to get classid for class= ", classTitle)
		log.Println("HandleTeacherTasks: Returned status code 400")
		respondWithError(w, http.StatusBadRequest, "Failed to get classid from path variable classTitle")
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
		log.Println("HandleTeacherTasks: Returning status code 500")
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
	log.Println("HandleTeacherTasks: Returning status code 200")

}

func (tc *TaskController) HandleAddTask(w http.ResponseWriter, r *http.Request) {
	_, err := util.TeacherLogin("HandleAddTask", tc.db, tc.ts, w, r)
	if err != nil {
		return
	}
	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Deadline    string `json:"deadline"`
		Class       string `json:"class"`
		Subject     string `json:"subject"`
	}

	log.Println("HandleAddTask: Getting data from body")
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		util.WriteErrorResponse(w, 400, "JSON parameters not provided")
	}

	log.Println("HandleAddTask: Adding task to database, task")
	err = tc.AddTask(req.Title, req.Description, req.Deadline, req.Subject, req.Class)
	if err != nil {
		if err.Error() == "subject does not exist" {
			log.Println("HandleAddTask: subject= ", req.Subject, "is not in database")
			util.WriteErrorResponse(w, 404, "Subject does not exist")
		}
		if err.Error() == "class does not exist" {
			log.Println("HandleAddTask: class= ", req.Class, "is not in database")
			util.WriteErrorResponse(w, 404, "Class does not exist")
		}
		log.Println("HandleAddTask: Failed to add task to database from request body", req)
		util.WriteErrorResponse(w, 500, "Failed to add task to database")
	}

	log.Println("HandleAddTask: Task added to database, task")
	util.WriteSuccessResponse(w, 201, "")
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
			{"as", "subjectDetails"},
		}},
	}

	unwindStage := bson.D{
		{"$unwind", "$subjectDetails"},
	}

	// Match the tasks that have the specified teacher and class
	matchStage := bson.D{
		{"$match", bson.D{
			{"class", class},
			{"subjectDetails.teacher", teacher},
		}},
	}

	// Pipeline for aggregation
	pipeline := mongo.Pipeline{lookupStage, unwindStage, matchStage}

	// Perform the aggregation
	cursor, err := collection.Aggregate(context.Background(), pipeline)
	if err != nil {
		log.Fatal(err)
	}

	var aggregationResults []model.TaskAggregationResult

	// Iterate over the cursor and print the filtered tasks
	defer cursor.Close(context.Background())

	if err := cursor.All(context.Background(), &aggregationResults); err != nil {
		log.Println("GetClassTasks: Failed saving tasks Error: ", err)
		return nil, err
	}

	var tasks []model.ClassTasks
	for _, result := range aggregationResults {
		tasks = append(tasks, model.ClassTasks{
			ID:          result.ID,
			Title:       result.Title,
			Description: result.Description,
			Deadline:    result.Deadline,
			Subject:     result.Subject.Title,
		})
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

	if !tc.ClassExists(classTitle) {
		return nil, fmt.Errorf("class does not exist")
	}

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

func (tc *TaskController) GetSubjectIdBySubjectTitle(subjectTitle string) (*primitive.ObjectID, error) {
	log.Println("Function GetSubjectIdBySubjectTitle called")
	collection := tc.db.Database("BrainBoard").Collection("subject")

	if !tc.SubjectExists(subjectTitle) {
		return nil, fmt.Errorf("subject does not exist")
	}
	filter := bson.M{"title": subjectTitle}
	//var id *primitive.ObjectID

	var id struct {
		ID primitive.ObjectID `bson:"_id"`
	}

	log.Println("GetSubjectIdBySubjectTitle: Searching for class= ", subjectTitle, " in database")
	err := collection.FindOne(context.Background(), filter).Decode(&id)
	if err != nil {
		log.Println("GetSubjectIdBySubjectTitle: Failed to find class= ", subjectTitle, " Error: ", err)
		return nil, err
	}

	log.Println("GetSubjectIdBySubjectTitle: Found class= ", subjectTitle, " in database and id= ", id)
	log.Println("GetSubjectIdBySubjectTitle: Returned values id= ", id, " error= nil")
	return &id.ID, nil
}

func (tc *TaskController) SubjectExists(title string) bool {
	log.Println("Function SubjectExists called")

	collection := tc.db.Database("BrainBoard").Collection("subject")

	_, err := collection.Find(context.Background(), bson.M{"title": title})
	if err == mongo.ErrNoDocuments {
		log.Println("SubjectExists: Returned false")
		return false
	}
	log.Println("SubjectExists: Returned true")
	return true

}

func (tc *TaskController) ClassExists(name string) bool {
	log.Println("Function ClassExists called")

	collection := tc.db.Database("BrainBoard").Collection("class")

	log.Println("ClassExists: Searching for class= ", name)
	_, err := collection.Find(context.Background(), bson.M{"name": name})
	if err == mongo.ErrNoDocuments {
		log.Println("ClassExists: Failed to find class=", name)
		log.Println("ClassExists: Returned false")
		return false
	}
	log.Println("ClassExists: Class= ", name, " is in database")
	log.Println("ClassExists: Returned true")
	return true

}

func (tc *TaskController) AddTask(title string, description string, deadline string, subject string, class string) error {
	log.Println("Function AddTask called")
	usercollection := tc.db.Database("BrainBoard").Collection("user")
	taskCollection := tc.db.Database("BrainBoard").Collection("task")

	classId, err := tc.GetClassIdByClassTitle(class)

	if err != nil {
		if err.Error() == "class does not exist" {
			log.Println("AddTask: Failed to get class id from class= ", class, " returned error=", err)
			return err
		}
		log.Println("AddTask: Failed to get class id from class= ", class, " returned error=", err)
		return err
	}

	subjectId, err := tc.GetSubjectIdBySubjectTitle(subject)
	if err != nil {
		if err.Error() == "subject does not exist" {
			log.Println("AddTask: Failed to get subject id from subject= ", subject, " returned error=", err)
			return err
		}
		log.Println("AddTask: Failed to get subject id from subject= ", subject)
		return err
	}
	var users []model.Student
	cursor, err := usercollection.Find(context.Background(), bson.M{"role": "student", "class": classId})
	if err != nil {
		log.Fatal(err)
		return err
	}
	if err = cursor.All(context.Background(), &users); err != nil {
		log.Fatal(err)
		return err
	}

	t, err := time.Parse("2006-01-02", deadline)
	if err != nil {
		log.Fatal(err)
		return err
	}

	date := primitive.NewDateTimeFromTime(t)
	var studentStatus []model.StudentStatus
	for _, user := range users {
		studentStatus = append(studentStatus, model.StudentStatus{StudentID: user.Id, Status: "0"})
	}

	task := model.NewTask{
		Title:       title,
		Description: description,
		Deadline:    date,
		Subject:     *subjectId,
		Students:    studentStatus,
		Class:       *classId,
	}

	_, err = taskCollection.InsertOne(context.Background(), task)
	if err != nil {
		log.Fatal(err)
		return err
	}
	log.Println("AddTask: added task to database, task= ", task)
	return nil
}
