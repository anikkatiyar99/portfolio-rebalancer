package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"portfolio-rebalancer/internal/models"
	"portfolio-rebalancer/internal/queue"
	"portfolio-rebalancer/internal/storage"
)

var ErrPortfolioNotFound = errors.New("portfolio not found")

type RebalanceService struct {
	store     storage.PortfolioStore
	publisher queue.MessagePublisher
}

func NewRebalanceService(store storage.PortfolioStore, publisher queue.MessagePublisher) *RebalanceService {
	return &RebalanceService{
		store:     store,
		publisher: publisher,
	}
}

func (s *RebalanceService) Rebalance(ctx context.Context, req models.UpdatedPortfolio) error {
	original, err := s.store.GetPortfolio(ctx, req.UserID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrPortfolioNotFound, err)
	}

	transactions := CalculateRebalance(original.Allocation, req.NewAllocation)
	for _, transaction := range transactions {
		transaction.UserID = req.UserID
		if err := s.store.SaveTransaction(ctx, transaction); err != nil {
			payload, marshalErr := json.Marshal(req)
			if marshalErr != nil {
				return fmt.Errorf("save transaction: %w; marshal fallback payload: %v", err, marshalErr)
			}
			if publishErr := s.publisher.PublishMessage(ctx, payload); publishErr != nil {
				return fmt.Errorf("save transaction: %w; publish fallback payload: %v", err, publishErr)
			}
			return fmt.Errorf("save transaction: %w", err)
		}
	}

	return nil
}

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
