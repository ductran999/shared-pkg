package middlewares

import (
	"github.com/ductran999/shared-pkg/transport/ginresp"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// PanicError is the error attached to the gin context when a handler panics.
// The Logger middleware reads it back to flag the access log line.
type PanicError struct {
	// Value is the recovered panic value.
	Value any
	// Stack holds the structured call stack at the panic site.
	Stack []stackFrame
}

// Error implements the error interface.
func (e *PanicError) Error() string {
	return "panic: " + panicMessage(e.Value)
}

// Recovery catches a panic raised by the handlers, writes the standard error
// envelope and logs the panic with its stack trace.
//
// The panic produces its own error log line, rich in detail, meant to
// investigate the bug. The access log stays lean and keeps one line per
// request, so the metrics never drift. Both lines carry the same request ID,
// so one ID leads to both.
//
// It must be registered inside the Logger middleware, otherwise a panicking
// request would lose its access log line:
//
//	r.Use(middlewares.RequestID())
//	r.Use(middlewares.Logger())
//	r.Use(middlewares.Recovery())
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			recovered := recover()
			if recovered == nil {
				return
			}

			stack := stackFrames(1)

			// The Logger reads it back to flag the access log line.
			_ = c.Error(&PanicError{Value: recovered, Stack: stack})

			log.Error().
				Str("request_id", ginresp.RequestIDFrom(c)).
				Str("method", c.Request.Method).
				Str("path", c.Request.URL.Path).
				Str("panic", panicMessage(recovered)).
				Interface("stack", stack).
				Msg("recovered from panic")

			// The response is already on its way, only the status code
			// can still be changed.
			if c.Writer.Written() {
				c.Abort()
				return
			}

			c.Abort()
			ginresp.InternalError(c, "internal server error")
		}()

		c.Next()
	}
}
