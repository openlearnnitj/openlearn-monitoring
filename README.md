# OpenLearn Monitoring Service

A production-ready monitoring system that tracks the health of OpenLearn's API services and displays a beautiful status page inspired by modern SaaS companies like Anthropic.


## Features

- **Automated Monitoring**: AWS Lambda function runs every 60 seconds
- **Concurrent Processing**: Parallel DynamoDB writes using goroutines
- **Beautiful Status Page**: Modern UI inspired by Anthropic's design
- **Uptime Tracking**: 24h, 7d, and 30d uptime statistics
- **Responsive Design**: Works perfectly on mobile and desktop
- **Visual History**: 90-day status history with color-coded bars
- **Easy Deployment**: Multiple deployment options (Lambda, Container, Binary)

## Environment Variables

The Lambda function requires the following environment variables:

- `MONITORING_API_URL`: Full URL to the health check endpoint (e.g., `https://api.openlearn.org.in/api/monitoring/health-status`)
- `MONITORING_API_SECRET`: Shared secret for API authentication
- `DYNAMODB_TABLE_NAME`: Name of the DynamoDB table (e.g., `OpenLearnStatus`)
- `AWS_REGION`: AWS region for DynamoDB (e.g., `ap-south-1`)

## DynamoDB Table Structure

The Lambda function stores data in a DynamoDB table with the following structure:

- **Partition Key**: `serviceName` (String)
- **Attributes**:
  - `status` (String): Service status (e.g., "OPERATIONAL")
  - `internalResponseTimeMs` (Number): Response time from the component
  - `totalResponseTimeMs` (Number): Total round-trip latency
  - `lastChecked` (String): ISO timestamp of the check

## Building and Deployment

### Prerequisites

- Go 1.22+ installed
- AWS CLI configured with appropriate permissions
- DynamoDB table created with `serviceName` as partition key

### Build

```bash
# Make the build script executable (if not already)
chmod +x build.sh

# Build the Lambda deployment package
./build.sh
```

This creates `monitoring-lambda.zip` ready for AWS Lambda deployment.

### Manual Build (Alternative)

```bash
# Set environment for Linux (Lambda runtime)
export GOOS=linux GOARCH=amd64 CGO_ENABLED=0

# Build the binary
go build -ldflags="-s -w" -o bootstrap ./cmd/lambda

# Create deployment package
zip monitoring-lambda.zip bootstrap

# Clean up
rm bootstrap
```

### AWS Lambda Deployment

1. **Create Lambda Function**:
   ```bash
   aws lambda create-function \
     --function-name openlearn-monitoring \
     --runtime provided.al2 \
     --role arn:aws:iam::YOUR-ACCOUNT:role/lambda-execution-role \
     --handler bootstrap \
     --zip-file fileb://monitoring-lambda.zip
   ```

2. **Set Environment Variables**:
   ```bash
   aws lambda update-function-configuration \
     --function-name openlearn-monitoring \
     --environment Variables='{
       "MONITORING_API_URL":"https://api.openlearn.org.in/api/monitoring/health-status",
       "MONITORING_API_SECRET":"your-secret-here",
       "DYNAMODB_TABLE_NAME":"OpenLearnStatus",
       "AWS_REGION":"ap-south-1"
     }'
   ```

3. **Update Function Code** (for subsequent deployments):
   ```bash
   aws lambda update-function-code \
     --function-name openlearn-monitoring \
     --zip-file fileb://monitoring-lambda.zip
   ```

### IAM Permissions

The Lambda execution role needs the following permissions:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "arn:aws:logs:*:*:*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "dynamodb:PutItem"
      ],
      "Resource": "arn:aws:dynamodb:*:*:table/OpenLearnStatus"
    }
  ]
}
```

### Scheduling

To run the monitoring function on a schedule, create a CloudWatch Events rule:

```bash
# Create rule to run every 5 minutes
aws events put-rule \
  --name openlearn-monitoring-schedule \
  --schedule-expression "rate(5 minutes)"

# Add Lambda function as target
aws events put-targets \
  --rule openlearn-monitoring-schedule \
  --targets "Id"="1","Arn"="arn:aws:lambda:REGION:ACCOUNT:function:openlearn-monitoring"

# Grant permission for CloudWatch Events to invoke Lambda
aws lambda add-permission \
  --function-name openlearn-monitoring \
  --statement-id allow-cloudwatch \
  --action lambda:InvokeFunction \
  --principal events.amazonaws.com \
  --source-arn arn:aws:events:REGION:ACCOUNT:rule/openlearn-monitoring-schedule
```

## Testing

You can test the Lambda function locally using the AWS SAM CLI or invoke it directly:

```bash
# Test invoke (requires AWS CLI and proper credentials)
aws lambda invoke \
  --function-name openlearn-monitoring \
  --payload '{}' \
  response.json

# Check the response
cat response.json
```

## Monitoring and Logs

View Lambda function logs:

```bash
aws logs describe-log-groups --log-group-name-prefix /aws/lambda/openlearn-monitoring
aws logs tail /aws/lambda/openlearn-monitoring --follow
```

## Troubleshooting

Common issues and solutions:

1. **Timeout**: Increase Lambda timeout if health checks take longer than expected
2. **DynamoDB Throttling**: Monitor DynamoDB metrics and adjust provisioned capacity
3. **Network Issues**: Ensure Lambda has internet access for API calls
4. **Permissions**: Verify IAM role has required DynamoDB and CloudWatch Logs permissions

## Development

To modify the monitoring service:

1. Make changes to the appropriate internal packages
2. Run tests: `go test ./...`
3. Build and redeploy: `./build.sh && aws lambda update-function-code ...`

The architecture is designed to be easily extensible - you can add new monitoring endpoints, change storage backends, or modify the data structure with minimal changes.
