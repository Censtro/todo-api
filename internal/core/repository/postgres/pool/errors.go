package core_postgres_pool

import "errors"

var (
	ErrNoRows             = errors.New("no rows in result set")
	ErrViolatesForeignKey = errors.New("violates foreign key")
)
