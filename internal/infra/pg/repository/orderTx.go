package repository

import (
	"context"
	"errors"

	"github.com/boliev/graphai/internal/domain/order"
	"github.com/jackc/pgx/v5"
)

type OrderTxRepo struct {
}

func NewOrderTxRepo() *OrderTxRepo {
	return &OrderTxRepo{}
}

func (o *OrderTxRepo) GetByVKOrderID(ctx context.Context, tx pgx.Tx, vkOrderId int64) (*order.Order, error) {
	query := `SELECT id, vk_order_id, user_id, product, created_at FROM orders WHERE vk_order_id = $1`
	row := tx.QueryRow(ctx, query, vkOrderId)
	var ord order.Order
	err := row.Scan(&ord.ID, &ord.VkOrderID, &ord.UserID, &ord.Product, &ord.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &ord, nil
}

func (o *OrderTxRepo) Upsert(ctx context.Context, tx pgx.Tx, ord *order.Order) (*order.Order, error) {
	var newOrder order.Order
	sql := "INSERT INTO orders (vk_order_id, user_id, product) VALUES ($1, $2, $3) ON CONFLICT (vk_order_id) DO NOTHING RETURNING id, vk_order_id, user_id, product, created_at"
	row := tx.QueryRow(ctx, sql, ord.VkOrderID, ord.UserID, ord.Product)
	err := row.Scan(&newOrder.ID, &newOrder.VkOrderID, &newOrder.UserID, &newOrder.Product, &newOrder.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &newOrder, nil
}
