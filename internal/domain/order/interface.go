package order

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type TxRepository interface {
	GetByVKOrderID(ctx context.Context, tx pgx.Tx, vkOrderId int64) (*Order, error)
	Upsert(ctx context.Context, tx pgx.Tx, ord *Order) (*Order, error)
}
