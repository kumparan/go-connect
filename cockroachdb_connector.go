package connect

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/jpillora/backoff"
	log "github.com/sirupsen/logrus"
	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// CockroachDBConnectionOptions options for the  CockroachDB connection
type CockroachDBConnectionOptions struct {
	PingInterval     time.Duration
	RetryAttempts    int
	MaxIdleConns     int
	MaxOpenConns     int
	ConnMaxLifetime  time.Duration
	UseOpenTelemetry bool
}

var (
	// CockroachDB represents gorm DB
	CockroachDB *gorm.DB
	// StopTickerCh signal for closing ticker channel
	StopTickerCh chan bool

	sqlRegexp = regexp.MustCompile(`(\$\d+)|\?`)

	defaultCockroachDBConnectionOptions = &CockroachDBConnectionOptions{
		PingInterval:     1 * time.Second,
		RetryAttempts:    5,
		MaxIdleConns:     2,
		MaxOpenConns:     5,
		ConnMaxLifetime:  1 * time.Hour,
		UseOpenTelemetry: false,
	}
)

// InitializeCockroachConn :nodoc:
func InitializeCockroachConn(databaseDSN, logLevel string, opt *CockroachDBConnectionOptions) {
	options := applyCockroachDBConnectionOptions(opt)

	conn, err := openCockroachConn(databaseDSN, options)
	if err != nil {
		log.WithField("databaseDSN", databaseDSN).Fatal("failed to connect cockroach database: ", err)
	}

	CockroachDB = conn
	StopTickerCh = make(chan bool)

	go checkConnection(databaseDSN, options, time.NewTicker(options.PingInterval))

	CockroachDB.Logger = NewGormCustomLogger()

	switch logLevel {
	case "error":
		CockroachDB.Logger = CockroachDB.Logger.LogMode(gormLogger.Error)
	case "warn":
		CockroachDB.Logger = CockroachDB.Logger.LogMode(gormLogger.Warn)
	default:
		CockroachDB.Logger = CockroachDB.Logger.LogMode(gormLogger.Info)

	}

	log.Info("Connection to Cockroach Server success...")
}

func checkConnection(databaseDSN string, options *CockroachDBConnectionOptions, ticker *time.Ticker) {
	for {
		select {
		case <-StopTickerCh:
			ticker.Stop()
			return
		case <-ticker.C:
			if _, err := CockroachDB.DB(); err != nil {
				reconnectCockroachConn(databaseDSN, options)
			}
		}
	}
}

func reconnectCockroachConn(databaseDSN string, options *CockroachDBConnectionOptions) {
	b := backoff.Backoff{
		Factor: 2,
		Jitter: true,
		Min:    100 * time.Millisecond,
		Max:    1 * time.Second,
	}

	for b.Attempt() < float64(options.RetryAttempts) {
		conn, err := openCockroachConn(databaseDSN, options)
		if err != nil {
			log.WithField("databaseDSN", databaseDSN).Error("failed to connect cockroach database: ", err)
		}

		if conn != nil {
			CockroachDB = conn
			break
		}
		time.Sleep(b.Duration())
	}

	if b.Attempt() >= float64(options.RetryAttempts) {
		log.Fatal("maximum retry to connect database")
	}
	b.Reset()
}

func openCockroachConn(dsn string, options *CockroachDBConnectionOptions) (*gorm.DB, error) {
	dialector := postgres.Open(dsn)
	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.Use(otelgorm.NewPlugin()); err != nil {
		log.Fatal(err)
	}

	conn, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}
	conn.SetMaxIdleConns(options.MaxIdleConns)
	conn.SetConnMaxLifetime(options.ConnMaxLifetime)
	conn.SetMaxOpenConns(options.MaxOpenConns)

	return db, nil
}

// GormCustomLogger override gorm logger
type GormCustomLogger struct {
	gormLogger.Config
}

// NewGormCustomLogger :nodoc:
func NewGormCustomLogger() *GormCustomLogger {
	return &GormCustomLogger{
		Config: gormLogger.Config{
			LogLevel: gormLogger.Info,
		},
	}
}

// LogMode :nodoc:
func (g *GormCustomLogger) LogMode(level gormLogger.LogLevel) gormLogger.Interface {
	g.LogLevel = level
	return g
}

// Info :nodoc:
func (g *GormCustomLogger) Info(ctx context.Context, message string, values ...interface{}) {
	if g.LogLevel >= gormLogger.Info {
		log.WithFields(log.Fields{"data": values}).Info(message)
	}
}

// Warn :nodoc:
func (g *GormCustomLogger) Warn(ctx context.Context, message string, values ...interface{}) {
	if g.LogLevel >= gormLogger.Warn {
		log.WithFields(log.Fields{"data": values}).Warn(message)
	}

}

// Error :nodoc:
func (g *GormCustomLogger) Error(ctx context.Context, message string, values ...interface{}) {
	if g.LogLevel >= gormLogger.Error {
		log.WithFields(log.Fields{"data": values}).Error(message)
	}
}

// Trace :nodoc:
func (g *GormCustomLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	sql, rows := fc()
	if g.LogLevel <= 0 {
		return
	}

	elapsed := time.Since(begin)
	logger := log.WithFields(log.Fields{
		"took": elapsed,
	})

	sqlLog := sqlRegexp.ReplaceAllString(sql, "%v")
	if rows >= 0 {
		logger.WithField("rows", rows)
	} else {
		logger.WithField("rows", "-")
	}

	switch {
	case err != nil && !errors.Is(err, gorm.ErrRecordNotFound) && g.LogLevel >= gormLogger.Error:
		logger.WithField("sql", sqlLog).Error(err)
	case elapsed > g.SlowThreshold && g.SlowThreshold != 0 && g.LogLevel >= gormLogger.Warn:
		slowLog := fmt.Sprintf("SLOW SQL >= %v", g.SlowThreshold)
		logger.WithField("sql", sqlLog).Warn(slowLog)
	case g.LogLevel >= gormLogger.Info:
		logger.Info(sqlLog)

	}
}

func applyCockroachDBConnectionOptions(opt *CockroachDBConnectionOptions) *CockroachDBConnectionOptions {
	if opt != nil {
		return opt
	}
	return defaultCockroachDBConnectionOptions
}
