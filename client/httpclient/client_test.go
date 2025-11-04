package httpclient_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ductran999/shared-pkg/client/httpclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewClientValidOptions(t *testing.T) {
	// Arrange
	c := httpclient.NewClient(
		httpclient.WithIdleConnTimeout(90*time.Second),
		httpclient.WithMaxIdleConns(50),
		httpclient.WithTimeout(5*time.Second),
		httpclient.WithMaxIdleConnsPerHost(10),
		httpclient.WithTLSHandshakeTimeout(5*time.Second),
	)

	// Assert
	require.NotNil(t, c, "Expected non-nil httpClient instance")
}

func Test_NewClientInvalidOptions(t *testing.T) {
	// Arrange negative or zero values for options
	c := httpclient.NewClient(
		httpclient.WithIdleConnTimeout(-90*time.Second),
		httpclient.WithMaxIdleConns(-50),
		httpclient.WithTimeout(-5*time.Second),
		httpclient.WithMaxIdleConnsPerHost(-10),
		httpclient.WithTLSHandshakeTimeout(-5*time.Second),
	)

	// Assert
	require.NotNil(t, c, "Expected non-nil httpClient instance")
}

func Test_HttpClientGet(t *testing.T) {
	t.Run("returns response on successful request", func(t *testing.T) {
		// Create a mock server
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/test", r.URL.Path)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message":"success"}`))
		}))
		defer mockServer.Close()

		// Initialize your HTTP client
		httpClient := httpclient.NewClient(
			httpclient.WithTimeout(2 * time.Second),
		)

		// Call your method
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		resp, err := httpClient.Get(ctx, mockServer.URL+"/test")
		require.NoError(t, err)
		require.JSONEq(t, `{"message":"success"}`, string(resp.Body))
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("returns error on non-200 response", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/test", r.URL.Path)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}))
		defer mockServer.Close()

		// Initialize your HTTP client
		client := httpclient.NewClient(
			httpclient.WithTimeout(2 * time.Second),
		)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		resp, err := client.Get(ctx, mockServer.URL+"/test")
		require.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("unreachable host returns error", func(t *testing.T) {
		// Initialize your HTTP client
		client := httpclient.NewClient(
			httpclient.WithTimeout(2 * time.Second),
		)

		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
		defer cancel()

		// non-routable IP to simulate connection failure
		_, err := client.Get(ctx, "http://10.255.255.1:12345")
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to do request")
	})
}
