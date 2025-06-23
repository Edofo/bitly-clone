
.PHONY: help build test lint clean run migrate create stats health full-test

BINARY_NAME=url-shortener
TEST_URL=https://www.google.com
SERVER_PORT=8080

GREEN=\033[0;32m
YELLOW=\033[1;33m
RED=\033[0;31m
NC=\033[0m

help:
	@echo "$(GREEN)Url-Shortener - Commands Available:$(NC)"
	@echo ""
	@echo "$(YELLOW)Build and Tests:$(NC)"
	@echo "  make build      - Compile the project"
	@echo "  make test       - Run unit tests"
	@echo "  make lint       - Check code quality"
	@echo "  make full-test  - Tests complets (lint + test + build + run)"
	@echo ""
	@echo "$(YELLOW)Database:$(NC)"
	@echo "  make migrate    - Run migrations"
	@echo ""
	@echo "$(YELLOW)Server:$(NC)"
	@echo "  make run        - Start the server"
	@echo "  make health     - Test the health endpoint"
	@echo ""
	@echo "$(YELLOW)CLI:$(NC)"
	@echo "  make create     - Create a short URL"
	@echo "  make stats      - Display statistics"
	@echo ""
	@echo "$(YELLOW)Maintenance:$(NC)"
	@echo "  make clean      - Clean temporary files"
	@echo "  make help       - Display this help"

build:
	@echo "$(GREEN)Building $(BINARY_NAME)...$(NC)"
	@go build -o $(BINARY_NAME) .
	@echo "$(GREEN)Build terminé!$(NC)"

test:
	@echo "$(GREEN)Running unit tests...$(NC)"
	@go test -v -race -coverprofile=coverage.out ./...
	@echo "$(GREEN)Tests done!$(NC)"
	@if command -v go tool cover >/dev/null 2>&1; then \
		echo "$(YELLOW)Generating coverage report...$(NC)"; \
		go tool cover -html=coverage.out -o coverage.html; \
		echo "$(GREEN)Coverage report generated: coverage.html$(NC)"; \
	fi

lint:
	@echo "$(GREEN)Checking code quality...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "$(YELLOW)golangci-lint not found, using go vet...$(NC)"; \
		go vet ./...; \
	fi
	@echo "$(GREEN)Lint done!$(NC)"

clean:
	@echo "$(GREEN)Cleaning temporary files...$(NC)"
	@rm -f $(BINARY_NAME) url_shortener.db coverage.out coverage.html
	@rm -f .test_short_code .test_full_url
	@go clean -cache -modcache -testcache
	@echo "$(GREEN)Cleaning done!$(NC)"

migrate: build
	@echo "$(GREEN)Running migrations...$(NC)"
	@./$(BINARY_NAME) migrate
	@echo "$(GREEN)Migrations done!$(NC)"

run: build
	@echo "$(GREEN)Starting server...$(NC)"
	@echo "$(YELLOW)The server starts on http://localhost:$(SERVER_PORT)$(NC)"
	@echo "$(YELLOW)Press Ctrl+C to stop$(NC)"
	@./$(BINARY_NAME) run-server

health:
	@echo "$(GREEN)Testing health endpoint...$(NC)"
	@curl -s http://localhost:$(SERVER_PORT)/health || echo "$(RED)Server not accessible$(NC)"

create: build
	@echo "$(GREEN)Creating short URL...$(NC)"
	@./$(BINARY_NAME) create --url="$(TEST_URL)"

stats: build
	@echo "$(GREEN)Displaying statistics...$(NC)"
	@if [ -f .test_short_code ]; then \
		./$(BINARY_NAME) stats --code=$$(cat .test_short_code); \
	else \
		echo "$(YELLOW)No test code found. Create a URL with 'make create'$(NC)"; \
	fi

full-test: lint test build migrate
	@echo "$(GREEN)All tests passed!$(NC)"

full-test-with-server: full-test
	@echo "$(GREEN)Tests with server...$(NC)"
	@echo "$(YELLOW)Make sure the server is running on port $(SERVER_PORT)$(NC)"
	@sleep 2
	@make health
	@echo "$(GREEN)All tests passed!$(NC)"

deps:
	@echo "$(GREEN)Downloading dependencies...$(NC)"
	@go mod tidy
	@go mod download
	@echo "$(GREEN)Dependencies installed!$(NC)"

dev-tools:
	@echo "$(GREEN)Installing development tools...$(NC)"
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "$(YELLOW)Installing golangci-lint...$(NC)"; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@if ! command -v jq >/dev/null 2>&1; then \
		echo "$(YELLOW)Installing jq...$(NC)"; \
		if command -v brew >/dev/null 2>&1; then \
			brew install jq; \
		elif command -v apt-get >/dev/null 2>&1; then \
			sudo apt-get install -y jq; \
		else \
			echo "$(RED)jq not installed. Install it manually.$(NC)"; \
		fi; \
	fi
	@echo "$(GREEN)Development tools installed!$(NC)"

dev: dev-tools deps full-test
	@echo "$(GREEN)Development environment ready!$(NC)"
	@echo "$(YELLOW)Run 'make run' to start the server$(NC)"

prod: clean deps full-test
	@echo "$(GREEN)Production build ready!$(NC)"
	@echo "$(YELLOW)The $(BINARY_NAME) binary is ready for deployment$(NC)" 