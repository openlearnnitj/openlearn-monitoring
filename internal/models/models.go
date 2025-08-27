package models

import "time"

// HealthStatusResponse represents the JSON response from the monitoring endpoint
type HealthStatusResponse struct {
	Timestamp     string      `json:"timestamp"`
	OverallStatus string      `json:"overallStatus"`
	Components    []Component `json:"components"`
}

// Component represents a single component in the health status response
type Component struct {
	Name           string  `json:"name"`
	Status         string  `json:"status"`
	ResponseTimeMs float64 `json:"responseTimeMs"`
}

// MonitoringResult represents the result of a monitoring check
type MonitoringResult struct {
	Components           []Component
	TotalResponseTimeMs  int64
	Timestamp           time.Time
}

// DynamoDBItem represents an item to be stored in DynamoDB
type DynamoDBItem struct {
	ServiceName             string    `dynamodbav:"serviceName"`
	Status                  string    `dynamodbav:"status"`
	InternalResponseTimeMs  float64   `dynamodbav:"internalResponseTimeMs"`
	TotalResponseTimeMs     int64     `dynamodbav:"totalResponseTimeMs"`
	LastChecked             string    `dynamodbav:"lastChecked"`
}
