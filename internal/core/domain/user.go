package domain

import (
	"fmt"
	"regexp"

	core_errors "github.com/Censtro/todo-api/internal/core/errors"
)

type User struct {
	ID      int
	Version int

	FullName    string
	PhoneNumber *string
}

func NewUserUninitialized(
	fullname string,
	phonenumber *string,
) User {
	return NewUser(
		UninitializedID,
		UninitializedVersion,
		fullname,
		phonenumber,
	)
}

func NewUser(
	id int,
	version int,
	fullname string,
	phonenumber *string,
) User {
	return User{
		ID:          id,
		Version:     version,
		FullName:    fullname,
		PhoneNumber: phonenumber,
	}
}

func (u *User) Validate() error {
	fullNameLength := len([]rune(u.FullName))
	if fullNameLength < 3 || fullNameLength > 100 {
		return fmt.Errorf(
			"Invalid FullName len: %d, %w",
			fullNameLength,
			core_errors.ErrInvalidArgument,
		)
	}
	if u.PhoneNumber != nil {
		phoneNumberLength := len([]rune(*u.PhoneNumber))
		if phoneNumberLength < 10 || phoneNumberLength > 15 {
			return fmt.Errorf(
				"Invalid PhoneNumber len: %d, %w",
				phoneNumberLength,
				core_errors.ErrInvalidArgument,
			)
		}
		re := regexp.MustCompile(`^\+[0-9]+$`)

		if !re.MatchString(*u.PhoneNumber) {
			return fmt.Errorf(
				"invalid PhoneNumber format: %w",
				core_errors.ErrInvalidArgument,
			)
		}
	}

	return nil
}

type UserPatch struct {
	FullName    Nullable[string]
	PhoneNumber Nullable[string]
}

func (p *UserPatch) Validate() error {
	if p.FullName.Set && p.FullName.Value == nil {
		return fmt.Errorf(
			"FullName can't be patched to null: %w",
			core_errors.ErrInvalidArgument,
		)
	}

	return nil
}

func (u *User) ApplyPatch(patch UserPatch) error {
	if err := patch.Validate(); err != nil {
		return fmt.Errorf("validate user patch: %w", err)
	}

	temp := *u

	if patch.FullName.Set {
		temp.FullName = *patch.FullName.Value
	}

	if patch.PhoneNumber.Set {
		temp.PhoneNumber = patch.PhoneNumber.Value
	}

	if err := temp.Validate(); err != nil {
		return fmt.Errorf("validate patched user: %w", err)
	}

	*u = temp

	return nil
}
