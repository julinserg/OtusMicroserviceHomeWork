BIN_CMD_ORDERS := "./bin/orders_service"

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

build:
	go build -v -o $(BIN_CMD_ORDERS) -ldflags "$(LDFLAGS)" ./cmd/orders_service

run: build
	$(BIN_CMD_ORDERS) -config ./configs/orders_config.toml

dbuild:
	docker compose -f ./deployments/docker-compose.yaml build --no-cache

up:
	docker compose -f ./deployments/docker-compose.yaml up --build

down:
	docker compose -f ./deployments/docker-compose.yaml down

version: build
	$(BIN_CMD_ORDERS) version

test:
	go test -race ./...

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.41.1

lint: install-lint-deps
	golangci-lint run ./...

.PHONY: build run build-img run-img version test lint
