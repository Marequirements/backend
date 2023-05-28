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

func (dc *DashBoardController) HandleTeacherDashBoard(w http.ResponseWriter, r *http.Request) {
	_, err := util.TeacherLogin("HandleTeacherDashBoard", dc.db, dc.ts, w, r)
	if err != nil {
		return
	}
	userCollection := dc.db.Database("BrainBoard").Collection("user")
	taskCollection := dc.db.Database("BrainBoard").Collection("task")

	log.Println("HandleTeacherDashBoard: Counting students in database")
	filter := bson.M{"role": "student"}
	count, err := userCollection.CountDocuments(context.Background(), filter)
	if err != nil {
		log.Println("HandleTeacherDashBoard: failed to get number of students")
		util.WriteErrorResponse(w, 500, "Failed to get number of students in database")
	}
	students := strconv.FormatInt(count, 10)
	log.Println("HandleTeacherDashBoard: number of students= ", students)

	log.Println("HandleTeacherDashBoard: Counting tasks in database")
	count, err = taskCollection.CountDocuments(context.Background(), bson.M{})
	if err != nil {
		log.Println("HandleTeacherDashBoard: failed to get number of students")
		util.WriteErrorResponse(w, 500, "Failed to get number of tasks in database")
	}
	tasks := strconv.FormatInt(count, 10)
	log.Println("HandleTeacherDashBoard: number of tasks= ", tasks)

	log.Println("HandleTeacherDashBoard: Counting number of students with tasks in state 3 for each task from database")

	pipeline := mongo.Pipeline{
		{{"$unwind", "$students"}},
		{{"$match", bson.D{{"students.status", "3"}}}},
		{{"$count", "count"}},
	}

	cursor, err := taskCollection.Aggregate(context.Background(), pipeline)

	if err != nil {
		log.Fatal(err)
	}

	defer cursor.Close(context.Background())
	if cursor.Next(context.Background()) {
		var result bson.M
		if err := cursor.Decode(&result); err != nil {
			log.Fatal(err)
		}

		count := result["count"]
		countInt := count.(int32)
		countStr := strconv.Itoa(int(countInt)) // convert integer count to string
		resp := struct {
			Students string `json:"students"`
			Tasks    string `json:"tasks"`
			Review   string `json:"review"`
		}{
			Students: students,
			Tasks:    tasks,
			Review:   countStr,
		}
		log.Println("HandleTeacherDashBoard: number of tasks with state 3= ", countStr)
		log.Println("HandleTeacherDashBoard: Sending all data= ", resp)
		log.Println("HandleTeacherDashBoard: Sent response 200")

		util.WriteSuccessResponse(w, 200, resp)
		if err = cursor.Err(); err != nil {
			log.Fatal(err)
		}
	}
}
