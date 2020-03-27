package connect

import (
	"context"
	"log"
	"time"

	"github.com/afex/hystrix-go/hystrix"

	"google.golang.org/grpc/metadata"

	"google.golang.org/grpc"

	"github.com/imdario/mergo"
	"github.com/kumparan/go-utils"
)

// GRPCUnaryInterceptorOptions wrapper options for the grpc connection
type GRPCUnaryInterceptorOptions struct {
	// UseCircuitBreaker flag if the connection will implement a circuit breaker
	UseCircuitBreaker bool

	// RetryCount retry the operation if found error.
	// When set to <= 1, then it means no retry
	RetryCount int

	// RetryInterval next interval for retry.
	RetryInterval time.Duration

	// Timeout value, will return context deadline exceeded when the operation exceeds the duration
	Timeout time.Duration
}

var defaultGRPCUnaryInterceptorOptions = &GRPCUnaryInterceptorOptions{
	UseCircuitBreaker: false,
	RetryCount:        0,
	RetryInterval:     20 * time.Millisecond,
	Timeout:           1 * time.Second,
}

func UnaryClientInterceptor(opts *GRPCUnaryInterceptorOptions) grpc.UnaryClientInterceptor {
	o := applyGRPCUnaryInterceptorOptions(opts)
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx, cancel := context.WithTimeout(ctx, o.Timeout)
		defer cancel()
		ctx = metadata.AppendToOutgoingContext(ctx, "caller", utils.MyCaller(5))

		if o.UseCircuitBreaker {
			success := make(chan bool, 1)
			errC := hystrix.GoC(ctx, method, func(ctx context.Context) error {
				err := utils.Retry(o.RetryCount, o.RetryInterval, func() error {
					return invoker(ctx, method, req, reply, cc, opts...)
				})

				if err == nil {
					success <- true
				}
				return err
			}, nil)

			select {
			case out := <-success:
				log.Printf("success %v", out)
			case err := <-errC:
				log.Printf("failed %s", err)
				return err
			}
		}

		return utils.Retry(o.RetryCount, o.RetryInterval, func() error {
			return invoker(ctx, method, req, reply, cc, opts...)
		})
	}
}

func applyGRPCUnaryInterceptorOptions(opts *GRPCUnaryInterceptorOptions) *GRPCUnaryInterceptorOptions {
	if opts == nil {
		return defaultGRPCUnaryInterceptorOptions
	}
	// if error occurs, also return options from input
	_ = mergo.Merge(opts, *defaultGRPCUnaryInterceptorOptions)
	return opts
}
