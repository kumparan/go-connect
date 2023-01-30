package connect

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/olivere/elastic/v7"
	log "github.com/sirupsen/logrus"
)

// ElasticsearchConnectionOptions options for the Elasticsearch connection
type ElasticsearchConnectionOptions struct {
	TLSHandshakeTimeout             time.Duration
	TLSInsecureSkipVerify           bool
	MaxElasticsearchIdleConnections int
	MaxElasticsearchConnsPerHost    int
	ElasticsearchSetSniff           bool
	ElasticsearchSetHealthcheck     bool
	UseOpenTelemetry                bool
}

var defaultElasticsearchConnectionOptions = &ElasticsearchConnectionOptions{
	TLSHandshakeTimeout:             5 * time.Second,
	TLSInsecureSkipVerify:           false,
	MaxElasticsearchIdleConnections: 2,
	MaxElasticsearchConnsPerHost:    10,
	ElasticsearchSetSniff:           false,
	ElasticsearchSetHealthcheck:     false,
	UseOpenTelemetry:                false,
}

// ESInfoLogger :nodoc:
type ESInfoLogger struct{}

// ESErrorLogger :nodoc:
type ESErrorLogger struct{}

// ESTraceLogger :nodoc:
type ESTraceLogger struct{}

// NewElasticsearchClient :nodoc:
func NewElasticsearchClient(url string, httpClient *http.Client, opt *ElasticsearchConnectionOptions) (*elastic.Client, error) {
	options := applyElasticsearchConnectionOptions(opt)

	httpTranspost := &http.Transport{
		TLSHandshakeTimeout: options.TLSHandshakeTimeout,
		// Set true on purpose
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
		MaxIdleConnsPerHost: options.MaxElasticsearchIdleConnections,
		MaxConnsPerHost:     options.MaxElasticsearchConnsPerHost,
	}
	httpClient.Transport = httpTranspost
	if options.UseOpenTelemetry {
		httpClient.Transport = NewTransport(WithRoundTripper(httpTranspost))
	}

	return elastic.NewClient(
		elastic.SetURL(url),
		elastic.SetScheme("https"),
		elastic.SetSniff(options.ElasticsearchSetSniff),
		elastic.SetHealthcheck(options.ElasticsearchSetHealthcheck),
		elastic.SetErrorLog(&ESErrorLogger{}),
		elastic.SetInfoLog(&ESInfoLogger{}),
		elastic.SetTraceLog(&ESTraceLogger{}),
		elastic.SetHttpClient(httpClient),
	)
}

// Printf :nodoc:
func (*ESTraceLogger) Printf(format string, values ...interface{}) {
	log.WithFields(log.Fields{"type": "elasticsearch-log"}).Debugf(format, values...)
}

// Printf :nodoc:
func (*ESInfoLogger) Printf(format string, values ...interface{}) {
	log.WithFields(log.Fields{"type": "elasticsearch-log"}).Infof(format, values...)
}

// Printf :nodoc:
func (*ESErrorLogger) Printf(format string, values ...interface{}) {
	log.WithFields(log.Fields{"type": "elasticsearch-log"}).Errorf(format, values...)
}

func applyElasticsearchConnectionOptions(opt *ElasticsearchConnectionOptions) *ElasticsearchConnectionOptions {
	if opt != nil {
		return opt
	}
	return defaultElasticsearchConnectionOptions
}
