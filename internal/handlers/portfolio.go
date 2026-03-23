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

// HandlePortfolio godoc
// @Summary Create portfolio
// @Description Create a user's portfolio with target allocation percentages
// @Tags portfolio
// @Accept json
// @Produce json
// @Param request body models.Portfolio true "Portfolio payload"
// @Success 201 {object} models.Portfolio
// @Failure 400 {object} models.ErrorResponse
// @Failure 405 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /portfolio [post]
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

	if err := p.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.store.SavePortfolio(r.Context(), p); err != nil {
		http.Error(w, "Failed to save portfolio", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(p)
}

// HandleRebalance godoc
// @Summary Rebalance portfolio
// @Description Calculate and persist rebalance transactions for a user's updated allocation
// @Tags portfolio
// @Accept json
// @Produce json
// @Param request body models.UpdatedPortfolio true "Updated portfolio payload"
// @Success 200 {object} models.MessageResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 405 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /rebalance [post]
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
		if errors.Is(err, services.ErrInvalidUserID) || errors.Is(err, services.ErrInvalidAllocation) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if errors.Is(err, services.ErrPortfolioNotFound) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		log.Println("Failed to rebalance portfolio:", err)
		http.Error(w, "Failed to rebalance portfolio", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.MessageResponse{Message: "Rebalance processed"})
}
