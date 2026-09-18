package display_test

import (
	"bytes"
	"testing"

	"github.com/ductran999/shared-pkg/pretty/display"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_SprintJSON(t *testing.T) {
	tests := []struct {
		name    string
		in      any
		want    string
		wantErr bool
	}{
		{
			name: "struct",
			in: struct {
				Name string `json:"name"`
				Age  int    `json:"age"`
			}{Name: "a", Age: 1},
			want: "{\n  \"name\": \"a\",\n  \"age\": 1\n}",
		},
		{
			name: "html chars kept raw",
			in:   map[string]string{"url": "https://x.test/?a=1&b=<2>"},
			want: "{\n  \"url\": \"https://x.test/?a=1&b=<2>\"\n}",
		},
		{
			name: "json string re-indented",
			in:   `{"x":1}`,
			want: "{\n  \"x\": 1\n}",
		},
		{
			name: "non-json string shown raw",
			in:   "hello",
			want: "hello",
		},
		{
			name: "nil",
			in:   nil,
			want: "null",
		},
		{
			name:    "unmarshalable value errors",
			in:      func() {},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := display.SprintJSON(tc.in)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func Test_FprintJSON(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, display.FprintJSON(&buf, map[string]int{"a": 1}))
	assert.Equal(t, "{\n  \"a\": 1\n}\n", buf.String())
}
