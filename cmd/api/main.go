package main

import (
	"log"
	"net/http"

	"github.com/Bharat1Rajput/task-management/internal/api"
	"github.com/Bharat1Rajput/task-management/internal/publisher"
)

func main() {
	mux := http.NewServeMux()

	publisher := publisher.NewDummyPublisher()
	handler := api.NewHandler(publisher)

	mux.HandleFunc("/jobs", handler.CreateJob)

	log.Println("API running on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
