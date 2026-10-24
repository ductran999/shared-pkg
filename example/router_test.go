package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ductran999/shared-pkg/transport/httpclient/v2"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	todos = newTodoService(&inMemoryTodoRepository{todos: []Todo{
		{ID: 1, Title: "write demo", Done: false},
		{ID: 2, Title: "commit", Done: true},
	}})

	return newRouter()
}

func unwrap(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	return body
}

// unwrapError returns the nested error object of the payload.
func unwrapError(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	body := unwrap(t, w)
	detail, ok := body["error"].(map[string]any)
	require.True(t, ok, "payload must contain a nested error object")

	return detail
}

func Test_Router_Todos(t *testing.T) {
	r := newTestRouter()

	t.Run("list", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/todos", nil)
		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)
		body := unwrap(t, w)
		assert.NotContains(t, body, "code")
		assert.NotEmpty(t, body["data"])
		metaJSON, _ := json.Marshal(body["meta"])
		assert.JSONEq(t, `{"page":1,"size":10,"total":2}`, string(metaJSON))
		assert.NotEmpty(t, body["request_id"])
	})

	t.Run("get one", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/todos/1", nil)
		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)
		body := unwrap(t, w)
		assert.Equal(t, float64(1), body["data"].(map[string]any)["id"])
	})

	t.Run("not found", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/todos/999", nil)
		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusNotFound, w.Code)
		detail := unwrapError(t, w)
		assert.Equal(t, "TODO_NOT_FOUND", detail["code"])
		assert.Equal(t, "todo not found", detail["message"])
	})

	t.Run("bad id", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/todos/abc", nil)
		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusBadRequest, w.Code)
		detail := unwrapError(t, w)
		assert.Equal(t, "INVALID_TODO_ID", detail["code"])
	})

	t.Run("create", func(t *testing.T) {
		w := httptest.NewRecorder()
		body, _ := json.Marshal(map[string]string{"title": "new task"})
		req := httptest.NewRequest(http.MethodPost, "/todos", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)
		got := unwrap(t, w)
		assert.Equal(t, "new task", got["data"].(map[string]any)["title"])
	})

	t.Run("create invalid body", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/todos", bytes.NewReader([]byte("not-json")))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusBadRequest, w.Code)
		detail := unwrapError(t, w)
		assert.Equal(t, "INVALID_BODY", detail["code"])
	})

	t.Run("create missing title", func(t *testing.T) {
		w := httptest.NewRecorder()
		body, _ := json.Marshal(map[string]string{})
		req := httptest.NewRequest(http.MethodPost, "/todos", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusBadRequest, w.Code)
		detail := unwrapError(t, w)
		assert.Equal(t, "TITLE_REQUIRED", detail["code"])
		detailsJSON, _ := json.Marshal(detail["details"])
		assert.JSONEq(t, `[{"field":"title","message":"is required"}]`, string(detailsJSON))
	})
}

func Test_Router_Upstream(t *testing.T) {
	// Point the package-level client at a local mock so /upstream is hermetic.
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"message":"hi","env":"test"}`))
	}))
	defer upstream.Close()

	upstreamClient = httpclient.New(httpclient.WithBaseURL(upstream.URL))
	defer func() { upstreamClient = nil }()

	r := newTestRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/upstream", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	body := unwrap(t, w)
	assert.Equal(t, "hi", body["data"].(map[string]any)["message"])
	assert.NotEmpty(t, body["request_id"])
}

func Test_Router_Panic(t *testing.T) {
	r := newTestRouter()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	r.ServeHTTP(w, req)

	// The recovery middleware turns the panic into the standard envelope.
	require.Equal(t, http.StatusInternalServerError, w.Code)
	detail := unwrapError(t, w)
	assert.Equal(t, "INTERNAL_ERROR", detail["code"])
	assert.Equal(t, "internal server error", detail["message"])
}
