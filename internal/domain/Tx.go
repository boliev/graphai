package domain

import (
	"context"
)

type Tx interface {
	Begin(ctx context.Context) (Tx, error)
	Commit() error
	Rollback() error
}
