# Makefile for URL Crawler Backend (Docker Only)

.PHONY: help docker-build docker-run docker-stop seed logs logs-backend logs-mysql restart status docker-clean

# Default target
help:
	@echo "Available targets:"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run with Docker Compose"
	@echo "  docker-stop  - Stop Docker Compose services"
	@echo "  logs         - View all logs"
	@echo "  logs-backend - View backend logs"
	@echo "  logs-mysql   - View MySQL logs"
	@echo "  restart      - Restart services"
	@echo "  status       - Show service status"
	@echo "  docker-clean - Clean Docker resources"

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t url-crawler .

# Run with Docker Compose
docker-run:
	@echo "Starting services with Docker Compose..."
	docker compose up

# Stop Docker Compose services
docker-stop:
	@echo "Stopping Docker Compose services..."
	docker compose down


# View logs
logs:
	docker compose logs -f

# View logs for specific service
logs-backend:
	docker compose logs -f backend

logs-mysql:
	docker compose logs -f mysql

# Restart services
restart:
	docker compose restart

# Show service status
status:
	docker compose ps

# Clean Docker resources
docker-clean:
	@echo "Cleaning Docker resources..."
	docker compose down -v --remove-orphans
	docker system prune -f 