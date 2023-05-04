package controller

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Task struct {
	Id          primitive.ObjectID   `json:"id" bson:"_id"`
	Deadline    string               `json:"deadline" bson:"deadline"`
	Priority    string               `json:"priority" bson:"priority"`
	Students    []primitive.ObjectID `json:"students" bson:"students"`
	Title       string               `json:"title" bson:"title"`
	Description string               `json:"description" bson:"description"`
	Lesson      primitive.ObjectID   `json:"lesson" bson:"lesson"`
}

func GetAllTasks() ([]Task, error) {
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

	var tasks []Task
	for cursor.Next(context.Background()) {
		var task Task
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

func AddTask(task Task) error {
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

func EditTask(task Task) error {
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
