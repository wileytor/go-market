package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DBstorage struct {
	Pool *pgxpool.Pool
}

// Создание нового пула соединений
func NewDB(pool *pgxpool.Pool) (*DBstorage, error) {
	_, err := pool.Exec(context.Background(), "SET search_path TO public")
	if err != nil {
		return nil, fmt.Errorf("failed to set schema : %w", err)
	}
	return &DBstorage{
		Pool: pool,
	}, nil
}

// Закрытие пула соединений
func (db *DBstorage) Close() {
	db.Pool.Close()
}
