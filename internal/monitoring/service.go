package monitoring

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/openlearnnitj/openlearn-monitoring/internal/models"
)

// Service handles monitoring operations
type Service struct {
	apiURL    string
	apiSecret string
	client    *http.Client
}

// NewService creates a new monitoring service
func NewService(apiURL, apiSecret string) *Service {
	return &Service{
		apiURL:    apiURL,
		apiSecret: apiSecret,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CheckHealth performs a health check against the monitoring endpoint
func (s *Service) CheckHealth() (*models.MonitoringResult, error) {
	// Create request
	req, err := http.NewRequest("GET", s.apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add required headers
	req.Header.Set("X-API-Secret", s.apiSecret)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "OpenLearn-Monitoring/1.0")

	// Measure response time
	start := time.Now()
	resp, err := s.client.Do(req)
	totalResponseTime := time.Since(start).Milliseconds()

	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status: %d", resp.StatusCode)
	}

	// Parse response
	var healthResp models.HealthStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&healthResp); err != nil {
		return nil, fmt.Errorf("failed to decode JSON response: %w", err)
	}

	return &models.MonitoringResult{
		Components:          healthResp.Components,
		TotalResponseTimeMs: totalResponseTime,
		Timestamp:          time.Now().UTC(),
	}, nil
}
