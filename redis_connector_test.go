package connect

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_applyRedisConnectionPoolOptions(t *testing.T) {
	t.Run("option is nil", func(t *testing.T) {
		option := applyRedisConnectionPoolOptions(nil)

		// should return default values
		assert.EqualValues(t, defaultRedisConnectionPoolOptions.PoolSize, option.PoolSize)
		assert.EqualValues(t, defaultRedisConnectionPoolOptions.IdleCount, option.IdleCount)
		assert.EqualValues(t, defaultRedisConnectionPoolOptions.IdleTimeout, option.IdleTimeout)
		assert.EqualValues(t, defaultRedisConnectionPoolOptions.ReadTimeout, option.ReadTimeout)
		assert.EqualValues(t, defaultRedisConnectionPoolOptions.WriteTimeout, option.WriteTimeout)
		assert.EqualValues(t, defaultRedisConnectionPoolOptions.DialTimeout, option.DialTimeout)
		assert.EqualValues(t, defaultRedisConnectionPoolOptions.MaxConnLifetime, option.MaxConnLifetime)
	})

	t.Run("merge struct", func(t *testing.T) {
		option := applyRedisConnectionPoolOptions(&RedisConnectionPoolOptions{
			WriteTimeout: 50,
			ReadTimeout:  3,
		})

		// all should set to default values except for WriteTimeout and ReadTimeout
		assert.EqualValues(t, defaultRedisConnectionPoolOptions.PoolSize, option.PoolSize)
		assert.EqualValues(t, defaultRedisConnectionPoolOptions.IdleCount, option.IdleCount)
		assert.EqualValues(t, defaultRedisConnectionPoolOptions.IdleTimeout, option.IdleTimeout)
		assert.EqualValues(t, 3, option.ReadTimeout)
		assert.EqualValues(t, 50, option.WriteTimeout)
		assert.EqualValues(t, defaultRedisConnectionPoolOptions.DialTimeout, option.DialTimeout)
		assert.EqualValues(t, defaultRedisConnectionPoolOptions.MaxConnLifetime, option.MaxConnLifetime)
	})
}
