package storage

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/openlearnnitj/openlearn-monitoring/internal/models"
)

// DynamoDBClient wraps the AWS DynamoDB client
type DynamoDBClient struct {
	client *dynamodb.Client
}

// GetClient returns the underlying DynamoDB client
func (d *DynamoDBClient) GetClient() *dynamodb.Client {
	return d.client
}

// NewDynamoDBClient creates a new DynamoDB client
func NewDynamoDBClient(region string) (*DynamoDBClient, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &DynamoDBClient{
		client: dynamodb.NewFromConfig(cfg),
	}, nil
}

// Service handles storage operations
type Service struct {
	client    *DynamoDBClient
	tableName string
}

// NewService creates a new storage service
func NewService(client *DynamoDBClient, tableName string) *Service {
	return &Service{
		client:    client,
		tableName: tableName,
	}
}

// StoreResults stores monitoring results concurrently in DynamoDB
func (s *Service) StoreResults(ctx context.Context, result *models.MonitoringResult) error {
	if len(result.Components) == 0 {
		return fmt.Errorf("no components to store")
	}

	// Use goroutines for concurrent writes
	var wg sync.WaitGroup
	errCh := make(chan error, len(result.Components))

	for _, component := range result.Components {
		wg.Add(1)
		go func(comp models.Component) {
			defer wg.Done()
			if err := s.storeComponent(ctx, comp, result); err != nil {
				errCh <- err
			}
		}(component)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errCh)

	// Collect any errors
	var errors []error
	for err := range errCh {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to store %d components: %v", len(errors), errors)
	}

	return nil
}

// storeComponent stores a single component in DynamoDB
func (s *Service) storeComponent(ctx context.Context, component models.Component, result *models.MonitoringResult) error {
	item := map[string]types.AttributeValue{
		"serviceName": &types.AttributeValueMemberS{
			Value: component.Name,
		},
		"status": &types.AttributeValueMemberS{
			Value: component.Status,
		},
		"internalResponseTimeMs": &types.AttributeValueMemberN{
			Value: fmt.Sprintf("%.2f", component.ResponseTimeMs),
		},
		"totalResponseTimeMs": &types.AttributeValueMemberN{
			Value: fmt.Sprintf("%d", result.TotalResponseTimeMs),
		},
		"lastChecked": &types.AttributeValueMemberS{
			Value: result.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
		},
	}

	_, err := s.client.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(s.tableName),
		Item:      item,
	})

	if err != nil {
		return fmt.Errorf("failed to store component %s: %w", component.Name, err)
	}

	return nil
}
