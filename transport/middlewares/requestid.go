// Package middlewares provides reusable gin middlewares.
//
// Example:
//
//	r.Use(middlewares.RequestID())
package middlewares

import (
	"github.com/ductran999/shared-pkg/transport/ginresp"
	httpclient "github.com/ductran999/shared-pkg/transport/httpclient/v2"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestIDHeader is the header carrying the request ID.
const RequestIDHeader = "X-Request-ID"

// maxRequestIDLength bounds the inbound request ID we accept, so a caller
// cannot flood the logs with an unbounded value.
const maxRequestIDLength = 128

// RequestID keeps the request ID sent by the caller when there is one, so the
// correlation survives across services, and generates a new UUIDv4 otherwise.
//
// The ID is stored in the gin context and in the request context, then echoed
// back in the X-Request-ID response header.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := validRequestID(c.GetHeader(RequestIDHeader))
		if requestID == "" {
			requestID = newRequestID()
		}

		c.Set(ginresp.RequestIDKey, requestID)
		c.Writer.Header().Set(RequestIDHeader, requestID)

		// The request context carries the ID too, so handlers using
		// c.Request.Context() can propagate it to downstream services.
		c.Request = c.Request.WithContext(httpclient.WithRequestID(c.Request.Context(), requestID))

		c.Next()
	}
}

// newRequestID returns a new request ID.
//
// A random UUIDv4 is used on purpose: a request ID is a correlation ID, its
// job is uniqueness and tracing, not ordering. It carries 122 random bits,
// stays unique across services and does not leak when the request happened
// nor depends on the system clock.
//
// UUIDv7 would only pay off when the ID doubles as a database primary key.
func newRequestID() string {
	return uuid.NewString()
}

// validRequestID returns the inbound request ID when it is safe to reuse it,
// or an empty string when it must be replaced by a generated one.
//
// Only the two things that can actually cause harm are rejected: a value that
// is too long, which would flood the logs, and a value holding control
// characters, which could break a log line. Everything else is kept as it is,
// because the request ID is an opaque correlation token, not necessarily a
// UUID: an upstream service may send a W3C trace id, a hex id or a short
// token.
func validRequestID(value string) string {
	if value == "" || len(value) > maxRequestIDLength {
		return ""
	}
	for _, char := range value {
		if char < ' ' || char == 0x7f {
			return ""
		}
	}

	return value
}
