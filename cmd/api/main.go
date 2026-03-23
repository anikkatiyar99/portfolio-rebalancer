package main

import (
	"log"
	"net/http"
	_ "portfolio-rebalancer/docs"
	"portfolio-rebalancer/internal/handlers"
	"portfolio-rebalancer/internal/queue"
	"portfolio-rebalancer/internal/services"
	"portfolio-rebalancer/internal/storage"

	httpSwagger "github.com/swaggo/http-swagger"
)

// @title Portfolio Rebalancer API
// @version 1.0
// @description APIs for managing portfolios and triggering rebalance transactions.
// @BasePath /
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
	http.Handle("/docs/", httpSwagger.Handler(
		httpSwagger.URL("/docs/doc.json"),
	))

	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
