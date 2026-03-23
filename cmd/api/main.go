package main

import (
	"log"
	"net/http"
	"portfolio-rebalancer/internal/handlers"
	"portfolio-rebalancer/internal/queue"
	"portfolio-rebalancer/internal/services"
	"portfolio-rebalancer/internal/storage"
)

func main() {
	if err := storage.InitElastic(); err != nil {
		log.Fatalf("Failed to initialize Elasticsearch: %v", err)
	}

	if err := queue.InitKafka(); err != nil {
		log.Fatalf("Failed to initialize Kafka: %v", err)
	}

	store := storage.NewElasticStore()
	publisher := queue.NewKafkaPublisher()
	rebalanceService := services.NewRebalanceService(store, publisher)
	h := handlers.NewHandler(store, rebalanceService)

	http.HandleFunc("/portfolio", h.HandlePortfolio)
	http.HandleFunc("/rebalance", h.HandleRebalance)

	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
