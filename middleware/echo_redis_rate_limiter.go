package middleware

import (
	"log"
	"net/http"
	"strconv"

	"github.com/ulule/limiter/v3"
	redisStore "github.com/ulule/limiter/v3/drivers/store/redis"

	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
)

// RedisIPRateLimiter is the redis store that implements IP Based rate limiter
type RedisIPRateLimiter struct {
	ipLimiter *limiter.Limiter
}

// NewRedisIPRateLimiter initializes RedisIPRateLimiter
func NewRedisIPRateLimiter(redisClient *redis.Client, rate limiter.Rate) (redisLimiter RedisIPRateLimiter, err error) {
	store, err := redisStore.NewStoreWithOptions(redisClient, limiter.StoreOptions{
		Prefix: "rate-limiter:",
	})
	if err != nil {
		return
	}

	return RedisIPRateLimiter{
		ipLimiter: limiter.New(store, rate),
	}, nil
}

func (r RedisIPRateLimiter) Limit() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			ip := c.RealIP()
			limiterCtx, err := r.ipLimiter.Get(c.Request().Context(), ip)
			if err != nil {
				log.Printf("IPRateLimit - ipRateLimiter.Get - err: %v, %s on %s", err, ip, c.Request().URL)
				return c.JSON(http.StatusInternalServerError, echo.Map{
					"success": false,
					"message": err,
				})
			}

			h := c.Response().Header()
			h.Set("X-RateLimit-Limit", strconv.FormatInt(limiterCtx.Limit, 10))
			h.Set("X-RateLimit-Remaining", strconv.FormatInt(limiterCtx.Remaining, 10))
			h.Set("X-RateLimit-Reset", strconv.FormatInt(limiterCtx.Reset, 10))

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
