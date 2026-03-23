package storage

import (
	"context"
	"portfolio-rebalancer/internal/models"
)

type PortfolioStore interface {
	SavePortfolio(ctx context.Context, p models.Portfolio) error
	GetPortfolio(ctx context.Context, userID string) (*models.Portfolio, error)
	SaveTransaction(ctx context.Context, t models.RebalanceTransaction) error
}
