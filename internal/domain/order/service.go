package order

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type Service struct {
	repository TxRepository
}

func NewService(repository TxRepository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) GetOrderByVkOrderIdTx(ctx context.Context, tx pgx.Tx, vkOrderId int64) (*Order, error) {
	return s.repository.GetByVKOrderID(ctx, tx, vkOrderId)
}

func (s *Service) UpsertTx(ctx context.Context, tx pgx.Tx, ord *Order) (*Order, error) {
	return s.repository.Upsert(ctx, tx, ord)
}
