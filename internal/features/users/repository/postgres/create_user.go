package user_postgres_repository

import (
	"context"
	"fmt"

	"github.com/Censtro/todo-api/internal/core/domain"
)

func (r *UserRepository) CreateUser(
	ctx context.Context,
	user domain.User,
) (domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.Pool.Timeout())
	defer cancel()

	query := `
	INSERT INTO todoapp.users (full_name, phone_number)
	VALUES ($1,$2)
	RETURNING id, version, full_name, phone_number;
	`
	row := r.Pool.QueryRow(ctx, query, user.FullName, user.PhoneNumber)

	var UserModel UserModel
	err := row.Scan(
		&UserModel.ID,
		&UserModel.Version,
		&UserModel.FullName,
		&UserModel.PhoneNumber,
	)
	if err != nil {
		return domain.User{}, fmt.Errorf("row scan: %w", err)
	}
	userDomain := domain.NewUser(
		UserModel.ID,
		UserModel.Version,
		UserModel.FullName,
		UserModel.PhoneNumber,
	)
	return userDomain, nil
}
