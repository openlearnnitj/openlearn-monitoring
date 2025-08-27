package status

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/openlearnnitj/openlearn-monitoring/internal/models"
)

// StatusService handles status page data operations
type StatusService struct {
	client    *dynamodb.Client
	tableName string
}

// NewStatusService creates a new status service
func NewStatusService(client *dynamodb.Client, tableName string) *StatusService {
	return &StatusService{
		client:    client,
		tableName: tableName,
	}
}

// ComponentStatus represents the current status of a component
type ComponentStatus struct {
	Name                   string    `json:"name"`
	Status                 string    `json:"status"`
	InternalResponseTimeMs float64   `json:"internalResponseTimeMs"`
	TotalResponseTimeMs    int64     `json:"totalResponseTimeMs"`
	LastChecked            time.Time `json:"lastChecked"`
	UptimePercent          float64   `json:"uptimePercent"`
	StatusHistory          []StatusPoint `json:"statusHistory"`
}

// StatusPoint represents a point in time status
type StatusPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Status    string    `json:"status"`
}

// SystemStatus represents the overall system status
type SystemStatus struct {
	OverallStatus string            `json:"overallStatus"`
	Components    []ComponentStatus `json:"components"`
	LastUpdated   time.Time         `json:"lastUpdated"`
	UptimeStats   UptimeStats       `json:"uptimeStats"`
}

// UptimeStats represents uptime statistics
type UptimeStats struct {
	Last24Hours float64 `json:"last24Hours"`
	Last7Days   float64 `json:"last7Days"`
	Last30Days  float64 `json:"last30Days"`
}

// GetCurrentStatus retrieves the current status of all components
func (s *StatusService) GetCurrentStatus(ctx context.Context) (*SystemStatus, error) {
	// Scan all items to get current status
	result, err := s.client.Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String(s.tableName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to scan DynamoDB: %w", err)
	}

	// Group items by service name and get latest status
	componentMap := make(map[string][]models.DynamoDBItem)
	
	for _, item := range result.Items {
		var dbItem models.DynamoDBItem
		
		// Manual parsing since we need to handle the timestamp
		if serviceName, ok := item["serviceName"].(*types.AttributeValueMemberS); ok {
			dbItem.ServiceName = serviceName.Value
		}
		if status, ok := item["status"].(*types.AttributeValueMemberS); ok {
			dbItem.Status = status.Value
		}
		if responseTime, ok := item["internalResponseTimeMs"].(*types.AttributeValueMemberN); ok {
			fmt.Sscanf(responseTime.Value, "%f", &dbItem.InternalResponseTimeMs)
		}
		if totalTime, ok := item["totalResponseTimeMs"].(*types.AttributeValueMemberN); ok {
			fmt.Sscanf(totalTime.Value, "%d", &dbItem.TotalResponseTimeMs)
		}
		if lastChecked, ok := item["lastChecked"].(*types.AttributeValueMemberS); ok {
			if timestamp, err := time.Parse(time.RFC3339, lastChecked.Value); err == nil {
				dbItem.LastChecked = timestamp.Format("2006-01-02T15:04:05Z07:00")
			}
		}
		
		componentMap[dbItem.ServiceName] = append(componentMap[dbItem.ServiceName], dbItem)
	}

	// Process each component
	var components []ComponentStatus
	overallOperational := true
	lastUpdated := time.Time{}

	for serviceName, items := range componentMap {
		// Sort by timestamp to get latest status
		sort.Slice(items, func(i, j int) bool {
			ti, _ := time.Parse("2006-01-02T15:04:05Z07:00", items[i].LastChecked)
			tj, _ := time.Parse("2006-01-02T15:04:05Z07:00", items[j].LastChecked)
			return ti.After(tj)
		})

		if len(items) == 0 {
			continue
		}

		latest := items[0]
		lastChecked, _ := time.Parse("2006-01-02T15:04:05Z07:00", latest.LastChecked)
		
		if lastChecked.After(lastUpdated) {
			lastUpdated = lastChecked
		}

		// Calculate uptime for different periods
		uptimeStats := s.calculateUptime(items)
		
		// Generate status history for the last 90 days (like Anthropic's style)
		statusHistory := s.generateStatusHistory(items, 90)

		component := ComponentStatus{
			Name:                   serviceName,
			Status:                 latest.Status,
			InternalResponseTimeMs: latest.InternalResponseTimeMs,
			TotalResponseTimeMs:    latest.TotalResponseTimeMs,
			LastChecked:            lastChecked,
			UptimePercent:          uptimeStats.Last24Hours,
			StatusHistory:          statusHistory,
		}

		if latest.Status != "OPERATIONAL" {
			overallOperational = false
		}

		components = append(components, component)
	}

	// Calculate overall uptime stats
	overallUptimeStats := s.calculateOverallUptime(components)

	overallStatus := "OPERATIONAL"
	if !overallOperational {
		overallStatus = "DEGRADED"
	}

	return &SystemStatus{
		OverallStatus: overallStatus,
		Components:    components,
		LastUpdated:   lastUpdated,
		UptimeStats:   overallUptimeStats,
	}, nil
}

// calculateUptime calculates uptime percentage for different time periods
func (s *StatusService) calculateUptime(items []models.DynamoDBItem) UptimeStats {
	now := time.Now()
	
	// Define time periods
	periods := map[string]time.Duration{
		"24h": 24 * time.Hour,
		"7d":  7 * 24 * time.Hour,
		"30d": 30 * 24 * time.Hour,
	}

	stats := UptimeStats{}

	for period, duration := range periods {
		cutoff := now.Add(-duration)
		total := 0
		operational := 0

		for _, item := range items {
			timestamp, err := time.Parse("2006-01-02T15:04:05Z07:00", item.LastChecked)
			if err != nil || timestamp.Before(cutoff) {
				continue
			}

			total++
			if item.Status == "OPERATIONAL" {
				operational++
			}
		}

		var uptime float64
		if total > 0 {
			uptime = float64(operational) / float64(total) * 100
		} else {
			uptime = 100.0 // Assume operational if no data
		}

		switch period {
		case "24h":
			stats.Last24Hours = uptime
		case "7d":
			stats.Last7Days = uptime
		case "30d":
			stats.Last30Days = uptime
		}
	}

	return stats
}

// generateStatusHistory generates a visual status history for the last N days
func (s *StatusService) generateStatusHistory(items []models.DynamoDBItem, days int) []StatusPoint {
	now := time.Now()
	cutoff := now.AddDate(0, 0, -days)

	var points []StatusPoint

	// Group by day and determine status for each day
	dayMap := make(map[string][]models.DynamoDBItem)

	for _, item := range items {
		timestamp, err := time.Parse("2006-01-02T15:04:05Z07:00", item.LastChecked)
		if err != nil || timestamp.Before(cutoff) {
			continue
		}

		dayKey := timestamp.Format("2006-01-02")
		dayMap[dayKey] = append(dayMap[dayKey], item)
	}

	// Generate points for each day
	for d := 0; d < days; d++ {
		day := now.AddDate(0, 0, -d)
		dayKey := day.Format("2006-01-02")

		status := "OPERATIONAL" // Default status
		if dayItems, exists := dayMap[dayKey]; exists && len(dayItems) > 0 {
			// Determine worst status for the day
			for _, item := range dayItems {
				if item.Status != "OPERATIONAL" {
					status = item.Status
					break
				}
			}
		}

		points = append([]StatusPoint{{
			Timestamp: day,
			Status:    status,
		}}, points...)
	}

	return points
}

// calculateOverallUptime calculates overall system uptime
func (s *StatusService) calculateOverallUptime(components []ComponentStatus) UptimeStats {
	if len(components) == 0 {
		return UptimeStats{
			Last24Hours: 100.0,
			Last7Days:   100.0,
			Last30Days:  100.0,
		}
	}

	var total24h, total7d, total30d float64

	for _, component := range components {
		total24h += component.UptimePercent
		// For now, using the same uptime for all periods
		// In a real implementation, you'd calculate these separately
		total7d += component.UptimePercent
		total30d += component.UptimePercent
	}

	count := float64(len(components))

	return UptimeStats{
		Last24Hours: total24h / count,
		Last7Days:   total7d / count,
		Last30Days:  total30d / count,
	}
}
