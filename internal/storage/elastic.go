package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"portfolio-rebalancer/internal/models"

	"github.com/elastic/go-elasticsearch/v8"
)

var esClient *elasticsearch.Client

// InitElastic initializes elasticsearch connection with retry logic
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
			log.Printf("Failed to create client: %v", err)
		} else {
			_, err = client.Info()
			if err == nil {
				log.Println("Connected to Elasticsearch")
				esClient = client
				return nil
			}
			log.Printf("Client created, but ES not ready: %v", err)
		}

		log.Printf("Retrying connection to Elasticsearch... (%d/5)", i)
		time.Sleep(5 * time.Second)
	}

	return fmt.Errorf("failed to connect to Elasticsearch after retries: %w", err)
}

func SavePortfolio(ctx context.Context, p models.Portfolio) error {
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

	log.Printf("Portfolio saved for user %s", p.UserID)
	return nil
}

func GetPortfolio(ctx context.Context, userID string) (*models.Portfolio, error) {
	res, err := esClient.Get("portfolios", userID)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("user not found")
	}

	var esResp struct {
		Source models.Portfolio `json:"_source"`
	}

	if err := json.NewDecoder(res.Body).Decode(&esResp); err != nil {
		return nil, err
	}

	return &esResp.Source, nil
}

func SaveTransaction(ctx context.Context, t models.RebalanceTransaction) error {
	body, err := json.Marshal(t)
	if err != nil {
		return err
	}

	res, err := esClient.Index("transactions", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error saving transaction: %s", res.String())
	}

	log.Printf("Transaction saved for user %s", t.UserID)
	return nil
}
