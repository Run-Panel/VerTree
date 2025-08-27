.PHONY: help build run dev test clean docker docker-dev migrate

# Default target
help:
	@echo "RunPanel Update Service - Available commands:"
	@echo "  make build       - Build the Go application"
	@echo "  make run         - Run the application locally"
	@echo "  make dev         - Run in development mode with hot reload"
	@echo "  make test        - Run tests"
	@echo "  make clean       - Clean build artifacts"
	@echo "  make docker      - Build and run with Docker Compose (production)"
	@echo "  make docker-dev  - Build and run with Docker Compose (development)"
	@echo "  make migrate     - Run database migrations"
	@echo "  make frontend    - Build frontend"

# Build the Go application
build:
	@echo "Building Go application..."
	CGO_ENABLED=1 go build -o bin/update-service cmd/server/main.go

# Run the application locally
run: build
	@echo "Starting application..."
	./bin/update-service

# Development mode
dev:
	@echo "Starting in development mode..."
	go run cmd/server/main.go

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -rf web/admin/
	rm -f *.db

# Build frontend
frontend:
	@echo "Building frontend..."
	cd frontend && npm install && npm run build

# Docker production
docker:
	@echo "Building and starting with Docker Compose (production)..."
	docker-compose down
	docker-compose build
	docker-compose up -d

# Docker development
docker-dev:
	@echo "Building and starting with Docker Compose (development)..."
	docker-compose -f docker-compose.dev.yml down
	docker-compose -f docker-compose.dev.yml build
	docker-compose -f docker-compose.dev.yml up -d

# Database migrations
migrate:
	@echo "Running database migrations..."
	go run cmd/server/main.go -migrate

# Stop Docker services
docker-stop:
	@echo "Stopping Docker services..."
	docker-compose down
	docker-compose -f docker-compose.dev.yml down

# View logs
logs:
	docker-compose logs -f update-service

# View development logs
logs-dev:
	docker-compose -f docker-compose.dev.yml logs -f update-service
