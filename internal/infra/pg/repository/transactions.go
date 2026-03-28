package repository

import (
	"context"

	"github.com/boliev/graphai/internal/domain/transactions"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionsRepo struct {
	pool *pgxpool.Pool
}

func NewTransactionsRepo(pool *pgxpool.Pool) *TransactionsRepo {
	return &TransactionsRepo{pool: pool}
}

func (t *TransactionsRepo) Create(ctx context.Context, tx *transactions.Transaction) error {
	sql := "INSERT INTO transactions (user_id, prompt, operation_type, amount) VALUES ($1, $2, $3, $4)"
	_, err := t.pool.Exec(ctx, sql, tx.UserID, tx.Prompt, tx.OperationType, tx.Amount)
	if err != nil {
		return err
	}

	return nil
}
