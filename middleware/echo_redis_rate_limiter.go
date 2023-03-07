package middleware

import (
	"net/http"
	"regexp"

	"github.com/kumparan/go-utils"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"

	"github.com/ulule/limiter/v3"
	redisStore "github.com/ulule/limiter/v3/drivers/store/redis"
)

var privateIPAddressRegex = regexp.MustCompile(`(10(?:\\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}$)|(192\\.168(?:\\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){2}$)|(172\\.(?:1[6-9]|2\\d|3[0-1])(?:\\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){2}$)`)

// RedisIPRateLimiter is the redis store that implements IP Based rate limiter
type RedisIPRateLimiter struct {
	ipLimiter   *limiter.Limiter
	excludedIPs []string
}

// NewRedisIPRateLimiter initializes RedisIPRateLimiter
func NewRedisIPRateLimiter(redisClient *redis.Client, rate limiter.Rate, excludedIPs []string) (redisLimiter RedisIPRateLimiter, err error) {
	store, err := redisStore.NewStoreWithOptions(redisClient, limiter.StoreOptions{
		Prefix: "rate-limiter:",
	})
	if err != nil {
		return
	}

	return RedisIPRateLimiter{
		ipLimiter:   limiter.New(store, rate),
		excludedIPs: excludedIPs,
	}, nil
}

// Limit limit request by IP
func (r RedisIPRateLimiter) Limit() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			ip := c.RealIP()
			if privateIPAddressRegex.MatchString(ip) || utils.Contains[string](r.excludedIPs, ip) {
				return next(c)
			}
			limiterCtx, err := r.ipLimiter.Get(c.Request().Context(), ip)
			if err != nil {
				log.Printf("IPRateLimit - ipRateLimiter.Get - err: %v, %s on %s", err, ip, c.Request().URL)
				return c.JSON(http.StatusInternalServerError, echo.Map{
					"success": false,
					"message": err,
				})
			}

			if limiterCtx.Reached {
				log.Printf("Too Many Requests from %s on %s", ip, c.Request().URL)
				return c.JSON(http.StatusTooManyRequests, echo.Map{
					"success": false,
					"message": "Too Many Requests on " + c.Request().URL.String(),
				})
			}

			return next(c)
		}
	}

}
