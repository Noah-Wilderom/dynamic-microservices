package main

import (
	"context"
	"encoding/json"
	"github.com/Noah-Wilderom/dynamic-microservices/logger-service/data"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"sync"
	"time"
)

const (
	mongoURL     = "mongodb://mongo:27017"
	mongoTimeout = 15 * time.Second
)

var client *mongo.Client

type Config struct {
	Models   data.Models
	GRPCPort int `json:"grpc_port"`
}

func main() {
	// connect to mongodb
	mongoClient, err := connectToMongo()
	if err != nil {
		log.Panic(err)
	}
	client = mongoClient

	// create a context in order to disconnect
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()

	// close connection
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	var app Config
	configFile, err := os.ReadFile("/config.json")
	if err != nil {
		log.Panicln("Cannot find config file", err)
	}

	err = json.Unmarshal(configFile, &app)
	if err != nil {
		log.Panicln("Cannot parse config file", err)
	}

	app.Models = data.New(client)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		app.gRPCListen()
	}()

	wg.Wait()
}

func connectToMongo() (*mongo.Client, error) {
	// create the connection options
	clientOptions := options.Client().ApplyURI(mongoURL)
	clientOptions.SetAuth(options.Credential{
		Username: os.Getenv("MONGO_USERNAME"),
		Password: os.Getenv("MONGO_PASSWORD"),
	})

	// connect
	conn, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Println("Error connection:", err)
		return nil, err
	}

	log.Println("Connected to mongodb")

	return conn, nil
}
