HOME_PATH := $(shell pwd)

DOCKER_DIR="./deployments/docker-compose.yaml"
DOCKER_COMPOSE_DEV_FILE="./deployments/docker-compose-dev.yaml"
MIGRATE_SQL := $(shell cat < ./migrations/specification.sql;)
BIN := "./bin/crypto_analyst"
VERSION :=$(shell date)

build:
	go build -o=$(BIN) -ldflags="-X 'main.version=${VERSION}' -X 'github.com/AlekseyPorandaykin/crypto_analyst/cmd.homeDir=${HOME_PATH}'" .

init:
	go install golang.org/x/tools/cmd/goimports@latest

run: build
	$(BIN) -config ./configs/default.toml

run-img: build-img
	docker run $(DOCKER_IMG)

up:
	docker-compose --file=$(DOCKER_DIR) up -d

down:
	docker-compose --file=$(DOCKER_DIR) down

recreate:
	docker-compose --file=$(DOCKER_DIR) rm -f
	docker-compose --file=$(DOCKER_DIR) pull
	docker-compose --file=$(DOCKER_DIR) up --build -d

ps:
	docker-compose --file=$(DOCKER_DIR) ps

linters:
	go vet .
	gofmt -w .
	goimports -w .
	gci write /app
	gofumpt -l -w /app
	golangci-lint run ./...
	gofmt -s -l $(git ls-files '*.go')

migrate:
	docker-compose --file=$(DOCKER_DIR) exec pg psql -U crypto_analyst -d crypto_analyst -c "$(MIGRATE_SQL)"

.PHONY: build run build-img run-img version test lint
