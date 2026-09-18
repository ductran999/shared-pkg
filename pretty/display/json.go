// Package display provides tiny helpers to pretty-print payloads
// for debugging (JSON responses, structs...).
package display

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// SprintJSON returns v as indented JSON.
// String and []byte inputs are treated as raw JSON: valid ones are
// re-indented, others are returned as-is.
func SprintJSON(v any) (string, error) {
	raw, err := toJSONBytes(v)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := json.Indent(&buf, raw, "", "  "); err != nil {
		// Not JSON after all: show raw content instead of failing,
		// this is a debug helper after all.
		return string(raw), nil
	}
	return buf.String(), nil
}

// FprintJSON writes indented JSON of v to w, followed by a newline.
func FprintJSON(w io.Writer, v any) error {
	s, err := SprintJSON(v)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, s)
	return err
}

// PrintJSON writes indented JSON of v to stdout.
func PrintJSON(v any) error {
	return FprintJSON(os.Stdout, v)
}

func toJSONBytes(v any) ([]byte, error) {
	switch t := v.(type) {
	case nil:
		return []byte("null"), nil
	case string:
		return []byte(t), nil
	case []byte:
		return t, nil
	default:
		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(v); err != nil {
			return nil, err
		}
		return bytes.TrimRight(buf.Bytes(), "\n"), nil
	}
}
