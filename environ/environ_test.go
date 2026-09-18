package environ_test

import (
	"testing"

	"github.com/ductran999/shared-pkg/environ"
	"github.com/stretchr/testify/assert"
)

func Test_Get(t *testing.T) {
	t.Setenv("APP_NAME", "demo")

	tests := []struct {
		name     string
		key      string
		fallback string
		want     string
	}{
		{name: "set", key: "APP_NAME", fallback: "fb", want: "demo"},
		{name: "unset", key: "APP_MISSING", fallback: "fb", want: "fb"},
		{name: "empty", key: "APP_EMPTY", fallback: "fb", want: "fb"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.name == "empty" {
				t.Setenv(tc.key, "")
			}
			assert.Equal(t, tc.want, environ.Get(tc.key, tc.fallback))
		})
	}
}

func Test_GetInt(t *testing.T) {
	t.Setenv("APP_PORT", "8080")
	t.Setenv("APP_EMPTY", "")
	t.Setenv("APP_BAD", "abc")

	tests := []struct {
		name     string
		key      string
		fallback int
		want     int
	}{
		{name: "valid", key: "APP_PORT", fallback: 80, want: 8080},
		{name: "unset", key: "APP_MISSING", fallback: 80, want: 80},
		{name: "empty", key: "APP_EMPTY", fallback: 80, want: 80},
		{name: "invalid", key: "APP_BAD", fallback: 80, want: 80},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, environ.GetInt(tc.key, tc.fallback))
		})
	}
}
