package database

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"time"
)

//var (
//	Client = DBInstance()
//)

// behave like init function to connect to mongo db

func DBInstance() *mongo.Client {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("error happened during loading the environment variables in .env file")
		return nil
	}
	mongoDB := os.Getenv("MONGODB_URL")
	// setting connection timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// connect to mongoDB
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoDB))
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("Connected to Database")
	}
	return client
}

func OpenCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	return client.Database("cluster0").Collection(collectionName)
}
