package core_pgx_pool

import (
	"errors"
	"fmt"

	core_postgres_pool "github.com/Censtro/todo-api/internal/core/repository/postgres/pool"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type PgxRow struct {
	pgx.Row
}

func (r PgxRow) Scan(dest ...any) error {
	const (
		pgxViolatesForeignKeyErrorCode = "23503"
	)
	err := r.Row.Scan(dest...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return core_postgres_pool.ErrNoRows
		}

		var PgErr *pgconn.PgError

		if errors.As(err, &PgErr) {
			if PgErr.Code == pgxViolatesForeignKeyErrorCode {
				return fmt.Errorf(
					"%v: %w",
					err,
					core_postgres_pool.ErrViolatesForeignKey,
				)
			}
		}

		return err
	}

	return nil
}

type PgxRows struct {
	pgx.Rows
}

type PgxCommandTag struct {
	pgconn.CommandTag
}
