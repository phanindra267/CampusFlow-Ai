package main

import (
	"errors"
	"fmt"
)

// Sentinel errors for start-up and shutdown. Each one names the stage that
// failed so the single line printed on stderr tells an operator what to fix
// without needing the full log stream.
var (
	errNilConfig = errors.New(
		"configuration returned no value and no error; refusing to start")
)

func errDatabaseUnavailable(cause error) error {
	return fmt.Errorf(
		"could not connect to PostgreSQL: %w; refusing to start without the source of truth "+
			"(set APP_ENV=development to allow a degraded start for local work)", cause)
}

func errListenFailed(cause error) error {
	return fmt.Errorf("HTTP server failed: %w", cause)
}

func errDrainFailed(cause error) error {
	return fmt.Errorf("graceful shutdown did not complete within the drain window: %w", cause)
}
