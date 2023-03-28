package connect

import (
	"context"
	"net"
	"time"

	"runtime/debug"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/ulule/limiter/v3"
	redisStore "github.com/ulule/limiter/v3/drivers/store/redis"

	"google.golang.org/grpc/metadata"

	"google.golang.org/grpc"

	"github.com/imdario/mergo"
	"github.com/kumparan/go-connect/internal"
	"github.com/kumparan/go-utils"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	otelcodes "go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"
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

type messageType attribute.KeyValue

// Event adds an event of the messageType to the span associated with the
// passed context with a message id.
func (m messageType) Event(ctx context.Context, id int, message interface{}) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}
	span.AddEvent("message", trace.WithAttributes(
		attribute.KeyValue(m),
		attribute.Key("message.id").Int(id),
	))
}

var (
	messageSent     = messageType(attribute.Key("message.type").String("SENT"))
	messageReceived = messageType(attribute.Key("message.type").String("RECEIVED"))
)

// RecoveryHandlerFunc is a function that recovers from the panic `p` by returning an `error`.
// The context can be used to extract request scoped metadata and context values.
type RecoveryHandlerFunc func(ctx context.Context, p interface{}) (err error)

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

	// UseOpenTelemetry flag if the connection will implement open telemetry
	UseOpenTelemetry bool

	// UseRateLimiter flag if the connection will implement rate limiter
	UseRateLimiter bool

	// RateLimiterLimit limit the request
	RateLimiterLimit int64

	// RateLimiterPeriod limit period
	RateLimiterPeriod time.Duration

	RecoveryHandlerFunc RecoveryHandlerFunc
}

var defaultGRPCUnaryInterceptorOptions = &GRPCUnaryInterceptorOptions{
	UseCircuitBreaker: false,
	RetryCount:        0,
	RetryInterval:     20 * time.Millisecond,
	Timeout:           1 * time.Second,
	UseOpenTelemetry:  false,
	UseRateLimiter:    false,
	RateLimiterLimit:  100,
	RateLimiterPeriod: 1 * time.Second,
	RecoveryHandlerFunc: func(ctx context.Context, p interface{}) (err error) {
		logrus.WithFields(logrus.Fields{
			"ctx":        utils.DumpIncomingContext(ctx),
			"stackTrace": string(debug.Stack()),
		}).Errorf("panic recovered: %v", p)
		return status.Error(codes.Internal, "internal server error")
	},
}

// UnaryClientInterceptor wrapper with circuit breaker, retry, timeout, open telemetry, and metadata logging
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

// spanInfo returns a span name and all appropriate attributes from the gRPC
// method and peer address.
func spanInfo(fullMethod, peerAddress string) (string, []attribute.KeyValue) {
	attrs := []attribute.KeyValue{semconv.RPCSystemKey.String("grpc")}
	name, mAttrs := internal.ParseFullMethod(fullMethod)
	attrs = append(attrs, mAttrs...)
	attrs = append(attrs, peerAttr(peerAddress)...)
	return name, attrs
}

// peerAttr returns attributes about the peer address.
func peerAttr(addr string) []attribute.KeyValue {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return []attribute.KeyValue(nil)
	}

	if host == "" {
		host = "127.0.0.1"
	}

	return []attribute.KeyValue{
		semconv.NetPeerIPKey.String(host),
		semconv.NetPeerPortKey.String(port),
	}
}

func (o *GRPCUnaryInterceptorOptions) retryableInvoke(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	return utils.Retry(o.RetryCount, o.RetryInterval, func() (err error) {
		if !o.UseOpenTelemetry {
			err = invoker(ctx, method, req, reply, cc, opts...)

			if status.Code(err) != codes.Unavailable { // stop retrying unless Unavailable
				return utils.NewRetryStopper(err)
			}

			return err
		}

		requestMetadata, _ := metadata.FromOutgoingContext(ctx)
		metadataCopy := requestMetadata.Copy()
		tracer := newConfig().TracerProvider.Tracer(
			instrumentationName,
		)

		name, attr := spanInfo(method, cc.Target())

		var span trace.Span
		ctx, span = tracer.Start(
			ctx,
			name,
			trace.WithSpanKind(trace.SpanKindClient),
			trace.WithAttributes(attr...),
		)
		defer span.End()

		inject(ctx, &metadataCopy)
		ctx = metadata.NewOutgoingContext(ctx, metadataCopy)

		messageSent.Event(ctx, defaultMessageID, req)

		err = invoker(ctx, method, req, reply, cc, opts...)

		messageReceived.Event(ctx, defaultMessageID, reply)

		switch {
		case span == nil:
			logrus.WithFields(logrus.Fields{
				"context": utils.DumpIncomingContext(ctx),
			}).Error("span is nil")
		case err != nil:
			s, _ := status.FromError(err)
			span.SetStatus(otelcodes.Error, s.Message())
			span.SetAttributes(statusCodeAttr(s.Code()))
		default:
			span.SetAttributes(statusCodeAttr(codes.OK))
		}

		if status.Code(err) != codes.Unavailable { // stop retrying unless Unavailable
			return utils.NewRetryStopper(err)
		}

		return err
	})
}

// peerFromCtx returns a peer address from a context, if one exists.
func peerFromCtx(ctx context.Context) string {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return ""
	}
	return p.Addr.String()
}

// statusCodeAttr returns status code attribute based on given gRPC code.
func statusCodeAttr(c codes.Code) attribute.KeyValue {
	return grpcStatusCodeKey.Int64(int64(c))
}

// UnaryServerInterceptor wrapper with open telemetry
func UnaryServerInterceptor(opts *GRPCUnaryInterceptorOptions, redisClient *redis.Client) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		panicked := true // default value, if not panic, this will be changed to false before the defer func called

		defer func() {
			if r := recover(); r != nil || panicked {
				err = recoverFrom(ctx, r, opts.RecoveryHandlerFunc)
			}
		}()

		ctx, cancel := context.WithTimeout(ctx, opts.Timeout)
		defer cancel()

		var span trace.Span
		if opts.UseOpenTelemetry {
			requestMetadata, _ := metadata.FromIncomingContext(ctx)
			metadataCopy := requestMetadata.Copy()

			bags, spanCtx := extract(ctx, &metadataCopy)
			ctx = baggage.ContextWithBaggage(ctx, bags)

			tracer := newConfig().TracerProvider.Tracer(
				instrumentationName,
			)

			name, attr := spanInfo(info.FullMethod, peerFromCtx(ctx))
			ctx, span = tracer.Start(
				trace.ContextWithRemoteSpanContext(ctx, spanCtx),
				name,
				trace.WithSpanKind(trace.SpanKindServer),
				trace.WithAttributes(attr...),
			)
			defer span.End()

			messageReceived.Event(ctx, defaultMessageID, req)

		}

		if opts.UseRateLimiter && redisClient != nil {
			meta, ok := metadata.FromIncomingContext(ctx)
			if !ok || len(meta.Get(string(ipAddressKey))) <= 0 {
				panicked = false
				return resp, status.Errorf(codes.Internal, "failed to get ip address from context")
			}

			ipAddress := meta.Get(string(ipAddressKey))[0]
			if ipAddress != "" && isRateLimited(ctx, redisClient, ipAddress, opts.RateLimiterLimit, opts.RateLimiterPeriod) {
				panicked = false
				return resp, status.Errorf(codes.ResourceExhausted, "too many requests")
			}
		}

		resp, err = handler(ctx, req)
		if opts.UseOpenTelemetry {
			if err != nil {
				s, _ := status.FromError(err)
				span.SetStatus(otelcodes.Error, s.Message())
				span.SetAttributes(statusCodeAttr(s.Code()))
				messageSent.Event(ctx, defaultMessageID, s.Proto())
			} else {
				span.SetAttributes(statusCodeAttr(codes.OK))
				messageSent.Event(ctx, defaultMessageID, resp)
			}
		}

		panicked = false
		return resp, err
	}
}

func recoverFrom(ctx context.Context, p interface{}, r RecoveryHandlerFunc) error {
	if r == nil {
		logrus.WithFields(logrus.Fields{
			"ctx":        utils.DumpIncomingContext(ctx),
			"stackTrace": string(debug.Stack()),
		}).Errorf("panic recovered: %v", p)
		return status.Errorf(codes.Internal, "%v", p)
	}
	return r(ctx, p)
}

func isRateLimited(ctx context.Context, redisClient *redis.Client, ip string, limit int64, period time.Duration) bool {
	store, err := redisStore.NewStoreWithOptions(redisClient, limiter.StoreOptions{
		Prefix: "grpc-rate-limiter:",
	})
	if err != nil {
		return false
	}
	limiterCtx, err := limiter.New(store, limiter.Rate{
		Period: period,
		Limit:  limit,
	}).Get(ctx, ip)
	if err != nil {
		return false
	}

	if limiterCtx.Reached {
		return true
	}

	return false
}
