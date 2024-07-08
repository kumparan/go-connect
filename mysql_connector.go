package connect

import (
	"context"
	"time"

	"github.com/imdario/mergo"

	"github.com/jpillora/backoff"
	log "github.com/sirupsen/logrus"
	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// MySQLConnectionOptions options for the  CockroachDB connection
type MySQLConnectionOptions struct {
	PingInterval      time.Duration
	RetryAttempts     int
	MaxIdleConns      int
	MaxOpenConns      int
	ConnMaxLifetime   time.Duration
	LogLevel          string
	UseOpenTelemetry  bool
	PingTimeout       time.Duration
	ReconnectCallback func(db *gorm.DB) // func will provide new *gorm.DB if current connection is broken and re-creation of new connection is succeeded.
}

var (
	defaultMySQLConnectionOptions = &MySQLConnectionOptions{
		PingInterval:     1 * time.Second,
		RetryAttempts:    5,
		MaxIdleConns:     2,
		MaxOpenConns:     5,
		ConnMaxLifetime:  1 * time.Hour,
		LogLevel:         "info",
		UseOpenTelemetry: false,
		PingTimeout:      5 * time.Second,
	}
)

// InitializeMySQLConn :nodoc:
func InitializeMySQLConn(databaseDSN string, opt *MySQLConnectionOptions) (conn *gorm.DB, healthCheckStopFunc func(), err error) {
	options := applyMySQLConnectionOptions(opt)

	conn, err = openMySQLConn(databaseDSN, options)
	if err != nil {
		log.WithField("databaseDSN", databaseDSN).Fatal("failed to connect to MySQL database: ", err)
	}

	stopTickerCh := make(chan bool)
	healthCheckStopFunc = func() {
		stopTickerCh <- true
	}

	go checkMySQLConnection(conn, databaseDSN, stopTickerCh, options, time.NewTicker(options.PingInterval))

	conn.Logger = NewGormCustomLogger()

	switch options.LogLevel {
	case "error":
		conn.Logger = conn.Logger.LogMode(gormLogger.Error)
	case "warn":
		conn.Logger = conn.Logger.LogMode(gormLogger.Warn)
	default:
		conn.Logger = conn.Logger.LogMode(gormLogger.Info)

	}

	log.Info("Connection to MySQL Server success...")
	return
}

func openMySQLConn(dsn string, opts *MySQLConnectionOptions) (*gorm.DB, error) {
	conn, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Error(err)
		return nil, err
	}

	if opts.UseOpenTelemetry {
		if err := conn.Use(otelgorm.NewPlugin()); err != nil {
			log.Error(err)
			return nil, err
		}
	}

	db, err := conn.DB()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	db.SetMaxIdleConns(opts.MaxIdleConns)
	db.SetConnMaxLifetime(opts.ConnMaxLifetime)
	db.SetMaxOpenConns(opts.MaxOpenConns)

	return conn, nil
}

func checkMySQLConnection(db *gorm.DB, databaseDSN string, stopTickerCh chan bool, options *MySQLConnectionOptions, ticker *time.Ticker) {

	for {
		select {
		case <-stopTickerCh:
			ticker.Stop()
			return
		case <-ticker.C:
			db, err := db.DB()
			if err != nil {
				log.Error("MySQL got disconnected!")
				if options.ReconnectCallback == nil {
					continue
				}
				reconnectMySQLConn(databaseDSN, options)
				continue
			}

			ctx, cancelFunc := context.WithTimeout(context.TODO(), options.PingTimeout)
			if err = db.PingContext(ctx); err != nil {
				log.Error("MySQL got disconnected!")
				if options.ReconnectCallback == nil {
					continue
				}
				reconnectMySQLConn(databaseDSN, options)
			}
			cancelFunc()
		}
	}
}

func reconnectMySQLConn(databaseDSN string, options *MySQLConnectionOptions) {
	b := backoff.Backoff{
		Factor: 2,
		Jitter: true,
		Min:    100 * time.Millisecond,
		Max:    1 * time.Second,
	}

	for b.Attempt() < float64(options.RetryAttempts) {
		conn, err := openMySQLConn(databaseDSN, options)
		if err != nil {
			log.WithField("databaseDSN", databaseDSN).Error("failed to connect to MySQL database: ", err)
		}

		if conn != nil {
			options.ReconnectCallback(conn)
			break
		}
		time.Sleep(b.Duration())
	}

	if b.Attempt() >= float64(options.RetryAttempts) {
		log.Fatal("maximum retry to connect database")
	}
	b.Reset()
}

func applyMySQLConnectionOptions(opt *MySQLConnectionOptions) *MySQLConnectionOptions {
	if opt == nil {
		return defaultMySQLConnectionOptions
	}

	// if error occurs, also return options from input
	_ = mergo.Merge(opt, *defaultMySQLConnectionOptions)
	return opt
}
