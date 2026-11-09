.PHONY: all build run test clean docker-up docker-down dev build-frontend dev-frontend

all: build build-frontend

# Build Go API Server binary
build:
	@echo "Building Go API Server binary..."
	@cd backend && go build -o bin/server ./cmd/server
	@echo "Build successful -> backend/bin/server"

# Build React SPA Production bundle
build-frontend:
	@echo "Building React SPA bundle..."
	@cd frontend && npm run build
	@echo "Build successful -> frontend/dist"

# Run Go API Server locally
run:
	@cd backend && go run ./cmd/server

# Run React SPA in Vite Dev Mode
dev-frontend:
	@cd frontend && npm run dev

# Run unit tests
test:
	@echo "Running backend unit tests..."
	@cd backend && go test -v ./...

# Start PostgreSQL 16 container
docker-up:
	docker compose up -d

# Stop PostgreSQL container
docker-down:
	docker compose down

# Clean build artifacts
clean:
	@rm -rf backend/bin frontend/dist
