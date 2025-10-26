.PHONY: help up down build restart logs clean test migrate-up migrate-down

# Default target
help:
	@echo "Available commands:"
	@echo "  make up              - Start all services"
	@echo "  make down            - Stop all services"
	@echo "  make build           - Build all services"
	@echo "  make restart         - Restart all services"
	@echo "  make logs            - View logs from all services"
	@echo "  make logs-backend    - View backend logs"
	@echo "  make logs-frontend   - View frontend logs"
	@echo "  make logs-ocr        - View OCR service logs"
	@echo "  make clean           - Clean up containers and volumes"
	@echo "  make test            - Run all tests"
	@echo "  make test-backend    - Run backend tests"
	@echo "  make test-frontend   - Run frontend tests"
	@echo "  make migrate-up      - Run database migrations"
	@echo "  make migrate-down    - Rollback database migrations"
	@echo "  make shell-backend   - Open shell in backend container"
	@echo "  make shell-ocr       - Open shell in OCR service container"

# Docker Compose commands
up:
	docker-compose up -d

down:
	docker-compose down

build:
	docker-compose build

restart:
	docker-compose restart

logs:
	docker-compose logs -f

logs-backend:
	docker-compose logs -f backend

logs-frontend:
	docker-compose logs -f frontend

logs-ocr:
	docker-compose logs -f ocr-service

# Clean up
clean:
	docker-compose down -v
	rm -rf storage/uploads/* storage/results/* storage/temp/* storage/thumbnails/*
	@echo "Cleaned up containers, volumes, and storage"

# Database migrations
migrate-up:
	docker-compose exec backend ./scripts/migrate up

migrate-down:
	docker-compose exec backend ./scripts/migrate down

# Testing
test:
	@echo "Running all tests..."
	@make test-backend
	@make test-frontend

test-backend:
	@echo "Running backend tests..."
	cd backend && go test -v ./...

test-frontend:
	@echo "Running frontend tests..."
	cd frontend && npm run test

# Shell access
shell-backend:
	docker-compose exec backend sh

shell-ocr:
	docker-compose exec ocr-service bash

shell-db:
	docker-compose exec postgres psql -U ocr_user -d ocr_db

# Development helpers
dev-backend:
	cd backend && go run cmd/server/main.go

dev-frontend:
	cd frontend && npm run dev

dev-ocr:
	cd ocr-service && python main.py

# Production
prod-up:
	docker-compose --profile production up -d

prod-build:
	docker-compose --profile production build

# Monitoring
stats:
	docker stats

ps:
	docker-compose ps
