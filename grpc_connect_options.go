package connect

import (
	"context"
	"net"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/sirupsen/logrus"

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
// passed context with id and size (if message is a proto message).
func (m messageType) Event(ctx context.Context, id int, message interface{}) {
	span := trace.SpanFromContext(ctx)
	if p, ok := message.(proto.Message); ok {
		span.AddEvent("message", trace.WithAttributes(
			attribute.KeyValue(m),
			attribute.Key("message.id").Int(id),
			attribute.Key("message.uncompressed_size").Int(proto.Size(p)),
		))
	} else {
		span.AddEvent("message", trace.WithAttributes(
			attribute.KeyValue(m),
			attribute.Key("message.id").Int(id),
		))
	}
}

var (
	messageSent     = messageType(attribute.Key("message.type").String("SENT"))
	messageReceived = messageType(attribute.Key("message.type").String("RECEIVED"))

	// Span is a component of a trace
	Span trace.Span
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

	// UseOpenTelemetry flag if the connection will implement open telemetry
	UseOpenTelemetry bool
}

var defaultGRPCUnaryInterceptorOptions = &GRPCUnaryInterceptorOptions{
	UseCircuitBreaker: false,
	RetryCount:        0,
	RetryInterval:     20 * time.Millisecond,
	Timeout:           1 * time.Second,
	UseOpenTelemetry:  false,
}

// UnaryClientInterceptor wrapper with circuit breaker, retry, timeout, open telemetry, and metadata logging
func UnaryClientInterceptor(opts *GRPCUnaryInterceptorOptions) grpc.UnaryClientInterceptor {
	o := applyGRPCUnaryInterceptorOptions(opts)
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx, cancel := context.WithTimeout(ctx, o.Timeout)
		defer cancel()

		ctx = metadata.AppendToOutgoingContext(ctx, "caller", utils.MyCaller(5))

		if o.UseOpenTelemetry {
			requestMetadata, _ := metadata.FromOutgoingContext(ctx)
			metadataCopy := requestMetadata.Copy()
			tracer := newConfig().TracerProvider.Tracer(
				instrumentationName,
			)

			name, attr := spanInfo(method, cc.Target())
			ctx, Span = tracer.Start(
				ctx,
				name,
				trace.WithSpanKind(trace.SpanKindClient),
				trace.WithAttributes(attr...),
			)
			defer Span.End()

			Inject(ctx, &metadataCopy)
			ctx = metadata.NewOutgoingContext(ctx, metadataCopy)

			messageSent.Event(ctx, 1, req)
		}

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
	return utils.Retry(o.RetryCount, o.RetryInterval, func() error {
		err := invoker(ctx, method, req, reply, cc, opts...)

		if o.UseOpenTelemetry {
			messageReceived.Event(ctx, 1, reply)

			if err != nil {
				s, _ := status.FromError(err)
				Span.SetStatus(otelcodes.Error, s.Message())
				Span.SetAttributes(statusCodeAttr(s.Code()))
			} else {
				Span.SetAttributes(statusCodeAttr(codes.OK))
			}
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
	return GRPCStatusCodeKey.Int64(int64(c))
}

// UnaryServerInterceptor wrapper with open telemetry
func UnaryServerInterceptor(opts *GRPCUnaryInterceptorOptions) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		ctx, cancel := context.WithTimeout(ctx, opts.Timeout)
		defer cancel()

		if opts.UseOpenTelemetry {
			requestMetadata, _ := metadata.FromIncomingContext(ctx)
			metadataCopy := requestMetadata.Copy()

			bags, spanCtx := Extract(ctx, &metadataCopy)
			ctx = baggage.ContextWithBaggage(ctx, bags)

			tracer := newConfig().TracerProvider.Tracer(
				instrumentationName,
			)

			name, attr := spanInfo(info.FullMethod, peerFromCtx(ctx))
			ctx, Span := tracer.Start(
				trace.ContextWithRemoteSpanContext(ctx, spanCtx),
				name,
				trace.WithSpanKind(trace.SpanKindServer),
				trace.WithAttributes(attr...),
			)
			defer Span.End()

			messageReceived.Event(ctx, 1, req)

		}

		resp, err := handler(ctx, req)
		if opts.UseOpenTelemetry {
			if err != nil {
				s, _ := status.FromError(err)
				Span.SetStatus(otelcodes.Error, s.Message())
				Span.SetAttributes(statusCodeAttr(s.Code()))
				messageSent.Event(ctx, 1, s.Proto())
			} else {
				Span.SetAttributes(statusCodeAttr(codes.OK))
				messageSent.Event(ctx, 1, resp)
			}
		}

		return resp, err
	}
}
