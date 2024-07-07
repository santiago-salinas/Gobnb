package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"mongo-server/controllers"
	"mongo-server/datasources"
	"mongo-server/mongo_models"
	"mongo-server/repositories"
	"mongo-server/workers"
)

const (
	mongoURL = "mongodb://localhost:27017"
	rabbitWorkerURL = "amqp://guest:guest@localhost:5672/"
)

func main() {
	fmt.Print("Starting server...")
	mongoClient, err := datasources.NewMongoDataSource(mongoURL)
	if err != nil {
		panic(err)
	}
	defer mongoClient.Disconnect(context.TODO())

	repo := repositories.NewReportsMongoRepo(mongoClient, "reports")
	controller := controllers.NewReportsController(repo)

	appWorker, err := workers.BuildRabbitWorker(rabbitWorkerURL)
	if err != nil {
		log.Fatal(err)
	}
	defer appWorker.Close()
	sensorWorker, err := workers.BuildRabbitWorker("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal(err)
	}
	defer sensorWorker.Close()

	appReportsHandler := func(message []byte) error {
		fmt.Println(string(message))

		var objResult mongo_models.AppReport
		reader := bytes.NewReader(message)
		if err := json.NewDecoder(reader).Decode(&objResult); err != nil {
			log.Printf("Error decoding message: %v", err)
			return err
		}

		fmt.Println(objResult)
		loadAppReport(controller, objResult)
		return nil
	}

	sensorReportsHandler := func(message []byte) error {
		fmt.Println(string(message))

		var objResult mongo_models.SensorReport
		reader := bytes.NewReader(message)
		if err := json.NewDecoder(reader).Decode(&objResult); err != nil {
			log.Printf("Error decoding message: %v", err)
			return err
		}

		fmt.Println(objResult)
		loadSensorReport(controller, objResult)
		return nil
	}

	go func() {
		appWorker.Listen(1, "appReports", appReportsHandler)
	}()

	go func() {
		sensorWorker.Listen(10, "sensorReports", sensorReportsHandler)
	}()

	// Create a channel to block the main function
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	// Block until a signal is received
	<-stopChan
	fmt.Println("Shutting down gracefully...")
}

func loadAppReport(controller controllers.ReportsController, report mongo_models.AppReport) {
	controller.AddAppReport(report)
}

func loadSensorReport(controller controllers.ReportsController, report mongo_models.SensorReport) {
	controller.AddSensorReport(report)
}
