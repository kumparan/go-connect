package connect

import (
	"crypto/tls"
	"net/http"
	"time"
)

// HTTPConnectionOptions options for the http connection
type HTTPConnectionOptions struct {
	TLSHandshakeTimeout   time.Duration
	TLSInsecureSkipVerify bool
	Timeout               time.Duration
	UseOpenTelemetry      bool
	EnableKeepAlives      bool
}

var defaultHTTPConnectionOptions = &HTTPConnectionOptions{
	TLSHandshakeTimeout:   5 * time.Second,
	TLSInsecureSkipVerify: false,
	Timeout:               200 * time.Second,
	UseOpenTelemetry:      false,
	EnableKeepAlives:      true,
}

// NewHTTPConnection new http client
func NewHTTPConnection(opt *HTTPConnectionOptions) *http.Client {
	options := applyHTTPConnectionOptions(opt)

	httpClient := &http.Client{
		Timeout: options.Timeout,
		Transport: &http.Transport{
			TLSHandshakeTimeout: options.TLSHandshakeTimeout,
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: options.TLSInsecureSkipVerify}, //nolint:gosec
			DisableKeepAlives:   !opt.EnableKeepAlives,
		},
	}

	if !options.UseOpenTelemetry {
		return httpClient
	}

	httpClient.Transport = NewTransport(WithRoundTripper(httpClient.Transport))

	return httpClient
}

func applyHTTPConnectionOptions(opt *HTTPConnectionOptions) *HTTPConnectionOptions {
	if opt != nil {
		return opt
	}
	return defaultHTTPConnectionOptions
}
