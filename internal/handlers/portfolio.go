package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"portfolio-rebalancer/internal/logging"
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

func writeJSONError(w http.ResponseWriter, status int, message, details string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(models.ErrorResponse{
		ErrorMessage: message,
		ErrorCode:    status,
		ErrorDetails: details,
	})
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
// @Failure 409 {object} models.ErrorResponse
// @Failure 405 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /portfolio [post]
func (h *Handler) HandlePortfolio(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed", "Only POST is supported for this endpoint")
		return
	}

	var p models.Portfolio
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	logging.Infof("received portfolio create request for user %s", p.UserID)

	if err := p.Validate(); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Portfolio validation failed", err.Error())
		return
	}

	if _, err := h.store.GetPortfolio(r.Context(), p.UserID); err == nil {
		writeJSONError(w, http.StatusConflict, "Portfolio already exists", services.ErrPortfolioAlreadyExists.Error())
		return
	} else if !errors.Is(err, storage.ErrPortfolioNotFound) {
		logging.Errorf("failed to check existing portfolio for user %s: %v", p.UserID, err)
		writeJSONError(w, http.StatusInternalServerError, "Failed to load portfolio", err.Error())
		return
	}

	if err := h.store.SavePortfolio(r.Context(), p); err != nil {
		logging.Errorf("failed to save portfolio for user %s: %v", p.UserID, err)
		writeJSONError(w, http.StatusInternalServerError, "Failed to save portfolio", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
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
		writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed", "Only POST is supported for this endpoint")
		return
	}

	var p models.UpdatedPortfolio
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if err := h.rebalanceService.Rebalance(r.Context(), p); err != nil {
		if errors.Is(err, services.ErrInvalidUserID) || errors.Is(err, services.ErrInvalidAllocation) {
			writeJSONError(w, http.StatusBadRequest, "Rebalance validation failed", err.Error())
			return
		}
		if errors.Is(err, services.ErrPortfolioNotFound) {
			writeJSONError(w, http.StatusNotFound, "User not found", err.Error())
			return
		}

		logging.Errorf("failed to rebalance portfolio for user %s: %v", p.UserID, err)
		writeJSONError(w, http.StatusInternalServerError, "Failed to rebalance portfolio", err.Error())
		return
	}

	logging.Infof("rebalance processed for user %s", p.UserID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.MessageResponse{Message: "Rebalance processed"})
}
