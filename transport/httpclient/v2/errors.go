package httpclient

import "errors"

// Sentinel errors returned by Client. Use errors.Is to distinguish
// failure causes without parsing error strings.
//
// Example:
//
//	resp, err := client.Get(ctx, url)
//	if errors.Is(err, httpclient.ErrDoRequest) {
//		// network failure, timeout, connection refused...
//	}
//	defer resp.Body.Close()
var (
	// ErrInvalidBaseURL is wrapped when baseURL set via WithBaseURL is not a valid URL.
	ErrInvalidBaseURL = errors.New("httpclient: invalid base URL")

	// ErrInvalidURL is wrapped when the request URL is not a valid URL.
	ErrInvalidURL = errors.New("httpclient: invalid URL")

	// ErrCreateRequest is wrapped when http.NewRequestWithContext fails.
	ErrCreateRequest = errors.New("httpclient: failed to create request")

	// ErrDoRequest is wrapped when the underlying http.Client.Do fails
	// (DNS, connection refused, timeout, context canceled...).
	// The underlying error (often *url.Error) stays in the chain for errors.As.
	ErrDoRequest = errors.New("httpclient: failed to execute request")
)
