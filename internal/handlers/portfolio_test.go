package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"portfolio-rebalancer/internal/models"
	"portfolio-rebalancer/internal/services"
	"portfolio-rebalancer/internal/storage"
)

type handlerStore struct {
	getPortfolioFunc  func(ctx context.Context, userID string) (*models.Portfolio, error)
	savePortfolioFunc func(ctx context.Context, p models.Portfolio) error
}

func (s *handlerStore) SavePortfolio(ctx context.Context, p models.Portfolio) error {
	if s.savePortfolioFunc != nil {
		return s.savePortfolioFunc(ctx, p)
	}
	return nil
}

func (s *handlerStore) GetPortfolio(ctx context.Context, userID string) (*models.Portfolio, error) {
	if s.getPortfolioFunc != nil {
		return s.getPortfolioFunc(ctx, userID)
	}
	return nil, storage.ErrPortfolioNotFound
}

func (s *handlerStore) SaveTransaction(ctx context.Context, t models.RebalanceTransaction) error {
	return nil
}

type handlerRebalancer struct {
	err error
}

func (r *handlerRebalancer) Rebalance(ctx context.Context, req models.UpdatedPortfolio) error {
	return r.err
}

func TestHandlePortfolioCreateSuccess(t *testing.T) {
	h := NewHandler(&handlerStore{}, &handlerRebalancer{})

	body := bytes.NewBufferString(`{"allocation":{"stocks":60,"bonds":30,"gold":10}}`)
	req := httptest.NewRequest(http.MethodPost, "/portfolio/user-1", body)
	rec := httptest.NewRecorder()

	h.HandlePortfolio(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rec.Code)
	}

	var got models.Portfolio
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if got.UserID != "user-1" {
		t.Fatalf("expected user_id to come from path, got %q", got.UserID)
	}
}

func TestHandlePortfolioDuplicateUser(t *testing.T) {
	store := &handlerStore{
		getPortfolioFunc: func(ctx context.Context, userID string) (*models.Portfolio, error) {
			return &models.Portfolio{UserID: userID}, nil
		},
	}
	h := NewHandler(store, &handlerRebalancer{})

	body := bytes.NewBufferString(`{"allocation":{"stocks":60,"bonds":30,"gold":10}}`)
	req := httptest.NewRequest(http.MethodPost, "/portfolio/user-1", body)
	rec := httptest.NewRecorder()

	h.HandlePortfolio(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d", rec.Code)
	}
}

func TestHandlePortfolioMissingPathUserID(t *testing.T) {
	h := NewHandler(&handlerStore{}, &handlerRebalancer{})

	body := bytes.NewBufferString(`{"allocation":{"stocks":60,"bonds":30,"gold":10}}`)
	req := httptest.NewRequest(http.MethodPost, "/portfolio", body)
	rec := httptest.NewRecorder()

	h.HandlePortfolio(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestHandleRebalanceValidationError(t *testing.T) {
	h := NewHandler(&handlerStore{}, &handlerRebalancer{err: &services.ValidationServiceError{
		Kind:    services.ErrInvalidAllocation,
		Details: "new_allocation: must not be empty",
	}})

	req := httptest.NewRequest(http.MethodPost, "/rebalance/user-1", bytes.NewBufferString(`{}`))
	rec := httptest.NewRecorder()

	h.HandleRebalance(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestHandleRebalanceNotFound(t *testing.T) {
	h := NewHandler(&handlerStore{}, &handlerRebalancer{err: services.ErrPortfolioNotFound})

	req := httptest.NewRequest(http.MethodPost, "/rebalance/user-404", bytes.NewBufferString(`{"new_allocation":{"stocks":100}}`))
	rec := httptest.NewRecorder()

	h.HandleRebalance(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}
}

func TestHandleRebalanceInternalError(t *testing.T) {
	h := NewHandler(&handlerStore{}, &handlerRebalancer{err: errors.New("boom")})

	req := httptest.NewRequest(http.MethodPost, "/rebalance/user-1", bytes.NewBufferString(`{"new_allocation":{"stocks":100}}`))
	rec := httptest.NewRecorder()

	h.HandleRebalance(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}
}
