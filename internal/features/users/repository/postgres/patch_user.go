package user_postgres_repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Censtro/todo-api/internal/core/domain"
	core_errors "github.com/Censtro/todo-api/internal/core/errors"
	core_postgres_pool "github.com/Censtro/todo-api/internal/core/repository/postgres/pool"
)

func (r *UserRepository) PatchUser(
	ctx context.Context,
	id int,
	user domain.User,
) (domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.Pool.Timeout())
	defer cancel()

	query := `
	UPDATE todoapp.users
	SET 
		full_name=$1,
		phone_number=$2,
		version=version+1
	WHERE id=$3 and version=$4
	RETURNING
		id,
		version,
		full_name,
		phone_number
	`
	row := r.Pool.QueryRow(
		ctx,
		query,
		user.FullName,
		user.PhoneNumber,
		user.ID,
		user.Version,
	)

	var userModel UserModel
	err := row.Scan(
		&userModel.ID,
		&userModel.Version,
		&userModel.FullName,
		&userModel.PhoneNumber,
	)
	if err != nil {
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return domain.User{}, fmt.Errorf(
				"user with id '%d' concurrently accesed: %w",
				id,
				core_errors.ErrConflit,
			)
		}

		return domain.User{}, fmt.Errorf("scan error: %w", err)
	}

	userDomain := domain.NewUser(
		userModel.ID,
		userModel.Version,
		userModel.FullName,
		userModel.PhoneNumber,
	)

	return userDomain, nil
}
