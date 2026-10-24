// Command demo showcases every shared package working together:
//
//	environ      - load .env + read config with fallbacks
//	fspath       - locate project root (go.mod)
//	pretty/display - pretty-print JSON for debug
//	server       - gin HTTP server with graceful shutdown
//	transport/httpclient/v2 - raw *http.Response client to call an upstream
//	transport/response      - Success/Error envelopes
//	transport/ginresp       - gin helpers writing envelopes + request ID
package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/ductran999/shared-pkg/environ"
	"github.com/ductran999/shared-pkg/errs"
	"github.com/ductran999/shared-pkg/fspath"
	"github.com/ductran999/shared-pkg/pretty/display"
	"github.com/ductran999/shared-pkg/server"
	"github.com/ductran999/shared-pkg/transport/ginresp"
	httpclient "github.com/ductran999/shared-pkg/transport/httpclient/v2"
	"github.com/ductran999/shared-pkg/transport/middlewares"
	"github.com/ductran999/shared-pkg/transport/response"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// upstreamClient is shared by /upstream handler to call the mock service.
var upstreamClient *httpclient.Client

// todos is the use case layer, wired in main.
var todos *todoService

func newRouter() *gin.Engine {
	gin.SetMode(gin.DebugMode)
	r := gin.New()
	r.Use(middlewares.RequestID())
	r.Use(middlewares.Logger())
	r.Use(middlewares.Recovery())

	r.GET("/todos", func(c *gin.Context) {
		items := todos.List(c.Request.Context())
		p := response.NewPagination(1, 10, len(items))
		ginresp.List(c, items, p)
	})

	r.GET("/todos/:id", func(c *gin.Context) {
		// Parsing an HTTP path param belongs to the adapter layer.
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			ginresp.WriteError(c, errs.Validation("INVALID_TODO_ID", "invalid todo id"))
			return
		}

		// The use case returns a classified error, the adapter only maps it.
		todo, err := todos.Get(c.Request.Context(), id)
		if err != nil {
			ginresp.WriteError(c, err)
			return
		}
		ginresp.OK(c, todo)
	})

	r.POST("/todos", func(c *gin.Context) {
		var t Todo
		if err := json.NewDecoder(c.Request.Body).Decode(&t); err != nil {
			ginresp.WriteError(c, errs.Validation("INVALID_BODY", "invalid body"))
			return
		}

		todo, err := todos.Create(c.Request.Context(), t)
		if err != nil {
			ginresp.WriteError(c, err)
			return
		}
		ginresp.Created(c, todo)
	})

	r.GET("/upstream", func(c *gin.Context) {
		// Call the mock upstream service with the raw client; caller decodes JSON.
		resp, err := upstreamClient.Get(c, "/hello",
			httpclient.WithBearerToken("demo-token"),
		)
		if err != nil {
			ginresp.InternalError(c, "upstream unreachable: "+err.Error())
			return
		}
		defer func() { _ = resp.Body.Close() }()

		var payload map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
			ginresp.InternalError(c, "upstream payload invalid: "+err.Error())
			return
		}
		ginresp.OK(c, payload)
	})

	r.GET("/debug", func(c *gin.Context) {
		// Pretty-print the current todos to stdout for local debugging.
		if err := display.PrintJSON(todos.List(c.Request.Context())); err != nil {
			ginresp.InternalError(c, "debug failed")
			return
		}
		ginresp.OK(c, gin.H{"debug": "printed to stdout"})
	})

	r.GET("/panic", func(_ *gin.Context) {
		// The recovery middleware turns the panic into a standard error
		// envelope, so the client never gets a broken response.
		panic("panic demo")
	})

	return r
}

func main() {
	// Load configuration from .env at project root.
	if err := environ.LoadEnv(); err != nil {
		log.Warn().Err(err).Msg("load .env")
	}
	port := environ.GetInt("APP_PORT", 9090)
	root, err := fspath.FindRootDir()
	if err != nil {
		log.Fatal().Err(err).Msg("find project root")
	}
	log.Info().Str("root", root).Int("port", port).Msg("demo starting")

	// Wire the layers: the use case only knows the repository port.
	todos = newTodoService(&inMemoryTodoRepository{todos: []Todo{
		{ID: 1, Title: "write demo", Done: false},
		{ID: 2, Title: "commit", Done: true},
	}})

	// Mock upstream service that /upstream calls via httpclient v2.
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"message":"hello from upstream","env":"demo"}`))
	}))
	defer upstream.Close()

	upstreamClient = httpclient.New(
		httpclient.WithBaseURL(upstream.URL),
		httpclient.WithDefaultHeader("X-Client", "shared-pkg-demo"),
	)

	router := newRouter()
	httpServer, err := server.NewGinHttpServer(router, server.ServerConfig{
		Host: "localhost",
		Port: port,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("init http server")
	}

	appCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := httpServer.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("start http server")
		}
	}()

	<-appCtx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Stop(ctx); err != nil {
		log.Error().Err(err).Msg("stop http server")
	} else {
		log.Info().Msg("http server stopped")
	}

	upstream.Close()
	upstreamClient.CloseIdleConnections()
}
