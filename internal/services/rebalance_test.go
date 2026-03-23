package services

import (
	"context"
	"errors"
	"testing"

	"portfolio-rebalancer/internal/models"
	"portfolio-rebalancer/internal/storage"
)

type stubStore struct {
	portfolio         *models.Portfolio
	getPortfolioErr   error
	saveTxnErr        error
	savedTransactions []models.RebalanceTransaction
}

func (s *stubStore) SavePortfolio(ctx context.Context, p models.Portfolio) error {
	return nil
}

func (s *stubStore) GetPortfolio(ctx context.Context, userID string) (*models.Portfolio, error) {
	if s.getPortfolioErr != nil {
		return nil, s.getPortfolioErr
	}
	return s.portfolio, nil
}

func (s *stubStore) SaveTransaction(ctx context.Context, t models.RebalanceTransaction) error {
	s.savedTransactions = append(s.savedTransactions, t)
	return s.saveTxnErr
}

type stubPublisher struct {
	published [][]byte
	err       error
}

func (s *stubPublisher) PublishMessage(ctx context.Context, payload []byte) error {
	s.published = append(s.published, payload)
	return s.err
}

func TestCalculateRebalanceHandlesAddedRemovedAndChangedAssets(t *testing.T) {
	current := map[string]float64{
		"stocks": 60,
		"bonds":  30,
		"gold":   10,
	}
	updated := map[string]float64{
		"stocks": 70,
		"bonds":  20,
		"crypto": 10,
	}

	got := CalculateRebalance(current, updated)
	if len(got) != 4 {
		t.Fatalf("expected 4 transactions, got %d", len(got))
	}
}

func TestCalculateRebalanceSkipsTinyDrift(t *testing.T) {
	current := map[string]float64{
		"stocks": 33.333333333,
		"bonds":  66.666666667,
	}
	updated := map[string]float64{
		"stocks": 33.3333333335,
		"bonds":  66.6666666665,
	}

	got := CalculateRebalance(current, updated)
	if len(got) != 0 {
		t.Fatalf("expected no transactions, got %d", len(got))
	}
}

func TestRebalanceValidationErrorIsClean(t *testing.T) {
	service := NewRebalanceService(&stubStore{}, &stubPublisher{})

	err := service.Rebalance(context.Background(), models.UpdatedPortfolio{
		UserID:        "user-1",
		NewAllocation: map[string]float64{},
	})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	if !errors.Is(err, ErrInvalidAllocation) {
		t.Fatalf("expected ErrInvalidAllocation, got %v", err)
	}
	if err.Error() != "new_allocation: must not be empty" {
		t.Fatalf("expected clean validation error, got %q", err.Error())
	}
}

func TestRebalanceReturnsNotFound(t *testing.T) {
	service := NewRebalanceService(&stubStore{getPortfolioErr: storage.ErrPortfolioNotFound}, &stubPublisher{})

	err := service.Rebalance(context.Background(), models.UpdatedPortfolio{
		UserID: "user-1",
		NewAllocation: map[string]float64{
			"stocks": 100,
		},
	})
	if !errors.Is(err, ErrPortfolioNotFound) {
		t.Fatalf("expected ErrPortfolioNotFound, got %v", err)
	}
}

func TestRebalanceSuccessSavesTransactions(t *testing.T) {
	store := &stubStore{
		portfolio: &models.Portfolio{
			UserID: "user-1",
			Allocation: map[string]float64{
				"stocks": 60,
				"bonds":  40,
			},
		},
	}
	service := NewRebalanceService(store, &stubPublisher{})

	err := service.Rebalance(context.Background(), models.UpdatedPortfolio{
		UserID: "user-1",
		NewAllocation: map[string]float64{
			"stocks": 70,
			"bonds":  30,
		},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(store.savedTransactions) != 2 {
		t.Fatalf("expected 2 saved transactions, got %d", len(store.savedTransactions))
	}
}
