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
| `environ` | Read env vars with fallbacks (`Get`, `GetInt`) and load `.env` from the project root (`LoadEnv`). |
| `fspath` | Locate directories by searching upwards (`FindDirOfFile`, `FindRootDir`). |
| `pretty/display` | Pretty-print payloads for debugging (`PrintJSON`, `SprintJSON`). |
| `cache` | Redis and in-memory cache. |
| `server` | Gin HTTP server with sane timeouts. |
| `client/httpclient` | Legacy v1 HTTP client (frozen, use `transport/httpclient/v2` for new code). |
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
