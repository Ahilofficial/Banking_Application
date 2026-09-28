.PHONY: all build test seed run-local docker-up docker-down

all: test build

build:
	@echo "Building all microservices..."
	go build ./...

test:
	@echo "Running full integration test suite..."
	go test -v ./tests

seed:
	@echo "Seeding MySQL database with banking actors and accounts..."
	go run ./cmd/seed/main.go

run-local:
	@echo "Starting all microservices and API Gateway concurrently..."
	go run ./cmd/runner/main.go

docker-up:
	docker-compose up --build -d

docker-down:
	docker-compose down -v
