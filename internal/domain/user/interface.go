package user

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type Repository interface {
	Upsert(ctx context.Context, user *User) (*User, error)
	ReduceCredits(ctx context.Context, id int64) error
	IncreaseCreditsTx(ctx context.Context, tx pgx.Tx, id, credits int64) error
	FindByVKID(ctx context.Context, id int64) (*User, error)
}
