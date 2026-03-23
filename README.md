# Portfolio Rebalancer

A Go backend for creating user portfolios and generating rebalance transactions when a provider sends updated allocations.

## What This Service Does

- Creates a portfolio for a user with a target allocation.
- Accepts updated market allocations for that user.
- Calculates the `BUY` and `SELL` transactions needed to rebalance back to the user's original target allocation.
- Persists portfolios and rebalance transactions in Elasticsearch.
- Publishes a Kafka fallback message if transaction persistence fails.
- Stores a dead-letter record in Elasticsearch if both transaction persistence and Kafka fallback publishing fail.
- Exposes Swagger UI for interactive API testing.

## Tech Stack

- Go
- Elasticsearch
- Kafka
- Docker Compose
- Swagger / OpenAPI

## Project Structure

- [`cmd/api/main.go`](/Users/anik/Desktop/wahed-test/portfolio-rebalancer/cmd/api/main.go): minimal entrypoint
- [`cmd/api/bootstrap.go`](/Users/anik/Desktop/wahed-test/portfolio-rebalancer/cmd/api/bootstrap.go): dependency initialization and HTTP server startup
- [`cmd/api/routes.go`](/Users/anik/Desktop/wahed-test/portfolio-rebalancer/cmd/api/routes.go): route registration and Swagger docs route
- [`internal/handlers/portfolio.go`](/Users/anik/Desktop/wahed-test/portfolio-rebalancer/internal/handlers/portfolio.go): HTTP handlers and error mapping
- [`internal/services/rebalance.go`](/Users/anik/Desktop/wahed-test/portfolio-rebalancer/internal/services/rebalance.go): rebalance calculation and orchestration
- [`internal/storage/elastic.go`](/Users/anik/Desktop/wahed-test/portfolio-rebalancer/internal/storage/elastic.go): Elasticsearch storage implementation
- [`internal/queue/producer.go`](/Users/anik/Desktop/wahed-test/portfolio-rebalancer/internal/queue/producer.go): Kafka publisher
- [`internal/logging/logger.go`](/Users/anik/Desktop/wahed-test/portfolio-rebalancer/internal/logging/logger.go): leveled application logging
- [`internal/models/dead_letter.go`](/Users/anik/Desktop/wahed-test/portfolio-rebalancer/internal/models/dead_letter.go): dead-letter model for failed fallback processing
- [`docs/`](/Users/anik/Desktop/wahed-test/portfolio-rebalancer/docs): generated Swagger docs

## Implemented So Far

These are the main changes now present in the codebase:

- Path-param APIs for both portfolio creation and rebalance requests
- Swagger/OpenAPI generation and a live docs URL
- Structured JSON error responses with `errorMessage`, `errorCode`, and `errorDetails`
- Duplicate portfolio protection with `409 Conflict`
- Model-level validation through `Validate()` methods
- Rebalance service layer instead of handler-driven business logic
- Rebalance calculation fixes for added assets, removed assets, and tiny floating-point drift
- Stable transaction IDs for idempotent transaction writes
- Leveled logging with `DEBUG`, `INFO`, `WARN`, and `ERROR`
- Dockerfile and Compose improvements for rebuild reliability and health checks
- Unit tests across handlers, models, and services
- Dead-letter persistence for failed fallback publish scenarios

## Startup

### Requirements

- Docker Desktop / Docker Engine with Compose

### Start With Docker

```bash
docker compose up --build
```

Run in the background:

```bash
docker compose up --build -d
```

Stop everything:

```bash
docker compose down
```

Rebuild just the API image:

```bash
docker compose build api
docker compose up -d --force-recreate api
```

### Service URLs

- API base URL: `http://localhost:8080`
- Swagger UI: `http://localhost:8080/docs/index.html`
- OpenAPI JSON: `http://localhost:8080/docs/doc.json`
- Elasticsearch: `http://localhost:9200`

### Health Checks

Check container status:

```bash
docker compose ps
```

View API logs:

```bash
docker compose logs -f api
```

## Environment

The Docker setup configures these values for the API container:

- `ELASTICSEARCH_URL=http://elasticsearch:9200`
- `KAFKA_BROKER=kafka:9092`
- `KAFKA_TOPIC=rebalance`

The app also supports:

- `LOG_LEVEL`

Supported log levels:

- `DEBUG`
- `INFO`
- `WARN`
- `ERROR`

## API Overview

### 1. Create Portfolio

`POST /portfolio/{user_id}`

Creates a new portfolio for a user.

Request body:

```json
{
  "allocation": {
    "stocks": 60,
    "bonds": 30,
    "gold": 10
  }
}
```

Success response: `201 Created`

Example:

```bash
curl -X POST http://localhost:8080/portfolio/user-1 \
  -H "Content-Type: application/json" \
  -d '{
    "allocation": {
      "stocks": 60,
      "bonds": 30,
      "gold": 10
    }
  }'
```

Behavior:

- Creating the same `user_id` again returns `409 Conflict`.
- `allocation` must not be empty.
- Asset percentages must be between `0` and `100`.
- Allocation total must sum to `100`.

### 2. Rebalance Portfolio

`POST /rebalance/{user_id}`

Accepts an updated allocation from a provider and calculates the transactions needed to rebalance back to the user's original allocation.

Request body:

```json
{
  "new_allocation": {
    "stocks": 70,
    "bonds": 20,
    "gold": 10
  }
}
```

Success response: `200 OK`

Example:

```bash
curl -X POST http://localhost:8080/rebalance/user-1 \
  -H "Content-Type: application/json" \
  -d '{
    "new_allocation": {
      "stocks": 70,
      "bonds": 20,
      "gold": 10
    }
  }'
```

Behavior:

- Returns `404 Not Found` if the user portfolio does not exist.
- Returns `400 Bad Request` for invalid payloads.
- Handles added assets and removed assets during rebalance calculation.
- Ignores tiny floating-point drift so it does not generate noise transactions.
- Saves deterministic transaction IDs so duplicate writes overwrite instead of duplicating.

## Rebalance Logic

The service compares the user's original saved allocation with the incoming updated allocation and emits `BUY` / `SELL` transactions for the difference.

Example:

- Original allocation: `{"stocks": 60, "bonds": 30, "gold": 10}`
- Updated allocation: `{"stocks": 70, "bonds": 20, "gold": 10}`

Generated transactions:

- `BUY 10% stocks`
- `SELL 10% bonds`

Edge cases handled:

- New asset added in the updated allocation
- Existing asset removed from the updated allocation
- No-op rebalance when allocations are effectively equal
- Tiny float drift near zero
- Empty transaction result returns success without unnecessary writes

## Error Response Format

All API errors return JSON in this format:

```json
{
  "errorMessage": "Portfolio validation failed",
  "errorCode": 400,
  "errorDetails": "allocation: total allocation must sum to 100"
}
```

Fields:

- `errorMessage`: high-level client-facing message
- `errorCode`: HTTP status code
- `errorDetails`: specific validation or internal failure details

Common examples:

- `400` for invalid payloads or missing path user IDs
- `404` when rebalancing a user that does not exist
- `409` when creating a portfolio for an existing user
- `500` for storage or downstream processing failures

## Swagger Docs

Swagger is generated and served from the running API.

Use:

- Swagger UI: `http://localhost:8080/docs/index.html`
- OpenAPI JSON: `http://localhost:8080/docs/doc.json`

The Swagger request models match the live API:

- `POST /portfolio/{user_id}` takes `user_id` from the path and `allocation` in the body
- `POST /rebalance/{user_id}` takes `user_id` from the path and `new_allocation` in the body

## Running Tests

Run all tests:

```bash
go test ./...
```

Run with coverage:

```bash
go test ./... -cover
```

Current test coverage is strongest in the core logic packages:

- handlers
- models
- services

The test suite currently covers:

- portfolio validation rules
- rebalance validation and not-found handling
- rebalance calculation edge cases
- duplicate portfolio creation
- handler error/status mapping
- dead-letter persistence when fallback publishing fails

## Current Behavior Summary

- `main` is intentionally small; bootstrapping and route registration are separated.
- Duplicate portfolio creation is rejected with `409 Conflict`.
- Validation lives close to the request models through `Validate()` methods.
- Error responses are returned as structured JSON.
- Rebalance transaction writes use stable transaction IDs for idempotent persistence.
- Failed Kafka fallback publishes are captured in a dead-letter index for later inspection.
- Logging is leveled and centralized.
- Docker startup includes API, Elasticsearch, Kafka, and Swagger access.

## Notes

- The API currently persists data in Elasticsearch directly.
- Kafka is used as a fallback path if transaction persistence fails.
- If Kafka fallback publishing also fails, a dead-letter document is written to the `dead_letters` index in Elasticsearch.
- This is a synchronous request flow intended to be correct and observable first.
