package services

import (
	"portfolio-rebalancer/internal/models"
)

func CalculateRebalance(updatedAllocation, currentAlloctaion map[string]float64) []models.RebalanceTransaction {
	var result []models.RebalanceTransaction

	// TODO: create rebalance transactions and update portfolio
	// For example, if the current allocation is {"stocks": 60, "bonds": 30, "gold": 10} and
	// the new allocation is {"stocks": 70, "bonds": 20, "gold": 10},
	// then we need to BUY 10% of stocks and SELL 10% of bonds.

	for asset, newPercent := range updatedAllocation {
		currentPercent := currentAlloctaion[asset]
		diff := newPercent - currentPercent

		if diff > 0 {
			result = append(result, models.RebalanceTransaction{
				Action:           "BUY",
				Asset:            asset,
				RebalancePercent: diff,
			})
		} else if diff < 0 {
			result = append(result, models.RebalanceTransaction{
				Action:           "SELL",
				Asset:            asset,
				RebalancePercent: -diff,
			})
		}
	}
	return result
}
