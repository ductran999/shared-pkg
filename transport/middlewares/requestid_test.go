package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ductran999/shared-pkg/transport/ginresp"
	httpclient "github.com/ductran999/shared-pkg/transport/httpclient/v2"
	"github.com/ductran999/shared-pkg/transport/middlewares"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middlewares.RequestID())

	r.GET("/ping", func(c *gin.Context) {
		ginresp.OK(c, gin.H{"message": "pong"})
	})

	return r
}

func Test_RequestID(t *testing.T) {
	r := newRouter()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	responseID := w.Header().Get("X-Request-ID")
	assert.NotEmpty(t, responseID)
	assert.Contains(t, w.Body.String(), `"request_id":"`+responseID+`"`)
}

func Test_RequestIDReusesInboundHeader(t *testing.T) {
	r := newRouter()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("X-Request-ID", "inbound-id-123")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "inbound-id-123", w.Header().Get("X-Request-ID"))
	assert.Contains(t, w.Body.String(), `"request_id":"inbound-id-123"`)
}

func Test_RequestIDGeneratesWhenInboundMissing(t *testing.T) {
	r := newRouter()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	r.ServeHTTP(w, req)

	responseID := w.Header().Get("X-Request-ID")
	assert.NotEmpty(t, responseID)
	assert.NotEqual(t, "inbound-id-123", responseID)
}

func Test_RequestIDReplacesUnsafeInboundHeader(t *testing.T) {
	tests := []struct {
		name       string
		inbound    string
		wantReused bool
	}{
		{
			name:       "too long",
			inbound:    string(make([]byte, 200)),
			wantReused: false,
		},
		{
			name:       "newline",
			inbound:    "bad\nid",
			wantReused: false,
		},
		{
			name:       "tab",
			inbound:    "bad\tid",
			wantReused: false,
		},
		{
			name:       "delete character",
			inbound:    "bad\x7fid",
			wantReused: false,
		},
		{
			name:       "space is kept",
			inbound:    "bad id",
			wantReused: true,
		},
		{
			name:       "quote and backslash are kept",
			inbound:    `bad"id\`,
			wantReused: true,
		},
		{
			name:       "slash is kept",
			inbound:    "bad/id",
			wantReused: true,
		},
		{
			name:       "unicode is kept",
			inbound:    "id-áê-日本",
			wantReused: true,
		},
		{
			name:       "safe",
			inbound:    "safe-id_123.456",
			wantReused: true,
		},
		{
			name:       "uuid",
			inbound:    "9c5b2c03-2a68-4da2-8d66-2275ad156a28",
			wantReused: true,
		},
		{
			name:       "w3c trace id",
			inbound:    "4bf92f3577b34da6a3ce929d0e0e4736",
			wantReused: true,
		},
		{
			name:       "aws x-ray id",
			inbound:    "1-5759e988-bd862e3fe1be46a994272793",
			wantReused: true,
		},
		{
			name:       "short base36 token",
			inbound:    "8f14e45f",
			wantReused: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := newRouter()

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/ping", nil)
			req.Header.Set("X-Request-ID", tc.inbound)
			r.ServeHTTP(w, req)

			responseID := w.Header().Get("X-Request-ID")
			assert.NotEmpty(t, responseID)
			if tc.wantReused {
				assert.Equal(t, tc.inbound, responseID)
			} else {
				assert.NotEqual(t, tc.inbound, responseID)
			}
		})
	}
}

func Test_RequestIDIsUniquePerRequest(t *testing.T) {
	r := newRouter()

	seen := make(map[string]struct{}, 2)
	for range 2 {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		r.ServeHTTP(w, req)

		id := w.Header().Get("X-Request-ID")
		require.NotEmpty(t, id)
		seen[id] = struct{}{}
	}

	assert.Len(t, seen, 2)
}

func Test_RequestIDReachesRequestContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middlewares.RequestID())

	var (
		fromRequestContext string
		fromGinContext     string
	)
	r.GET("/ping", func(c *gin.Context) {
		fromGinContext = ginresp.RequestIDFrom(c)
		id, _ := httpclient.RequestIDFromContext(c.Request.Context())
		fromRequestContext = id
		ginresp.OK(c, gin.H{"message": "pong"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("X-Request-ID", "inbound-id-123")
	r.ServeHTTP(w, req)

	assert.Equal(t, "inbound-id-123", fromGinContext)
	assert.Equal(t, "inbound-id-123", fromRequestContext)
}

func Test_HeaderNameMatchesHTTPClient(t *testing.T) {
	// The server header and the client propagation header must not drift.
	assert.Equal(t, httpclient.DefaultRequestIDHeader, middlewares.RequestIDHeader)
}

func Test_GeneratedRequestIDIsUUIDv4(t *testing.T) {
	r := newRouter()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	r.ServeHTTP(w, req)

	parsed, err := uuid.Parse(w.Header().Get("X-Request-ID"))
	require.NoError(t, err)
	assert.Equal(t, uuid.Version(4), parsed.Version())
}
