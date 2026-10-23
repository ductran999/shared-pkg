package middlewares_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ductran999/shared-pkg/transport/middlewares"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// captureLog redirects the global logger to a buffer for the duration of a test.
func captureLog(t *testing.T) *bytes.Buffer {
	t.Helper()

	var buf bytes.Buffer
	original := log.Logger
	log.Logger = zerolog.New(&buf)
	t.Cleanup(func() { log.Logger = original })

	return &buf
}

// logRecords decodes every log line written to the buffer.
func logRecords(t *testing.T, buf *bytes.Buffer) []map[string]any {
	t.Helper()

	var records []map[string]any
	for _, line := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
		if line == "" {
			continue
		}

		var fields map[string]any
		require.NoError(t, json.Unmarshal([]byte(line), &fields))
		records = append(records, fields)
	}

	return records
}

// logRecord returns the first log line whose message matches.
func logRecord(t *testing.T, buf *bytes.Buffer, message string) map[string]any {
	t.Helper()

	for _, record := range logRecords(t, buf) {
		if record["message"] == message {
			return record
		}
	}

	t.Fatalf("no log record with message %q", message)

	return nil
}

// logFields decodes the last written log line.
func logFields(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()

	var fields map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &fields))

	return fields
}

func Test_Logger(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middlewares.RequestID())
	r.Use(middlewares.Logger())

	r.GET("/todos", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "ok"}) })

	buf := captureLog(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/todos?page=1", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	fields := logFields(t, buf)
	assert.Equal(t, "http request", fields["message"])
	assert.Equal(t, http.MethodGet, fields["method"])
	assert.Equal(t, "/todos", fields["path"])
	assert.Equal(t, "page=1", fields["query"])
	assert.Equal(t, float64(http.StatusOK), fields["status"])
	assert.NotEmpty(t, fields["request_id"])
	assert.Contains(t, fields, "latency")
}

func Test_Logger_LogLevelFollowsStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name   string
		status int
		level  string
	}{
		{name: "success", status: http.StatusOK, level: "info"},
		{name: "client error", status: http.StatusBadRequest, level: "warn"},
		{name: "server error", status: http.StatusInternalServerError, level: "error"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := gin.New()
			r.Use(middlewares.RequestID())
			r.Use(middlewares.Logger())
			r.GET("/x", func(c *gin.Context) { c.Status(tc.status) })

			buf := captureLog(t)

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/x", nil)
			r.ServeHTTP(w, req)

			fields := logFields(t, buf)
			assert.Equal(t, tc.level, fields["level"])
			assert.Equal(t, float64(tc.status), fields["status"])
		})
	}
}

func Test_Logger_SkipPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middlewares.RequestID())
	r.Use(middlewares.Logger("/healthz", "/metrics"))

	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	r.GET("/todos", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "ok"}) })

	buf := captureLog(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	r.ServeHTTP(w, req)

	assert.Empty(t, buf.String(), "skipped path must not be logged")

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/todos", nil)
	r.ServeHTTP(w, req)

	assert.NotEmpty(t, buf.String(), "other paths must be logged")
}
