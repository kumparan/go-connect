package connect

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRoundTripper struct {
	resp *http.Response
	err  error
}

func (m *mockRoundTripper) RoundTrip(_ *http.Request) (*http.Response, error) {
	return m.resp, m.err
}

func makeResponse(statusCode int) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader("")),
	}
}

func Test_applyHTTPConnectionOptions(t *testing.T) {
	t.Run("nil returns defaults", func(t *testing.T) {
		opt := applyHTTPConnectionOptions(nil)
		assert.Equal(t, defaultHTTPConnectionOptions.TLSHandshakeTimeout, opt.TLSHandshakeTimeout)
		assert.Equal(t, defaultHTTPConnectionOptions.TLSInsecureSkipVerify, opt.TLSInsecureSkipVerify)
		assert.Equal(t, defaultHTTPConnectionOptions.Timeout, opt.Timeout)
		assert.Equal(t, defaultHTTPConnectionOptions.UseOpenTelemetry, opt.UseOpenTelemetry)
		assert.Equal(t, defaultHTTPConnectionOptions.UseCircuitBreaker, opt.UseCircuitBreaker)
		assert.Equal(t, defaultHTTPConnectionOptions.EnableKeepAlives, opt.EnableKeepAlives)
		assert.Equal(t, defaultHTTPConnectionOptions.Name, opt.Name)
	})

	t.Run("non-nil returns same pointer", func(t *testing.T) {
		in := &HTTPConnectionOptions{Name: "custom"}
		out := applyHTTPConnectionOptions(in)
		assert.Same(t, in, out)
	})
}

func TestNewHTTPConnection(t *testing.T) {
	t.Run("no flags - base transport", func(t *testing.T) {
		client := NewHTTPConnection(&HTTPConnectionOptions{
			UseCircuitBreaker: false,
			UseOpenTelemetry:  false,
		})
		assert.IsType(t, &http.Transport{}, client.Transport)
	})

	t.Run("circuit breaker only - wrapped in CircuitBreakerTransport", func(t *testing.T) {
		client := NewHTTPConnection(&HTTPConnectionOptions{
			UseCircuitBreaker: true,
			Name:              "cb-only",
		})
		cb, ok := client.Transport.(*CircuitBreakerTransport)
		require.True(t, ok)
		assert.IsType(t, &http.Transport{}, cb.rt)
	})

	t.Run("otel only - wrapped in Transport", func(t *testing.T) {
		client := NewHTTPConnection(&HTTPConnectionOptions{
			UseOpenTelemetry: true,
			Name:             "otel-only",
		})
		assert.IsType(t, &Transport{}, client.Transport)
	})

	t.Run("both flags - CircuitBreakerTransport wrapping Transport", func(t *testing.T) {
		client := NewHTTPConnection(&HTTPConnectionOptions{
			UseCircuitBreaker: true,
			UseOpenTelemetry:  true,
			Name:              "both",
		})
		cb, ok := client.Transport.(*CircuitBreakerTransport)
		require.True(t, ok)
		assert.IsType(t, &Transport{}, cb.rt)
	})
}

func TestCircuitBreakerTransport_RoundTrip(t *testing.T) {
	newCBTransport := func(name string, mock *mockRoundTripper) *CircuitBreakerTransport {
		return &CircuitBreakerTransport{commandName: name, rt: mock}
	}

	makeReq := func() *http.Request {
		req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
		return req
	}

	t.Run("2xx - success, response returned", func(t *testing.T) {
		mock := &mockRoundTripper{resp: makeResponse(200)}
		cb := newCBTransport(t.Name(), mock)

		resp, err := cb.RoundTrip(makeReq())
		assert.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("3xx - success, response returned", func(t *testing.T) {
		mock := &mockRoundTripper{resp: makeResponse(301)}
		cb := newCBTransport(t.Name(), mock)

		resp, err := cb.RoundTrip(makeReq())
		assert.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, 301, resp.StatusCode)
	})

	t.Run("4xx - response returned, circuit not tripped", func(t *testing.T) {
		mock := &mockRoundTripper{resp: makeResponse(404)}
		cb := newCBTransport(t.Name(), mock)

		resp, err := cb.RoundTrip(makeReq())
		assert.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, 404, resp.StatusCode)
	})

	t.Run("5xx - error returned", func(t *testing.T) {
		mock := &mockRoundTripper{resp: makeResponse(500)}
		cb := newCBTransport(t.Name(), mock)

		resp, err := cb.RoundTrip(makeReq())
		assert.Nil(t, resp)
		assert.ErrorContains(t, err, "server error: 500")
	})

	t.Run("network error - error returned", func(t *testing.T) {
		mock := &mockRoundTripper{err: fmt.Errorf("dial failed")}
		cb := newCBTransport(t.Name(), mock)

		resp, err := cb.RoundTrip(makeReq())
		assert.Nil(t, resp)
		assert.ErrorContains(t, err, "dial failed")
	})

	t.Run("circuit open after repeated 5xx", func(t *testing.T) {
		cmdName := t.Name()
		hystrix.ConfigureCommand(cmdName, hystrix.CommandConfig{
			RequestVolumeThreshold: 1,
			ErrorPercentThreshold:  1,
			SleepWindow:            5000,
		})

		mock := &mockRoundTripper{resp: makeResponse(500)}
		cb := newCBTransport(cmdName, mock)

		// first call trips the circuit; hystrix updates metrics asynchronously
		_, _ = cb.RoundTrip(makeReq())
		time.Sleep(50 * time.Millisecond)

		// subsequent call should get circuit-open error without hitting the mock
		mock.resp = makeResponse(200) // would return success if mock were called
		_, err := cb.RoundTrip(makeReq())
		assert.Error(t, err)
		assert.ErrorContains(t, err, "circuit open")
	})
}
