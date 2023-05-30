package controller

import (
	"back-end/token"
	"back-end/util"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"net/http"
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

type DashboardResponse struct {
	Students string `json:"students"`
	Tasks    string `json:"tasks"`
	Review   string `json:"review"`
}

func (dc *DashBoardController) HandleTeacherDashBoard(w http.ResponseWriter, r *http.Request) {
	_, err := util.TeacherLogin("HandleTeacherDashBoard", dc.db, dc.ts, w, r)
	if err != nil {
		util.WriteErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userCollection := dc.db.Database("BrainBoard").Collection("user")
	taskCollection := dc.db.Database("BrainBoard").Collection("task")

	// Counting students in the database
	studentsCount, err := userCollection.CountDocuments(context.Background(), bson.M{"role": "student"})
	if err != nil {
		log.Println("HandleTeacherDashBoard: failed to get number of students")
		util.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to get number of students in database")
		return
	}

	// Counting tasks in the database
	tasksCount, err := taskCollection.CountDocuments(context.Background(), bson.M{})
	if err != nil {
		log.Println("HandleTeacherDashBoard: failed to get number of tasks")
		util.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to get number of tasks in database")
		return
	}

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

	resp := DashboardResponse{
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
