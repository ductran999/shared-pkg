package httpclient

import (
	"net/http"
	"net/url"
)

// requestConfig holds per-request settings built from RequestOption.
type requestConfig struct {
	header       http.Header
	query        url.Values
	basicUser    string
	basicPass    string
	hasBasicAuth bool
}

// RequestOption configures a single request.
type RequestOption func(*requestConfig)

// WithHeader sets a single request header.
func WithHeader(key, value string) RequestOption {
	return func(rc *requestConfig) {
		if key == "" {
			return
		}
		rc.header.Set(key, value)
	}
}

// WithHeaders merges headers into the request.
func WithHeaders(h http.Header) RequestOption {
	return func(rc *requestConfig) {
		for k, vs := range h {
			for _, v := range vs {
				rc.header.Add(k, v)
			}
		}
	}
}

// WithQueryParam sets a single query parameter.
func WithQueryParam(key, value string) RequestOption {
	return func(rc *requestConfig) {
		if key == "" {
			return
		}
		rc.query.Set(key, value)
	}
}

// WithQuery merges query values into the request URL.
func WithQuery(q url.Values) RequestOption {
	return func(rc *requestConfig) {
		for k, vs := range q {
			for _, v := range vs {
				rc.query.Add(k, v)
			}
		}
	}
}

// WithBearerToken sets Authorization: Bearer <token>.
func WithBearerToken(token string) RequestOption {
	return func(rc *requestConfig) {
		if token == "" {
			return
		}
		rc.header.Set("Authorization", "Bearer "+token)
	}
}

// WithBasicAuth sets basic auth on the request.
func WithBasicAuth(username, password string) RequestOption {
	return func(rc *requestConfig) {
		rc.basicUser = username
		rc.basicPass = password
		rc.hasBasicAuth = true
	}
}

// WithContentType sets Content-Type for the request.
func WithContentType(contentType string) RequestOption {
	return func(rc *requestConfig) {
		if contentType == "" {
			return
		}
		rc.header.Set("Content-Type", contentType)
	}
}

// WithAccept sets Accept for the request.
func WithAccept(accept string) RequestOption {
	return func(rc *requestConfig) {
		if accept == "" {
			return
		}
		rc.header.Set("Accept", accept)
	}
}

func newRequestConfig(opts []RequestOption) *requestConfig {
	rc := &requestConfig{
		header: make(http.Header),
		query:  make(url.Values),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(rc)
		}
	}
	return rc
}
