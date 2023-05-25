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
	//doeas not work
	filter = bson.M{
		"students.status": "3",
	}
	projectStage := bson.D{
		{"$project", bson.D{
			{"numStudents", bson.D{
				{"$size", bson.D{
					{"$filter", bson.D{
						{"input", "$students"},
						{"as", "student"},
						{"cond", bson.D{
							{"$eq", bson.A{"$$student.status", "3"}},
						}},
					}},
				}},
			}},
		}},
	}

	pipeline := mongo.Pipeline{bson.D{{"$match", filter}}, projectStage}
	cursor, err := taskCollection.Aggregate(context.Background(), pipeline)
	defer cursor.Close(context.Background())

	if cursor.Next(context.Background()) {
		var result struct {
			NumStudents int64 `bson:"numStudents"`
		}
		err := cursor.Decode(&result)
		if err != nil {
			log.Println("HandleTeacherDashBoard: failed to get number of students")
			util.WriteErrorResponse(w, 500, "Failed to get number of tasks in database")
		}
		state3 := strconv.FormatInt(result.NumStudents, 10)
		resp := struct {
			Students string `json:"students"`
			Tasks    string `json:"tasks"`
			Review   string `json:"review"`
		}{
			Students: students,
			Tasks:    tasks,
			Review:   state3,
		}
		util.WriteSuccessResponse(w, 200, resp)
	}

}
