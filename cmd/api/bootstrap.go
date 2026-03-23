package main

import (
	"net/http"
	_ "portfolio-rebalancer/docs"
	"portfolio-rebalancer/internal/handlers"
	"portfolio-rebalancer/internal/logging"
	"portfolio-rebalancer/internal/queue"
	"portfolio-rebalancer/internal/services"
	"portfolio-rebalancer/internal/storage"
)

// @title Portfolio Rebalancer API
// @version 1.0
// @description APIs for managing portfolios and triggering rebalance transactions.
// @BasePath /
func run() error {
	logging.SetLevel("")

	deps, err := initializeDependencies()
	if err != nil {
		return err
	}

	server := &http.Server{
		Addr:    ":8080",
		Handler: registerRoutes(deps),
	}

	logging.Infof("server started at :8080")
	return server.ListenAndServe()
}

type appDependencies struct {
	handler *handlers.Handler
}

func initializeDependencies() (*appDependencies, error) {
	if err := storage.InitElastic(); err != nil {
		return nil, err
	}

	if err := queue.InitKafka(); err != nil {
		return nil, err
	}

	store := storage.NewElasticStore()
	publisher := queue.NewKafkaPublisher()
	rebalanceService := services.NewRebalanceService(store, publisher)

	return &appDependencies{
		handler: handlers.NewHandler(store, rebalanceService),
	}, nil
}
