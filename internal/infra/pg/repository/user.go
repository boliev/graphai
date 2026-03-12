package repository

import "github.com/jackc/pgx/v5/pgxpool"

type User struct {
	pool *pgxpool.Pool
}

func NewUser(pool *pgxpool.Pool) *User {
	return &User{pool: pool}
}
