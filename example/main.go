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
	"github.com/ductran999/shared-pkg/fspath"
	"github.com/ductran999/shared-pkg/pretty/display"
	"github.com/ductran999/shared-pkg/server"
	"github.com/ductran999/shared-pkg/transport/ginresp"
	httpclient "github.com/ductran999/shared-pkg/transport/httpclient/v2"
	"github.com/ductran999/shared-pkg/transport/response"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// Todo is the demo resource.
type Todo struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

var store = []Todo{
	{ID: 1, Title: "write demo", Done: false},
	{ID: 2, Title: "commit", Done: true},
}

// upstreamClient is shared by /upstream handler to call the mock service.
var upstreamClient *httpclient.Client

func newRouter() *gin.Engine {
	gin.SetMode(gin.DebugMode)
	r := gin.New()
	r.Use(ginresp.RequestID())

	r.GET("/todos", func(c *gin.Context) {
		p := response.NewPagination(1, 10, len(store))
		ginresp.List(c, store, p)
	})

	r.GET("/todos/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			ginresp.BadRequest(c, "invalid todo id")
			return
		}
		for _, t := range store {
			if t.ID == id {
				ginresp.OK(c, t)
				return
			}
		}
		ginresp.NotFound(c, "todo not found")
	})

	r.POST("/todos", func(c *gin.Context) {
		var t Todo
		if err := json.NewDecoder(c.Request.Body).Decode(&t); err != nil {
			ginresp.BadRequest(c, "invalid body")
			return
		}
		t.ID = len(store) + 1
		store = append(store, t)
		ginresp.Created(c, t)
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
		// Pretty-print the current store to stdout for local debugging.
		if err := display.PrintJSON(store); err != nil {
			ginresp.InternalError(c, "debug failed")
			return
		}
		ginresp.OK(c, gin.H{"debug": "printed to stdout"})
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
