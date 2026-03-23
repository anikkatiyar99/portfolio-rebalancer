package main

import (
	"net/http"
	_ "portfolio-rebalancer/docs"
	"portfolio-rebalancer/internal/handlers"
	"portfolio-rebalancer/internal/logging"
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
	logging.SetLevel("")

	if err := storage.InitElastic(); err != nil {
		logging.Errorf("failed to initialize Elasticsearch: %v", err)
		panic(err)
	}

	if err := queue.InitKafka(); err != nil {
		logging.Errorf("failed to initialize Kafka: %v", err)
		panic(err)
	}

	store := storage.NewElasticStore()
	publisher := queue.NewKafkaPublisher()
	rebalanceService := services.NewRebalanceService(store, publisher)
	h := handlers.NewHandler(store, rebalanceService)

	http.HandleFunc("/portfolio", h.HandlePortfolio)
	http.HandleFunc("/portfolio/", h.HandlePortfolio)
	http.HandleFunc("/rebalance", h.HandleRebalance)
	http.HandleFunc("/rebalance/", h.HandleRebalance)
	http.Handle("/docs/", httpSwagger.Handler(
		httpSwagger.URL("/docs/doc.json"),
	))

	logging.Infof("server started at :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		logging.Errorf("server failed: %v", err)
		panic(err)
	}
}
