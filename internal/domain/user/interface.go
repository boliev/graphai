package user

import "context"

type Repository interface {
	Upsert(ctx context.Context, user *User) (*User, error)
	ReduceFreeUsages(ctx context.Context, id int64) error
}
