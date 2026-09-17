package httpclient_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	httpclient "github.com/ductran999/shared-pkg/transport/httpclient/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// readBody reads and closes a raw response body, returning it as string.
func readBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	defer func() { _ = resp.Body.Close() }()
	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return string(b)
}

func Test_NewDefaults(t *testing.T) {
	c := httpclient.New()
	require.NotNil(t, c)

	// Invalid options must be ignored, not panic.
	c2 := httpclient.New(
		httpclient.WithTimeout(-1),
		httpclient.WithMaxIdleConns(0),
		httpclient.WithMaxIdleConnsPerHost(-5),
		httpclient.WithIdleConnTimeout(0),
		httpclient.WithTLSHandshakeTimeout(-1),
		httpclient.WithDialTimeout(0),
		httpclient.WithDialKeepAlive(-1),
		httpclient.WithTransport(nil),
		httpclient.WithDefaultHeader("", "x"),
		nil,
	)
	require.NotNil(t, c2)
}

func Test_GetSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/test", r.URL.Path)
		assert.Equal(t, "Bearer token123", r.Header.Get("Authorization"))
		assert.Equal(t, "bar", r.URL.Query().Get("foo"))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message":"success"}`))
	}))
	defer srv.Close()

	c := httpclient.New(httpclient.WithTimeout(2 * time.Second))
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resp, err := c.Get(ctx, srv.URL+"/test",
		httpclient.WithBearerToken("token123"),
		httpclient.WithQueryParam("foo", "bar"),
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.JSONEq(t, `{"message":"success"}`, readBody(t, resp))
}

func Test_StdlibInterop(t *testing.T) {
	// Raw *http.Response works with the stdlib directly, no helpers needed.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"hello":"world"}`))
	}))
	defer srv.Close()

	c := httpclient.New()
	resp, err := c.Get(context.Background(), srv.URL)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	var out map[string]string
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&out))
	assert.Equal(t, "world", out["hello"])
}

func Test_PostWithBody(t *testing.T) {
	// Caller marshals the body with whatever library/format they choose.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Contains(t, r.Header.Get("Content-Type"), "application/json")
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.JSONEq(t, `{"name":"a"}`, string(body))
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":1}`))
	}))
	defer srv.Close()

	payload, err := json.Marshal(map[string]string{"name": "a"})
	require.NoError(t, err)

	c := httpclient.New()
	resp, err := c.Post(context.Background(), srv.URL, bytes.NewReader(payload),
		httpclient.WithContentType("application/json"),
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.JSONEq(t, `{"id":1}`, readBody(t, resp))
}

func Test_BaseURLAndDefaultHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/users", r.URL.Path)
		if r.URL.Query().Get("case") == "override" {
			assert.Equal(t, "req", r.Header.Get("X-Tenant"))
		} else {
			assert.Equal(t, "default", r.Header.Get("X-Tenant"))
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`ok`))
	}))
	defer srv.Close()

	c := httpclient.New(
		httpclient.WithBaseURL(srv.URL+"/v1"),
		httpclient.WithDefaultHeader("X-Tenant", "default"),
	)

	resp, err := c.Get(context.Background(), "/users")
	require.NoError(t, err)
	assert.Equal(t, "ok", readBody(t, resp))

	resp, err = c.Get(context.Background(), "/users",
		httpclient.WithHeader("X-Tenant", "req"),
		httpclient.WithQueryParam("case", "override"),
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	_ = resp.Body.Close()
}

func Test_BasicAuth(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		require.True(t, ok)
		assert.Equal(t, "user", u)
		assert.Equal(t, "pass", p)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := httpclient.New()
	resp, err := c.Get(context.Background(), srv.URL, httpclient.WithBasicAuth("user", "pass"))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	_ = resp.Body.Close()
}

func Test_Methods(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(r.Method))
	}))
	defer srv.Close()

	c := httpclient.New()
	ctx := context.Background()

	for method, fn := range map[string]func() (*http.Response, error){
		"GET":    func() (*http.Response, error) { return c.Get(ctx, srv.URL) },
		"DELETE": func() (*http.Response, error) { return c.Delete(ctx, srv.URL) },
		"POST":   func() (*http.Response, error) { return c.Post(ctx, srv.URL, nil) },
		"PUT":    func() (*http.Response, error) { return c.Put(ctx, srv.URL, nil) },
		"PATCH":  func() (*http.Response, error) { return c.Patch(ctx, srv.URL, nil) },
	} {
		resp, err := fn()
		require.NoError(t, err, method)
		assert.Equal(t, method, readBody(t, resp), method)
	}
}

func Test_Unreachable(t *testing.T) {
	c := httpclient.New(httpclient.WithTimeout(500 * time.Millisecond))
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	_, err := c.Get(ctx, "http://10.255.255.1:12345")
	require.Error(t, err)
	require.Contains(t, err.Error(), "httpclient:")
	assert.ErrorIs(t, err, httpclient.ErrDoRequest)
}
