#!/bin/bash

# Deploy script for AWS Lambda using SAM
set -e

echo "Deploying OpenLearn Monitoring Service..."

# Check if SAM CLI is installed
if ! command -v sam &> /dev/null; then
    echo "AWS SAM CLI is not installed. Please install it first:"
    echo "   https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/serverless-sam-cli-install.html"
    exit 1
fi

# Check if AWS CLI is configured
if ! aws sts get-caller-identity &> /dev/null; then
    echo "AWS CLI is not configured. Please run 'aws configure' first."
    exit 1
fi

# Set default values
STACK_NAME="openlearn-monitoring"
REGION="ap-south-1"
API_SECRET=""

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --stack-name)
            STACK_NAME="$2"
            shift 2
            ;;
        --region)
            REGION="$2"
            shift 2
            ;;
        --api-secret)
            API_SECRET="$2"
            shift 2
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --stack-name   CloudFormation stack name (default: openlearn-monitoring)"
            echo "  --region       AWS region (default: ap-south-1)"
            echo "  --api-secret   API secret for monitoring endpoint (required)"
            echo "  --help         Show this help message"
            exit 0
            ;;
        *)
            echo "Unknown option $1"
            exit 1
            ;;
    esac
done

# Check if API secret is provided
if [ -z "$API_SECRET" ]; then
    echo "❌ API secret is required. Use --api-secret flag."
    exit 1
fi

echo "Configuration:"
echo "   Stack Name: $STACK_NAME"
echo "   Region: $REGION"
echo "   API Secret: ***hidden***"
echo ""

# Build the Lambda function
echo "Building Lambda function..."
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o bootstrap ./cmd/lambda

# Deploy using SAM
echo "Deploying to AWS..."
sam deploy \
    --template-file template.yaml \
    --stack-name "$STACK_NAME" \
    --region "$REGION" \
    --capabilities CAPABILITY_IAM \
    --parameter-overrides \
        MonitoringAPISecret="$API_SECRET" \
    --no-confirm-changeset \
    --no-fail-on-empty-changeset

# Clean up
rm -f bootstrap

echo ""
echo "✅ Deployment completed successfully!"
echo ""
echo "📊 You can check the monitoring status in:"
echo "   - Lambda Console: https://$REGION.console.aws.amazon.com/lambda/home?region=$REGION#/functions/openlearn-monitoring"
echo "   - DynamoDB Console: https://$REGION.console.aws.amazon.com/dynamodb/home?region=$REGION#tables"
echo "   - CloudWatch Logs: https://$REGION.console.aws.amazon.com/cloudwatch/home?region=$REGION#logsV2:log-groups/log-group/%2Faws%2Flambda%2Fopenlearn-monitoring"
echo ""
echo "🎯 The Lambda function will run every minute automatically."
