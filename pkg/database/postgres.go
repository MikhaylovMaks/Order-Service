package database

import (
	"context"
	"fmt"

	"github.com/MikhaylovMaks/wb_techl0/internal/config"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Storage struct {
	Pool *pgxpool.Pool
}

func (s *Storage) Close() {
	if s.Pool != nil {
		s.Pool.Close()
	}
}

func NewPostgres(ctx context.Context, cfg config.Postgres) (*Storage, error) {
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
	)

	pool, err := pgxpool.Connect(ctx, connStr)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}
	return &Storage{Pool: pool}, nil
}
