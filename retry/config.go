package retry

import "github.com/ductran999/shared-pkg/retry/backoff"

// Config defines the configuration settings for retry behavior.
type Config struct {
	// MaxAttempts specifies the total number of attempts for the operation,
	// including the initial attempt and all retries.
	MaxAttempts int

	// Backoff specifies the strategy used to wait between retry attempts,
	// such as constant, linear, or exponential delays.
	Backoff backoff.BackoffStrategy

	// Logging specifies the retry attempt number at which logs should be generated.
	// For example, if Logging is 2, the 2nd, 3rd, etc., attempts will be logged.
	Logging int
}
