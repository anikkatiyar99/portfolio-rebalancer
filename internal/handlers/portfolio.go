package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"portfolio-rebalancer/internal/models"
	"portfolio-rebalancer/internal/services"
	"portfolio-rebalancer/internal/storage"
)

type Handler struct {
	store            storage.PortfolioStore
	rebalanceService services.Rebalancer
}

func NewHandler(store storage.PortfolioStore, rebalanceService services.Rebalancer) *Handler {
	return &Handler{
		store:            store,
		rebalanceService: rebalanceService,
	}
}

func (h *Handler) HandlePortfolio(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var p models.Portfolio
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Println("HandlePortfolio==", p)

	for _, percent := range p.Allocation {
		if percent < 0 || percent > 100 {
			http.Error(w, "Allocation percentages must be between 0 and 100", http.StatusBadRequest)
			return
		}
	}

	var sum float64
	for _, percent := range p.Allocation {
		sum += percent
	}
	if sum != 100.0 {
		http.Error(w, "Total allocation must sum to 100%", http.StatusBadRequest)
		return
	}

	if err := h.store.SavePortfolio(r.Context(), p); err != nil {
		http.Error(w, "Failed to save portfolio", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(p)
}

func (h *Handler) HandleRebalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var p models.UpdatedPortfolio
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.rebalanceService.Rebalance(r.Context(), p); err != nil {
		if errors.Is(err, services.ErrPortfolioNotFound) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		log.Println("Failed to rebalance portfolio:", err)
		http.Error(w, "Failed to rebalance portfolio", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
