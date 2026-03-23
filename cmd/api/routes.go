package main

import (
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger"
)

func registerRoutes(deps *appDependencies) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/portfolio", deps.handler.HandlePortfolio)
	mux.HandleFunc("/portfolio/", deps.handler.HandlePortfolio)
	mux.HandleFunc("/rebalance", deps.handler.HandleRebalance)
	mux.HandleFunc("/rebalance/", deps.handler.HandleRebalance)
	mux.Handle("/docs/", httpSwagger.Handler(
		httpSwagger.URL("/docs/doc.json"),
	))

	return mux
}
