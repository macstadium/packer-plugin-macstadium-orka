package orka

import (
	"context"
	"errors"
	"time"
)

// WatcherError is a custom error type that wraps another error. It is used to indicate
// errors that should trigger a retry mechanism in the RetryOnWatcherErrorWithTimeout function.
// This type implements the error interface and provides error unwrapping capabilities.
type WatcherError struct {
	Err error
}

func (e WatcherError) Error() string {
	return e.Err.Error()
}

func (e WatcherError) Unwrap() error {
	return e.Err
}

// RetryOnWatcherErrorWithTimeout executes the given function repeatedly until it succeeds or encounters a non-WatcherError,
// with an overall timeout. It will retry only on WatcherError types, returning immediately for other errors.
//
// Parameters:
//   - ctx: The parent context for timeout and cancellation
//   - timeout: Maximum total duration to keep retrying
//   - fn: The function to execute and potentially retry (should respect context cancellation)
//   - delay: Duration to wait between retries
//
// Returns:
//   - nil if the function succeeds
//   - context.Err() if the timeout is reached
//   - the error from fn if it's not a WatcherError
//
// Note: The provided function fn must properly check and respect the context,
// otherwise the timeout may not be honored correctly.
func RetryOnWatcherErrorWithTimeout(ctx context.Context, timeout time.Duration, fn func(ctx context.Context) error, delay time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		if err := fn(ctx); err == nil {
			return nil
		} else {
			var watcherErr WatcherError
			if !errors.As(err, &watcherErr) {
				return err
			}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}
}
