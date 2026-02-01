.PHONY: help proto build test run docker-build

help:
	@echo "Database Service - Commands"
	@echo "==========================="
	@echo "proto        - Generate proto code (run from shared dir: cd ../../shared && buf generate)"
	@echo "build        - Build the database service"
	@echo "test         - Run unit tests"
	@echo "run          - Run the service locally"
	@echo "docker-build - Build Docker image"

proto:
	@echo "Run: cd ../../shared && buf generate"
	cd ../../shared && buf generate

build:
	go build -o bin/database ./cmd/database

test:
	go test -v ./...

run: build
	./bin/database

docker-build:
	docker build -t database-service:latest .
