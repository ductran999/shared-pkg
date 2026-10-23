package middlewares_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ductran999/shared-pkg/transport/middlewares"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Recovery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middlewares.RequestID())
	r.Use(middlewares.Logger())
	r.Use(middlewares.Recovery())

	r.GET("/panic", func(_ *gin.Context) { panic("boom") })
	r.GET("/ok", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "pong"}) })

	t.Run("panic becomes the error envelope", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/panic", nil)
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusInternalServerError, w.Code)

		var body map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))

		detail, ok := body["error"].(map[string]any)
		require.True(t, ok, "payload must contain a nested error object")
		assert.Equal(t, "INTERNAL_ERROR", detail["code"])
		assert.Equal(t, "internal server error", detail["message"])
		assert.NotEmpty(t, body["request_id"])
	})

	t.Run("normal request is untouched", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/ok", nil)
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"message":"pong"}`, w.Body.String())
	})
}

// Test_RecoveryProducesTwoLogLines checks the shape of the logging: one
// detailed error line for the investigation and one lean access line.
func Test_RecoveryProducesTwoLogLines(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middlewares.RequestID())
	r.Use(middlewares.Logger())
	r.Use(middlewares.Recovery())
	r.GET("/panic", func(_ *gin.Context) { panic("boom") })

	buf := captureLog(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	r.ServeHTTP(w, req)

	records := logRecords(t, buf)
	require.Len(t, records, 2, "a panicking request must produce an error line and an access line")

	errorLine := logRecord(t, buf, "recovered from panic")
	accessLine := logRecord(t, buf, "http request")

	// Both lines point at the same request.
	assert.NotEmpty(t, errorLine["request_id"])
	assert.Equal(t, errorLine["request_id"], accessLine["request_id"])

	// The error line is the detailed one.
	assert.Equal(t, "error", errorLine["level"])
	assert.Equal(t, "boom", errorLine["panic"])
	assert.Equal(t, http.MethodGet, errorLine["method"])
	assert.Equal(t, "/panic", errorLine["path"])
	assert.Contains(t, errorLine, "stack")

	// The access line stays lean, it only flags the panic.
	assert.Equal(t, "error", accessLine["level"])
	assert.Equal(t, float64(http.StatusInternalServerError), accessLine["status"])
	assert.Equal(t, true, accessLine["panic"])
	assert.NotContains(t, accessLine, "stack")
	assert.NotContains(t, accessLine, "panic_value")
}

func Test_RecoveryLogsStructuredStack(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middlewares.RequestID())
	r.Use(middlewares.Logger())
	r.Use(middlewares.Recovery())
	r.GET("/panic", func(_ *gin.Context) { panic("boom") })

	buf := captureLog(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	r.ServeHTTP(w, req)

	fields := logRecord(t, buf, "recovered from panic")

	// The stack must be a list of frames, not a blob of text with newlines.
	stack, ok := fields["stack"].([]any)
	require.True(t, ok, "stack must be a list of frames")

	frames := make([]string, 0, len(stack))
	for _, frame := range stack {
		entry, isObject := frame.(map[string]any)
		require.True(t, isObject, "each stack frame must be an object")
		assert.NotEmpty(t, entry["function"])
		assert.Contains(t, entry, "file")
		assert.Contains(t, entry, "line")

		function, _ := entry["function"].(string)
		frames = append(frames, function)
	}

	// The frame of the panicking handler must be part of the stack.
	assert.Contains(t, strings.Join(frames, "\n"), "Test_RecoveryLogsStructuredStack")
}

// Test_RecoveryInsideLoggerKeepsAccessLine checks the middleware order: the
// Recovery must be registered inside the Logger, otherwise a panicking
// request would lose its access log line and the metrics would drift.
func Test_RecoveryInsideLoggerKeepsAccessLine(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middlewares.RequestID())
	r.Use(middlewares.Logger())
	r.Use(middlewares.Recovery())
	r.GET("/panic", func(_ *gin.Context) { panic("boom") })

	buf := captureLog(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	r.ServeHTTP(w, req)

	accessLine := logRecord(t, buf, "http request")
	assert.Equal(t, float64(http.StatusInternalServerError), accessLine["status"])
}

func Test_RecoveryKeepsWrittenResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middlewares.RequestID())
	r.Use(middlewares.Logger())
	r.Use(middlewares.Recovery())

	// The handler writes a response first, then panics. The already written
	// body must be kept, only the status code can still be changed.
	r.GET("/partial", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "partial"})
		panic("boom")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/partial", nil)
	r.ServeHTTP(w, req)

	assert.JSONEq(t, `{"message":"partial"}`, w.Body.String())
}
