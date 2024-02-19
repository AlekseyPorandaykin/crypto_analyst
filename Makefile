HOME_PATH := $(shell pwd)

MIGRATE_SQL := $(shell cat < ./migrations/specification.sql;)
BIN := "./bin/crypto_analyst"
VERSION :=$(shell date)

build:
	go build -o=$(BIN) -ldflags="-X 'main.version=${VERSION}' -X 'github.com/AlekseyPorandaykin/crypto_analyst/cmd.homeDir=${HOME_PATH}'" .

init:
	go install golang.org/x/tools/cmd/goimports@latest

run: build
	$(BIN) -config ./configs/default.toml

linters:
	go vet .
	gofmt -w .
	goimports -w .
	gci write /app
	gofumpt -l -w /app
	golangci-lint run ./...
	gofmt -s -l $(git ls-files '*.go')


go-fix:
	go mod tidy
	gci write ./
	gofumpt -l -w ./


.PHONY: build run build-img run-img version test lint
