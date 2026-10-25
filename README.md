# shared-pkg

Shared Go utilities used across services: HTTP clients, config helpers,
retry logic, caching, display helpers and more. Standard library first,
small API surface, no surprises.

## Coverage
[![codecov](https://codecov.io/gh/ductran999/shared-pkg/branch/master/graph/badge.svg)](https://codecov.io/gh/ductran999/shared-pkg/75%25)

## Requirements

- Go 1.23+

## Packages

| Package | Description |
|---|---|
| `transport/httpclient/v2` | HTTP client with sensible defaults, functional options, base URL, request-ID propagation and sentinel errors. Returns raw `*http.Response`. |
| `transport/response` | Standard API envelope with constructors (`OK`, `Created`, `BadRequest`, ...). Re-exports `errs.Code` and `errs.FieldError`. |
| `errs` | Application error classification (`Kind`, `AppError`, `NotFound`, `Validation`, ...) and the shared `Code` vocabulary. |
| `environ` | Read env vars with fallbacks (`Get`, `GetInt`) and load `.env` from the project root (`LoadEnv`). |
| `fspath` | Locate directories by searching upwards (`FindDirOfFile`, `FindRootDir`). |
| `pretty/display` | Pretty-print payloads for debugging (`PrintJSON`, `SprintJSON`). |
| `transport/ginresp` | Gin helpers that write envelopes (`OK`, `Created`, `List`, `Error`, ...) + `WriteError` mapping `errs` to HTTP. |
| `transport/middlewares` | Reusable gin middlewares (`RequestID` generates a UUIDv4 and echoes it in `X-Request-ID`, `Logger` writes a structured access log, `Recovery` turns a panic into the error envelope). |
| `cache` | Redis and in-memory cache. |
| `server` | Gin HTTP server with sane timeouts. |
| `client/httpclient` | **Deprecated** legacy v1 HTTP client. Use `transport/httpclient/v2` for new code. |
| `alertmanager`, `scrypto` | Alerting and simple crypto helpers. |

## Quick start

```go
import (
	httpclient "github.com/ductran999/shared-pkg/transport/httpclient/v2"
	"github.com/ductran999/shared-pkg/environ"
	"github.com/ductran999/shared-pkg/pretty/display"
)

// HTTP client: verbs return the raw *http.Response, you own the body.
client := httpclient.New(
	httpclient.WithBaseURL("https://api.example.com"),
	httpclient.WithTimeout(5*time.Second),
)

resp, err := client.Get(ctx, "/todos")
if err != nil {
	return err
}
defer resp.Body.Close()

// Env with fallback.
port := environ.GetInt("PORT", 8080)

// Pretty-print anything for debugging.
_ = display.PrintJSON(map[string]int{"port": port})
```

## Development

```bash
make check   # tidy + vet + lint (stops on first failure)
make tidy    # go mod tidy
make vet     # go vet ./...
make lint    # golangci-lint
```

## Request ID

`transport/middlewares` holds the cross-cutting gin middlewares. The request
ID middleware reuses the ID sent by the caller when there is one, so the
correlation survives across services, and generates a new UUIDv4 otherwise.
The ID is stored in the gin context and in the request context, then echoed
back in the `X-Request-ID` header.

```go
r.Use(middlewares.RequestID())
r.GET("/todos", func(c *gin.Context) {
	ginresp.OK(c, todos)
})
```

```bash
curl -H "X-Request-ID: trace-abc-123" localhost:9090/todos
# X-Request-ID: trace-abc-123
# {"data":[...],"request_id":"trace-abc-123"}
```

A random UUIDv4 is used on purpose: a request ID is a correlation ID, so it
only has to be unique and traceable. It carries 122 random bits, does not
leak when the request happened and does not depend on the system clock.
UUIDv7 would only pay off when the ID doubles as a database primary key.

Because the ID also lives in the request context, the HTTP client propagates
it to downstream services automatically:

```go
resp, err := client.Get(c.Request.Context(), "/users") // sends X-Request-ID
```

An inbound ID is reused as it is. Only the two things that can actually cause
harm are rejected: a value longer than 128 characters, which would flood the
logs, and a value holding control characters, which could break a log line.
Everything else is kept, because the request ID is an opaque correlation
token, not necessarily a UUID: a W3C trace id, an OpenTelemetry trace id, an
AWS X-Ray id or a short internal token are all accepted.

### Logger

`gin.Logger` writes plain text, which is painful to ship to a log aggregator,
and it does not log the request ID. `middlewares.Logger` writes a structured
JSON line with the request ID, so a request can be traced from the access log
to the application logs.

```go
r.Use(middlewares.RequestID())
r.Use(middlewares.Logger())
```

```json
{"level":"info","request_id":"9c5b2c03-2a68-4da2-8d66-2275ad156a28","method":"GET",
 "path":"/todos","query":"","status":200,"size":42,"latency":155212,
 "client_ip":"127.0.0.1","message":"http request"}
```

The level follows the status: `info` for 2xx and 3xx, `warn` for 4xx, `error`
for 5xx. Paths can be skipped, e.g. health checks:

```go
r.Use(middlewares.Logger("/healthz", "/metrics"))
```

If you prefer gin's own logger, it is customizable too:

```go
gin.LoggerWithConfig(gin.LoggerConfig{
	Formatter: func(p gin.LogFormatterParams) string { return myFormat(p) },
	Output:    os.Stdout,
	SkipPaths: []string{"/healthz"},
	Skip:      func(c *gin.Context) bool { return false },
})
```

### Recovery

`middlewares.Recovery` catches a panic raised by the handlers, writes the
standard error envelope and logs the panic with its stack trace.

A panicking request produces **two log lines**, each with its own purpose,
linked by the same `request_id`:

| Line | Level | Content | Purpose |
|---|---|---|---|
| `recovered from panic` | error | panic value, structured stack, method, path | investigate the bug |
| `http request` | info/warn/error | status, latency, size, `panic: true` | access log, metrics, SLO |

The access log stays lean on purpose: it must keep exactly one line per
request, so the error rate and the latency stay accurate. The stack trace is
heavy and lives in its own line.

**The order matters.** The Recovery must be registered *inside* the Logger,
otherwise a panicking request would lose its access log line and the metrics
would drift:

```go
r.Use(middlewares.RequestID())   // 1. generate or take the request ID
r.Use(middlewares.Logger())       // 2. write the access log after c.Next()
r.Use(middlewares.Recovery())     // 3. recover, set the status, log the panic
// panic -> 500 {"error":{"code":"INTERNAL_ERROR","message":"internal server error"},"request_id":"..."}
```

```json
{"level":"error","request_id":"9b5fbc70...","method":"GET","path":"/panic",
 "panic":"panic demo","stack":[{"function":"runtime.gopanic","file":"...","line":783}],
 "message":"recovered from panic"}
{"level":"error","request_id":"9b5fbc70...","method":"GET","path":"/panic","status":500,
 "latency":322115,"panic":true,"message":"http request"}
```

The stack is a list of frames, not a blob of text, so it stays greppable and
each frame carries its own file and line.

## Error handling

`errs` classifies failures in business terms; the transport layer maps the
kind to an HTTP status. The use case never mentions HTTP.

`errs` owns the `Code` type and the common codes, and `transport/response`
re-exports them, so a code means the same thing in every layer:

```go
// use case / domain layer
err := errs.NotFound(errs.Code("USER_NOT_FOUND"), "user not found")
// or with your own service code
const CodeUserNotFound errs.Code = "USER_NOT_FOUND"

// gin handler (interface adapter)
if err != nil {
    ginresp.WriteError(c, err)
    return
}
// 404 {"error":{"code":"USER_NOT_FOUND","message":"user not found"},"request_id":"..."}
```

Field level details use `errs.FieldError`, so the domain layer never imports
the transport package:

```go
err := errs.Validation("TITLE_REQUIRED", "title is required",
    errs.FieldError{Field: "title", Message: "is required"},
)
// 400 {"error":{"code":"TITLE_REQUIRED","message":"title is required",
//      "details":[{"field":"title","message":"is required"}]},"request_id":"..."}
```

### Envelopes

Two shapes only, both carrying the request ID:

```jsonc
// success
{"data": {...}, "meta": {...}, "request_id": "..."}   // meta is optional

// error
{"error": {"code": "...", "message": "...", "details": [...]}, "request_id": "..."}
```

`details` is always a list of `errs.FieldError` when present, and the key is
omitted entirely when there is none. It is never an object, so clients can
parse it the same way for every error:

```jsonc
{"error": {"code": "TODO_NOT_FOUND", "message": "todo not found"}, "request_id": "..."}
{"error": {"code": "TITLE_REQUIRED", "message": "title is required",
           "details": [{"field": "title", "message": "is required"}]}, "request_id": "..."}
```

| Kind | Status |
|---|---|
| `KindValidation` | 400 |
| `KindUnauthorized` | 401 |
| `KindForbidden` | 403 |
| `KindNotFound` | 404 |
| `KindConflict` | 409 |
| `KindInternal` | 500 |

Errors that are not `*errs.AppError` are reported as `INTERNAL_ERROR` so
unexpected failures never leak to the client. The root cause stays in `Cause`
for logging only.

## Example

`example/` is a demo server that wires every shared package together
(`environ`, `fspath`, `pretty/display`, `server`, `transport/httpclient/v2`,
`transport/response`, `transport/ginresp`, `transport/middlewares`, `errs`).

```bash
go run ./example            # listens on :9090 (override with APP_PORT)
curl http://localhost:9090/todos
```

It exposes `GET/POST /todos`, `GET /todos/:id`, `GET /upstream` (calls a mock
service through `httpclient v2`) and `GET /debug` (pretty-prints the todos).

The example is layered the way a service should be:

| File | Layer | Knows about |
|---|---|---|
| `example/todo.go` | domain + use case | business rules, `errs`, no HTTP |
| `example/main.go` | interface adapter | gin, envelopes, HTTP status |

The use case returns `*errs.AppError`; the handler only maps it to a status
and an envelope, so the same use case can be reused by a CLI or a gRPC
service without touching the transport layer.
