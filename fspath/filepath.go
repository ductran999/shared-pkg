// Package fspath provides tiny helpers to locate directories
// by searching upwards on the filesystem.
package fspath

import (
	"fmt"
	"os"
	"path/filepath"
)

// FindDirOfFile searches upwards from the working directory until it finds
// a directory containing the specified file or folder name.
func FindDirOfFile(name string) (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	if found, ok := searchUpward(dir, name); ok {
		return found, nil
	}
	return "", fmt.Errorf("fspath: could not find directory containing: %s", name)
}

// FindRootDir searches upwards until it finds the project root,
// defined as the nearest directory containing go.mod.
func FindRootDir() (string, error) {
	return FindDirOfFile("go.mod")
}

// searchUpward reports the nearest directory at or above start
// containing target.
func searchUpward(start, target string) (string, bool) {
	dir := start
	for {
		if _, err := os.Stat(filepath.Join(dir, target)); err == nil {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}
