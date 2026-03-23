package services

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"portfolio-rebalancer/internal/logging"
	"portfolio-rebalancer/internal/models"
	"portfolio-rebalancer/internal/queue"
	"portfolio-rebalancer/internal/storage"
	"sort"
	"strconv"
	"strings"
	"time"
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
		if errors.Is(err, storage.ErrPortfolioNotFound) {
			return ErrPortfolioNotFound
		}
		return fmt.Errorf("%w: %v", ErrPortfolioNotFound, err)
	}

	transactions := CalculateRebalance(original.Allocation, req.NewAllocation)
	if len(transactions) == 0 {
		return nil
	}

	errCh := make(chan error, len(transactions))
	for _, transaction := range transactions {
		transaction.UserID = req.UserID
		transaction.TransactionID = buildTransactionID(req.UserID, original.Allocation, req.NewAllocation, transaction)
		transaction := transaction // capture range variable

		go func() {
			errCh <- s.store.SaveTransaction(ctx, transaction)
		}()
	}

	for range transactions {
		if err := <-errCh; err != nil {
			payload, marshalErr := json.Marshal(req)
			if marshalErr != nil {
				if dlqErr := s.saveDeadLetter(ctx, req.UserID, "", fmt.Sprintf("marshal fallback payload: %v", marshalErr)); dlqErr != nil {
					return fmt.Errorf("save transaction: %w; marshal fallback payload: %v; save dead letter: %v", err, marshalErr, dlqErr)
				}
				return fmt.Errorf("save transaction: %w; marshal fallback payload: %v", err, marshalErr)
			}
			if publishErr := s.publisher.PublishMessage(ctx, payload); publishErr != nil {
				logging.Errorf("failed to publish Kafka fallback for user %s: %v", req.UserID, publishErr)
				if dlqErr := s.saveDeadLetter(ctx, req.UserID, string(payload), fmt.Sprintf("publish fallback payload: %v", publishErr)); dlqErr != nil {
					return fmt.Errorf("save transaction: %w; publish fallback payload: %v; save dead letter: %v", err, publishErr, dlqErr)
				}
				return fmt.Errorf("save transaction: %w; publish fallback payload: %v", err, publishErr)
			}
			return fmt.Errorf("save transaction: %w", err)
		}
	}

	return nil
}

func (s *RebalanceService) saveDeadLetter(ctx context.Context, userID, payload, reason string) error {
	dlq := models.DeadLetterMessage{
		ID:        buildDeadLetterID(userID, payload, reason),
		UserID:    userID,
		Stage:     "rebalance_fallback_publish",
		Reason:    reason,
		Payload:   payload,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	return s.store.SaveDeadLetter(ctx, dlq)
}

func classifyValidationError(err error) error {
	var validationErr *models.ValidationError
	if !errors.As(err, &validationErr) {
		return &ValidationServiceError{
			Kind:    ErrInvalidAllocation,
			Details: err.Error(),
		}
	}
	if validationErr.Field == "user_id" {
		return &ValidationServiceError{
			Kind:    ErrInvalidUserID,
			Details: validationErr.Error(),
		}
	}
	return &ValidationServiceError{
		Kind:    ErrInvalidAllocation,
		Details: validationErr.Error(),
	}
}

func buildTransactionID(userID string, currentAllocation, updatedAllocation map[string]float64, transaction models.RebalanceTransaction) string {
	var builder strings.Builder
	builder.WriteString(userID)
	builder.WriteString("|current:")
	builder.WriteString(canonicalAllocation(currentAllocation))
	builder.WriteString("|updated:")
	builder.WriteString(canonicalAllocation(updatedAllocation))
	builder.WriteString("|asset:")
	builder.WriteString(transaction.Asset)
	builder.WriteString("|action:")
	builder.WriteString(transaction.Action)
	builder.WriteString("|percent:")
	builder.WriteString(strconv.FormatFloat(transaction.RebalancePercent, 'f', -1, 64))

	sum := sha256.Sum256([]byte(builder.String()))
	return fmt.Sprintf("%x", sum)
}

func buildDeadLetterID(userID, payload, reason string) string {
	var builder strings.Builder
	builder.WriteString(userID)
	builder.WriteString("|payload:")
	builder.WriteString(payload)
	builder.WriteString("|reason:")
	builder.WriteString(reason)

	sum := sha256.Sum256([]byte(builder.String()))
	return fmt.Sprintf("%x", sum)
}

func canonicalAllocation(allocation map[string]float64) string {
	if len(allocation) == 0 {
		return ""
	}

	keys := make([]string, 0, len(allocation))
	for asset := range allocation {
		keys = append(keys, asset)
	}
	sort.Strings(keys)

	var builder strings.Builder
	for i, asset := range keys {
		if i > 0 {
			builder.WriteString(",")
		}
		builder.WriteString(asset)
		builder.WriteString("=")
		builder.WriteString(strconv.FormatFloat(allocation[asset], 'f', -1, 64))
	}

	return builder.String()
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
