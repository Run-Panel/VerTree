.PHONY: help build run dev test clean docker docker-postgres docker-dev docker-china docker-china-redis docker-china-prod docker-stop docker-stop-sqlite docker-stop-postgres docker-stop-china logs logs-postgres logs-dev logs-china migrate frontend

# Default target
help:
	@echo "RunPanel Update Service - Available commands:"
	@echo "  make build              - Build the Go application"
	@echo "  make run                - Run the application locally"
	@echo "  make dev                - Run in development mode with hot reload"
	@echo "  make test               - Run tests"
	@echo "  make clean              - Clean build artifacts"
	@echo "  make docker             - Build and run with Docker Compose (SQLite)"
	@echo "  make docker-postgres    - Build and run with Docker Compose (PostgreSQL)"
	@echo "  make docker-dev         - Build and run with Docker Compose (development)"
	@echo "  make docker-china       - Build and run with Docker Compose (中国大陆优化版本)"
	@echo "  make docker-china-redis - Build and run with Redis (中国大陆版本)"
	@echo "  make docker-china-prod  - Build and run with Nginx+Redis (中国大陆生产版本)"
	@echo "  make docker-stop        - Stop all Docker services"
	@echo "  make docker-stop-sqlite - Stop SQLite Docker services"
	@echo "  make docker-stop-postgres - Stop PostgreSQL Docker services"
	@echo "  make docker-stop-china  - Stop China optimized Docker services"
	@echo "  make logs               - View SQLite container logs"
	@echo "  make logs-postgres      - View PostgreSQL container logs"
	@echo "  make logs-dev           - View development container logs"
	@echo "  make logs-china         - View China optimized container logs"
	@echo "  make migrate            - Run database migrations"
	@echo "  make frontend           - Build frontend"

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
	rm -f data/*.db

# Build frontend
frontend:
	@echo "Building frontend..."
	cd frontend && npm install && npm run build

# Docker with SQLite (default for development)
docker:
	@echo "Building and starting with Docker Compose (SQLite)..."
	docker-compose -f docker-compose.sqlite.yml down
	docker-compose -f docker-compose.sqlite.yml build
	docker-compose -f docker-compose.sqlite.yml up -d

# Docker with PostgreSQL (production)
docker-postgres:
	@echo "Building and starting with Docker Compose (PostgreSQL)..."
	docker-compose -f docker-compose.postgres.yml down
	docker-compose -f docker-compose.postgres.yml build
	docker-compose -f docker-compose.postgres.yml up -d

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
	docker-compose -f docker-compose.sqlite.yml down
	docker-compose -f docker-compose.postgres.yml down
	docker-compose -f docker-compose.dev.yml down
	docker-compose -f docker-compose.china.yml down

# Stop SQLite Docker services
docker-stop-sqlite:
	@echo "Stopping SQLite Docker services..."
	docker-compose -f docker-compose.sqlite.yml down

# Stop PostgreSQL Docker services
docker-stop-postgres:
	@echo "Stopping PostgreSQL Docker services..."
	docker-compose -f docker-compose.postgres.yml down

# View logs (SQLite version)
logs:
	docker-compose -f docker-compose.sqlite.yml logs -f vertree-app

# View PostgreSQL logs
logs-postgres:
	docker-compose -f docker-compose.postgres.yml logs -f update-service

# View development logs
logs-dev:
	docker-compose -f docker-compose.dev.yml logs -f update-service

# Docker with China optimization (SQLite + China mirrors)
docker-china:
	@echo "Building and starting with Docker Compose (China optimized)..."
	docker-compose -f docker-compose.china.yml down
	docker-compose -f docker-compose.china.yml build
	docker-compose -f docker-compose.china.yml up -d

# Docker with China optimization + Redis
docker-china-redis:
	@echo "Building and starting with Docker Compose (China optimized + Redis)..."
	docker-compose -f docker-compose.china.yml down
	docker-compose -f docker-compose.china.yml build
	docker-compose -f docker-compose.china.yml --profile with-redis up -d

# Docker with China optimization + Redis + Nginx (Production)
docker-china-prod:
	@echo "Building and starting with Docker Compose (China production)..."
	docker-compose -f docker-compose.china.yml down
	docker-compose -f docker-compose.china.yml build
	docker-compose -f docker-compose.china.yml --profile production --profile with-redis up -d

# Stop China optimized Docker services
docker-stop-china:
	@echo "Stopping China optimized Docker services..."
	docker-compose -f docker-compose.china.yml down

# View China optimized logs
logs-china:
	docker-compose -f docker-compose.china.yml logs -f vertree-app
