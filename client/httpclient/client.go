// Package httpclient is the legacy v1 HTTP client.
//
// Deprecated: use github.com/ductran999/shared-pkg/transport/httpclient/v2
// instead. It returns the raw *http.Response, has functional options and
// propagates the request ID across services. This package is kept frozen for
// backward compatibility and will be removed in a future major version.
package httpclient

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	DefaultClientTimeout time.Duration = 5 * time.Second

	DefaultMaxIdleConns        int = 100
	DefaultMaxIdleConnsPerHost int = 10

	DefaultDialTimeout   time.Duration = 5 * time.Second
	DefaultDialKeepAlive time.Duration = 5 * time.Second

	DefaultIdleConnTimeout     time.Duration = 90 * time.Second
	DefaultTLSHandshakeTimeout time.Duration = 5 * time.Second
)

type httpClient struct {
	client *http.Client
}

type response struct {
	StatusCode  int
	Status      string
	Body        []byte
	RawResponse *http.Response
}

// NewClient creates and returns a new instance of httpClient with pre-configured
// timeout and transport settings suitable for most HTTP client use cases.
// The returned httpClient can be used to perform HTTP requests with sensible defaults.
func NewClient(options ...Option) *httpClient {
	// Default transport settings for the HTTP Client.
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   DefaultDialTimeout,
			KeepAlive: DefaultDialKeepAlive,
		}).DialContext,

		MaxIdleConns:        DefaultMaxIdleConns,
		MaxIdleConnsPerHost: DefaultMaxIdleConnsPerHost,
		IdleConnTimeout:     DefaultIdleConnTimeout,
		TLSHandshakeTimeout: DefaultTLSHandshakeTimeout,
	}

	c := &httpClient{
		client: &http.Client{
			Timeout:   DefaultClientTimeout,
			Transport: transport,
		},
	}

	// Apply options to the client
	for _, opt := range options {
		opt(c)
	}

	return c
}

// Get performs an HTTP GET request to the specified URL using the provided context.
//
// Example usage:
//
//	client := NewClient()
//	data, err := client.Get(context.Background(), "http://example.com")
func (h *httpClient) Get(ctx context.Context, url string) (*response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	rawResp, doErr := h.client.Do(req)
	if doErr != nil {
		return nil, fmt.Errorf("failed to do request: %w", doErr)
	}

	// Ensure body is closed after reading to prevent resource leaks.
	defer func() {
		if cErr := rawResp.Body.Close(); cErr != nil {
			log.Warn().Msg("failed to close response body")
		}
	}()

	body, err := io.ReadAll(rawResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	resp := &response{
		StatusCode:  rawResp.StatusCode,
		Status:      rawResp.Status,
		Body:        body,
		RawResponse: rawResp,
	}

	return resp, nil
}
