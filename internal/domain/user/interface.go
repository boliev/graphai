package user

import (
	"context"
)

type Repository interface {
	Upsert(ctx context.Context, user *User) (*User, error)
	ReduceCredits(ctx context.Context, id int64) error
	FindByVKID(ctx context.Context, id int64) (*User, error)
}
