package retry

import (
	"time"

	"github.com/ductran999/shared-pkg/retry/backoff"
	"github.com/rs/zerolog/log"
)

// RetryableFunc defines a function that determines whether an error is retryable.
type RetryableFunc func(err error) bool

// BusinessFunc defines the core business logic that may need to be retried.
type BusinessFunc func() error

// Retry defines the interface for executing retryable operations.
type Retry interface {
	// Do executes the given business function with the default retry configuration.
	// It uses the retryable function to determine whether to retry on error.
	Do(fn BusinessFunc, retryable RetryableFunc) error

	// DoWithConfig executes the given business function using a custom retry configuration.
	DoWithConfig(fn BusinessFunc, retryable RetryableFunc, config Config) error
}

// retry implements the Retry interface and holds the retry configuration.
type retry struct {
	config Config
}

// NewRetry creates a new retry instance with the provided configuration.
//
// Default values are used for any field not explicitly set:
//
//   - MaxAttempts: 3
//   - Backoff: ExponentialBackoff
//   - Logging: 0
//
// This allows the user to only override what they care about,
// while sensible defaults ensure the retry mechanism is functional.
func NewRetry(config Config) *retry {
	r := &retry{
		config: Config{
			MaxAttempts: 3,
			Backoff:     backoff.NewExponentialBackoff(),
			Logging:     0,
		},
	}

	// Override default config with custom values if provided
	if config.MaxAttempts > 0 {
		r.config.MaxAttempts = config.MaxAttempts
	}

	if config.Backoff != nil {
		r.config.Backoff = config.Backoff
	}

	r.config.Logging = config.Logging

	return r
}

// DefaultRetry returns a retry instance with predefined default configuration.
// This is useful when you want to quickly use retry logic without custom setup.
//
// Default Config:
//   - MaxAttempts: 3 (retries the operation up to 3 times)
//   - Backoff: Exponential backoff strategy (wait time increases exponentially between attempts)
//   - Logging: 0 (logs each retry attempt and final result)
func DefaultRetry() *retry {
	return &retry{
		config: Config{
			MaxAttempts: 3,
			Backoff:     backoff.NewExponentialBackoff(),
			Logging:     1,
		},
	}
}

// Do executes the given retryable function using the initialize retry configuration.
func (r *retry) Do(fn BusinessFunc, retryable RetryableFunc) error {
	var err error

	for attempt := 1; attempt <= r.config.MaxAttempts; attempt++ {
		err = fn()
		if err == nil || !retryable(err) {
			return err
		}

		if attempt >= r.config.Logging {
			log.Warn().Int("attempt", attempt).Err(err).Msg("retry attempt failed")
		}

		if attempt < r.config.MaxAttempts {
			time.Sleep(r.config.Backoff.Next(attempt))
		}
	}

	log.Error().Int("all attempts", r.config.MaxAttempts).Err(err).Msg("all retry attempts failed")

	return err
}

// DoWithConfig executes the given retryable function using a custom retry configuration.
func (r retry) DoWithConfig(fn BusinessFunc, retryable RetryableFunc, custom Config) error {
	var err error

	for attempt := 1; attempt <= custom.MaxAttempts; attempt++ {
		err = fn()
		if err == nil || !retryable(err) {
			return err
		}

		if attempt >= custom.Logging {
			log.Warn().Int("attempt", attempt).Err(err).Msg("retry attempt failed")
		}

		if attempt < custom.MaxAttempts {
			time.Sleep(custom.Backoff.Next(attempt))
		}
	}

	log.Error().Int("all attempts", custom.MaxAttempts).Err(err).Msg("all retry attempts failed")

	return err
}
