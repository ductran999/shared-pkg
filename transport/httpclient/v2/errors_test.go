package httpclient_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	httpclient "github.com/ductran999/shared-pkg/transport/httpclient/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_SentinelErrors(t *testing.T) {
	// Closed server forces connection refused immediately (no timeout wait).
	refused := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	refusedURL := refused.URL
	refused.Close()

	tests := []struct {
		name    string
		run     func() error
		wantErr error
	}{
		{
			name: "invalid baseURL",
			run: func() error {
				c := httpclient.New(httpclient.WithBaseURL("://bad"))
				_, err := c.Get(context.Background(), "/users")
				return err
			},
			wantErr: httpclient.ErrInvalidBaseURL,
		},
		{
			name: "invalid request URL",
			run: func() error {
				c := httpclient.New()
				_, err := c.Get(context.Background(), "://bad")
				return err
			},
			wantErr: httpclient.ErrInvalidURL,
		},
		{
			name: "request failed",
			run: func() error {
				c := httpclient.New()
				_, err := c.Get(context.Background(), refusedURL)
				return err
			},
			wantErr: httpclient.ErrDoRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.run()
			require.Error(t, err)
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}

func Test_DoRequestKeepsUnderlyingChain(t *testing.T) {
	// Multiple %w must preserve the underlying *url.Error for errors.As.
	refused := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	refusedURL := refused.URL
	refused.Close()

	c := httpclient.New()
	_, err := c.Get(context.Background(), refusedURL)
	require.Error(t, err)
	assert.ErrorIs(t, err, httpclient.ErrDoRequest)

	var urlErr *url.Error
	assert.True(t, errors.As(err, &urlErr), "expected *url.Error in chain, got %T", err)
}
