package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/ductran999/shared-pkg/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewCache(t *testing.T) {
	tests := []struct {
		name    string
		config  cache.Config
		wantErr error
	}{
		{
			name:    "missing host",
			config:  cache.Config{Host: "", Port: 6379},
			wantErr: cache.ErrMissingHost,
		},
		{
			name:   "in memory",
			config: cache.Config{IsCacheOnMemory: true},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := cache.NewCache(tc.config)
			if tc.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.wantErr)
				assert.Nil(t, got)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)
			assert.NoError(t, got.Close())
		})
	}
}

func Test_RistrettoCache(t *testing.T) {
	c, err := cache.NewRistrettoCache()
	require.NoError(t, err)

	t.Cleanup(func() { _ = c.Close() })

	ctx := context.Background()

	t.Run("set and get", func(t *testing.T) {
		require.NoError(t, c.Set(ctx, "key", "value", time.Minute))

		got, err := c.Get(ctx, "key")
		require.NoError(t, err)
		assert.Equal(t, "value", got)
	})

	t.Run("set with a struct value", func(t *testing.T) {
		require.NoError(t, c.Set(ctx, "struct", map[string]int{"id": 7}, time.Minute))

		got, err := c.Get(ctx, "struct")
		require.NoError(t, err)
		assert.Equal(t, "map[id:7]", got)
	})

	t.Run("missing key", func(t *testing.T) {
		_, err := c.Get(ctx, "missing")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "key not found")
	})

	t.Run("empty value is rejected", func(t *testing.T) {
		assert.Error(t, c.Set(ctx, "empty", "", time.Minute))
	})

	t.Run("delete", func(t *testing.T) {
		require.NoError(t, c.Set(ctx, "to-delete", "value", time.Minute))
		require.NoError(t, c.Del(ctx, "to-delete"))

		_, err := c.Get(ctx, "to-delete")
		assert.Error(t, err)
	})

	t.Run("ping", func(t *testing.T) {
		assert.NoError(t, c.Ping(ctx))
	})
}

func Test_DefaultRistrettoConfig(t *testing.T) {
	got := cache.DefaultRistrettoConfig()

	assert.Equal(t, int64(1e7), got.NumCounters)
	assert.Equal(t, int64(1<<30), got.MaxCost)
	assert.Equal(t, int64(64), got.BufferItems)
}

func Test_NewRedisCache_Unreachable(t *testing.T) {
	// Port 1 is not reachable, so the constructor must fail.
	c, err := cache.NewRedisCache(cache.Config{Host: "127.0.0.1", Port: 1})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect to redis")
	assert.Nil(t, c)
}
