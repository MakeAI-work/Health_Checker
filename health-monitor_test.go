package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

// TestConfig tests the configuration loading functionality
func TestConfig(t *testing.T) {
	// Test default values
	config, err := LoadConfig("")
	if err != nil {
		t.Fatalf("Failed to load default config: %v", err)
	}

	if config.CheckInterval != 30 {
		t.Errorf("Expected default CheckInterval to be 30, got %d", config.CheckInterval)
	}
	if config.TimeoutSeconds != 5 {
		t.Errorf("Expected default TimeoutSeconds to be 5, got %d", config.TimeoutSeconds)
	}
	if config.SlowThreshold != 1000 {
		t.Errorf("Expected default SlowThreshold to be 1000, got %d", config.SlowThreshold)
	}

	// Test environment variables
	os.Setenv("HEALTH_CHECK_URLS", "https://test.com")
	os.Setenv("CHECK_INTERVAL", "60")
	os.Setenv("TIMEOUT_SECONDS", "10")
	os.Setenv("SLOW_THRESHOLD", "2000")

	config, err = LoadConfig("")
	if err != nil {
		t.Fatalf("Failed to load config from env: %v", err)
	}

	if len(config.URLs) != 1 || config.URLs[0] != "https://test.com" {
		t.Errorf("Expected URL to be https://test.com, got %v", config.URLs)
	}
	if config.CheckInterval != 60 {
		t.Errorf("Expected CheckInterval to be 60, got %d", config.CheckInterval)
	}
	if config.TimeoutSeconds != 10 {
		t.Errorf("Expected TimeoutSeconds to be 10, got %d", config.TimeoutSeconds)
	}
	if config.SlowThreshold != 2000 {
		t.Errorf("Expected SlowThreshold to be 2000, got %d", config.SlowThreshold)
	}
}

// TestChecker tests the health checker functionality
func TestChecker(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Test successful check
	checker := NewChecker(server.URL, 1, 100)
	result := checker.Ping(context.Background())

	if result.Error != nil {
		t.Errorf("Expected no error, got %v", result.Error)
	}
	if result.Status != http.StatusOK {
		t.Errorf("Expected status 200, got %d", result.Status)
	}
	if result.ResponseTime < 50*time.Millisecond {
		t.Errorf("Expected response time > 50ms, got %v", result.ResponseTime)
	}

	// Test timeout
	checker.Timeout = 1 * time.Millisecond
	result = checker.Ping(context.Background())
	if result.Error == nil {
		t.Error("Expected timeout error, got nil")
	}
}

// TestReporter tests the reporter functionality
func TestReporter(t *testing.T) {
	reporter := NewConsoleReporter()

	// Test successful result
	result := Result{
		URL:         "https://test.com",
		Status:      200,
		ResponseTime: 500 * time.Millisecond,
		Error:       nil,
		Timestamp:   time.Now(),
	}

	err := reporter.Report(result)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	stats := reporter.GetStats()
	if stats.SuccessCount != 1 {
		t.Errorf("Expected 1 success, got %d", stats.SuccessCount)
	}
	if stats.FailureCount != 0 {
		t.Errorf("Expected 0 failures, got %d", stats.FailureCount)
	}

	// Test failed result
	result.Error = http.ErrServerClosed
	err = reporter.Report(result)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	stats = reporter.GetStats()
	if stats.SuccessCount != 1 {
		t.Errorf("Expected 1 success, got %d", stats.SuccessCount)
	}
	if stats.FailureCount != 1 {
		t.Errorf("Expected 1 failure, got %d", stats.FailureCount)
	}
}

// TestAggregator tests the aggregator functionality
func TestAggregator(t *testing.T) {
	// Create test servers
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server2.Close()

	// Create checkers
	checkers := []*Checker{
		NewChecker(server1.URL, 1, 100),
		NewChecker(server2.URL, 1, 100),
	}

	// Create reporter
	reporter := NewConsoleReporter()

	// Create aggregator
	aggregator := NewAggregator(checkers, reporter, 2)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Start aggregator
	aggregator.Start(ctx)

	// Wait for some results
	time.Sleep(1 * time.Second)

	// Stop aggregator
	aggregator.Stop()

	// Check stats
	stats := reporter.GetStats()
	if stats.TotalChecks == 0 {
		t.Error("Expected some checks to be performed")
	}
}

// TestMain tests the main program flow
func TestMain(t *testing.T) {
	config, err := LoadConfig("config.json")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify URLs
	if len(config.URLs) == 0 {
		t.Error("Expected at least one URL in config")
	}

	// Verify other settings
	if config.CheckInterval < 1 {
		t.Error("Expected CheckInterval to be at least 1 second")
	}
	if config.TimeoutSeconds < 1 {
		t.Error("Expected TimeoutSeconds to be at least 1 second")
	}
	if config.SlowThreshold < 1 {
		t.Error("Expected SlowThreshold to be at least 1 millisecond")
	}
} 