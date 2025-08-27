package handler

import (
	"context"
	"fmt"
	"log"

	"github.com/openlearnnitj/openlearn-monitoring/internal/monitoring"
	"github.com/openlearnnitj/openlearn-monitoring/internal/storage"
)

// Handler manages the Lambda function execution
type Handler struct {
	monitoringService *monitoring.Service
	storageService    *storage.Service
}

// NewHandler creates a new handler instance
func NewHandler(monitoringService *monitoring.Service, storageService *storage.Service) *Handler {
	return &Handler{
		monitoringService: monitoringService,
		storageService:    storageService,
	}
}

// Handle processes the Lambda function request
func (h *Handler) Handle(ctx context.Context) error {
	log.Println("Starting monitoring service execution")

	// Perform health check
	result, err := h.monitoringService.CheckHealth()
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	log.Printf("Health check completed successfully. Found %d components. Total response time: %dms",
		len(result.Components), result.TotalResponseTimeMs)

	// Store results in DynamoDB
	if err := h.storageService.StoreResults(ctx, result); err != nil {
		return fmt.Errorf("failed to store results: %w", err)
	}

	log.Printf("Successfully stored %d component statuses to DynamoDB", len(result.Components))

	return nil
}
