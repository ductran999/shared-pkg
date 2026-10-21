package response_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ductran999/shared-pkg/transport/response"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/todos", func(c *gin.Context) {
		c.JSON(http.StatusOK, response.OK([]string{"a", "b"}))
	})

	r.POST("/todos", func(c *gin.Context) {
		c.JSON(http.StatusCreated, response.Created(gin.H{"id": 7}))
	})

	r.GET("/list", func(c *gin.Context) {
		c.JSON(http.StatusOK, response.List([]string{"a"}, response.NewPagination(1, 10, 100)))
	})

	r.GET("/bad", func(c *gin.Context) {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid id"))
	})

	r.GET("/custom", func(c *gin.Context) {
		c.JSON(http.StatusBadRequest, response.NewError("OBJECT_ALREADY_EXISTED", "object already existed", nil))
	})

	return r
}

func Test_GinHandlers(t *testing.T) {
	r := setupRouter()

	tests := []struct {
		name   string
		method string
		path   string
		status int
		want   string
	}{
		{
			name:   "ok",
			method: http.MethodGet,
			path:   "/todos",
			status: http.StatusOK,
			want:   `{"data":["a","b"],"request_id":""}`,
		},
		{
			name:   "created",
			method: http.MethodPost,
			path:   "/todos",
			status: http.StatusCreated,
			want:   `{"data":{"id":7},"request_id":""}`,
		},
		{
			name:   "list",
			method: http.MethodGet,
			path:   "/list",
			status: http.StatusOK,
			want:   `{"data":["a"],"meta":{"page":1,"size":10,"total":100},"request_id":""}`,
		},
		{
			name:   "bad request",
			method: http.MethodGet,
			path:   "/bad",
			status: http.StatusBadRequest,
			want:   `{"error":{"code":"BAD_REQUEST","message":"invalid id"},"request_id":""}`,
		},
		{
			name:   "custom code",
			method: http.MethodGet,
			path:   "/custom",
			status: http.StatusBadRequest,
			want:   `{"error":{"code":"OBJECT_ALREADY_EXISTED","message":"object already existed"},"request_id":""}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, tc.path, nil)
			r.ServeHTTP(w, req)

			require.Equal(t, tc.status, w.Code)
			assert.JSONEq(t, tc.want, w.Body.String())
		})
	}
}
