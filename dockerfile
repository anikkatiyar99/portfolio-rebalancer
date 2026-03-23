# syntax=docker/dockerfile:1.7
FROM golang:1.20-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -o /portfolio-rebalancer ./cmd/api

FROM alpine:3.20

RUN adduser -D -H appuser

COPY --from=builder /portfolio-rebalancer /portfolio-rebalancer

EXPOSE 8080

USER appuser

ENTRYPOINT ["/portfolio-rebalancer"]
