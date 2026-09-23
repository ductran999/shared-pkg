// Package environ provides tiny helpers to read environment variables
// with fallbacks. Standard library only, no new dependencies.
package environ

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/ductran999/shared-pkg/fspath"
	"github.com/joho/godotenv"
)

// Get returns the value of key, or fallback when unset or empty.
func Get(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}

// GetInt returns key parsed as int, or fallback when unset, empty or invalid.
func GetInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}

	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}

	return n
}

// LoadEnv loads the .env file from the project root (nearest directory
// containing go.mod), falling back to the working directory.
// It returns the error for the caller instead of exiting the process.
func LoadEnv() error {
	rootDir, err := fspath.FindRootDir()
	if err != nil {
		return godotenv.Load()
	}
	return godotenv.Load(filepath.Join(rootDir, ".env"))
}
