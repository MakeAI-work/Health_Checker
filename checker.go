package main

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Result represents the outcome of a health check
type Result struct {
	URL          string
	Status       int
	ResponseTime time.Duration
	Error        error
	Timestamp    time.Time
	Details      map[string]interface{}
	IsSlow       bool
}

// Checker performs health checks on a URL
type Checker struct {
	URL           string
	Timeout       time.Duration
	SlowThreshold time.Duration
	Client        *http.Client
}

// NewChecker creates a new Checker instance
func NewChecker(url string, timeoutSeconds, slowThresholdMs int) *Checker {
	return &Checker{
		URL:           url,
		Timeout:       time.Duration(timeoutSeconds) * time.Second,
		SlowThreshold: time.Duration(slowThresholdMs) * time.Millisecond,
		Client: &http.Client{
			Timeout: time.Duration(timeoutSeconds) * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				IdleConnTimeout:     90 * time.Second,
				DisableCompression:  true,
			},
		},
	}
}

// Ping performs a health check on the URL
func (c *Checker) Ping(ctx context.Context) Result {
	start := time.Now()
	result := Result{
		URL:       c.URL,
		Timestamp: start,
		Details:   make(map[string]interface{}),
	}

	req, err := http.NewRequestWithContext(ctx, "GET", c.URL, nil)
	if err != nil {
		result.Error = fmt.Errorf("failed to create request: %v", err)
		return result
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		result.Error = fmt.Errorf("request failed: %v", err)
		return result
	}
	defer resp.Body.Close()

	result.Status = resp.StatusCode
	result.ResponseTime = time.Since(start)
	result.Details["content_length"] = resp.ContentLength
	result.Details["content_type"] = resp.Header.Get("Content-Type")
	result.IsSlow = result.ResponseTime > c.SlowThreshold

	if resp.StatusCode >= 400 {
		result.Error = fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	return result
} 