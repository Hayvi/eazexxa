package db

import (
	"context"
	"fmt"
	"time"

	"github.com/betpro/server/internal/config"
	"github.com/betpro/server/pkg/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool *pgxpool.Pool
	cfg  *config.Config
}

func New(ctx context.Context, cfg *config.Config) (*DB, error) {
	connStr := buildConnString(cfg)
	
	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("parse pool config: %w", err)
	}

	poolConfig.MaxConns = int32(cfg.DBPoolMax)
	poolConfig.MinConns = int32(cfg.DBPoolMin)
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = cfg.DBPoolIdleTimeout
	poolConfig.HealthCheckPeriod = 30 * time.Second
	poolConfig.ConnConfig.ConnectTimeout = cfg.DBPoolConnTimeout
	poolConfig.ConnConfig.RuntimeParams["application_name"] = cfg.DBAppName
	poolConfig.ConnConfig.RuntimeParams["statement_timeout"] = fmt.Sprintf("%d", cfg.DBStatementTimeout.Milliseconds())
	poolConfig.ConnConfig.RuntimeParams["idle_in_transaction_session_timeout"] = fmt.Sprintf("%d", cfg.DBIdleInTxTimeout.Milliseconds())

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	logger.Info("Database connection pool initialized",
		"max_conns", cfg.DBPoolMax,
		"min_conns", cfg.DBPoolMin,
		"db_name", cfg.DBName,
	)

	return &DB{Pool: pool, cfg: cfg}, nil
}

func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
		logger.Info("Database connection pool closed")
	}
}

func (db *DB) Health(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}

func (db *DB) Stats() *pgxpool.Stat {
	return db.Pool.Stat()
}

func (db *DB) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return db.Pool.Query(ctx, sql, args...)
}

func (db *DB) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return db.Pool.QueryRow(ctx, sql, args...)
}

func (db *DB) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	return db.Pool.Exec(ctx, sql, args...)
}

func (db *DB) Begin(ctx context.Context) (pgx.Tx, error) {
	return db.Pool.Begin(ctx)
}

func (db *DB) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error) {
	return db.Pool.BeginTx(ctx, txOptions)
}

func buildConnString(cfg *config.Config) string {
	if cfg.DBPassword != "" {
		return fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=disable",
			cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBUser, cfg.DBPassword)
	}
	return fmt.Sprintf("host=%s port=%d dbname=%s user=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBUser)
}
