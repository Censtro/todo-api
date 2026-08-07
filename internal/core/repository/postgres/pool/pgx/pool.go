package core_pgx_pool

import (
	"context"
	"fmt"
	"time"

	core_postgres_pool "github.com/Censtro/todo-api/internal/core/repository/postgres/pool"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	core_postgres_pool.BaseConfig               //`envconfig:",inline"`
	Timeout                       time.Duration `envconfig:"TIMEOUT" default:"30s"`
}

type ConnectionPool struct {
	*pgxpool.Pool
	timeout time.Duration
}

func NewConfig() (Config, error) {
	var cfg Config
	baseCfg, err := core_postgres_pool.NewConfig()
	if err != nil {
		return Config{}, err
	}
	cfg.BaseConfig = baseCfg
	if err := envconfig.Process("POSTGRES", &cfg); err != nil {
		return Config{}, fmt.Errorf("process envconfig: %w", err)
	}
	return cfg, nil
}

func NewConfigMust() Config {
	cfg, err := NewConfig()
	if err != nil {
		panic(err)
	}
	return cfg
}

func NewConnectionPool(
	ctx context.Context,
	cfg Config,
) (*ConnectionPool, error) {
	connectionString := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
	)
	pgxconfig, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		return nil, fmt.Errorf("parse pgxconfig: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, pgxconfig)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("pgxpool ping: %w", err)
	}

	return &ConnectionPool{
		Pool:    pool,
		timeout: cfg.Timeout,
	}, nil

}

func (p *ConnectionPool) Timeout() time.Duration {
	return p.timeout
}

func (p *ConnectionPool) Query(ctx context.Context, sql string, args ...any) (core_postgres_pool.Rows, error) {
	rows, err := p.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	return PgxRows{rows}, nil
}

func (p *ConnectionPool) Exec(ctx context.Context, sql string, args ...any) (core_postgres_pool.CommandTag, error) {
	tag, err := p.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return PgxCommandTag{tag}, nil
}

func (p *ConnectionPool) QueryRow(ctx context.Context, sql string, args ...any) core_postgres_pool.Row {
	row := p.Pool.QueryRow(ctx, sql, args...)
	return PgxRow{row}
}

func (p *ConnectionPool) Close() {
	p.Pool.Close()
}
