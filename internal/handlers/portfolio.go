package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"portfolio-rebalancer/internal/kafka"
	"portfolio-rebalancer/internal/models"
	"portfolio-rebalancer/internal/services"
	"portfolio-rebalancer/internal/storage"
)

// HandlePortfolio handles new portfolio creation requests (feel free to update the request parameter/model)
// Sample Request (POST /portfolio):
//
//	{
//	    "user_id": "1",
//	    "allocation": {"stocks": 60, "bonds": 30, "gold": 10}
//	}
func HandlePortfolio(w http.ResponseWriter, r *http.Request) {
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

	const totalAllocation = 100.0
	var sum float64
	for _, percent := range p.Allocation {
		sum += percent
	}

	if sum != totalAllocation {
		http.Error(w, "Total allocation must sum to 100%", http.StatusBadRequest)
		return
	}

	err = storage.SavePortfolio(r.Context(), p)
	if err != nil {
		http.Error(w, "Failed to save portfolio", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(p)
}

// HandleRebalance handles portfolio rebalance requests from 3rd party provider (feel free to update the request parameter/model)
// Sample Request (POST /rebalance):
//
//	{
//	    "user_id": "1",
//	    "new_allocation": {"stocks": 70, "bonds": 20, "gold": 10}
//	}
func HandleRebalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.UpdatedPortfolio
	json.NewDecoder(r.Body).Decode(&req)

	log.Println("HandleRebalance==", req)

	original, err := storage.GetPortfolio(r.Context(), req.UserID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	transactions := services.CalculateRebalance(original.Allocation, req.NewAllocation)

	for _, t := range transactions {
		err := storage.SaveTransaction(r.Context(), t)
		if err != nil {
			// after calculating transactions, publish the raw request to Kafka in case of any DB failure
			payload, _ := json.Marshal(req)
			err := kafka.PublishMessage(r.Context(), payload)
			if err != nil {
				log.Println("Failed to publish message to Kafka:", err)
			}
		}
	}

	w.WriteHeader(http.StatusOK)
}
