package core_postgres_pool

import (
	"context"
	"time"
)

type Pool interface {
	Begin(ctx context.Context) (Tx, error)
	Query(ctx context.Context, sql string, args ...any) (Rows, error)
	Exec(ctx context.Context, sql string, arguments ...any) (CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) Row
	Close()

	Timeout() time.Duration
}

type Tx interface {
	Query(ctx context.Context, sql string, args ...any) (Rows, error)
	Exec(ctx context.Context, sql string, args ...any) (CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) Row
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type Rows interface {
	Close()
	Err() error
	Next() bool
	Scan(dest ...any) error
}

type Row interface {
	Scan(dest ...any) error
}

type CommandTag interface {
	RowsAffected() int64
}
