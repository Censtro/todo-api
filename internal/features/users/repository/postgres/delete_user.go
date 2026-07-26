package user_postgres_repository

import (
	"context"
	"fmt"

	core_errors "github.com/Censtro/todo-api/internal/core/errors"
)

func (r *UserRepository) DeleteUser(
	ctx context.Context,
	id int,
) error {
	ctx, cancel := context.WithTimeout(ctx, r.Pool.Timeout())
	defer cancel()

	query := `
	DELETE from todoapp.users
	WHERE id=$1
	`

	cmdTag, err := r.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("exec query: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("user with id='%d': %w", id, core_errors.ErrNotFound)
	}

	return nil
}
