package database

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // nolintlint
)

var (
	// ErrDBConnAttemptsFailed ошибка для случаев невозможности подключения к БД
	ErrDBConnAttemptsFailed = errors.New("все попытки подключения к базе провалились")
	ErrStopBySignal         = errors.New("service was stopped by signal")
	dbDriverName            = "postgres"                     // nolint: gochecknoglobals
	connMaxIdleTime         = time.Duration(5) * time.Minute // nolint: gochecknoglobals, gomnd
	connMaxLifetime         = time.Duration(5) * time.Minute // nolint: gochecknoglobals, gomnd
	maxConnectionAttempts   = 30                             // nolint: gochecknoglobals, gomnd
)

// checkConnectionMethod описывает метод проверки доступности БД
type checkConnectionMethod string

const (
	// pingCheckMethod проверим базу с помощью ping
	pingCheckMethod = checkConnectionMethod("ping")
	// queryCheckMethod проверим базу с помощью запроса
	queryCheckMethod = checkConnectionMethod("query")
)

// DB интерфейс работы с подключением к БД
type DB interface {
	QueryExecutor
}

// QueryExecutor исполнитель запросов
type QueryExecutor interface {
	sqlx.Execer
	sqlx.ExecerContext
	sqlx.QueryerContext
}

type Logger interface {
	Error(err error, msg string, keysAndValues ...interface{})
}

// Database - компонент для подключения к БД
type Database struct {
	db     *sqlx.DB
	logger Logger
}

// NewDB инициализирует подключение к БД
func NewDB(ctx context.Context, dsn string, l Logger) (*Database, error) {
	db, err := sqlx.Open(dbDriverName, dsn)
	if err != nil {
		return nil, fmt.Errorf("cannot init connection to DB: %w", err)
	}

	db.SetConnMaxLifetime(connMaxLifetime)

	res := &Database{
		db:     db,
		logger: l,
	}

	err = res.connect(ctx, pingCheckMethod)
	if err != nil {
		return nil, fmt.Errorf("cannot init connection to DB, dsn='%s': %w", dsn, err)
	}
	return res, nil
}

type PoolSettings struct {
	ConnMaxIdleTime time.Duration
	ConnMaxLifetime time.Duration
	MaxIdleConns    int
	MaxOpenConns    int
}

var DefaultPoolSettings = PoolSettings{
	ConnMaxIdleTime: connMaxIdleTime,
	ConnMaxLifetime: connMaxLifetime,
	MaxIdleConns:    2,
	MaxOpenConns:    2,
}

type databaseOptions struct {
	logger              Logger
	binaryParamsEnabled bool
}

func newDatabaseOptions(opts ...DatabaseOption) databaseOptions {
	def := defaultDatabaseOptions()
	for _, opt := range opts {
		opt(&def)
	}
	return def
}

func defaultDatabaseOptions() databaseOptions {
	return databaseOptions{}
}

type DatabaseOption func(*databaseOptions)

// DatabaseWithBinaryParams опция добавляет binary_parameters=yes к DSN query параметрам.
func DatabaseWithBinaryParams() DatabaseOption {
	return func(do *databaseOptions) {
		do.binaryParamsEnabled = true
	}
}

func DatabaseWithLogger(l Logger) DatabaseOption {
	return func(do *databaseOptions) {
		do.logger = l
	}
}

func applyBinaryParamsToDSN(dsn string) (string, error) {
	parsedDSN, err := url.Parse(dsn)
	if err != nil {
		return "", fmt.Errorf("parse dns: %w", err)
	}

	query := parsedDSN.Query()
	query.Set("binary_parameters", "yes")
	parsedDSN.RawQuery = query.Encode()

	return parsedDSN.String(), nil
}

func NewDBWithPoolSettings(ctx context.Context, dsn string, poolSettings PoolSettings, opts ...DatabaseOption) (*Database, error) {
	options := newDatabaseOptions(opts...)

	if options.binaryParamsEnabled {
		modifiedDSN, err := applyBinaryParamsToDSN(dsn)
		if err != nil {
			return nil, err
		}
		dsn = modifiedDSN
	}

	db, err := NewDB(ctx, dsn, options.logger)
	if err != nil {
		return nil, fmt.Errorf("init connection pool to postgres: %w", err)
	}

	db.db.SetConnMaxIdleTime(poolSettings.ConnMaxIdleTime)
	db.db.SetConnMaxLifetime(poolSettings.ConnMaxLifetime)
	db.db.SetMaxIdleConns(poolSettings.MaxIdleConns)
	db.db.SetMaxOpenConns(poolSettings.MaxOpenConns)

	return db, nil
}

// connect осуществляет попытку подключения к БД
func (d *Database) connect(ctx context.Context, checkDBMethod checkConnectionMethod) error {
	var dbError error

	for attempt := 1; attempt <= maxConnectionAttempts; attempt++ {
		switch checkDBMethod {
		case pingCheckMethod:
			dbError = d.db.PingContext(ctx)
		case queryCheckMethod:
			_, dbError = d.db.ExecContext(ctx, "SELECT 1;")
		}
		if errors.Is(dbError, context.Canceled) {
			return ErrStopBySignal
		}
		if dbError == nil {
			break
		}

		nextAttemptWait := time.Duration(attempt) * time.Second
		d.logger.Error(dbError,
			"не удалось установить соединение с базой",
			"attempt", attempt,
			"nextAttemptWait", nextAttemptWait,
		)

		timer := time.NewTimer(nextAttemptWait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ErrStopBySignal
		case <-timer.C:
		}
	}

	if dbError != nil {
		return ErrDBConnAttemptsFailed
	}

	return nil
}

// Close закрывает соединение с БД
func (d *Database) Close() error {
	if err := d.db.Close(); err != nil {
		return fmt.Errorf("не удалось закрыть соединение с базой: %w", err)
	}

	return nil
}

// DB возвращает указатель на пул коннектов к БД.
func (d *Database) DB() DB {
	return d.db
}
