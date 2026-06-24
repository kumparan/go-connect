package connect

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/afex/hystrix-go/hystrix"
)

// HTTPConnectionOptions options for the http connection
type HTTPConnectionOptions struct {
	TLSHandshakeTimeout   time.Duration
	TLSInsecureSkipVerify bool
	Timeout               time.Duration
	UseOpenTelemetry      bool
	UseCircuitBreaker     bool
	CircuitBreakerConfig  *CircuitSetting
	EnableKeepAlives      bool
	Name                  string
}

var defaultHTTPConnectionOptions = &HTTPConnectionOptions{
	TLSHandshakeTimeout:   5 * time.Second,
	TLSInsecureSkipVerify: false,
	Timeout:               200 * time.Second,
	UseOpenTelemetry:      false,
	UseCircuitBreaker:     false,
	EnableKeepAlives:      true,
	Name:                  "HTTPRequest",
}

// NewHTTPConnection new http client
func NewHTTPConnection(opt *HTTPConnectionOptions) *http.Client {
	options := applyHTTPConnectionOptions(opt)

	var rt http.RoundTripper = &http.Transport{
		TLSHandshakeTimeout: options.TLSHandshakeTimeout,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: options.TLSInsecureSkipVerify}, //nolint:gosec
		DisableKeepAlives:   !options.EnableKeepAlives,
	}

	if options.UseOpenTelemetry {
		rt = NewTransport(options.Name, WithRoundTripper(rt))
	}

	if options.UseCircuitBreaker {
		rt = &CircuitBreakerTransport{commandName: options.Name, rt: rt}
		if options.CircuitBreakerConfig == nil {
			options.CircuitBreakerConfig = &defaultCircuitBreakerConfig
		}
		hystrix.ConfigureCommand(options.Name, hystrix.CommandConfig{
			Timeout:                opt.CircuitBreakerConfig.Timeout,
			MaxConcurrentRequests:  opt.CircuitBreakerConfig.MaxConcurrentRequests,
			RequestVolumeThreshold: opt.CircuitBreakerConfig.RequestVolumeThreshold,
			SleepWindow:            opt.CircuitBreakerConfig.SleepWindow,
			ErrorPercentThreshold:  opt.CircuitBreakerConfig.ErrorPercentThreshold,
		})
	}

	return &http.Client{Timeout: options.Timeout, Transport: rt}
}

func applyHTTPConnectionOptions(opt *HTTPConnectionOptions) *HTTPConnectionOptions {
	if opt != nil {
		return opt
	}
	return defaultHTTPConnectionOptions
}
