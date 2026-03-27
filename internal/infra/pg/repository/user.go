package repository

import (
	"context"

	"github.com/boliev/graphai/internal/domain/user"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

func (u *UserRepo) Upsert(ctx context.Context, usr *user.User) (*user.User, error) {
	var newUser user.User
	sql := "INSERT INTO users (user_vk_id, peer_id) VALUES ($1, $2) ON CONFLICT (user_vk_id) DO UPDATE SET peer_id = EXCLUDED.peer_id, last_action = now() RETURNING id, user_vk_id, peer_id, balance, free_usages, last_action, last_notify, created_at"
	row := u.pool.QueryRow(ctx, sql, usr.UserVKID, usr.PeerID)
	err := row.Scan(&newUser.ID, &newUser.UserVKID, &newUser.PeerID, &newUser.Balance, &newUser.FreeUsages, &newUser.LastAction, &newUser.LastNotify, &newUser.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &newUser, nil
}

func (u *UserRepo) FindByVKID(ctx context.Context, id int64) (*user.User, error) {
	sql := "SELECT id, user_vk_id, peer_id, balance, free_usages, last_action, last_notify, created_at FROM users WHERE user_vk_id=$1"
	row := u.pool.QueryRow(ctx, sql, id)
	var usr user.User
	err := row.Scan(&usr.ID, &usr.UserVKID, &usr.PeerID, &usr.Balance, &usr.FreeUsages, &usr.LastAction, &usr.LastNotify, &usr.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &usr, nil
}

func (u *UserRepo) RefreshLastAction(ctx context.Context, id int64) error {
	sql := "UPDATE users SET last_action=now() WHERE id=$1"
	_, err := u.pool.Exec(ctx, sql, id)
	return err
}
