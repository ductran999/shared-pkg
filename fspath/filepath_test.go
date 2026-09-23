package fspath

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_SearchUpward(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "a", "b")
	require.NoError(t, os.MkdirAll(nested, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "marker.txt"), []byte("x"), 0o600))

	tests := []struct {
		name   string
		start  string
		target string
		want   string
		wantOK bool
	}{
		{name: "found from nested dir", start: nested, target: "marker.txt", want: root, wantOK: true},
		{name: "found from root itself", start: root, target: "marker.txt", want: root, wantOK: true},
		{name: "missing", start: nested, target: "no-such-file-xyz", wantOK: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := searchUpward(tc.start, tc.target)
			assert.Equal(t, tc.wantOK, ok)
			assert.Equal(t, tc.want, got)
		})
	}
}

func Test_FindRootDir(t *testing.T) {
	dir, err := FindRootDir()
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(dir, "go.mod"))
	assert.NoError(t, err, "returned dir should contain go.mod")
}

func Test_FindDirOfFileMissing(t *testing.T) {
	_, err := FindDirOfFile("definitely-not-exist-xyz-123")
	require.Error(t, err)
}
