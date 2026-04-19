package repository

import (
	"context"

	"github.com/boliev/graphai/internal/domain/prompt"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PromptsRepo struct {
	pool *pgxpool.Pool
}

func NewPromptsRepo(pool *pgxpool.Pool) *PromptsRepo {
	return &PromptsRepo{pool: pool}
}

func (t *PromptsRepo) Create(ctx context.Context, prompt *prompt.Prompt) error {
	sql := "INSERT INTO prompts (user_id, prompt) VALUES ($1, $2)"

	_, err := t.pool.Exec(ctx, sql, prompt.UserID, prompt.Prompt)
	if err != nil {
		return err
	}

	return nil
}
