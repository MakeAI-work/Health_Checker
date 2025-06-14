# Health Monitor

A simple Go application to periodically check the health and performance of a list of URLs.

## Features

- Configurable list of URLs to monitor
- Concurrent checks with worker pool
- Timeout and slow-response detection
- Console-based reporting with statistics

## Configuration

The application reads settings from `config.json`:

```json
{
  "urls": [
    "https://example.com",
    "https://another-service/status"
  ],
  "check_interval": 30,
  "timeout_seconds": 10,
  "slow_threshold": 1000
}
```

- `urls`: Array of URLs to check.
- `check_interval`: Interval (in seconds) between each round of checks.
- `timeout_seconds`: HTTP request timeout (in seconds).
- `slow_threshold`: Threshold (in milliseconds) to count a response as slow.

## Installation & Usage

1. Clone the repo:
   ```sh
   git clone <repo-url>
   cd health-monitor
   ```

2. Build:
   ```sh
   go build -o health-monitor
   ```

3. Run:
   ```sh
   ./health-monitor
   ```
   The program will load `config.json`, start checking URLs, and print results to the console.

4. Stop:
   Press `Ctrl+C` to gracefully shut down and display final statistics.

## Testing

Run unit tests for checkers and real URLs:
```sh
go test ./...
```

## Project Structure

- `aggregator.go`: Manages worker pool and dispatches URL checks.
- `checker.go`: Performs HTTP requests and measures response time.
- `reporter.go`: Logs results and aggregates statistics.
- `config.go`: Loads and validates configuration.
- `main.go`: Entry point, sets up components and handles shutdown.

---

Feel free to open an issue or submit a pull request for improvements.
