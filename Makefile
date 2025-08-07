# Makefile for Golang Load Balancer Project

# Variables
BINARY_NAME=backend-server
BACKEND_DIR=cmd/backend-server
FRONTEND_DIR=frontend
PORT=8080
FE_PORT=5173

# Default target
.PHONY: help
help:
	@echo "Available commands:"
	@echo ""
	@echo "Backend:"
	@echo "  make be-start   - Start the backend server"
	@echo "  make be-build   - Build the backend binary"
	@echo "  make be-dev     - Run backend in development mode with auto-restart"
	@echo "  make be-stop    - Stop running backend processes"
	@echo ""
	@echo "Frontend:"
	@echo "  make fe-start   - Start the frontend development server"
	@echo "  make fe-install - Install frontend dependencies"
	@echo "  make fe-build   - Build frontend for production"
	@echo "  make fe-lint    - Lint frontend code"
	@echo "  make fe-preview - Preview production build"
	@echo ""
	@echo "General:"
	@echo "  make deps       - Download all dependencies"
	@echo "  make clean      - Clean all build artifacts"
	@echo "  make test       - Run all tests"
	@echo "  make start      - Start both frontend and backend"

# Backend Commands

# Start the backend server
.PHONY: be-start
be-start: be-deps
	@echo "Starting backend server on port $(PORT)..."
	cd $(BACKEND_DIR) && go run main.go middleware.go

# Build the backend binary
.PHONY: be-build
be-build: be-deps
	@echo "Building backend server..."
	cd $(BACKEND_DIR) && go build -o ../../$(BINARY_NAME) .
	@echo "Binary built: $(BINARY_NAME)"

# Run the built binary
.PHONY: be-run-binary
be-run-binary: be-build
	@echo "Starting backend server from binary..."
	./$(BINARY_NAME)

# Backend development mode with file watching (requires air)
.PHONY: be-dev
be-dev:
	@if command -v air > /dev/null; then \
		echo "Starting backend development server with hot reload..."; \
		cd $(BACKEND_DIR) && air; \
	else \
		echo "Air not installed. Installing..."; \
		go install github.com/cosmtrek/air@latest; \
		echo "Starting backend development server with hot reload..."; \
		cd $(BACKEND_DIR) && air; \
	fi

# Stop backend server processes
.PHONY: be-stop
be-stop:
	@echo "Stopping backend server processes..."
	@pkill -f "$(BINARY_NAME)" || true
	@pkill -f "go run.*main.go" || true
	@echo "Backend server processes stopped"

# Download backend dependencies
.PHONY: be-deps
be-deps:
	@echo "Downloading backend dependencies..."
	go mod download
	go mod tidy

# Frontend Commands

# Start the frontend development server
.PHONY: fe-start
fe-start: fe-install
	@echo "Starting frontend development server on port $(FE_PORT)..."
	cd $(FRONTEND_DIR) && npm run dev

# Install frontend dependencies
.PHONY: fe-install
fe-install:
	@echo "Installing frontend dependencies..."
	cd $(FRONTEND_DIR) && npm install

# Build frontend for production
.PHONY: fe-build
fe-build: fe-install
	@echo "Building frontend for production..."
	cd $(FRONTEND_DIR) && npm run build

# Lint frontend code
.PHONY: fe-lint
fe-lint: fe-install
	@echo "Linting frontend code..."
	cd $(FRONTEND_DIR) && npm run lint

# Preview frontend production build
.PHONY: fe-preview
fe-preview: fe-build
	@echo "Starting frontend preview server..."
	cd $(FRONTEND_DIR) && npm run preview

# Stop frontend server processes
.PHONY: fe-stop
fe-stop:
	@echo "Stopping frontend server processes..."
	@pkill -f "vite" || true
	@echo "Frontend server processes stopped"

# General Commands

# Download all dependencies
.PHONY: deps
deps: be-deps fe-install
	@echo "All dependencies installed"

# Clean all build artifacts
.PHONY: clean
clean:
	@echo "Cleaning all build artifacts..."
	rm -f $(BINARY_NAME)
	go clean
	cd $(FRONTEND_DIR) && rm -rf dist node_modules/.vite
	@echo "Clean completed"

# Run all tests
.PHONY: test
test:
	@echo "Running backend tests..."
	go test ./...
	@echo "Running frontend lint check..."
	cd $(FRONTEND_DIR) && npm run lint

# Start both frontend and backend
.PHONY: start
start:
	@echo "Starting both frontend and backend..."
	@echo "Backend will start on port $(PORT), frontend on port $(FE_PORT)"
	@make be-start &
	@sleep 2
	@make fe-start

# Stop all processes
.PHONY: stop
stop: be-stop fe-stop
	@echo "All processes stopped"

# Development Tools

# Format backend code
.PHONY: be-fmt
be-fmt:
	@echo "Formatting backend code..."
	go fmt ./...

# Lint backend code (requires golangci-lint)
.PHONY: be-lint
be-lint:
	@if command -v golangci-lint > /dev/null; then \
		echo "Linting backend code..."; \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Install development tools
.PHONY: install-tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/cosmtrek/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Development tools installed"

# Monitoring

# Check backend server health
.PHONY: be-health
be-health:
	@echo "Checking backend server health..."
	@curl -s http://localhost:$(PORT)/health || echo "Backend server not responding"

# Show backend server status
.PHONY: be-status
be-status:
	@echo "Getting backend server status..."
	@curl -s http://localhost:$(PORT)/api/v1/status | jq . || echo "Backend server not responding or jq not installed" 