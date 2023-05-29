package main

import (
	"back-end/controller"
	"back-end/token"
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"log"
	"net/http"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//var ts *token.TokenStorage

func main() {
	mongoURL := os.Getenv("MONGO_URL")

	router := chi.NewRouter()
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"X-PINGOTHER", "Accept", "Authorization", "Content-Type", "X-CSRF-Token", "All"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
	//conection to database
	client, err := getDatabase(mongoURL)
	if err != nil {
		log.Fatal("Error connecting to MongoDB: ", err)
	}
	defer client.Disconnect(context.Background())

	ts := token.GetTokenStorageInstance()

	//Created user controller
	uc := controller.NewStudentController(client, ts)
	subjectc := controller.NewSubjectController(client, ts, uc)
	taskController := controller.NewTaskController(client, ts, uc)
	dashBoardController := controller.NewDashBoardController(client, ts, uc, taskController)

	router.Post("/login", uc.HandleLogin)

	router.Post("/logout", uc.HandleLogout)

	router.Post("/student", uc.HandleAddStudent)
	router.Put("/student", uc.HandleEditStudent)
	router.Delete("/student", uc.HandleDeleteStudent)
	router.Get("/student/{classTitle}", uc.HandleGetStudentsFromClass)

	router.Post("/subject", subjectc.HandleNewSubject)
	router.Delete("/subject", subjectc.HandleDeleteSubject)

	router.Get("/tasks", taskController.GetTasks)
	router.Get("/task", subjectc.HandleGetFormSubjects)
	router.Get("/task/{classTitle}", taskController.HandleTeacherTasks)
	router.Post("/task", taskController.HandleAddTask)

	router.Get("/teacher/dashboard", dashBoardController.HandleTeacherDashBoard)

	log.Println("Starting server...")
	err = http.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatal("Error starting server: ", err)
	}

	log.Println("Server started!")
}

func getDatabase(url string) (*mongo.Client, error) {
	log.Println("this is mongo url", url)
	clientOptions := options.Client().ApplyURI(url)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, err
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	log.Println("Connected to mongodb")
	return client, nil
}
