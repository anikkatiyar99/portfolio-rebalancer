package services

import (
	"context"
	"portfolio-rebalancer/internal/models"
)

type Rebalancer interface {
	Rebalance(ctx context.Context, req models.UpdatedPortfolio) error
}
