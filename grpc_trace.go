package connect

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/metadata"
)

type contextKey string

const (
	// instrumentationName is the name of this instrumentation package.
	instrumentationName = "github.com/kumparan/go-connect"
	// grpcStatusCodeKey is convention for numeric status code of a gRPC request.
	grpcStatusCodeKey = attribute.Key("rpc.grpc.status_code")
	// defaultMessageID is default id for event message
	defaultMessageID = 1

	// ipAddressKey represents key to get ip address
	ipAddressKey = contextKey("ip_address")
)

// config is a group of options for this instrumentation.
type config struct {
	Propagators    propagation.TextMapPropagator
	TracerProvider trace.TracerProvider
}

// newConfig returns a config.
func newConfig() *config {
	return &config{
		Propagators:    otel.GetTextMapPropagator(),
		TracerProvider: otel.GetTracerProvider(),
	}
}

type metadataSupplier struct {
	metadata *metadata.MD
}

// assert that metadataSupplier implements the TextMapCarrier interface.
var _ propagation.TextMapCarrier = &metadataSupplier{}

func (s *metadataSupplier) Get(key string) string {
	values := s.metadata.Get(key)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func (s *metadataSupplier) Set(key string, value string) {
	s.metadata.Set(key, value)
}

func (s *metadataSupplier) Keys() []string {
	out := make([]string, 0, len(*s.metadata))
	for key := range *s.metadata {
		out = append(out, key)
	}
	return out
}

// inject injects correlation context and span context into the gRPC
// metadata object. This function is meant to be used on outgoing
// requests.
func inject(ctx context.Context, md *metadata.MD) {
	c := newConfig()
	c.Propagators.Inject(ctx, &metadataSupplier{
		metadata: md,
	})
}

// extract returns the correlation context and span context that
// another service encoded in the gRPC metadata object with Inject.
// This function is meant to be used on incoming requests.
func extract(ctx context.Context, md *metadata.MD) (baggage.Baggage, trace.SpanContext) {
	c := newConfig()
	ctx = c.Propagators.Extract(ctx, &metadataSupplier{
		metadata: md,
	})

	return baggage.FromContext(ctx), trace.SpanContextFromContext(ctx)
}
