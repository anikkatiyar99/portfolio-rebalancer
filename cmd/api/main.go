package main

import (
	"log"
	"net/http"
	"portfolio-rebalancer/internal/handlers"
	"portfolio-rebalancer/internal/kafka"
	"portfolio-rebalancer/internal/storage"
)

func main() {

	if err := storage.InitElastic(); err != nil {
		log.Fatalf("Failed to initialize Elasticsearch: %v", err)
	}

	if err := kafka.InitKafka(); err != nil {
		log.Fatalf("Failed to initialize Kafka: %v", err)
	}

	http.HandleFunc("/portfolio", handlers.HandlePortfolio)
	http.HandleFunc("/rebalance", handlers.HandleRebalance)

	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
