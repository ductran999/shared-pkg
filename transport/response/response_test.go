package response_test

import (
	"encoding/json"
	"testing"

	"github.com/ductran999/shared-pkg/transport/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Success(t *testing.T) {
	tests := []struct {
		name string
		got  any
		want string
	}{
		{
			name: "OK",
			got:  response.OK([]string{"a"}),
			want: `{"data":["a"],"request_id":""}`,
		},
		{
			name: "Created",
			got:  response.Created(map[string]int{"id": 7}),
			want: `{"data":{"id":7},"request_id":""}`,
		},
		{
			name: "List",
			got:  response.List([]string{"a"}, response.NewPagination(2, 10, 95)),
			want: `{"data":["a"],"meta":{"page":2,"size":10,"total":95},"request_id":""}`,
		},
		{
			name: "with request id",
			got:  response.OK("x").WithRequestID("req-123"),
			want: `{"data":"x","request_id":"req-123"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b, err := json.Marshal(tc.got)
			require.NoError(t, err)
			assert.JSONEq(t, tc.want, string(b))
		})
	}
}

func Test_Error(t *testing.T) {
	tests := []struct {
		name string
		got  response.Error
		want string
	}{
		{
			name: "BadRequest default",
			got:  response.BadRequest(),
			want: `{"error":{"code":"BAD_REQUEST","message":"Bad Request"},"request_id":""}`,
		},
		{
			name: "BadRequest custom message",
			got:  response.BadRequest("invalid id"),
			want: `{"error":{"code":"BAD_REQUEST","message":"invalid id"},"request_id":""}`,
		},
		{
			name: "NewError custom code",
			got:  response.NewError("OBJECT_ALREADY_EXISTED", "object already existed", nil),
			want: `{"error":{"code":"OBJECT_ALREADY_EXISTED","message":"object already existed"},"request_id":""}`,
		},
		{
			name: "NewError with details",
			got: response.NewError(response.CodeInvalidParam, "validation failed",
				[]response.FieldError{
					{Field: "email", Message: "invalid"},
					{Field: "age", Message: "must be positive"},
				},
			),
			want: `{"error":{"code":"INVALID_PARAM","message":"validation failed","details":[{"field":"email","message":"invalid"},{"field":"age","message":"must be positive"}]},"request_id":""}`,
		},
		{
			name: "Unauthorized",
			got:  response.Unauthorized(),
			want: `{"error":{"code":"UNAUTHORIZED","message":"Unauthorized"},"request_id":""}`,
		},
		{
			name: "Forbidden",
			got:  response.Forbidden(),
			want: `{"error":{"code":"FORBIDDEN","message":"Forbidden"},"request_id":""}`,
		},
		{
			name: "NotFound",
			got:  response.NotFound(),
			want: `{"error":{"code":"NOT_FOUND","message":"Not Found"},"request_id":""}`,
		},
		{
			name: "InternalError",
			got:  response.InternalError("db down"),
			want: `{"error":{"code":"INTERNAL_ERROR","message":"db down"},"request_id":""}`,
		},
		{
			name: "with request id",
			got:  response.BadRequest().WithRequestID("req-123"),
			want: `{"error":{"code":"BAD_REQUEST","message":"Bad Request"},"request_id":"req-123"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b, err := json.Marshal(tc.got)
			require.NoError(t, err)
			assert.JSONEq(t, tc.want, string(b))
		})
	}
}
