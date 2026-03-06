.PHONY: run build test migrate seed clean deps lint help

APP_NAME = task-manager
GO_FILES = $(shell find . -type f -name '*.go' -not -path "./vendor/*")
GO_PKGS = $(shell go list ./...)

# Colors for output
GREEN := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
RESET := $(shell tput -Txterm sgr0)

## help: Show this help message
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## run: Run the application
run:
	@echo "${YELLOW}Starting application...${RESET}"
	go run ./cmd/api/main.go

## build: Build the application
build:
	@echo "${YELLOW}Building application...${RESET}"
	go build -o bin/$(APP_NAME) ./cmd/api/main.go

## test: Run tests
test:
	@echo "${YELLOW}Running tests...${RESET}"
	go test -v -cover ./...

## test-race: Run tests with race detection
test-race:
	@echo "${YELLOW}Running tests with race detection...${RESET}"
	go test -race -v ./...

## test-integration: Run integration tests
test-integration:
	@echo "${YELLOW}Running integration tests...${RESET}"
	go test -tags=integration -v ./test/integration/...

## coverage: Run tests with coverage
coverage:
	@echo "${YELLOW}Running tests with coverage...${RESET}"
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "${GREEN}Coverage report generated: coverage.html${RESET}"

## migrate: Run database migrations
migrate:
	@echo "${YELLOW}Running migrations...${RESET}"
	go run scripts/migrate.go

## seed: Seed the database with sample data
seed:
	@echo "${YELLOW}Seeding database...${RESET}"
	go run scripts/seed.go

## deps: Download dependencies
deps:
	@echo "${YELLOW}Downloading dependencies...${RESET}"
	go mod download
	go mod tidy

## lint: Run linter
lint:
	@echo "${YELLOW}Running linter...${RESET}"
	golangci-lint run ./...

## fmt: Format code
fmt:
	@echo "${YELLOW}Formatting code...${RESET}"
	go fmt ./...

## vet: Run go vet
vet:
	@echo "${YELLOW}Running go vet...${RESET}"
	go vet ./...

## clean: Clean build artifacts
clean:
	@echo "${YELLOW}Cleaning...${RESET}"
	rm -rf bin/
	rm -f coverage.out coverage.html
	go clean

## dev: Run with hot reload (requires air)
dev:
	@echo "${YELLOW}Starting development server with hot reload...${RESET}"
	air -c .air.toml

## create-migration: Create a new migration file
create-migration:
	@read -p "Enter migration name: " name; \
	timestamp=$$(date +%Y%m%d%H%M%S); \
	filename=scripts/migrations/$${timestamp}_$${name}.sql; \
	touch $$filename; \
	echo "Created migration: $$filename"

## db-drop: Drop database (use with caution!)
db-drop:
	@echo "${RED}WARNING: This will drop the database!${RESET}"
	@read -p "Are you sure? (y/N): " confirm; \
	if [ "$$confirm" = "y" ]; then \
		echo "${YELLOW}Dropping database...${RESET}"; \
		dropdb --if-exists $(shell go run scripts/get_db_name.go); \
		createdb $(shell go run scripts/get_db_name.go); \
		echo "${GREEN}Database recreated${RESET}"; \
	fi

## db-reset: Reset database (drop, create, migrate, seed)
db-reset: db-drop migrate seed
	@echo "${GREEN}Database reset complete${RESET}"

## swagger: Generate Swagger documentation
swagger:
	@echo "${YELLOW}Generating Swagger documentation...${RESET}"
	swag init -g cmd/api/main.go -o api/docs

## install-tools: Install development tools
install-tools:
	@echo "${YELLOW}Installing development tools...${RESET}"
	go install github.com/cosmtrek/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/swaggo/swag/cmd/swag@latest

## all: Run all checks
all: deps fmt vet lint test build
	@echo "${GREEN}All checks passed!${RESET}"

.DEFAULT_GOAL := help