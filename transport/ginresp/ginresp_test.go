package ginresp_test

import (
	"encoding/json"
	stderrors "errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ductran999/shared-pkg/errs"
	"github.com/ductran999/shared-pkg/transport/ginresp"
	"github.com/ductran999/shared-pkg/transport/middlewares"
	"github.com/ductran999/shared-pkg/transport/response"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middlewares.RequestID())

	r.GET("/ok", func(c *gin.Context) { ginresp.OK(c, []string{"a"}) })
	r.GET("/created", func(c *gin.Context) { ginresp.Created(c, gin.H{"id": 1}) })
	r.GET("/list", func(c *gin.Context) {
		ginresp.List(c, []string{"a"}, response.Pagination{Page: 2, Size: 10, Total: 95})
	})
	r.GET("/bad", func(c *gin.Context) { ginresp.BadRequest(c, "invalid id") })
	r.GET("/invalid-param", func(c *gin.Context) {
		ginresp.BadRequestCode(c, response.CodeInvalidParam, "invalid id",
			response.FieldError{Field: "email", Message: "invalid"},
		)
	})
	r.GET("/custom", func(c *gin.Context) {
		ginresp.Error(c, http.StatusBadRequest, "OBJECT_ALREADY_EXISTED", "object already existed")
	})
	r.GET("/boom", func(c *gin.Context) { ginresp.InternalError(c, "db down") })

	r.GET("/app-not-found", func(c *gin.Context) {
		ginresp.WriteError(c, errs.NotFound("USER_NOT_FOUND", "user not found"))
	})
	r.GET("/app-validation", func(c *gin.Context) {
		ginresp.WriteError(c, errs.Validation("INVALID_PARAM", "validation failed",
			errs.FieldError{Field: "email", Message: "invalid"},
		))
	})
	r.GET("/app-conflict", func(c *gin.Context) {
		ginresp.WriteError(c, errs.Conflict("OBJECT_ALREADY_EXISTED", "object already existed"))
	})
	r.GET("/app-unauthorized", func(c *gin.Context) {
		ginresp.WriteError(c, errs.Unauthorized("UNAUTHORIZED", "unauthorized"))
	})
	r.GET("/app-forbidden", func(c *gin.Context) {
		ginresp.WriteError(c, errs.Forbidden("FORBIDDEN", "forbidden"))
	})
	r.GET("/app-internal", func(c *gin.Context) {
		ginresp.WriteError(c, errs.Internal("BOOM", "boom"))
	})
	r.GET("/app-plain", func(c *gin.Context) {
		ginresp.WriteError(c, stderrors.New("unexpected"))
	})

	return r
}

// errorBody returns the nested error object of the payload.
func errorBody(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))

	detail, ok := body["error"].(map[string]any)
	require.True(t, ok, "payload must contain a nested error object")

	return detail
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
			name:        "invalid param",
			path:        "/invalid-param",
			status:      http.StatusBadRequest,
			wantCode:    "INVALID_PARAM",
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
			assert.NotEmpty(t, body["request_id"])

			if tc.wantData != "" {
				b, _ := json.Marshal(body["data"])
				assert.JSONEq(t, tc.wantData, string(b))
			}
			if tc.wantMeta != "" {
				b, _ := json.Marshal(body["meta"])
				assert.JSONEq(t, tc.wantMeta, string(b))
			}
			if tc.wantCode != "" {
				detail := errorBody(t, w)
				assert.Equal(t, tc.wantCode, detail["code"])
				assert.Equal(t, tc.wantMessage, detail["message"])
			} else {
				assert.NotContains(t, body, "error")
			}
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

func Test_WriteError(t *testing.T) {
	r := setupRouter()

	tests := []struct {
		name        string
		path        string
		status      int
		wantCode    string
		wantMessage string
		wantDetails string
	}{
		{
			name:        "not found",
			path:        "/app-not-found",
			status:      http.StatusNotFound,
			wantCode:    "USER_NOT_FOUND",
			wantMessage: "user not found",
		},
		{
			name:        "validation",
			path:        "/app-validation",
			status:      http.StatusBadRequest,
			wantCode:    "INVALID_PARAM",
			wantMessage: "validation failed",
			wantDetails: `[{"field":"email","message":"invalid"}]`,
		},
		{
			name:        "conflict",
			path:        "/app-conflict",
			status:      http.StatusConflict,
			wantCode:    "OBJECT_ALREADY_EXISTED",
			wantMessage: "object already existed",
		},
		{
			name:        "unauthorized",
			path:        "/app-unauthorized",
			status:      http.StatusUnauthorized,
			wantCode:    "UNAUTHORIZED",
			wantMessage: "unauthorized",
		},
		{
			name:        "forbidden",
			path:        "/app-forbidden",
			status:      http.StatusForbidden,
			wantCode:    "FORBIDDEN",
			wantMessage: "forbidden",
		},
		{
			name:        "internal",
			path:        "/app-internal",
			status:      http.StatusInternalServerError,
			wantCode:    "BOOM",
			wantMessage: "boom",
		},
		{
			name:        "plain error becomes internal",
			path:        "/app-plain",
			status:      http.StatusInternalServerError,
			wantCode:    "INTERNAL_ERROR",
			wantMessage: "internal server error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			r.ServeHTTP(w, req)

			require.Equal(t, tc.status, w.Code)

			detail := errorBody(t, w)
			assert.Equal(t, tc.wantCode, detail["code"])
			assert.Equal(t, tc.wantMessage, detail["message"])
			if tc.wantDetails != "" {
				b, _ := json.Marshal(detail["details"])
				assert.JSONEq(t, tc.wantDetails, string(b))
			} else {
				assert.NotContains(t, detail, "details")
			}
		})
	}
}

func Test_StatusFor(t *testing.T) {
	tests := []struct {
		kind errs.Kind
		want int
	}{
		{kind: errs.KindInternal, want: http.StatusInternalServerError},
		{kind: errs.KindNotFound, want: http.StatusNotFound},
		{kind: errs.KindValidation, want: http.StatusBadRequest},
		{kind: errs.KindConflict, want: http.StatusConflict},
		{kind: errs.KindUnauthorized, want: http.StatusUnauthorized},
		{kind: errs.KindForbidden, want: http.StatusForbidden},
	}

	for _, tc := range tests {
		t.Run(tc.kind.String(), func(t *testing.T) {
			assert.Equal(t, tc.want, ginresp.StatusFor(tc.kind))
		})
	}
}
