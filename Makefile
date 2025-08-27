.PHONY: build clean test deploy local-test build-status-page run-status-page docker-status-page

# Build the Lambda function for deployment
build:
	@echo "Building Lambda function..."
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o bootstrap ./cmd/lambda
	@zip -r monitoring-lambda.zip bootstrap
	@rm bootstrap
	@echo "Build completed: monitoring-lambda.zip"

# Build the status page application
build-status-page:
	@echo "Building status page..."
	@go build -ldflags="-s -w" -o status-page ./cmd/status-page
	@echo "Build completed: status-page"

# Run the status page locally
run-status-page: build-status-page
	@echo "Starting status page..."
	@./status-page

# Build Docker image for status page
docker-status-page:
	@echo "Building Docker image for status page..."
	@docker build -f Dockerfile.status-page -t openlearn-status-page .
	@echo "Docker image built: openlearn-status-page"

# Run status page in Docker
docker-run-status-page: docker-status-page
	@echo "Running status page in Docker..."
	@docker run -p 8080:8080 \
		-e MONITORING_API_URL="$$MONITORING_API_URL" \
		-e MONITORING_API_SECRET="$$MONITORING_API_SECRET" \
		-e DYNAMODB_TABLE_NAME="$$DYNAMODB_TABLE_NAME" \
		-e AWS_REGION="$$AWS_REGION" \
		-e AWS_ACCESS_KEY_ID="$$AWS_ACCESS_KEY_ID" \
		-e AWS_SECRET_ACCESS_KEY="$$AWS_SECRET_ACCESS_KEY" \
		openlearn-status-page

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f bootstrap monitoring-lambda.zip status-page
	@echo "Clean completed"

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Build and test
build-test: test build

# Deploy to AWS Lambda (requires AWS CLI configured)
deploy: build
	@echo "Deploying to AWS Lambda..."
	@aws lambda update-function-code \
		--function-name openlearn-monitoring \
		--zip-file fileb://monitoring-lambda.zip
	@echo "Deployment completed"

# Local development build (for testing compilation)
local-build:
	@echo "Building for local testing..."
	@go build -o monitoring-local ./cmd/lambda
	@echo "Local build completed: monitoring-local"

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Formatting completed"

# Lint code (requires golangci-lint)
lint:
	@echo "Linting code..."
	@golangci-lint run
	@echo "Linting completed"

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	@go mod tidy
	@echo "Dependencies tidied"

# Full check (format, lint, test, build)
check: fmt lint test build

# Development setup
dev-setup:
	@echo "Setting up development environment..."
	@go mod download
	@echo "Development setup completed"
