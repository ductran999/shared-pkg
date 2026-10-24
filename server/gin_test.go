package server

import (
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_setupConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  ServerConfig
		wantErr error
		want    *ServerConfig
	}{
		{
			name:    "port zero",
			config:  ServerConfig{Host: "localhost", Port: 0},
			wantErr: ErrInvalidPort,
		},
		{
			name:    "port negative",
			config:  ServerConfig{Host: "localhost", Port: -1},
			wantErr: ErrInvalidPort,
		},
		{
			name:    "port too large",
			config:  ServerConfig{Host: "localhost", Port: 65536},
			wantErr: ErrInvalidPort,
		},
		{
			name:   "defaults",
			config: ServerConfig{Host: "localhost", Port: 8080},
			want: &ServerConfig{
				Host:              "localhost",
				Port:              8080,
				ReadHeaderTimeout: 500 * time.Millisecond,
				ReadTimeout:       10 * time.Second,
				WriteTimeout:      10 * time.Second,
				IdleTimeout:       60 * time.Second,
			},
		},
		{
			name: "custom values are kept",
			config: ServerConfig{
				Host:              "localhost",
				Port:              9090,
				ReadHeaderTimeout: time.Second,
				ReadTimeout:       2 * time.Second,
				WriteTimeout:      3 * time.Second,
				IdleTimeout:       4 * time.Second,
			},
			want: &ServerConfig{
				Host:              "localhost",
				Port:              9090,
				ReadHeaderTimeout: time.Second,
				ReadTimeout:       2 * time.Second,
				WriteTimeout:      3 * time.Second,
				IdleTimeout:       4 * time.Second,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := setupConfig(tc.config)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.wantErr)
				assert.Nil(t, got)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func Test_NewGinHttpServer(t *testing.T) {
	router := gin.New()

	srv, err := NewGinHttpServer(router, ServerConfig{Host: "localhost", Port: 8080})
	require.NoError(t, err)
	require.NotNil(t, srv)

	assert.Equal(t, "localhost:8080", srv.server.Addr)
	assert.NotNil(t, srv.server.Handler)
	assert.Equal(t, 500*time.Millisecond, srv.server.ReadHeaderTimeout)
	assert.Equal(t, 10*time.Second, srv.server.ReadTimeout)
	assert.Equal(t, 10*time.Second, srv.server.WriteTimeout)
	assert.Equal(t, 60*time.Second, srv.server.IdleTimeout)
}

func Test_NewGinHttpServer_InvalidPort(t *testing.T) {
	srv, err := NewGinHttpServer(gin.New(), ServerConfig{Host: "localhost", Port: 0})

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidPort)
	assert.Nil(t, srv)
}
