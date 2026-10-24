package httpclient_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	httpclient "github.com/ductran999/shared-pkg/transport/httpclient/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Head(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodHead, r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := httpclient.New()
	resp, err := c.Head(context.Background(), srv.URL)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func Test_CloseIdleConnections(t *testing.T) {
	c := httpclient.New()
	assert.NotPanics(t, func() { c.CloseIdleConnections() })
}

func Test_WithDefaultHeaders(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "a", r.Header.Get("X-One"))
		assert.Equal(t, "b", r.Header.Get("X-Two"))
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := httpclient.New(httpclient.WithDefaultHeaders(http.Header{
		"X-One": []string{"a"},
		"X-Two": []string{"b"},
	}))

	resp, err := c.Get(context.Background(), srv.URL)
	require.NoError(t, err)
	_ = resp.Body.Close()
}

func Test_WithHeaders(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "a", r.Header.Get("X-One"))
		assert.Equal(t, "b", r.Header.Get("X-Two"))
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := httpclient.New()
	resp, err := c.Get(context.Background(), srv.URL, httpclient.WithHeaders(http.Header{
		"X-One": []string{"a"},
		"X-Two": []string{"b"},
	}))
	require.NoError(t, err)
	_ = resp.Body.Close()
}

func Test_WithQuery(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "1", r.URL.Query().Get("page"))
		assert.Equal(t, "2", r.URL.Query().Get("size"))
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := httpclient.New()
	resp, err := c.Get(context.Background(), srv.URL, httpclient.WithQuery(url.Values{
		"page": []string{"1"},
		"size": []string{"2"},
	}))
	require.NoError(t, err)
	_ = resp.Body.Close()
}

func Test_WithAccept(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/json", r.Header.Get("Accept"))
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := httpclient.New()
	resp, err := c.Get(context.Background(), srv.URL, httpclient.WithAccept("application/json"))
	require.NoError(t, err)
	_ = resp.Body.Close()
}

func Test_TransportOptions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := httpclient.New(
		httpclient.WithDisableCompression(true),
		httpclient.WithDisableKeepAlives(true),
		httpclient.WithMaxIdleConns(42),
		httpclient.WithMaxIdleConnsPerHost(7),
		httpclient.WithIdleConnTimeout(11*time.Second),
		httpclient.WithTLSHandshakeTimeout(12*time.Second),
		httpclient.WithDialTimeout(13*time.Second),
		httpclient.WithDialKeepAlive(14*time.Second),
	)

	resp, err := c.Get(context.Background(), srv.URL)
	require.NoError(t, err)
	_ = resp.Body.Close()
}

func Test_RequestIDPropagationDisabled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Empty(t, r.Header.Get("X-Request-ID"))
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ctx := httpclient.WithRequestID(context.Background(), "ctx-request-id")
	c := httpclient.New(httpclient.WithRequestIDPropagation(false))

	resp, err := c.Get(ctx, srv.URL)
	require.NoError(t, err)
	_ = resp.Body.Close()
}

func Test_InvalidBaseURL(t *testing.T) {
	c := httpclient.New(httpclient.WithBaseURL("://bad-url"))

	_, err := c.Get(context.Background(), "/users")
	require.Error(t, err)
	assert.ErrorIs(t, err, httpclient.ErrInvalidBaseURL)
}
