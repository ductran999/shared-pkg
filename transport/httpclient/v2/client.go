package httpclient

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// dialerConfig is a small helper to mutate the default *net.Dialer
// without exposing net internals in options.go.
type dialerConfig struct {
	timeout   time.Duration
	keepAlive time.Duration
}

func applyDialer(c *Client, fn func(*dialerConfig)) {
	t, ok := c.http.Transport.(*http.Transport)
	if !ok || t.DialContext == nil {
		return
	}
	// Default transport created by New uses *net.Dialer; rebuild it
	// preserving the other field.
	cfg := &dialerConfig{
		timeout:   DefaultDialTimeout,
		keepAlive: DefaultDialKeepAlive,
	}
	fn(cfg)
	t.DialContext = (&net.Dialer{
		Timeout:   cfg.timeout,
		KeepAlive: cfg.keepAlive,
	}).DialContext
}

// Client is an exported HTTP client with defaults suitable for services.
type Client struct {
	http          *http.Client
	baseURL       string
	defaultHeader http.Header

	// propagateRequestID controls whether Do copies request/trace ID
	// from ctx into the outgoing header. Default true.
	propagateRequestID bool
	// requestIDHeader is the header name used for propagation.
	// Default DefaultRequestIDHeader ("X-Request-ID").
	requestIDHeader string
}

// New creates a Client with sensible defaults, then applies options.
func New(options ...Option) *Client {
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

	c := &Client{
		http: &http.Client{
			Timeout:   DefaultTimeout,
			Transport: transport,
		},
		defaultHeader:      make(http.Header),
		propagateRequestID: true,
		requestIDHeader:    DefaultRequestIDHeader,
	}

	for _, opt := range options {
		if opt != nil {
			opt(c)
		}
	}

	return c
}

// CloseIdleConnections closes idle connections, same as http.Client.
func (c *Client) CloseIdleConnections() {
	c.http.CloseIdleConnections()
}

// resolveURL joins baseURL with a relative path. Absolute URLs pass through.
// Join is slash-aware: base "https://api/x/v1" + "/users" => "https://api/x/v1/users".
func (c *Client) resolveURL(rawURL string) (string, error) {
	if c.baseURL == "" {
		return rawURL, nil
	}
	if strings.HasPrefix(rawURL, "http://") || strings.HasPrefix(rawURL, "https://") {
		return rawURL, nil
	}
	if _, err := url.Parse(c.baseURL); err != nil {
		return "", fmt.Errorf("httpclient: invalid baseURL %q: %w: %w", c.baseURL, ErrInvalidBaseURL, err)
	}
	if _, err := url.Parse(rawURL); err != nil {
		return "", fmt.Errorf("httpclient: invalid request URL %q: %w: %w", rawURL, ErrInvalidURL, err)
	}
	return strings.TrimSuffix(c.baseURL, "/") + "/" + strings.TrimPrefix(rawURL, "/"), nil
}

// Do performs an HTTP request with context, optional body and request options.
//
// It returns the raw *http.Response. The caller owns resp.Body and must
// close it (usually via defer), exactly like the stdlib:
//
//	resp, err := c.Do(ctx, http.MethodGet, url, nil)
//	if err != nil {
//		return err
//	}
//	defer resp.Body.Close()
//
// Handle resp.Body directly with the stdlib (json.Decoder, io.Copy...)
// using whatever format the service speaks (JSON, YAML, ...).
// This keeps streaming and large bodies working without forced buffering.
func (c *Client) Do(ctx context.Context, method, rawURL string, body io.Reader, opts ...RequestOption) (*http.Response, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	resolved, err := c.resolveURL(rawURL)
	if err != nil {
		return nil, err
	}

	rc := newRequestConfig(opts)

	u, err := url.Parse(resolved)
	if err != nil {
		return nil, fmt.Errorf("httpclient: invalid URL %q: %w: %w", resolved, ErrInvalidURL, err)
	}
	q := u.Query()
	for k, vs := range rc.query {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return nil, fmt.Errorf("httpclient: create request: %w: %w", ErrCreateRequest, err)
	}

	for k, vs := range c.defaultHeader {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}
	for k, vs := range rc.header {
		// Per-request headers override defaults.
		req.Header.Del(k)
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}
	if rc.hasBasicAuth {
		req.SetBasicAuth(rc.basicUser, rc.basicPass)
	}

	// Propagate request/trace ID from ctx unless already set explicitly.
	// Explicit (default or per-request) always wins over auto.
	if c.propagateRequestID {
		headerName := c.requestIDHeader
		if headerName == "" {
			headerName = DefaultRequestIDHeader
		}
		if req.Header.Get(headerName) == "" {
			if id, ok := RequestIDFromContext(ctx); ok {
				req.Header.Set(headerName, id)
			}
		}
	}

	rawResp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("httpclient: do request: %w: %w", ErrDoRequest, err)
	}

	return rawResp, nil
}

// Get performs an HTTP GET request. See Do for body ownership.
func (c *Client) Get(ctx context.Context, rawURL string, opts ...RequestOption) (*http.Response, error) {
	return c.Do(ctx, http.MethodGet, rawURL, nil, opts...)
}

// Head performs an HTTP HEAD request. See Do for body ownership.
func (c *Client) Head(ctx context.Context, rawURL string, opts ...RequestOption) (*http.Response, error) {
	return c.Do(ctx, http.MethodHead, rawURL, nil, opts...)
}

// Delete performs an HTTP DELETE request. See Do for body ownership.
func (c *Client) Delete(ctx context.Context, rawURL string, opts ...RequestOption) (*http.Response, error) {
	return c.Do(ctx, http.MethodDelete, rawURL, nil, opts...)
}

// Post performs an HTTP POST request with an optional body. See Do for body ownership.
func (c *Client) Post(ctx context.Context, rawURL string, body io.Reader, opts ...RequestOption) (*http.Response, error) {
	return c.Do(ctx, http.MethodPost, rawURL, body, opts...)
}

// Put performs an HTTP PUT request with an optional body. See Do for body ownership.
func (c *Client) Put(ctx context.Context, rawURL string, body io.Reader, opts ...RequestOption) (*http.Response, error) {
	return c.Do(ctx, http.MethodPut, rawURL, body, opts...)
}

// Patch performs an HTTP PATCH request with an optional body. See Do for body ownership.
func (c *Client) Patch(ctx context.Context, rawURL string, body io.Reader, opts ...RequestOption) (*http.Response, error) {
	return c.Do(ctx, http.MethodPatch, rawURL, body, opts...)
}
