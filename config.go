package main

import (
	"encoding/json"
	"os"
	"strconv"
)

// Config holds the application configuration
type Config struct {
	URLs           []string `json:"urls"`
	CheckInterval  int      `json:"check_interval"` // in seconds
	TimeoutSeconds int      `json:"timeout_seconds"`
	SlowThreshold  int      `json:"slow_threshold"` // in milliseconds
}

// LoadConfig loads configuration from a JSON file or environment variables
func LoadConfig(configPath string) (*Config, error) {
	config := &Config{
		CheckInterval:  30,  // default: check every 30 seconds
		TimeoutSeconds: 5,   // default: 5 second timeout
		SlowThreshold:  1000, // default: 1 second is considered slow
	}

	// Try to load from file first
	if configPath != "" {
		file, err := os.ReadFile(configPath)
		if err == nil {
			if err := json.Unmarshal(file, config); err != nil {
				return nil, err
			}
		}
	}

	// Override with environment variables if they exist
	if urls := os.Getenv("HEALTH_CHECK_URLS"); urls != "" {
		config.URLs = []string{urls}
	}
	if interval := os.Getenv("CHECK_INTERVAL"); interval != "" {
		if val, err := strconv.Atoi(interval); err == nil {
			config.CheckInterval = val
		}
	}
	if timeout := os.Getenv("TIMEOUT_SECONDS"); timeout != "" {
		if val, err := strconv.Atoi(timeout); err == nil {
			config.TimeoutSeconds = val
		}
	}
	if threshold := os.Getenv("SLOW_THRESHOLD"); threshold != "" {
		if val, err := strconv.Atoi(threshold); err == nil {
			config.SlowThreshold = val
		}
	}

	return config, nil
} 