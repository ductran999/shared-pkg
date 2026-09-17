package httpclient

import (
	"net/http"
	"time"
)

const (
	DefaultTimeout = 5 * time.Second

	DefaultMaxIdleConns        = 100
	DefaultMaxIdleConnsPerHost = 10

	DefaultDialTimeout   = 5 * time.Second
	DefaultDialKeepAlive = 5 * time.Second

	DefaultIdleConnTimeout     = 90 * time.Second
	DefaultTLSHandshakeTimeout = 5 * time.Second
)

// Option configures a Client. Invalid values are ignored so defaults are kept.
type Option func(*Client)

// WithTimeout sets the total timeout for a request, including body read.
// Defaults to DefaultTimeout. Ignored if d <= 0.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		if d <= 0 {
			return
		}
		c.http.Timeout = d
	}
}

// WithTransport replaces the underlying RoundTripper entirely.
func WithTransport(rt http.RoundTripper) Option {
	return func(c *Client) {
		if rt == nil {
			return
		}
		c.http.Transport = rt
	}
}

// WithBaseURL sets a base URL prepended to relative request paths.
// Example: WithBaseURL("https://api.example.com") + Get(ctx, "/users").
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// WithDefaultHeader adds a default header sent on every request.
// Per-request headers set via RequestOption override it.
func WithDefaultHeader(key, value string) Option {
	return func(c *Client) {
		if key == "" {
			return
		}
		c.defaultHeader.Set(key, value)
	}
}

// WithDefaultHeaders merges a set of default headers sent on every request.
func WithDefaultHeaders(h http.Header) Option {
	return func(c *Client) {
		for k, vs := range h {
			for _, v := range vs {
				c.defaultHeader.Add(k, v)
			}
		}
	}
}

// WithDisableCompression disables transparent gzip handling.
func WithDisableCompression(disable bool) Option {
	return func(c *Client) {
		if t, ok := c.http.Transport.(*http.Transport); ok {
			t.DisableCompression = disable
		}
	}
}

// WithDisableKeepAlives disables HTTP keep-alives.
func WithDisableKeepAlives(disable bool) Option {
	return func(c *Client) {
		if t, ok := c.http.Transport.(*http.Transport); ok {
			t.DisableKeepAlives = disable
		}
	}
}

// WithMaxIdleConns sets MaxIdleConns on the default transport.
// Ignored if n <= 0 or a custom RoundTripper was set.
func WithMaxIdleConns(n int) Option {
	return func(c *Client) {
		if n <= 0 {
			return
		}
		if t, ok := c.http.Transport.(*http.Transport); ok {
			t.MaxIdleConns = n
		}
	}
}

// WithMaxIdleConnsPerHost sets MaxIdleConnsPerHost on the default transport.
// Ignored if n <= 0 or a custom RoundTripper was set.
func WithMaxIdleConnsPerHost(n int) Option {
	return func(c *Client) {
		if n <= 0 {
			return
		}
		if t, ok := c.http.Transport.(*http.Transport); ok {
			t.MaxIdleConnsPerHost = n
		}
	}
}

// WithIdleConnTimeout sets IdleConnTimeout on the default transport.
// Ignored if d <= 0 or a custom RoundTripper was set.
func WithIdleConnTimeout(d time.Duration) Option {
	return func(c *Client) {
		if d <= 0 {
			return
		}
		if t, ok := c.http.Transport.(*http.Transport); ok {
			t.IdleConnTimeout = d
		}
	}
}

// WithTLSHandshakeTimeout sets TLSHandshakeTimeout on the default transport.
// Ignored if d <= 0 or a custom RoundTripper was set.
func WithTLSHandshakeTimeout(d time.Duration) Option {
	return func(c *Client) {
		if d <= 0 {
			return
		}
		if t, ok := c.http.Transport.(*http.Transport); ok {
			t.TLSHandshakeTimeout = d
		}
	}
}

// WithDialTimeout sets the dialer Timeout on the default transport.
// Ignored if d <= 0 or a custom RoundTripper was set.
func WithDialTimeout(d time.Duration) Option {
	return func(c *Client) {
		if d <= 0 {
			return
		}
		applyDialer(c, func(dialer *dialerConfig) { dialer.timeout = d })
	}
}

// WithDialKeepAlive sets the dialer KeepAlive on the default transport.
// Ignored if d <= 0 or a custom RoundTripper was set.
func WithDialKeepAlive(d time.Duration) Option {
	return func(c *Client) {
		if d <= 0 {
			return
		}
		applyDialer(c, func(dialer *dialerConfig) { dialer.keepAlive = d })
	}
}

// WithRequestIDPropagation enables or disables copying request/trace ID
// from ctx into the outgoing X-Request-ID header. Default enabled.
func WithRequestIDPropagation(enabled bool) Option {
	return func(c *Client) {
		c.propagateRequestID = enabled
	}
}

// WithRequestIDHeader customizes the header name used for request ID
// propagation. Default DefaultRequestIDHeader ("X-Request-ID").
// Ignored if name is empty.
func WithRequestIDHeader(name string) Option {
	return func(c *Client) {
		if name == "" {
			return
		}
		c.requestIDHeader = name
	}
}
