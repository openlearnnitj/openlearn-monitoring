package config

import (
	"fmt"
	"os"
)

// Config holds all configuration values for the monitoring service
type Config struct {
	MonitoringAPIURL    string
	MonitoringAPISecret string
	DynamoDBTableName   string
	AWSRegion           string
	Port                string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	cfg := &Config{}

	var err error
	if cfg.MonitoringAPIURL, err = getEnvVar("MONITORING_API_URL"); err != nil {
		return nil, err
	}

	if cfg.MonitoringAPISecret, err = getEnvVar("MONITORING_API_SECRET"); err != nil {
		return nil, err
	}

	if cfg.DynamoDBTableName, err = getEnvVar("DYNAMODB_TABLE_NAME"); err != nil {
		return nil, err
	}

	if cfg.AWSRegion, err = getEnvVar("AWS_REGION"); err != nil {
		return nil, err
	}

	// Port is optional, default will be used if not set
	cfg.Port = os.Getenv("PORT")

	return cfg, nil
}

// getEnvVar retrieves an environment variable or returns an error if not set
func getEnvVar(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("required environment variable %s is not set", key)
	}
	return value, nil
}
