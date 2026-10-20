package ginresp_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ductran999/shared-pkg/transport/ginresp"
	"github.com/ductran999/shared-pkg/transport/response"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(ginresp.RequestID())

	r.GET("/ok", func(c *gin.Context) { ginresp.OK(c, []string{"a"}) })
	r.GET("/created", func(c *gin.Context) { ginresp.Created(c, gin.H{"id": 1}) })
	r.GET("/list", func(c *gin.Context) {
		ginresp.List(c, []string{"a"}, response.Pagination{Page: 2, Size: 10, Total: 95})
	})
	r.GET("/bad", func(c *gin.Context) { ginresp.BadRequest(c, "invalid id") })
	r.GET("/custom", func(c *gin.Context) {
		ginresp.Error(c, http.StatusBadRequest, "OBJECT_ALREADY_EXISTED", "object already existed")
	})
	r.GET("/boom", func(c *gin.Context) { ginresp.InternalError(c, "db down") })

	return r
}

func Test_GinResp(t *testing.T) {
	r := setupRouter()

	tests := []struct {
		name        string
		path        string
		status      int
		wantCode    string
		wantMessage string
		wantData    string
		wantMeta    string
	}{
		{
			name:     "ok",
			path:     "/ok",
			status:   http.StatusOK,
			wantData: `["a"]`,
		},
		{
			name:     "created",
			path:     "/created",
			status:   http.StatusCreated,
			wantData: `{"id":1}`,
		},
		{
			name:     "list",
			path:     "/list",
			status:   http.StatusOK,
			wantData: `["a"]`,
			wantMeta: `{"page":2,"size":10,"total":95}`,
		},
		{
			name:        "bad request",
			path:        "/bad",
			status:      http.StatusBadRequest,
			wantCode:    "BAD_REQUEST",
			wantMessage: "invalid id",
		},
		{
			name:        "custom code",
			path:        "/custom",
			status:      http.StatusBadRequest,
			wantCode:    "OBJECT_ALREADY_EXISTED",
			wantMessage: "object already existed",
		},
		{
			name:        "internal error",
			path:        "/boom",
			status:      http.StatusInternalServerError,
			wantCode:    "INTERNAL_ERROR",
			wantMessage: "db down",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			r.ServeHTTP(w, req)

			require.Equal(t, tc.status, w.Code)

			var body map[string]any
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))

			if tc.wantCode != "" {
				assert.Equal(t, tc.wantCode, body["code"])
			} else {
				assert.NotContains(t, body, "code")
			}
			if tc.wantMessage != "" {
				assert.Equal(t, tc.wantMessage, body["message"])
			}
			if tc.wantData != "" {
				b, _ := json.Marshal(body["data"])
				assert.JSONEq(t, tc.wantData, string(b))
			}
			if tc.wantMeta != "" {
				b, _ := json.Marshal(body["meta"])
				assert.JSONEq(t, tc.wantMeta, string(b))
			}
			assert.NotEmpty(t, body["request_id"])
		})
	}
}

func Test_RequestIDHeader(t *testing.T) {
	r := setupRouter()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	r.ServeHTTP(w, req)

	id := w.Header().Get("X-Request-ID")
	assert.NotEmpty(t, id)
	assert.Contains(t, w.Body.String(), `"request_id":"`+id+`"`)
}
