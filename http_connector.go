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
}

var defaultHTTPConnectionOptions = &HTTPConnectionOptions{
	TLSHandshakeTimeout:   5 * time.Second,
	TLSInsecureSkipVerify: false,
	Timeout:               200 * time.Second,
}

// NewHTTPConnection new http client
func NewHTTPConnection(opt *HTTPConnectionOptions) *http.Client {
	options := applyHTTPConnectionOptions(opt)

	return &http.Client{
		Timeout: options.Timeout,
		Transport: &http.Transport{
			TLSHandshakeTimeout: options.TLSHandshakeTimeout,
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: options.TLSInsecureSkipVerify},
		},
	}
}

func applyHTTPConnectionOptions(opt *HTTPConnectionOptions) *HTTPConnectionOptions {
	if opt != nil {
		return opt
	}
	return defaultHTTPConnectionOptions
}
