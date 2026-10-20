package response_test

import (
	"encoding/json"
	"testing"

	"github.com/ductran999/shared-pkg/transport/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Constructors(t *testing.T) {
	tests := []struct {
		name string
		got  response.Envelope
		want string
	}{
		{
			name: "OK",
			got:  response.OK([]string{"a"}),
			want: `{"code":200,"message":"OK","data":["a"]}`,
		},
		{
			name: "OK with meta",
			got:  response.OK(map[string]int{"id": 1}, map[string]any{"total": 10}),
			want: `{"code":200,"message":"OK","data":{"id":1},"meta":{"total":10}}`,
		},
		{
			name: "Created",
			got:  response.Created(map[string]int{"id": 7}),
			want: `{"code":201,"message":"Created","data":{"id":7}}`,
		},
		{
			name: "BadRequest default",
			got:  response.BadRequest(),
			want: `{"code":400,"message":"Bad Request"}`,
		},
		{
			name: "BadRequest custom",
			got:  response.BadRequest("invalid id"),
			want: `{"code":400,"message":"invalid id"}`,
		},
		{
			name: "Unauthorized",
			got:  response.Unauthorized(),
			want: `{"code":401,"message":"Unauthorized"}`,
		},
		{
			name: "Forbidden",
			got:  response.Forbidden(),
			want: `{"code":403,"message":"Forbidden"}`,
		},
		{
			name: "NotFound",
			got:  response.NotFound(),
			want: `{"code":404,"message":"Not Found"}`,
		},
		{
			name: "InternalError",
			got:  response.InternalError("db down"),
			want: `{"code":500,"message":"db down"}`,
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

func Test_MergeMeta(t *testing.T) {
	t.Run("none is omitted", func(t *testing.T) {
		b, _ := json.Marshal(response.OK("x"))
		assert.NotContains(t, string(b), "meta")
	})

	t.Run("several maps merged", func(t *testing.T) {
		env := response.OK("x", map[string]any{"a": 1}, map[string]any{"b": 2})
		m, ok := env.Meta.(map[string]any)
		require.True(t, ok)
		assert.Equal(t, 1, m["a"])
		assert.Equal(t, 2, m["b"])
	})
}
