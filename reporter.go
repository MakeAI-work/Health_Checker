package main

import (
	"fmt"
	"sync/atomic"
	"time"
)

// HealthReporter defines the interface for reporting health check results
type HealthReporter interface {
	Report(Result) error
	GetStats() Stats
}

// Stats holds the aggregated statistics
type Stats struct {
	TotalChecks    uint64
	SuccessCount   uint64
	FailureCount   uint64
	SlowCount      uint64
	AverageLatency time.Duration
}

// ConsoleReporter implements HealthReporter for console output
type ConsoleReporter struct {
	stats Stats
	lastResults map[string]Result
}

// NewConsoleReporter creates a new ConsoleReporter
func NewConsoleReporter() *ConsoleReporter {
	return &ConsoleReporter{
		lastResults: make(map[string]Result),
	}
}

// Report implements the HealthReporter interface
func (r *ConsoleReporter) Report(result Result) error {
	atomic.AddUint64(&r.stats.TotalChecks, 1)

	// Update success/failure counts
	if result.Error == nil {
		atomic.AddUint64(&r.stats.SuccessCount, 1)
	} else {
		atomic.AddUint64(&r.stats.FailureCount, 1)
		fmt.Printf("❌ %s is DOWN: %v\n", result.URL, result.Error)
	}

	// Check for slow responses
	if result.ResponseTime > time.Second {
		atomic.AddUint64(&r.stats.SlowCount, 1)
		fmt.Printf("⚠️ %s is SLOW: %v\n", result.URL, result.ResponseTime)
	}

	// Check for status changes
	if lastResult, exists := r.lastResults[result.URL]; exists {
		if (lastResult.Error == nil && result.Error != nil) ||
			(lastResult.Error != nil && result.Error == nil) {
			fmt.Printf("🔄 %s status changed: %v -> %v\n",
				result.URL,
				lastResult.Error,
				result.Error)
		}
	}

	// Update last result
	r.lastResults[result.URL] = result

	return nil
}

// GetStats returns the current statistics
func (r *ConsoleReporter) GetStats() Stats {
	return r.stats
} 