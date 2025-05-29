// Package ticker_server provides a flexible and robust ticker-based task scheduling system.
// It implements the transport.Server interface from go-kratos framework and offers
// a convenient way to manage and execute periodic tasks with configurable intervals.
//
// The package provides two main components:
//   - Ticker: A single ticker that executes a task at specified intervals
//   - Tickers: A manager for multiple tickers with automatic ID management
//
// Example usage of a single Ticker:
//
//	task := &TickTask{
//		Name:      "my-task",
//		Fn:        func(ctx context.Context, isStop bool) error { return nil },
//		Timeout:   time.Second * 5,
//		Interval:  time.Minute,
//		Immediate: true,
//	}
//	ticker := NewTicker(time.Minute, task, WithTickerImmediate(true))
//	ticker.Start(context.Background())
//
// Example usage of Tickers manager:
//
//	tickers := NewTickers(
//		WithTickersLogger(logger),
//		WithTickersTasks(task1, task2),
//	)
//	tickers.Start(context.Background())
//
// The Ticker provides:
//   - Configurable task execution interval
//   - Immediate execution option
//   - Task timeout control
//   - Graceful shutdown support
//
// The Tickers manager provides:
//   - Automatic ID management for multiple tickers
//   - Add new tickers with unique IDs
//   - Remove tickers by ID
//   - Bulk start/stop of all tickers
//   - ID recycling for efficient memory usage
//
// Each TickTask requires:
//   - Name: A unique identifier for the task
//   - Fn: The function to execute
//   - Timeout: Maximum execution time (defaults to 10s)
//   - Interval: Time between executions
//   - Immediate: Whether to run immediately on start
package ticker_server
