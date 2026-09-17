// Package httpclient provides a lightweight wrapper around net/http
// with sensible defaults, functional options and request-scoped options.
//
// Verbs (Do/Get/Post/Put/Patch/Delete/Head) return the raw *http.Response.
// The caller owns resp.Body, closes it, and decodes it with whatever format
// and library the service speaks (encoding/json, yaml, ...) — the package
// stays format-agnostic on purpose.
//
// Import path:
//
//	import httpclientv2 "github.com/ductran999/shared-pkg/transport/httpclient/v2"
package httpclient
