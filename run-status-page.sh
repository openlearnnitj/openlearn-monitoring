#!/bin/bash

# Local development script for the status page
set -e

echo "🌐 Starting OpenLearn Status Page..."

# Check if required environment variables are set
required_vars=("MONITORING_API_URL" "MONITORING_API_SECRET" "DYNAMODB_TABLE_NAME" "AWS_REGION")
missing_vars=()

for var in "${required_vars[@]}"; do
    if [ -z "${!var}" ]; then
        missing_vars+=("$var")
    fi
done

if [ ${#missing_vars[@]} -ne 0 ]; then
    echo "❌ Missing required environment variables:"
    for var in "${missing_vars[@]}"; do
        echo "   - $var"
    done
    echo ""
    echo "💡 Example:"
    echo "export MONITORING_API_URL='https://api.openlearn.org.in/api/monitoring/health-status'"
    echo "export MONITORING_API_SECRET='your-secret-here'"
    echo "export DYNAMODB_TABLE_NAME='OpenLearnStatus'"
    echo "export AWS_REGION='ap-south-1'"
    echo "export PORT='8080'  # Optional, defaults to 8080"
    exit 1
fi

# Check if AWS credentials are configured
if ! aws sts get-caller-identity &> /dev/null; then
    echo "⚠️  Warning: AWS CLI is not configured. The status page may not work properly."
    echo "   Please run 'aws configure' or set AWS credentials in environment variables."
fi

echo "📋 Configuration:"
echo "   API URL: $MONITORING_API_URL"
echo "   Table: $DYNAMODB_TABLE_NAME"
echo "   Region: $AWS_REGION"
echo "   Port: ${PORT:-8080}"
echo ""

# Build and run the status page
echo "🔨 Building status page..."
go build -o status-page ./cmd/status-page

echo "🚀 Starting server..."
echo "📱 Visit: http://localhost:${PORT:-8080}"
echo "🔄 API endpoint: http://localhost:${PORT:-8080}/api/status"
echo ""
echo "Press Ctrl+C to stop the server"
echo ""

./status-page
