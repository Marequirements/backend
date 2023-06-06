package controller

import (
	"back-end/model"
	"back-end/token"
	"back-end/util"
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"net/http"
	"sort"
	"strconv"
)

type DashBoardController struct {
	db *mongo.Client
	ts *token.Storage
	uc *StudentController
	tc *TaskController
}

func NewDashBoardController(db *mongo.Client, ts *token.Storage, uc *StudentController, tc *TaskController) *DashBoardController {
	return &DashBoardController{db: db, ts: ts, uc: uc, tc: tc}
}

type TeacherDashboardResponse struct {
	Students string `json:"students"`
	Tasks    string `json:"tasks"`
	Review   string `json:"review"`
}

type StudentDashboardResponse struct {
	Subjects     []string                     `json:"subjects,omitempty"`
	SubjectTasks string                       `json:"subjectTasks,omitempty"`
	Todo         []model.StudentDashboardTask `json:"todo"`
	InProgress   []model.StudentDashboardTask `json:"inProgress"`
	Review       []model.StudentDashboardTask `json:"review"`
	Done         []model.StudentDashboardTask `json:"done"`
}

func (dc *DashBoardController) HandleTeacherDashBoard(w http.ResponseWriter, r *http.Request) {
	_, err := util.TeacherLogin(dc.db, dc.ts, w, r)
	if err != nil {
		return
	}

	userCollection := dc.db.Database("BrainBoard").Collection("user")
	taskCollection := dc.db.Database("BrainBoard").Collection("task")

	log.Println("HandleTeacherDashBoard: counting number of students")
	// Counting students in the database
	studentsCount, err := userCollection.CountDocuments(context.Background(), bson.M{"role": "student"})
	if err != nil {
		log.Println("HandleTeacherDashBoard: failed to get number of students")
		util.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to get number of students in database")
		return
	}

	log.Println("HandleTeacherDashBoard: Counting number of tasks")
	// Counting tasks in the database
	tasksCount, err := taskCollection.CountDocuments(context.Background(), bson.M{})
	if err != nil {
		log.Println("HandleTeacherDashBoard: failed to get number of tasks")
		util.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to get number of tasks in database")
		return
	}

	log.Println("HandleTeacherDashBoard: Counting students with tasks in state 3 for each task from the database")

	// Counting number of students with tasks in state 3 for each task from the database
	pipeline := mongo.Pipeline{
		{{"$unwind", "$students"}},
		{{"$match", bson.D{{"students.status", "3"}}}},
		{{"$count", "count"}},
	}

	cursor, err := taskCollection.Aggregate(context.Background(), pipeline)
	if err != nil {
		log.Println("HandleTeacherDashBoard: failed to aggregate tasks")
		util.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to aggregate tasks in database")
		return
	}
	defer cursor.Close(context.Background())

	resp := TeacherDashboardResponse{
		Students: strconv.FormatInt(studentsCount, 10),
		Tasks:    strconv.FormatInt(tasksCount, 10),
		Review:   "0", // Default value if no tasks with state 3 found
	}

	if cursor.Next(context.Background()) {
		var result bson.M
		if err := cursor.Decode(&result); err != nil {
			log.Println("HandleTeacherDashBoard: failed to decode cursor result")
			util.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to process task result")
			return
		}

		count := result["count"]
		countInt, ok := count.(int32)
		if !ok {
			log.Println("HandleTeacherDashBoard: invalid count type")
			util.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to process task count")
			return
		}

		resp.Review = strconv.Itoa(int(countInt))
	}

	log.Println("HandleTeacherDashBoard: Sending response:", resp)
	util.WriteSuccessResponse(w, http.StatusOK, resp)
}

// allowed status numbers{1,2,3,4}
func (dc *DashBoardController) HandleStatusChange(w http.ResponseWriter, r *http.Request) {
	username, err := util.StudentLogin(dc.db, dc.ts, w, r)
	if err != nil {
		return
	}

	var task model.UpdateStatusTask

	log.Println("HandleStatusChange:Decoding request body")
	err = json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		log.Println("HandleStatusChange: Failed to decode body")
		util.WriteErrorResponse(w, 400, "Wrong JSON format")
		return
	}
	if task.ID == "" {
		log.Println("HandleStatusChange: task id not in body")
		util.WriteErrorResponse(w, 400, "Wrong JSON format")
		return
	}

	log.Println("HandleStatusChange: Validating format of task id")
	taskID, err := primitive.ObjectIDFromHex(task.ID)
	if err != nil {
		log.Println("HandleStatusChange: Failed to get taskID from JSON")
		util.WriteErrorResponse(w, 400, "wrong taskID format")
		return
	}
	//if task.ID.IsZero() {
	//	log.Println("HandleStatusChange: Failed to get taskID from JSON")
	//	util.WriteErrorResponse(w, 400, "wrong taskID format")
	//	return
	//}

	log.Println("HandleStatusChange: Checking if task exist in database")
	exist := dc.TaskExist(taskID)
	if !exist {
		log.Println("HandleStatusChange: Task not in database")
		util.WriteErrorResponse(w, 404, " taskID not in database")
		return
	}

	log.Println("HandleStatusChange: Converting status to int and validating the value")
	status, err := strconv.Atoi(task.Status)
	if err != nil {
		log.Println("HandleStatusChange: Failed to convert status to int")
		util.WriteErrorResponse(w, 400, "Wrong status code")
		return
	}

	if status < 1 || status > 4 {
		log.Println("HandleStatusChange: Incorrect status code", status)
		util.WriteErrorResponse(w, 400, "Wrong status code")
		return
	}

	log.Println("HandleStatusChange: Getting student id")
	studentID, err := dc.GetIdByUsername(username)
	if err != nil {
		log.Println("HandleStatusChange: Studedent id not found in database")
		util.WriteErrorResponse(w, 500, "Failed to get student id")
		return
	}

	log.Println("HandleStatusChange: Updating status")
	err = dc.ChangeStatus(taskID, *studentID, task.Status)
	if err != nil {
		log.Println("HandleStatusChange: Failed to update the status")
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(204)
	log.Println("HandleStatusChange: Status updated successfully")
}

func (dc *DashBoardController) HandleStudentDashboard(w http.ResponseWriter, r *http.Request) {
	username, err := util.StudentLogin(dc.db, dc.ts, w, r)
	if err != nil {
		return
	}

	userCollection := dc.db.Database("BrainBoard").Collection("user")
	taskCollection := dc.db.Database("BrainBoard").Collection("task")
	subjectCollection := dc.db.Database("BrainBoard").Collection("subject")

	log.Println("HandleStudentDashboard: getting id of username= ", username)
	studentId, err := dc.GetIdByUsername(username)
	if err != nil {
		util.WriteErrorResponse(w, 500, "Failed to get user id for username= ")
	}
	log.Println("HandleStudentDashboard: found user= ", username, " id=", studentId)

	log.Println("HandleStudentDashboard: getting user from database")
	//get user and class id
	var student model.Student
	userCollection.FindOne(context.Background(), bson.M{"_id": studentId}).Decode(&student)
	log.Println("HandleStudentDashboard: Got user from database")

	log.Println("HandleStudentDashboard: Getting Path variable")
	SubjTitleURL := chi.URLParam(r, "subjectTitle")
	if SubjTitleURL == "" {
		log.Println("HandleStudentDashboard: Path variable is empty/default dashboard")

		log.Println("HandleStudentDashboard: Getting subjects by user class")
		subjects, err := dc.GetSubjectsFromCLassID(student.Class, subjectCollection)
		if err != nil {
			util.WriteErrorResponse(w, 500, "Failed to get subjects")
		}
		log.Println("HandleStudentDashboard: subjects = ", subjects, " class", student.Class)

		log.Println("HandleStudentDashboard: Sorting Subjects alphabetically")
		// Sort subjects by title
		sort.Slice(subjects, func(i, j int) bool {
			return subjects[i].Title < subjects[j].Title
		})

		log.Println("HandleStudentDashboard: subject sorted")

		log.Println("HandleStudentDashboard: Extracting subjects titles from subject structs")
		// Extract subject titles to separate slice
		subjectTitles := make([]string, len(subjects))
		for i, s := range subjects {
			subjectTitles[i] = s.Title
		}
		log.Println("HandleStudentDashboard: got titles= ", subjectTitles)

		// Construct response with placeholders for tasks
		response := &StudentDashboardResponse{
			Subjects:     subjectTitles,
			SubjectTasks: subjectTitles[0],
			Todo:         make([]model.StudentDashboardTask, 0),
			InProgress:   make([]model.StudentDashboardTask, 0),
			Review:       make([]model.StudentDashboardTask, 0),
			Done:         make([]model.StudentDashboardTask, 0),
		}

		log.Println("HandleStudentDashboard: Getting tasks that student has")
		tasks, err := dc.GetTasksFromUserID(subjects[0].Id, student.Id, taskCollection)
		if err != nil {
			log.Println("HandleStudentDashboard: Failed to get tasks")
			util.WriteErrorResponse(w, 500, "Failed to get tasks")
		}
		log.Println("HandleStudentDashboard: Got tasks")

		log.Println("HandleStudentDashboard: Sorting tasks by state")
		dc.SortTasks(response, tasks, studentId)

		log.Println("HandleStudentDashboard: sending response")
		w.Header().Set("Content-Type", "application/json")
		// Encode and write response
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}
	log.Println("HandleStudentDashboard: Path variable= ", SubjTitleURL)
	exists, id := dc.SubjectExists(SubjTitleURL)
	if exists {
		if id != primitive.NilObjectID {
			log.Println("HandleStudentDashboard: Getting subjects by user class")
			subjects, err := dc.GetSubjectsFromCLassID(student.Class, subjectCollection)
			if err != nil {
				util.WriteErrorResponse(w, 500, "Failed to get subjects")
			}
			log.Println("HandleStudentDashboard: subjects = ", subjects, " class", student.Class)

			log.Println("HandleStudentDashboard: Sorting Subjects alphabetically")
			// Sort subjects by title
			sort.Slice(subjects, func(i, j int) bool {
				return subjects[i].Title < subjects[j].Title
			})

			log.Println("HandleStudentDashboard: subject sorted")

			log.Println("HandleStudentDashboard: Extracting subjects titles from subject structs")
			// Extract subject titles to separate slice
			subjectTitles := make([]string, len(subjects))
			for i, s := range subjects {
				subjectTitles[i] = s.Title
			}
			log.Println("HandleStudentDashboard: got titles= ", subjectTitles)

			// Construct response with placeholders for tasks
			response := &StudentDashboardResponse{
				Subjects:     subjectTitles,
				SubjectTasks: SubjTitleURL,
				Todo:         make([]model.StudentDashboardTask, 0),
				InProgress:   make([]model.StudentDashboardTask, 0),
				Review:       make([]model.StudentDashboardTask, 0),
				Done:         make([]model.StudentDashboardTask, 0),
			}
			// Use subjectID here
			log.Println("SubjectID: ", id.Hex())
			tasks, err := dc.GetTasksFromUserID(id, student.Id, taskCollection)
			if err != nil {
				util.WriteErrorResponse(w, 500, "Failed to get tasks")
			}

			dc.SortTasks(response, tasks, studentId)

			w.Header().Set("Content-Type", "application/json")
			// Encode and write response
			err = json.NewEncoder(w).Encode(response)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			//subject exists but ID could not be retrieved
			log.Println("Subject exists but ID could not be retrieved.")
			util.WriteErrorResponse(w, 500, "Failed to get id of subject from url")
		}
	} else {
		log.Println("Subject does not exist.")
		util.WriteErrorResponse(w, 404, "Subject does not exists")
	}

}

func (dc *DashBoardController) GetIdByUsername(username string) (*primitive.ObjectID, error) {
	log.Println("Function GetIdByUsername called")
	collection := dc.db.Database("BrainBoard").Collection("user")

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

// get array of subjects by class id
func (dc *DashBoardController) GetSubjectsFromCLassID(classID primitive.ObjectID, collection *mongo.Collection) ([]model.Subject, error) {
	log.Println("Function GetSubjectsFromCLassID called")

	log.Println("GetSubjectsFromCLassID: Finding all subjects for class id= ", classID)
	// Find subjects for user's class
	cursor, err := collection.Find(context.Background(), bson.M{"class": classID})
	if err != nil {
		log.Println("GetSubjectsFromCLassID: Failed to get subject for class id= ", classID)
		return nil, err
	}

	var subjects []model.Subject
	for cursor.Next(context.Background()) {
		var subject model.Subject
		err := cursor.Decode(&subject)
		if err != nil {
			return nil, err
		}
		subjects = append(subjects, subject)
	}

	if err := cursor.Err(); err != nil {
		log.Println("GetSubjectsFromCLassID: Failed to save subjects for class id= ", classID)
		return nil, err
	}

	// Close the cursor once finished
	cursor.Close(context.Background())
	log.Println("GetSubjectsFromCLassID: Returned subjects=", subjects)
	return subjects, nil
}

// get array of tasks by user id
func (dc *DashBoardController) GetTasksFromUserID(subjectID primitive.ObjectID, userID primitive.ObjectID, collection *mongo.Collection) ([]model.StudentDashboardTask, error) {
	log.Println("Function GetTasksFromUserID called")

	log.Println("GetTasksFromUserID: Finding tasks in database")
	cursor, err := collection.Find(context.Background(), bson.M{"subject": subjectID, "students.studentid": userID})
	if err != nil {
		log.Println("GetTasksFromUserID: Failed to find tasks in database")
		return nil, err
	}

	log.Println("GetTasksFromUserID: Saving found tasks")
	var tasks []model.StudentDashboardTask
	for cursor.Next(context.Background()) {
		var task model.StudentDashboardTask
		err := cursor.Decode(&task)
		if err != nil {
			log.Println("GetTasksFromUserID: Failed to save tasks")
			return nil, err
		}
		tasks = append(tasks, task)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	log.Println("GetTasksFromUserID: Tasks returned,tasks= ", tasks)
	return tasks, nil
}

// Distribute tasks by status
func (dc *DashBoardController) SortTasks(response *StudentDashboardResponse, tasks []model.StudentDashboardTask, studentID *primitive.ObjectID) {
	log.Println("Function SortTasks called")
	for _, task := range tasks {
		for _, student := range task.Students {
			if student.StudentID == *studentID {
				switch student.Status {

				case "1":
					response.Todo = append(response.Todo, task)
				case "2":
					response.InProgress = append(response.InProgress, task)
				case "3":
					response.Review = append(response.Review, task)
				case "4":
					response.Done = append(response.Done, task)
				case "5":
					response.Todo = append(response.Todo, task)
				}
				break
			}
		}
	}
	log.Println("SortTasks: tasks sorted")
}

func (dc *DashBoardController) SubjectExists(title string) (bool, primitive.ObjectID) {
	log.Println("Function SubjectExists called")

	collection := dc.db.Database("BrainBoard").Collection("subject")

	log.Println("SubjectExists: Searching for subject =", title)
	cur, err := collection.Find(context.Background(), bson.M{"title": title})
	if err != nil {
		log.Println("SubjectExists: Failed to execute the find query:", err)
		return false, primitive.NilObjectID
	}
	defer cur.Close(context.Background())

	for cur.Next(context.Background()) {
		var result struct {
			ID primitive.ObjectID `bson:"_id"`
		}
		err := cur.Decode(&result)
		if err != nil {
			log.Println("SubjectExists: Failed to decode document:", err)
			return false, primitive.NilObjectID
		}
		// Document found, return true
		log.Println("SubjectExists: Subject =", title, "is in the database")
		log.Println("SubjectExists: Returned true")
		return true, result.ID
	}

	if cur.Err() != nil {
		log.Println("SubjectExists: Error occurred while iterating over the cursor:", cur.Err())
	}

	// No matching documents found, return false
	log.Println("SubjectExists: Failed to find subject =", title)
	log.Println("SubjectExists: Returned false")
	return false, primitive.NilObjectID
}

func (dc *DashBoardController) TaskExist(taskID primitive.ObjectID) bool {
	log.Println("Function TaskExist called")
	taskCollection := dc.db.Database("BrainBoard").Collection("task")

	filter := bson.M{"_id": taskID}

	n, err := taskCollection.CountDocuments(context.Background(), filter)
	if err != nil {
		log.Println("TaskExist: FAiled to execute querry")
	}
	if n == 0 {
		return false
	}
	return true
}

func (dc *DashBoardController) ChangeStatus(taskID primitive.ObjectID, studentID primitive.ObjectID, status string) error {
	log.Println("Function ChangeStatus called")
	taskCollection := dc.db.Database("BrainBoard").Collection("task")

	filter := bson.M{
		"_id":                taskID,
		"students.studentid": studentID,
	}

	update := bson.M{
		"$set": bson.M{
			"students.$.status": status,
		},
	}

	log.Println("ChangeStatus: Updating status code for task")
	_, err := taskCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Println("ChangeStatus: Failed to update status for task= ", taskID, " student= ", studentID)
		return err
	}

	log.Println("ChangeStatus: Status for student updated")
	return nil
}
