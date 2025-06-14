package main

import (
	"context"
	"testing"
	"time"
)

// TestRealURLs tests the health monitor with real, public APIs
func TestRealURLs(t *testing.T) {
	// Load configuration
	config, err := LoadConfig("config.json")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Create checkers for each URL
	var checkers []*Checker
	for _, url := range config.URLs {
		checker := NewChecker(url, config.TimeoutSeconds, config.SlowThreshold)
		checkers = append(checkers, checker)
	}

	// Create reporter and aggregator
	reporter := NewConsoleReporter()
	aggregator := NewAggregator(checkers, reporter, 3)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Start the aggregator
	aggregator.Start(ctx)
	time.Sleep(5 * time.Second)
	aggregator.Stop()

	// Get and verify results
	stats := reporter.GetStats()
	if stats.TotalChecks == 0 {
		t.Error("Expected some health checks to be performed")
	}

	t.Logf("\nTest Results:")
	t.Logf("Total Checks: %d", stats.TotalChecks)
	t.Logf("Successful: %d", stats.SuccessCount)
	t.Logf("Failed: %d", stats.FailureCount)
	t.Logf("Slow Responses: %d", stats.SlowCount)
}

// TestSpecificURLs tests specific URLs with different behaviors
func TestSpecificURLs(t *testing.T) {
	testCases := []struct {
		name           string
		url            string
		expectSuccess  bool
		expectSlow     bool
		timeoutSeconds int
	}{
		{
			name:           "Fast Success",
			url:            "https://httpbin.org/status/200",
			expectSuccess:  true,
			expectSlow:     false,
			timeoutSeconds: 5,
		},
		{
			name:           "Expected Failure",
			url:            "https://httpbin.org/status/404",
			expectSuccess:  false,
			expectSlow:     false,
			timeoutSeconds: 5,
		},
		{
			name:           "Slow Response",
			url:            "https://httpbin.org/delay/1",
			expectSuccess:  true,
			expectSlow:     true,
			timeoutSeconds: 5,
		},
		{
			name:           "Timeout",
			url:            "https://httpbin.org/delay/3",
			expectSuccess:  false,
			expectSlow:     false,
			timeoutSeconds: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			checker := NewChecker(tc.url, tc.timeoutSeconds, 1000)
			reporter := NewConsoleReporter()
			aggregator := NewAggregator([]*Checker{checker}, reporter, 1)

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			aggregator.Start(ctx)
			time.Sleep(3 * time.Second)
			aggregator.Stop()

			stats := reporter.GetStats()

			if tc.expectSuccess && stats.SuccessCount == 0 {
				t.Errorf("Expected successful check for %s", tc.url)
			}

			if !tc.expectSuccess && stats.FailureCount == 0 {
				t.Errorf("Expected failed check for %s", tc.url)
			}

			if tc.expectSlow && stats.SlowCount == 0 {
				t.Errorf("Expected slow response for %s", tc.url)
			}

			t.Logf("URL: %s", tc.url)
			t.Logf("Success: %d, Failures: %d, Slow: %d",
				stats.SuccessCount,
				stats.FailureCount,
				stats.SlowCount)
		})
	}
} 