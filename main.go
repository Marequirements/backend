package main

import (
	"back-end/controller"
	"back-end/token"
	"context"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//var ts *token.TokenStorage

func main() {
	router := chi.NewRouter()

	//coonection to database
	client, err := getDatabase()
	if err != nil {
		log.Fatal("Error connecting to MongoDB: ", err)
	}
	defer client.Disconnect(context.Background())

	ts := token.GetTokenStorageInstance()
	//Created user controller
	uc := controller.NewStudentController(client, token.GetTokenStorageInstance())
	subjectc := controller.NewSubjectController(client, ts, uc)
	//tsc := controller.NewTaskController(client, token.GetTokenStorageInstance())
	taskController := controller.NewTaskController(client, ts, uc)

	router.Post("/login", uc.HandleLogin)

	router.Post("/logout", uc.HandleLogout)

	router.Post("/add-student", uc.HandleAddStudent)
	router.Post("/add-subject", subjectc.HandleNewSubject)

	router.Delete("/delete-student", uc.HandleDeleteStudent)

	router.Put("/edit-student", uc.HandleEditStudent)

	router.Get("/tasks", taskController.GetTasks)

	log.Println("Starting server...")
	err = http.ListenAndServe(":3000", router)
	if err != nil {
		log.Fatal("Error starting server: ", err)
	}

	log.Println("Server started!")
}

func getDatabase() (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI("mongodb+srv://mareklescinsky:EUFZTs6jcdkqqEUk@brainboard.lrpc8h3.mongodb.net/?retryWrites=true&w=majority")
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
