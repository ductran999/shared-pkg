package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// freePort reserves a port and releases it, so the server can bind it.
func freePort(t *testing.T) int {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	port := listener.Addr().(*net.TCPAddr).Port
	require.NoError(t, listener.Close())

	return port
}

func Test_StartAndStop(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/ping", func(c *gin.Context) { c.String(http.StatusOK, "pong") })

	port := freePort(t)
	srv, err := NewGinHttpServer(router, ServerConfig{Host: "127.0.0.1", Port: port})
	require.NoError(t, err)

	go func() {
		_ = srv.Start()
	}()

	url := fmt.Sprintf("http://127.0.0.1:%d/ping", port)
	require.Eventually(t, func() bool {
		resp, err := http.Get(url) //nolint:noctx
		if err != nil {
			return false
		}
		defer func() { _ = resp.Body.Close() }()

		return resp.StatusCode == http.StatusOK
	}, 5*time.Second, 50*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	require.NoError(t, srv.Stop(ctx))

	// The server must not accept connections anymore.
	_, err = http.Get(url) //nolint:noctx
	assert.Error(t, err)
}

func Test_StopWithoutStart(t *testing.T) {
	srv, err := NewGinHttpServer(gin.New(), ServerConfig{Host: "127.0.0.1", Port: freePort(t)})
	require.NoError(t, err)

	// There is no listener and no active connection, so the shutdown
	// completes right away even with an already cancelled context.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	assert.NoError(t, srv.Stop(ctx))
}

func Test_tcpPing(t *testing.T) {
	t.Run("reachable", func(t *testing.T) {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		defer func() { _ = listener.Close() }()

		srv := &httpServer{}
		assert.NoError(t, srv.tcpPing(listener.Addr().String(), time.Second))
	})

	t.Run("unreachable", func(t *testing.T) {
		srv := &httpServer{}
		assert.Error(t, srv.tcpPing("127.0.0.1:1", 500*time.Millisecond))
	})
}
