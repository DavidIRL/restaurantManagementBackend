package database

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

func DBinstance() *mongo.Client {
	port := 27017
	MongoDB := fmt.Sprintf("mongodb://localhost:%d", port)
	fmt.Print(MongoDB)

	client, err := mongo.NewClient(options.Client().ApplyURI(MongoDB))
	if err != nil {
		log.Fatal(err)
	}
	contxt, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()
	err = client.Connect(contxt)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Sprintf("connected to mongodb on %d\n", port)
}

var Client *mongo.Client = DBinstance()

func OpenCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	var collection *mongo.Client = client.Database("restaurant").Collection(collectionName)

	return collection
}
