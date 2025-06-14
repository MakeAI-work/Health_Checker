package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Load configuration
	config, err := LoadConfig("config.json")
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Create checkers for each URL
	var checkers []*Checker
	for _, url := range config.URLs {
		checker := NewChecker(url, config.TimeoutSeconds, config.SlowThreshold)
		checkers = append(checkers, checker)
	}

	// Create reporter
	reporter := NewConsoleReporter()

	// Create aggregator
	aggregator := NewAggregator(checkers, reporter, 5) // 5 workers

	// Create context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the aggregator
	aggregator.Start(ctx)

	// Print startup message
	fmt.Printf("Starting health monitor for %d URLs...\n", len(config.URLs))
	fmt.Println("Press Ctrl+C to stop")

	// Wait for shutdown signal
	<-sigChan
	fmt.Println("\nShutting down gracefully...")

	// Stop the aggregator
	aggregator.Stop()

	// Print final stats
	stats := reporter.GetStats()
	fmt.Printf("\nFinal Statistics:\n")
	fmt.Printf("Total Checks: %d\n", stats.TotalChecks)
	fmt.Printf("Successful: %d\n", stats.SuccessCount)
	fmt.Printf("Failed: %d\n", stats.FailureCount)
	fmt.Printf("Slow Responses: %d\n", stats.SlowCount)
} 