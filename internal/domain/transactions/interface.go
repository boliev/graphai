package transactions

import "context"

type Repository interface {
	Create(ctx context.Context, tx *Transaction) error
}
