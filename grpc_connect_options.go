package connect

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/sirupsen/logrus"

	"google.golang.org/grpc/metadata"

	"google.golang.org/grpc"

	"github.com/imdario/mergo"
	"github.com/kumparan/go-utils"
)

// NewUnaryGRPCConnection establish a new grpc connection
func NewUnaryGRPCConnection(target string, dialOptions ...grpc.DialOption) (*grpc.ClientConn, error) {
	conn, err := grpc.Dial(target, dialOptions...)
	if err != nil {
		logrus.Errorf("Error : %v", err)
		return nil, err
	}

	return conn, err
}

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

// UnaryClientInterceptor wrapper with circuit breaker, retry, timeout and metadata logging
func UnaryClientInterceptor(opts *GRPCUnaryInterceptorOptions) grpc.UnaryClientInterceptor {
	o := applyGRPCUnaryInterceptorOptions(opts)
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx, cancel := context.WithTimeout(ctx, o.Timeout)
		defer cancel()

		ctx = metadata.AppendToOutgoingContext(ctx, "caller", utils.MyCaller(5))

		if o.UseCircuitBreaker {
			success := make(chan bool, 1)
			errC := hystrix.GoC(ctx, method, func(ctx context.Context) error {
				err := o.retryableInvoke(ctx, method, req, reply, cc, invoker, opts...)
				if err == nil {
					success <- true
				}
				return err
			}, nil)

			select {
			case out := <-success:
				logrus.Debugf("success %v", out)
				return nil
			case err := <-errC:
				logrus.Warnf("failed %s", err)
				return err
			}
		}

		return o.retryableInvoke(ctx, method, req, reply, cc, invoker, opts...)
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

func (o *GRPCUnaryInterceptorOptions) retryableInvoke(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	return utils.Retry(o.RetryCount, o.RetryInterval, func() error {
		err := invoker(ctx, method, req, reply, cc, opts...)

		if status.Code(err) != codes.Unavailable { // stop retrying unless Unavailable
			return utils.NewRetryStopper(err)
		}
		return err
	})
}
