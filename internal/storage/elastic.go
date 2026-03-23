package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"portfolio-rebalancer/internal/logging"
	"time"

	"portfolio-rebalancer/internal/models"

	"github.com/elastic/go-elasticsearch/v8"
)

type ElasticStore struct{}

func NewElasticStore() *ElasticStore {
	return &ElasticStore{}
}

var esClient *elasticsearch.Client

func InitElastic() error {
	cfg := elasticsearch.Config{
		Addresses: []string{
			os.Getenv("ELASTICSEARCH_URL"),
		},
	}

	var client *elasticsearch.Client
	var err error

	for i := 1; i <= 5; i++ {
		client, err = elasticsearch.NewClient(cfg)
		if err != nil {
			logging.Errorf("failed to create elasticsearch client: %v", err)
		} else {
			_, err = client.Info()
			if err == nil {
				logging.Infof("connected to Elasticsearch")
				esClient = client
				return nil
			}
			logging.Warnf("elasticsearch not ready yet: %v", err)
		}

		logging.Infof("retrying connection to Elasticsearch (%d/5)", i)
		time.Sleep(5 * time.Second)
	}

	return fmt.Errorf("failed to connect to Elasticsearch after retries: %w", err)
}

func (e *ElasticStore) SavePortfolio(ctx context.Context, p models.Portfolio) error {
	body, err := json.Marshal(p)
	if err != nil {
		return err
	}

	res, err := esClient.Index("portfolios", bytes.NewReader(body), esClient.Index.WithDocumentID(p.UserID))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error saving portfolio: %s", res.String())
	}

	logging.Infof("portfolio saved for user %s", p.UserID)
	return nil
}

func (e *ElasticStore) GetPortfolio(ctx context.Context, userID string) (*models.Portfolio, error) {
	res, err := esClient.Get("portfolios", userID)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, ErrPortfolioNotFound
	}

	var esResp struct {
		Source models.Portfolio `json:"_source"`
	}

	if err := json.NewDecoder(res.Body).Decode(&esResp); err != nil {
		return nil, err
	}

	return &esResp.Source, nil
}

func (e *ElasticStore) SaveTransaction(ctx context.Context, t models.RebalanceTransaction) error {
	body, err := json.Marshal(t)
	if err != nil {
		return err
	}

	res, err := esClient.Index("transactions", bytes.NewReader(body), esClient.Index.WithDocumentID(t.TransactionID))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error saving transaction: %s", res.String())
	}

	logging.Infof("transaction saved for user %s", t.UserID)
	return nil
}
