.PHONY: build build-api build-cli build-worker test vet lint tidy frontend-install frontend-build up down logs

BIN := bin

build: build-api build-cli build-worker

build-api:
	go build -o $(BIN)/custodian-api ./cmd/custodian-api

build-cli:
	go build -o $(BIN)/custodian ./cmd/custodian-cli

build-worker:
	go build -o $(BIN)/custodian-worker ./cmd/custodian-worker

test:
	go test ./...

vet:
	go vet ./...

lint:
	golangci-lint run

tidy:
	go mod tidy

frontend-install:
	cd frontend && npm ci

frontend-build:
	cd frontend && npm run build

# Control plane stack (requires .env — see deploy/.env.example)
up:
	docker compose -f deploy/docker-compose.yml --env-file deploy/.env up -d

down:
	docker compose -f deploy/docker-compose.yml --env-file deploy/.env down

logs:
	docker compose -f deploy/docker-compose.yml --env-file deploy/.env logs -f api worker
