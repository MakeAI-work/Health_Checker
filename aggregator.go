package main

import (
	"context"
	"sync"
)

// Aggregator manages health checks and result aggregation
type Aggregator struct {
	checkers  []*Checker
	reporter  HealthReporter
	workers   int
	results   chan Result
	done      chan struct{}
}

// NewAggregator creates a new Aggregator
func NewAggregator(checkers []*Checker, reporter HealthReporter, workers int) *Aggregator {
	return &Aggregator{
		checkers: checkers,
		reporter: reporter,
		workers:  workers,
		results:  make(chan Result),
		done:     make(chan struct{}),
	}
}

// Start begins the health check process
func (a *Aggregator) Start(ctx context.Context) {
	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < a.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case <-a.done:
					return
				default:
					// Process all checkers
					for _, checker := range a.checkers {
						select {
						case <-ctx.Done():
							return
						case <-a.done:
							return
						default:
							result := checker.Ping(ctx)
							a.results <- result
						}
					}
				}
			}
		}()
	}

	// Start result processor
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-a.done:
				return
			case result := <-a.results:
				a.reporter.Report(result)
			}
		}
	}()

	// Wait for all workers to finish
	go func() {
		wg.Wait()
		close(a.results)
	}()
}

// Stop gracefully stops the aggregator
func (a *Aggregator) Stop() {
	close(a.done)
} 