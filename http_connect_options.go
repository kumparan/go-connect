package connect

import (
	"bytes"
	"io"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// Transport for tracing HTTP operations.
type Transport struct {
	rt             http.RoundTripper
	connectionName string
}

// Option signature for specifying options, e.g. WithRoundTripper.
type Option func(t *Transport)

// WithRoundTripper specifies the http.RoundTripper to call
// next after this transport. If it is nil (default), the
// transport will use http.DefaultTransport.
func WithRoundTripper(rt http.RoundTripper) Option {
	return func(t *Transport) {
		t.rt = rt
	}
}

// NewTransport specifies a transport that will trace HTTP
// and report back via OpenTracing.
func NewTransport(connectionName string, opts ...Option) *Transport {
	t := &Transport{
		connectionName: connectionName,
	}
	for _, o := range opts {
		o(t)
	}
	return t
}

// RoundTrip captures the request and starts an OpenTracing span
// for HTTP operation.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx, span := otel.Tracer("HTTP").Start(req.Context(), t.connectionName)
	defer span.End()

	// See General (https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/semantic_conventions/span-general.md)
	// and HTTP (https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/semantic_conventions/http.md)
	attributes := []attribute.KeyValue{
		attribute.String("http.url", req.URL.Redacted()),
		attribute.String("http.method", req.Method),
		attribute.String("http.scheme", req.URL.Scheme),
		attribute.String("http.host", req.URL.Hostname()),
		attribute.String("http.path", req.URL.Path),
		attribute.String("http.user_agent", req.UserAgent()),
	}

	req = req.WithContext(ctx)

	var (
		buf    []byte
		err    error
		reader io.ReadCloser
	)
	if req.Body == nil {
		goto SetAttribute
	}

	buf, err = io.ReadAll(req.Body)
	if err == nil {
		attributes = append(attributes, attribute.String("http.body", string(buf)))
	}

	reader = io.NopCloser(bytes.NewBuffer(buf))
	req.Body = reader

SetAttribute:
	span.SetAttributes(
		attributes...,
	)

	var (
		resp *http.Response
	)
	if t.rt != nil {
		resp, err = t.rt.RoundTrip(req)
	} else {
		resp, err = http.DefaultTransport.RoundTrip(req)
	}
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	if resp != nil {
		span.SetAttributes(attribute.Int64("http.status_code", int64(resp.StatusCode)))
	}

	return resp, err
}
