# retry

`retry` is a lightweight and extensible Go package that provides a retry mechanism with support for configurable backoff strategies.

## Features

- Retry any function that returns an error
- Customizable retry configuration:
  - Max attempts
  - Logging
  - Backoff strategies (constant, linear, exponential)
- Jitter support to prevent thundering herd

## Installation

```bash

go get github.com/ductran999/shared-pkg

```

## Usage

### Basic Usage
```go
r := retry.DefaultRetry()

err := r.Do(func() error {
    // your operation here
    return nil
})
```

### Custom Configuration
```go
r := retry.NewRetry(retry.Config{
    MaxAttempts: 5,
    Backoff:     backoff.NewLinearBackoff(),
    Logging:     3,
})
```

### Override Configuration Per Call
```go
err := r.DoWithConfig(func() error {
    return someOperation()
},func(err error) bool {
    ErrMustRetry := errors.New("error must retry")
    if errors.Is(err, ErrMustRetry) {
        return true
    }
    return false
}, retry.NewRetry(retry.Config{
    MaxAttempts: 4,
    Backoff:     backoff.NewLinearBackoff(WithBase(1 * time.Second)),
    Logging:     3,
})
```

## Backoff Strategies
- ConstantBackoff: fixed interval between retries

- LinearBackoff: grows linearly with each attempt

- ExponentialBackoff: grows exponentially, with optional jitter

Each strategy supports options like WithJitter, WithCap, WithStep, etc.

### Example
```go
r := retry.NewRetry(retry.Config{
    MaxAttempts: 4,
    Backoff: backoff.NewExponentialBackoff(
        backoff.WithBase(500 * time.Millisecond),
        backoff.WithCap(5 * time.Second),
        backoff.WithJitter(true),
    ),
    Logging: 3,
})
```