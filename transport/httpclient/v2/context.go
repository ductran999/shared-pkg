package httpclient

import "context"

// DefaultRequestIDHeader is the header used to propagate request correlation
// across distributed services.
const DefaultRequestIDHeader = "X-Request-ID"

// requestIDKey is the private context key for request IDs set via WithRequestID.
// Private to avoid collisions; RequestIDFromContext also falls back to common
// string keys for interop with other libs/middlewares.
type requestIDKey struct{}

// WithRequestID returns a context carrying the given request/trace ID.
// Use it on the server side (after generating or extracting the inbound ID)
// so downstream Client.Do calls can propagate it.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, requestIDKey{}, requestID)
}

// RequestIDFromContext extracts a request/trace ID from ctx.
// Lookup order:
//  1. typed key set by WithRequestID (preferred, collision-free)
//  2. common string keys for interop: request_id, request-id,
//     trace_id, trace-id, x-request-id, X-Request-ID
//
// Returns ("", false) when nothing is found.
func RequestIDFromContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}

	if v, ok := ctx.Value(requestIDKey{}).(string); ok && v != "" {
		return v, true
	}

	for _, k := range []string{
		"request_id", "request-id",
		"trace_id", "trace-id",
		"x-request-id", "X-Request-ID", "X-Request-Id",
	} {
		if v, ok := ctx.Value(k).(string); ok && v != "" {
			return v, true
		}
		if v, ok := ctx.Value(k).([]byte); ok && len(v) > 0 {
			return string(v), true
		}
		if s, ok := ctx.Value(k).(interface{ String() string }); ok && s != nil {
			if str := s.String(); str != "" {
				return str, true
			}
		}
	}

	return "", false
}
