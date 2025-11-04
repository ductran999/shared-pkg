package retry_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/ductran999/shared-pkg/retry"
	"github.com/ductran999/shared-pkg/retry/backoff"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var ErrMock = errors.New("mock error")

var tryablefunc = func(err error) bool {
	return errors.Is(err, ErrMock)
}

func Test_Do(t *testing.T) {
	retryer := retry.NewRetry(retry.Config{
		MaxAttempts: 5,
		Backoff:     backoff.NewLinearBackoff(),
		Logging:     3,
	})

	t.Run("immediate success", func(t *testing.T) {
		t.Parallel()
		actualErr := retryer.Do(func() error {
			return nil
		}, tryablefunc)
		require.NoError(t, actualErr)
	})

	t.Run("error but no retry", func(t *testing.T) {
		t.Parallel()
		actualErr := retryer.Do(func() error {
			return errors.New("unexpected error")
		}, tryablefunc)
		require.EqualError(t, actualErr, "unexpected error")
	})

	t.Run("success after retry", func(t *testing.T) {
		t.Parallel()
		attempts := 0

		actualErr := retryer.Do(func() error {
			attempts++
			if attempts == 3 {
				return nil
			}

			return fmt.Errorf("wrap %w", ErrMock)
		}, tryablefunc)

		require.NoError(t, actualErr)
		assert.Equal(t, 3, attempts, "should have retried 2 times before success")
	})

	t.Run("failing all attempts", func(t *testing.T) {
		t.Parallel()
		actualErr := retryer.Do(func() error {
			return fmt.Errorf("wrap %w", ErrMock)
		}, tryablefunc)

		assert.EqualError(t, actualErr, "wrap mock error")
	})
}

func Test_DoWithConfig(t *testing.T) {
	retryer := retry.NewRetry(retry.Config{
		MaxAttempts: 5,
		Backoff:     backoff.NewLinearBackoff(),
		Logging:     2,
	})

	t.Run("immediate success", func(t *testing.T) {
		t.Parallel()
		actualErr := retryer.DoWithConfig(func() error {
			return nil
		}, tryablefunc, retry.Config{
			MaxAttempts: 3,
			Backoff:     backoff.NewExponentialBackoff(),
			Logging:     1,
		})
		require.NoError(t, actualErr)
	})

	t.Run("success after retry", func(t *testing.T) {
		t.Parallel()
		attempts := 0

		actualErr := retryer.DoWithConfig(func() error {
			attempts++
			if attempts == 3 {
				return nil
			}
			return fmt.Errorf("wrap %w", ErrMock)
		}, tryablefunc, retry.Config{
			MaxAttempts: 4,
			Backoff:     backoff.NewLinearBackoff(),
			Logging:     2,
		})

		require.NoError(t, actualErr)
		assert.Equal(t, 3, attempts, "should have retried 2 times before success")
	})

	t.Run("failing all attempts", func(t *testing.T) {
		t.Parallel()
		actualErr := retryer.DoWithConfig(func() error {
			return fmt.Errorf("wrap %w", ErrMock)
		}, tryablefunc, retry.Config{
			MaxAttempts: 4,
			Backoff:     backoff.NewConstantBackoff(),
			Logging:     1,
		})

		assert.EqualError(t, actualErr, "wrap mock error")
	})
}
