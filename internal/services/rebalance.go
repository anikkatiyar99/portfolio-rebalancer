package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"portfolio-rebalancer/internal/models"
	"portfolio-rebalancer/internal/queue"
	"portfolio-rebalancer/internal/storage"
)

const rebalanceTolerance = 1e-9

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
	if err := req.Validate(); err != nil {
		return classifyValidationError(err)
	}

	original, err := s.store.GetPortfolio(ctx, req.UserID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrPortfolioNotFound, err)
	}

	transactions := CalculateRebalance(original.Allocation, req.NewAllocation)
	if len(transactions) == 0 {
		return nil
	}

	errCh := make(chan error, len(transactions))
	for _, transaction := range transactions {
		transaction.UserID = req.UserID
		transaction := transaction // capture range variable

		go func() {
			errCh <- s.store.SaveTransaction(ctx, transaction)
		}()
	}

	for range transactions {
		if err := <-errCh; err != nil {
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

func classifyValidationError(err error) error {
	var validationErr *models.ValidationError
	if !errors.As(err, &validationErr) {
		return fmt.Errorf("%w: %v", ErrInvalidAllocation, err)
	}
	if validationErr.Field == "user_id" {
		return fmt.Errorf("%w: %s", ErrInvalidUserID, validationErr.Error())
	}
	return fmt.Errorf("%w: %s", ErrInvalidAllocation, validationErr.Error())
}

func CalculateRebalance(currentAllocation, updatedAllocation map[string]float64) []models.RebalanceTransaction {
	var result []models.RebalanceTransaction
	seenAssets := make(map[string]struct{})

	// TODO: create rebalance transactions and update portfolio
	// For example, if the current allocation is {"stocks": 60, "bonds": 30, "gold": 10} and
	// the new allocation is {"stocks": 70, "bonds": 20, "gold": 10},
	// then we need to BUY 10% of stocks and SELL 10% of bonds.

	for asset := range currentAllocation {
		seenAssets[asset] = struct{}{}
	}
	for asset := range updatedAllocation {
		seenAssets[asset] = struct{}{}
	}

	for asset := range seenAssets {
		currentPercent := currentAllocation[asset]
		newPercent := updatedAllocation[asset]
		diff := newPercent - currentPercent

		if math.Abs(diff) <= rebalanceTolerance {
			continue
		}

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
