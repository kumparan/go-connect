package connect

import (
	"errors"
	"time"

	"github.com/imdario/mergo"

	goredis "github.com/go-redis/redis"
	redigo "github.com/gomodule/redigo/redis"
)

// RedisConnectionPoolOptions options for the redis connection
type RedisConnectionPoolOptions struct {
	// Dial timeout for establishing new connections.
	// Default is 5 seconds. Only for go-redis.
	DialTimeout time.Duration

	// Timeout for socket reads. If reached, commands will fail
	// with a timeout instead of blocking. Use value -1 for no timeout and 0 for default.
	// Default is 3 seconds. Only for go-redis.
	ReadTimeout time.Duration

	// Timeout for socket writes. If reached, commands will fail
	// with a timeout instead of blocking.
	// Default is ReadTimeout. Only for go-redis.
	WriteTimeout time.Duration

	// Number of idle connections in the pool.
	IdleCount int

	// Maximum number of connections allocated by the pool at a given time.
	// When zero, there is no limit on the number of connections in the pool.
	PoolSize int

	// Close connections after remaining idle for this duration. If the value
	// is zero, then idle connections are not closed. Applications should set
	// the timeout to a value less than the server's timeout.
	IdleTimeout time.Duration

	// Close connections older than this duration. If the value is zero, then
	// the pool does not close connections based on age.
	MaxConnLifetime time.Duration

	// Enables read-only commands on slave nodes.
	ReadOnly bool

	// Allows routing read-only commands to the closest master or slave node.
	// It automatically enables ReadOnly.
	RouteByLatency bool

	// Allows routing read-only commands to the random master or slave node.
	// It automatically enables ReadOnly.
	RouteRandomly bool
}

var defaultRedisConnectionPoolOptions = &RedisConnectionPoolOptions{
	IdleCount:       20,
	PoolSize:        100,
	IdleTimeout:     60 * time.Second,
	MaxConnLifetime: 0,
	DialTimeout:     5 * time.Second,
	WriteTimeout:    2 * time.Second,
	ReadTimeout:     2 * time.Second,
	ReadOnly:        false,
	RouteRandomly:   false,
	RouteByLatency:  false,
}

// NewRedigoRedisConnectionPool uses redigo library to establish the redis connection pool
func NewRedigoRedisConnectionPool(url string, opt *RedisConnectionPoolOptions) (*redigo.Pool, error) {
	if !isValidRedisStandaloneURL(url) {
		return nil, errors.New("invalid redis URL: " + url)
	}

	options := applyRedisConnectionPoolOptions(opt)

	return &redigo.Pool{
		MaxIdle:     options.IdleCount,
		MaxActive:   options.PoolSize,
		IdleTimeout: options.IdleTimeout,
		Dial: func() (redigo.Conn, error) {
			c, err := redigo.DialURL(url)
			if err != nil {
				return nil, err
			}
			return c, err
		},
		MaxConnLifetime: options.MaxConnLifetime,
		TestOnBorrow: func(c redigo.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}, nil
}

// NewGoRedisConnectionPool uses goredis library to establish the redis connection pool
func NewGoRedisConnectionPool(url string, opt *RedisConnectionPoolOptions) (*goredis.Client, error) {
	options, err := goredis.ParseURL(url)
	if err != nil {
		return nil, err
	}

	myOptions := applyRedisConnectionPoolOptions(opt)
	options.MinIdleConns = myOptions.IdleCount
	options.PoolSize = myOptions.PoolSize
	options.IdleTimeout = myOptions.IdleTimeout
	options.MaxConnAge = myOptions.MaxConnLifetime
	options.DialTimeout = myOptions.DialTimeout
	options.WriteTimeout = myOptions.WriteTimeout
	options.ReadTimeout = myOptions.ReadTimeout

	return goredis.NewClient(options), nil
}

// NewGoRedisClusterConnectionPool uses goredis library to establish the redis cluster connection pool
func NewGoRedisClusterConnectionPool(urls []string, opt *RedisConnectionPoolOptions) (*goredis.ClusterClient, error) {
	for _, url := range urls {
		if isValidRedisStandaloneURL(url) {
			return nil, errors.New("invalid redis-cluster URL: " + url)
		}
	}
	options := applyRedisConnectionPoolOptions(opt)

	return goredis.NewClusterClient(&goredis.ClusterOptions{
		Addrs:        urls,
		IdleTimeout:  options.IdleTimeout,
		MinIdleConns: options.IdleCount,
		MaxConnAge:   options.MaxConnLifetime,
		PoolSize:     options.PoolSize,
		DialTimeout:  options.DialTimeout,
		WriteTimeout: options.WriteTimeout,
		ReadTimeout:  options.ReadTimeout,
		OnConnect: func(conn *goredis.Conn) error {
			return conn.Ping().Err()
		},
		ReadOnly:       options.ReadOnly,
		RouteByLatency: options.RouteByLatency,
		RouteRandomly:  options.RouteRandomly,
	}), nil
}

func applyRedisConnectionPoolOptions(opt *RedisConnectionPoolOptions) *RedisConnectionPoolOptions {
	if opt == nil {
		return defaultRedisConnectionPoolOptions
	}
	// if error occurs, also return options from input
	_ = mergo.Merge(opt, *defaultRedisConnectionPoolOptions)
	return opt
}

func isValidRedisStandaloneURL(url string) bool {
	_, err := goredis.ParseURL(url)
	return err == nil
}
