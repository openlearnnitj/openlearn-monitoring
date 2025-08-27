#!/bin/bash

# Build script for AWS Lambda deployment
set -e

echo "Building AWS Lambda function..."

# Set build environment for Linux (Lambda runtime)
export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=0

# Build the Lambda function
go build -ldflags="-s -w" -o bootstrap ./cmd/lambda

# Create deployment package
zip -r monitoring-lambda.zip bootstrap

echo "Build completed successfully!"
echo "Deployment package: monitoring-lambda.zip"

# Clean up
rm bootstrap

echo "Ready for AWS Lambda deployment!"
