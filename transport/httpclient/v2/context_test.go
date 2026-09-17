package httpclient_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	httpclient "github.com/ductran999/shared-pkg/transport/httpclient/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_RequestIDPropagation(t *testing.T) {
	tests := []struct {
		name       string
		ctx        func() context.Context
		clientOpts []httpclient.Option
		reqOpts    []httpclient.RequestOption
		want       string
	}{
		{
			name: "typed context propagates",
			ctx: func() context.Context {
				return httpclient.WithRequestID(context.Background(), "req-123")
			},
			want: "req-123",
		},
		{
			name: "trace_id string key propagates",
			ctx: func() context.Context {
				//nolint:staticcheck // intentional: simulate external libs using string keys
				return context.WithValue(context.Background(), "trace_id", "trace-abc")
			},
			want: "trace-abc",
		},
		{
			name: "request_id string key propagates",
			ctx: func() context.Context {
				//nolint:staticcheck // intentional: simulate external libs using string keys
				return context.WithValue(context.Background(), "request_id", "req-xyz")
			},
			want: "req-xyz",
		},
		{
			name: "no id sends no header",
			ctx: func() context.Context {
				return context.Background()
			},
			want: "",
		},
		{
			name: "explicit header wins over context",
			ctx: func() context.Context {
				return httpclient.WithRequestID(context.Background(), "from-ctx")
			},
			reqOpts: []httpclient.RequestOption{
				httpclient.WithHeader("X-Request-ID", "explicit"),
			},
			want: "explicit",
		},
		{
			name: "disabled propagation",
			ctx: func() context.Context {
				return httpclient.WithRequestID(context.Background(), "req-123")
			},
			clientOpts: []httpclient.Option{
				httpclient.WithRequestIDPropagation(false),
			},
			want: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tc.want, r.Header.Get("X-Request-ID"))
				w.WriteHeader(http.StatusOK)
			}))
			defer srv.Close()

			c := httpclient.New(tc.clientOpts...)
			resp, err := c.Get(tc.ctx(), srv.URL, tc.reqOpts...)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			_ = resp.Body.Close()
		})
	}

	t.Run("custom header name", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "", r.Header.Get("X-Request-ID"))
			assert.Equal(t, "req-123", r.Header.Get("X-Correlation-ID"))
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		c := httpclient.New(httpclient.WithRequestIDHeader("X-Correlation-ID"))
		ctx := httpclient.WithRequestID(context.Background(), "req-123")
		resp, err := c.Get(ctx, srv.URL)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		_ = resp.Body.Close()
	})
}

func Test_RequestIDFromContext(t *testing.T) {
	tests := []struct {
		name   string
		ctx    context.Context
		wantID string
		wantOK bool
	}{
		{
			name:   "nil context",
			ctx:    nil,
			wantID: "",
			wantOK: false,
		},
		{
			name:   "empty context",
			ctx:    context.Background(),
			wantID: "",
			wantOK: false,
		},
		{
			name:   "typed key",
			ctx:    httpclient.WithRequestID(context.Background(), "a"),
			wantID: "a",
			wantOK: true,
		},
		{
			name: "trace_id key",
			//nolint:staticcheck // intentional: simulate external libs using string keys
			ctx:    context.WithValue(context.Background(), "trace_id", "t-1"),
			wantID: "t-1",
			wantOK: true,
		},
		{
			name: "request_id key",
			//nolint:staticcheck // intentional: simulate external libs using string keys
			ctx:    context.WithValue(context.Background(), "request_id", "r-1"),
			wantID: "r-1",
			wantOK: true,
		},
		{
			name: "empty string ignored",
			//nolint:staticcheck // intentional: simulate external libs using string keys
			ctx:    context.WithValue(context.Background(), "request_id", ""),
			wantID: "",
			wantOK: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			id, ok := httpclient.RequestIDFromContext(tc.ctx)
			assert.Equal(t, tc.wantOK, ok)
			assert.Equal(t, tc.wantID, id)
		})
	}
}
